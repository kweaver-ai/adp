package data_view

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dtype "data-model/interfaces/data_type"
	"data-model/logics"
	"data-model/logics/permission"
)

var (
	dvServiceOnce sync.Once
	dvService     interfaces.DataViewService
)

type dataViewService struct {
	appSetting *common.AppSetting
	ps         interfaces.PermissionService
	dvrcrs     interfaces.DataViewRowColumnRuleService
	db         *sql.DB
	dsa        interfaces.DataSourceAccess
	dva        interfaces.DataViewAccess
	iba        interfaces.IndexBaseAccess
	dmja       interfaces.DataModelJobAccess
	dvga       interfaces.DataViewGroupAccess
	ua         interfaces.UniqueryAccess
}

func NewDataViewService(appSetting *common.AppSetting) interfaces.DataViewService {
	dvServiceOnce.Do(func() {
		dvService = &dataViewService{
			appSetting: appSetting,
			ps:         permission.NewPermissionService(appSetting),
			dvrcrs:     NewDataViewRowColumnRuleService(appSetting),
			db:         logics.DB,
			dsa:        logics.DSA,
			dva:        logics.DVA,
			iba:        logics.IBA,
			dmja:       logics.DMJA,
			dvga:       logics.DVGA,
			ua:         logics.UA,
		}
	})

	return dvService
}

// 批量创建数据视图, 加个是否校验权限的参数, 内部同步库表不校验权限
func (dvs *dataViewService) CreateDataViews(ctx context.Context, views []*interfaces.DataView, mode string, checkPermission bool) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Create data views")
	defer span.End()

	// 判断userid是否有创建数据视图的权限（策略决策）
	if checkPermission {
		err := dvs.ps.CheckPermission(ctx,
			interfaces.Resource{
				Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
				ID:   interfaces.RESOURCE_ID_ALL,
			},
			[]string{interfaces.OPERATION_TYPE_CREATE},
		)
		if err != nil {
			span.SetStatus(codes.Error, "Check permission failed")
			return nil, err
		}
	}

	viewIDs := make([]string, 0, len(views))

	for _, view := range views {
		// 如果视图ID为空，则生成一个
		if view.ViewID == "" {
			view.ViewID = xid.New().String()
		}

		viewIDs = append(viewIDs, view.ViewID)

		httpErr := dvs.commonForCreateAndUpdate(ctx, view)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Common operation for creating and updating failed")
			return nil, httpErr
		}
	}

	needRollback := false
	// 开始事务
	tx, err := dvs.db.Begin()
	if err != nil {
		logger.Errorf("CreateDataViews begin DB transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "CreateDataViews begin DB transaction failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).WithErrorDetails(err.Error())
	}

	defer func() {
		if !needRollback {
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateDataViews Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "create data views transaction commit failed")
			}
			logger.Debugf("CreateDataViews Transaction Commit Success, viewIDs: %v", viewIDs)
		} else {
			err = tx.Rollback()
			if err != nil {
				logger.Errorf("CreateDataViews Transaction Rollback Error:%v", err)
				span.SetStatus(codes.Error, "create data views transaction rollback failed")
			}
		}
	}()

	createdViews, updatedViews, httpErr := dvs.handleDataViewImportMode(ctx, tx, mode, views)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Handle data view import mode failed")
		needRollback = true
		return nil, httpErr
	}

	// 循环调用更新函数，其中校验对象的更新权限，如果没有更新权限则报错
	for _, uView := range updatedViews {
		httpErr = dvs.UpdateDataView(ctx, tx, uView)
		if err != nil {
			span.SetStatus(codes.Error, "Update data view failed")
			return nil, httpErr
		}
	}

	currentTime := time.Now().UnixMilli()
	createSrcs := make([]interfaces.Resource, 0, len(createdViews))
	for i, cView := range createdViews {
		// 初始化视图分组对象
		initedGroup := initViewGroupReq(cView)
		// 更新视图的分组ID
		groupID, isBuilitinGroup, httpErr := dvs.RetriveGroupIDByGroupName(ctx, tx, initedGroup)
		if httpErr != nil {
			needRollback = true
			logger.Errorf("RetriveGroupIDByGroupName error: %s", httpErr.Error())
			span.SetStatus(codes.Error, "Retrive group id by group name failed")
			return nil, httpErr
		}

		// 内置视图使用内置分组，非内置视图使用非内置分组
		if cView.Builtin != isBuilitinGroup {
			needRollback = true
			errDetails := "Built-in views must use built-in groups, non-built-in views must use non-built-in groups"
			logger.Error(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_InvalidBuiltinGroupMatch).
				WithErrorDetails(errDetails)
		}

		// 补充创建者、更新者
		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		cView.Creator = accountInfo
		cView.Updater = accountInfo
		cView.GroupID = groupID
		cView.CreateTime = currentTime
		cView.UpdateTime = currentTime
		// 原子视图也存字段，因为要存储字段的状态信息
		createdViews[i] = cView

		createSrcs = append(createSrcs, interfaces.Resource{
			ID:   cView.ViewID,
			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
			Name: common.ProcessUngroupedName(ctx, cView.GroupName, cView.ViewName),
		})
	}

	//调用driven层创建数据视图
	if len(createdViews) > 0 {
		// mysql限制占位符65535个，分批插入，每批2000个，可支持32个字段，当前字段24个
		batchSize := 2000
		logger.Infof("Insert %d data views into DB, batchSize is %d", len(createdViews), batchSize)
		for i := 0; i < len(createdViews); i += batchSize {
			end := i + batchSize
			if end > len(createdViews) {
				end = len(createdViews)
			}

			viewBatch := createdViews[i:end]
			if len(viewBatch) == 0 {
				continue
			}

			// 向数据库创建视图
			err = dvs.dva.CreateDataViews(ctx, tx, viewBatch)
			if err != nil {
				needRollback = true
				logger.Errorf("Create data views error: %s", err.Error())

				span.SetStatus(codes.Error, "Create data views failed")
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataView_InternalError_CreateDataViewsFailed).WithErrorDetails(err.Error())
			}

			resourcesBatch := createSrcs[i:end]
			if len(resourcesBatch) == 0 {
				continue
			}

			// 注册资源策略
			err = dvs.ps.CreateResources(ctx, resourcesBatch, interfaces.COMMON_OPERATIONS)
			if err != nil {
				needRollback = true
				logger.Errorf("Create resources error: %s", err.Error())

				span.SetStatus(codes.Error, "Create resources failed")
				o11y.Error(ctx, err.Error())
				return nil, err
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return viewIDs, nil
}

// 批量删除数据视图
func (dvs *dataViewService) DeleteDataViews(ctx context.Context, viewIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Delete data views")
	defer span.End()

	// // 获取视图信息
	// views, err := dvs.dva.GetDataViews(ctx, viewIDs)
	// if err != nil {
	// 	return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
	// 		WithErrorDetails(err.Error())
	// }

	// 先获取资源序列
	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range viewIDs {
		if _, exist := matchResouces[mID]; !exist {
			return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data view's delete operation")
		}
	}

	// 使用数据库事务
	tx, err := dvs.db.Begin()
	if err != nil {
		logger.Errorf("DeleteDataViews begin DB transaction failed: %s", err.Error())
		span.SetStatus(codes.Error, "DeleteDataViews begin DB transaction failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).
			WithErrorDetails(err.Error())
	}

	needRollback := false
	defer func() {
		if !needRollback {
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteDataViews commit DB transaction failed: %s", err.Error())
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				logger.Errorf("DeleteDataViews rollback DB transaction failed: %s", err.Error())
			}
		}
	}()

	// 删除数据视图表中的视图
	err = dvs.dva.DeleteDataViews(ctx, tx, viewIDs)
	if err != nil {
		needRollback = true
		logger.Errorf("Delete DataViews error: %s", err.Error())
		span.SetStatus(codes.Error, "Delete data views failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_DeleteDataViewsFailed).WithErrorDetails(err.Error())
	}

	// 删除视图下面的行列规则
	err = dvs.dvrcrs.DeleteRowColumnRulesByViewIDs(ctx, tx, viewIDs)
	if err != nil {
		needRollback = true
		logger.Errorf("Delete DataViewRowColumnRules error: %s", err.Error())
		span.SetStatus(codes.Error, "Delete data view row column rules failed")
		return err
	}

	//  清除资源策略
	err = dvs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs)
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改单个数据视图
func (dvs *dataViewService) UpdateDataView(ctx context.Context, tx *sql.Tx, view *interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update a data view")
	defer span.End()

	// 判断userid是否有修改数据视图的权限（策略决策）
	err := dvs.ps.CheckPermission(ctx,
		interfaces.Resource{
			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
			ID:   view.ViewID,
		},
		[]string{interfaces.OPERATION_TYPE_MODIFY},
	)
	if err != nil {
		span.SetStatus(codes.Error, "Check permission failed")
		return err
	}

	// 从数据库查询旧的视图信息
	oldViews, err := dvs.dva.GetDataViews(ctx, []string{view.ViewID})
	if err != nil {
		logger.Errorf("GetDataViews error: %s", err.Error())
		span.SetStatus(codes.Error, "GetDataViews failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsFailed).WithErrorDetails(err.Error())
	}

	if len(oldViews) == 0 {
		span.SetStatus(codes.Error, "Data view not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
			WithErrorDetails(fmt.Sprintf("Data view '%s' does not exist", view.ViewID))
	}
	oldView := oldViews[0]

	// 原子视图不支持修改分组名称
	if oldView.Type == interfaces.ViewType_Atomic {
		if view.GroupName != oldView.GroupName {
			span.SetStatus(codes.Error, "Atomic data view cannot change group")
			return rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("Atomic data view '%s' cannot change group", view.ViewID))
		}
	}

	// 检查索引库是否存在，字段类型是否冲突
	httpErr := dvs.commonForCreateAndUpdate(ctx, view)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Common operation for creating and updating failed")
		return httpErr
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	view.Updater = accountInfo
	view.UpdateTime = time.Now().UnixMilli()

	needRollback := false
	// 使用数据库事务
	if tx == nil {
		tx, err = dvs.db.Begin()
		if err != nil {
			logger.Errorf("UpdateDataView begin DB transaction failed: %s", err.Error())
			span.SetStatus(codes.Error, "UpdateDataView begin DB transaction failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).
				WithErrorDetails(err.Error())
		}

		defer func() {
			if !needRollback {
				err = tx.Commit()
				if err != nil {
					span.SetStatus(codes.Error, "UpdateDataView commit DB transaction failed")
					logger.Errorf("UpdateDataView commit DB transaction failed: %s", err.Error())
				}
			} else {
				err = tx.Rollback()
				if err != nil {
					span.SetStatus(codes.Error, "UpdateDataView rollback DB transaction failed")
					logger.Errorf("UpdateDataView rollback DB transaction failed: %s", err.Error())
				}
			}
		}()
	}

	// 获取分组ID，如果分组不存在，则创建分组
	groupID, isBuilitinGroup, httpErr := dvs.RetriveGroupIDByGroupName(ctx, tx, initViewGroupReq(view))
	if httpErr != nil {
		needRollback = true
		logger.Errorf(fmt.Sprintf("Retrive group id by group name %s failed", view.GroupName))
		span.SetStatus(codes.Error, "Retrive group id by group name failed")
		return httpErr
	}

	// 内置视图使用内置分组，非内置视图使用非内置分组
	if oldView.Builtin != isBuilitinGroup {
		needRollback = true
		errDetails := "Built-in views must use built-in groups, non-built-in views must use non-built-in groups"
		logger.Error(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_InvalidBuiltinGroupMatch).
			WithErrorDetails(errDetails)
	}

	// 更新视图的分组ID
	view.GroupID = groupID

	oldGroupID := oldView.GroupID
	oldViewName := oldView.ViewName
	newGroupID := groupID
	newViewName := view.ViewName

	// 校验视图名称在分组内是否已存在
	if newGroupID != oldGroupID || newViewName != oldViewName {
		_, exist, httpErr := dvs.CheckDataViewExistByName(ctx, tx, view.ViewName, view.GroupName)
		if httpErr != nil {
			needRollback = true
			span.SetStatus(codes.Error, "Check data view exist by name failed")
			return httpErr
		}

		if exist {
			needRollback = true
			errDetails := fmt.Sprintf("Data view '%s' already exists in group '%s'", view.ViewName, view.GroupName)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
				WithDescription(map[string]any{"ViewName": view.ViewName, "GroupName": view.GroupName}).
				WithErrorDetails(errDetails)
		}
	}

	// 更新数据库的视图信息
	err = dvs.dva.UpdateDataView(ctx, tx, view)
	if err != nil {
		needRollback = true
		logger.Errorf("Update a data view error: %s", err.Error())
		span.SetStatus(codes.Error, "Update a data view failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_UpdateDataViewFailed).WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = dvs.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   view.ViewID,
		Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
		Name: common.ProcessUngroupedName(ctx, view.GroupName, view.ViewName),
	})
	if err != nil {
		span.SetStatus(codes.Error, "Update resource name failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部修改单个数据视图，不校验权限, 同步库表为原子视图时使用
func (dvs *dataViewService) UpdateDataViewInternal(ctx context.Context, view *interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update a data view internal")
	defer span.End()

	// 从数据库查询旧的视图信息
	oldViews, err := dvs.dva.GetDataViews(ctx, []string{view.ViewID})
	if err != nil {
		logger.Errorf("GetDataViews error: %s", err.Error())
		span.SetStatus(codes.Error, "GetDataViews failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsFailed).WithErrorDetails(err.Error())
	}

	if len(oldViews) == 0 {
		// 不报错，如果同步过程中更新视图时，用户手动删除了视图，直接跳过当前视图的更新
		span.SetStatus(codes.Error, "Data view not found")
		logger.Warnf("Update data view id '%s' name '%s' not found, skip", view.ViewID, view.ViewName)
		return nil
		// return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
		// 	WithErrorDetails(fmt.Sprintf("Data view '%s' does not exist", view.ViewID))
	}
	oldView := oldViews[0]

	// 原子视图不支持修改分组名称
	if oldView.Type == interfaces.ViewType_Atomic {
		if view.GroupName != oldView.GroupName {
			span.SetStatus(codes.Error, "Atomic data view cannot change group")
			return rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("Atomic data view '%s' cannot change group", view.ViewID))
		}
	}

	// 检查索引库是否存在，字段类型是否冲突
	httpErr := dvs.commonForCreateAndUpdate(ctx, view)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Common operation for creating and updating failed")
		return httpErr
	}

	// accountInfo := interfaces.AccountInfo{}
	// if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
	// 	accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	// }
	// 同步库表到原子视图时，如果是更新视图，不改变视图的更新者
	view.Updater = oldView.Updater
	view.UpdateTime = time.Now().UnixMilli()

	needRollback := false
	// 使用数据库事务
	tx, err := dvs.db.Begin()
	if err != nil {
		logger.Errorf("UpdateDataView begin DB transaction failed: %s", err.Error())
		span.SetStatus(codes.Error, "UpdateDataView begin DB transaction failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).
			WithErrorDetails(err.Error())
	}

	defer func() {
		if !needRollback {
			err = tx.Commit()
			if err != nil {
				span.SetStatus(codes.Error, "UpdateDataView commit DB transaction failed")
				logger.Errorf("UpdateDataView commit DB transaction failed: %s", err.Error())
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				span.SetStatus(codes.Error, "UpdateDataView rollback DB transaction failed")
				logger.Errorf("UpdateDataView rollback DB transaction failed: %s", err.Error())
			}
		}
	}()

	// 获取分组ID，如果分组不存在，则创建分组
	groupID, isBuilitinGroup, httpErr := dvs.RetriveGroupIDByGroupName(ctx, tx, initViewGroupReq(view))
	if httpErr != nil {
		needRollback = true
		logger.Errorf(fmt.Sprintf("Retrive group id by group name %s failed", view.GroupName))
		span.SetStatus(codes.Error, "Retrive group id by group name failed")
		return httpErr
	}

	// 内置视图使用内置分组，非内置视图使用非内置分组
	if oldView.Builtin != isBuilitinGroup {
		needRollback = true
		errDetails := "Built-in views must use built-in groups, non-built-in views must use non-built-in groups"
		logger.Error(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_InvalidBuiltinGroupMatch).
			WithErrorDetails(errDetails)
	}

	// 更新视图的分组ID
	view.GroupID = groupID

	oldGroupID := oldView.GroupID
	oldViewName := oldView.ViewName
	newGroupID := groupID
	newViewName := view.ViewName

	// 校验视图名称在分组内是否已存在
	if newGroupID != oldGroupID || newViewName != oldViewName {
		_, exist, httpErr := dvs.CheckDataViewExistByName(ctx, tx, view.ViewName, view.GroupName)
		if httpErr != nil {
			needRollback = true
			span.SetStatus(codes.Error, "Check data view exist by name failed")
			return httpErr
		}

		if exist {
			needRollback = true
			errDetails := fmt.Sprintf("Data view '%s' already exists in group '%s'", view.ViewName, view.GroupName)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
				WithDescription(map[string]any{"ViewName": view.ViewName, "GroupName": view.GroupName}).
				WithErrorDetails(errDetails)
		}
	}

	// 更新数据库的视图信息
	err = dvs.dva.UpdateDataView(ctx, tx, view)
	if err != nil {
		needRollback = true
		logger.Errorf("Update a data view error: %s", err.Error())
		span.SetStatus(codes.Error, "Update a data view failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_UpdateDataViewFailed).WithErrorDetails(err.Error())
	}

	// 对于删了源表又重新创建的视图，需要更新视图的状态为正常
	if oldView.DeleteTime > 0 {
		// 更新数据库的视图信息
		err = dvs.dva.UpdateViewStatus(ctx, tx, []string{view.ViewID}, &interfaces.UpdateViewStatus{
			ViewStatus: interfaces.ViewScanStatus_New,
			DeleteTime: 0,
		})
		if err != nil {
			needRollback = true
			logger.Errorf("Update a data view status and delete time error: %s", err.Error())
			span.SetStatus(codes.Error, "Update a data view status and delete time failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataView_InternalError_UpdateDataViewFailed).WithErrorDetails(err.Error())
		}

		logger.Infof("Update data view %s status to new success", view.ViewName)
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = dvs.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   view.ViewID,
		Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
		Name: common.ProcessUngroupedName(ctx, view.GroupName, view.ViewName),
	})
	if err != nil {
		span.SetStatus(codes.Error, "Update resource name failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 单个获取视图详情
func (dvs *dataViewService) GetDataView(ctx context.Context, viewID string) (*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get a data view")
	defer span.End()

	views, err := dvs.GetDataViews(ctx, []string{viewID}, false)
	if err != nil {
		logger.Errorf("Get data views failed, %s", err.Error())

		span.SetStatus(codes.Error, "Get data views failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsFailed).WithErrorDetails(err.Error())
	}

	if len(views) == 0 {
		errDetails := fmt.Sprintf("The data view %s was not found", viewID)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, errDetails)
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
			WithErrorDetails(errDetails)
	}

	view := views[0]

	span.SetStatus(codes.Ok, "")
	return view, nil
}

// 批量获取视图详情
func (dvs *dataViewService) GetDataViews(ctx context.Context, viewIDs []string, includeDataScopeViews bool) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get data views")
	defer span.End()

	views, err := dvs.dva.GetDataViews(ctx, viewIDs)
	if err != nil {
		logger.Errorf("Get data views failed, %s", err.Error())

		span.SetStatus(codes.Error, "Get data views failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsFailed).WithErrorDetails(err.Error())
	}

	// 找到不存在的视图 id，如果有视图 id 不存在，则返回错误
	if len(views) < len(viewIDs) {
		viewsMap := make(map[string]struct{}, len(views))
		for _, view := range views {
			viewsMap[view.ViewID] = struct{}{}
		}

		for _, viewID := range viewIDs {
			if _, ok := viewsMap[viewID]; !ok {
				errDetails := fmt.Sprintf("The data view %s was not found", viewID)
				logger.Errorf(errDetails)

				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, errDetails)
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
					WithErrorDetails(errDetails)
			}
		}
	}

	// 先获取资源序列
	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range viewIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data view's view_detail operation.")
		}
	}

	dataSources, err := dvs.dsa.ListDataSources(ctx)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("List data sources error, %v", err))
	}

	// 数据源id转数据源名称
	dataSourceMap := make(map[string]*interfaces.DataSource)
	for _, dataSource := range dataSources.Entries {
		dataSourceMap[dataSource.ID] = dataSource
	}

	for index, view := range views {
		// 补充视图的可操作权限、数据源名称
		view.Operations = matchResouces[view.ViewID].Operations
		if dataSource, ok := dataSourceMap[view.DataSourceID]; ok {
			view.DataSourceName = dataSource.Name
			view.DataSourceCatalog = dataSource.BinData.CatalogName
			if view.MetaTableName == "" {
				schema := dataSource.BinData.Schema
				// 先用schema，没有再用database
				if schema == "" {
					schema = dataSource.BinData.DataBaseName
				}
				// database也没有使用默认值 default
				if schema == "" {
					schema = interfaces.DefaultSchema
				}

				view.MetaTableName = fmt.Sprintf(`%s."%s"."%s"`, dataSource.BinData.CatalogName, schema, view.TechnicalName)
			}
		}

		switch view.Type {
		case interfaces.ViewType_Atomic:
			// 索引库视图实时获取字段
			if view.QueryType == interfaces.QueryType_IndexBase {
				baseInfos, err := dvs.GetIndexBases(ctx, &view.SimpleDataView)
				if err != nil {
					span.SetStatus(codes.Error, "Get index bases info failed")
					return nil, err
				}

				viewFields := []*interfaces.ViewField{}
				fieldsMap := make(map[string]*interfaces.ViewField)
				for _, base := range baseInfos {
					allBaseFields := mergeIndexBaseFields(base.Mappings)

					for _, field := range allBaseFields {
						displayName := field.DisplayName
						if displayName == "" {
							displayName = field.Field
						}

						// fieldType := field.Type
						// if fieldType == dtype.DataType_Date {
						// 	fieldType = dtype.DataType_Datetime
						// }

						fieldType, ok := dtype.IndexBase_DataType_Map[field.Type]
						if !ok {
							fieldType = field.Type
						}

						vf := &interfaces.ViewField{
							Name:         field.Field,
							Type:         fieldType,
							DisplayName:  displayName,
							OriginalName: field.Field,
						}

						// 索引库的字段转为视图字段格式
						viewFields = append(viewFields, vf)
						// 索引库的字段转为视图字段格式，以字段名作为 key
						fieldsMap[field.Field] = vf
					}
				}

				view.Fields = viewFields
				view.FieldsMap = fieldsMap
			} else {
				// SQL 类
				fieldsMap := make(map[string]*interfaces.ViewField)
				for _, field := range view.Fields {
					fieldsMap[field.Name] = field
				}

				view.FieldsMap = fieldsMap
			}
		case interfaces.ViewType_Custom:
			// 给每个原子视图添加对应的技术名称（DSL类视图技术名称对应的是来源索引库），uniquery查询数据时需要
			for _, node := range view.DataScope {
				if node.Type != interfaces.DataScopeNodeType_View {
					continue
				}

				var viewID string
				var ok bool
				if viewID, ok = node.Config["view_id"].(string); !ok {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
						WithErrorDetails("view_id is not string")
				}

				// 获取原子视图的信息
				atomicViews, err := dvs.GetDataViews(ctx, []string{viewID}, true)
				if err != nil {
					return nil, err
				}

				// 上面有判断 not found 情况，这里可以放心取第一个
				atomicView := atomicViews[0]
				if includeDataScopeViews {
					fieldsMap := make(map[string]*interfaces.ViewField)
					for _, vf := range atomicView.Fields {
						fieldsMap[vf.Name] = vf
					}
					atomicView.FieldsMap = fieldsMap

					// 给自定义视图的来源视图加上表名
					if atomicView.QueryType == interfaces.QueryType_SQL {
						if atomicView.MetaTableName == "" {
							if dataSource, ok := dataSourceMap[atomicView.DataSourceID]; ok {
								schema := dataSource.BinData.Schema
								// 先用schema，没有再用database
								if schema == "" {
									schema = dataSource.BinData.DataBaseName
								}
								atomicView.MetaTableName = fmt.Sprintf(`%s."%s"."%s"`, dataSource.BinData.CatalogName, schema, atomicView.TechnicalName)
							}
						}
					}

					node.Config["view"] = atomicView
				}
			}

			fieldsMap := make(map[string]*interfaces.ViewField)
			for _, vf := range view.Fields {
				// name 作为 key
				fieldsMap[vf.Name] = vf
			}

			view.FieldsMap = fieldsMap

		default:
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails("unsupported view type")
		}

		views[index] = view
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

// 按分组导出数据视图
func (dvs *dataViewService) GetDataViewsByGroupID(ctx context.Context, groupID string) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get data views by group id")
	defer span.End()

	views, err := dvs.dva.GetDataViewsByGroupID(ctx, groupID)
	if err != nil {
		logger.Errorf("Get data views by group id error: %s", err.Error())
		span.SetStatus(codes.Error, "Get data views by group id failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsByGroupIDFailed).WithErrorDetails(err.Error())
	}

	if len(views) == 0 {
		return views, nil
	}

	viewIDs := make([]string, 0)
	for _, view := range views {
		viewIDs = append(viewIDs, view.ViewID)
	}

	// 先获取资源序列
	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return nil, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能导出
	if len(matchResouces) != len(viewIDs) {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for data view's view operation.")
	}

	dataSources, err := dvs.dsa.ListDataSources(ctx)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("List data sources error, %v", err))
	}

	// 数据源id转数据源名称
	dataSourceNameMap := make(map[string]*interfaces.DataSource)
	for _, dataSource := range dataSources.Entries {
		dataSourceNameMap[dataSource.ID] = dataSource
	}

	for index, view := range views {
		// 补充视图的可操作权限、数据源名称
		view.Operations = matchResouces[view.ViewID].Operations
		if dataSource, ok := dataSourceNameMap[view.DataSourceID]; ok {
			view.DataSourceName = dataSource.Name
			if view.MetaTableName == "" {
				schema := dataSource.BinData.Schema
				// 先用schema，没有再用database
				if schema == "" {
					schema = dataSource.BinData.DataBaseName
				}
				view.MetaTableName = fmt.Sprintf(`%s."%s"."%s"`, dataSource.BinData.CatalogName, schema, view.TechnicalName)
			}
		}

		switch view.Type {
		case interfaces.ViewType_Atomic:
			// 索引库视图实时获取字段
			if view.QueryType == interfaces.QueryType_IndexBase {
				baseInfos, err := dvs.GetIndexBases(ctx, &view.SimpleDataView)
				if err != nil {
					span.SetStatus(codes.Error, "Get index bases info failed")
					return nil, err
				}

				viewFields := []*interfaces.ViewField{}
				fieldsMap := make(map[string]*interfaces.ViewField)
				for _, base := range baseInfos {
					allBaseFields := mergeIndexBaseFields(base.Mappings)

					for _, field := range allBaseFields {
						displayName := field.DisplayName
						if displayName == "" {
							displayName = field.Field
						}

						fieldType, ok := dtype.IndexBase_DataType_Map[field.Type]
						if !ok {
							fieldType = field.Type
						}

						vf := &interfaces.ViewField{
							Name:         field.Field,
							Type:         fieldType,
							DisplayName:  displayName,
							OriginalName: field.Field,
						}

						// 索引库的字段转为视图字段格式
						viewFields = append(viewFields, vf)
						// 索引库的字段转为视图字段格式，以字段名作为 key
						fieldsMap[field.Field] = vf
					}
				}

				view.Fields = viewFields
				view.FieldsMap = fieldsMap
			}
		case interfaces.ViewType_Custom:
			// 给每个原子视图添加对应的技术名称（DSL类视图技术名称对应的是来源索引库），uniquery查询数据时需要
			for _, node := range view.DataScope {
				if node.Type != interfaces.DataScopeNodeType_View {
					continue
				}

				var viewID string
				var ok bool
				if viewID, ok = node.Config["view_id"].(string); !ok {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
						WithErrorDetails("view_id is not string")
				}

				// 获取原子视图的信息
				atomicViews, err := dvs.GetDataViews(ctx, []string{viewID}, true)
				if err != nil {
					return nil, err
				}

				// 上面有判断 not found 情况，这里可以放心取第一个
				atomicView := atomicViews[0]
				techName := atomicView.TechnicalName
				node.Config["technical_name"] = techName
				// 给自定义视图的来源视图加上表名
				if atomicView.QueryType == interfaces.QueryType_SQL {
					node.Config["meta_table_name"] = atomicView.MetaTableName
					if atomicView.MetaTableName == "" {
						if dataSource, ok := dataSourceNameMap[atomicView.DataSourceID]; ok {
							schema := dataSource.BinData.Schema
							// 先用schema，没有再用database
							if schema == "" {
								schema = dataSource.BinData.DataBaseName
							}
							node.Config["meta_table_name"] = fmt.Sprintf(`%s."%s"."%s"`, dataSource.BinData.CatalogName, schema, techName)
						}
					}
				}
			}

		default:
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails("unsupported view type")
		}

		views[index] = view
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

// 按数据源获取数据视图
func (dvs *dataViewService) GetDataViewsBySourceID(ctx context.Context, sourceID string) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get data views by data source")
	defer span.End()

	views, err := dvs.dva.GetDataViewsBySourceID(ctx, sourceID)
	if err != nil {
		logger.Errorf("Get data views by data source error: %s", err.Error())
		span.SetStatus(codes.Error, "Get data views by data source failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return views, nil
}

// 分页查询数据视图
func (dvs *dataViewService) ListDataViews(ctx context.Context, param *interfaces.ListViewQueryParams) ([]*interfaces.SimpleDataView, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: List data views")
	defer span.End()

	views, err := dvs.dva.ListDataViews(ctx, param)
	if err != nil {
		logger.Errorf("ListDataViews error: %s", err.Error())

		span.SetStatus(codes.Error, "List data views failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_ListDataViewsFailed).WithErrorDetails(err.Error())
	}

	if len(views) == 0 {
		return views, 0, nil
	}

	// 获取数据源列表
	dataSources, err := dvs.dsa.ListDataSources(ctx)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("List data sources error, %v", err))
	}

	// 数据源id转数据源名称
	dataSourceMap := make(map[string]*interfaces.DataSource)
	for _, dataSource := range dataSources.Entries {
		dataSourceMap[dataSource.ID] = dataSource
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	resMids := make([]string, 0)
	for _, v := range views {
		resMids = append(resMids, v.ViewID)
	}

	// 分批处理，每批1万个resmids, fix权限接口报错prepared statement contains too many placeholders
	batchSize := 10000
	// 所有有权限的视图id
	matchResouceIDMap := make(map[string]interfaces.ResourceOps)

	for i := 0; i < len(resMids); i += batchSize {
		end := i + batchSize
		if end > len(resMids) {
			end = len(resMids)
		}
		batchResMids := resMids[i:end]

		var batchMatchResources map[string]interfaces.ResourceOps
		if len(param.Operations) > 0 {
			batchMatchResources, err = dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
				batchResMids, param.Operations, true)
			if err != nil {
				return nil, 0, err
			}
		} else {
			batchMatchResources, err = dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
				batchResMids, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
			if err != nil {
				return nil, 0, err
			}
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResouceIDMap[resourceOps.ResourceID] = resourceOps
		}
	}

	// 遍历对象
	results := make([]*interfaces.SimpleDataView, 0)
	for _, v := range views {
		if resrc, exist := matchResouceIDMap[v.ViewID]; exist {
			v.Operations = resrc.Operations // 用户当前有权限的操作
			if dataSource, ok := dataSourceMap[v.DataSourceID]; ok {
				v.DataSourceName = dataSource.Name
				v.DataSourceCatalog = dataSource.BinData.CatalogName
			}

			results = append(results, v)
		}
	}

	// limit = -1,则返回所有
	if param.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if param.Offset < 0 || param.Offset >= len(results) {
		return []*interfaces.SimpleDataView{}, 0, nil
	}
	// 计算结束位置
	end := param.Offset + param.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[param.Offset:end], len(results), nil
}

// 根据名称检查数据视图是否存在，暴露 exist 参数，方便内部模块调用时根据exist决定后续行为
func (dvs *dataViewService) CheckDataViewExistByName(ctx context.Context, tx *sql.Tx, viewName, groupName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Check data view exist by name")
	defer span.End()

	viewID, exist, err := dvs.dva.CheckDataViewExistByName(ctx, tx, viewName, groupName)
	if err != nil {
		logger.Errorf("CheckDataViewExistByName %s in group %s error: %s", viewName, groupName, err.Error())

		span.SetStatus(codes.Error, "Check data view exist by name failed")
		return viewID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return viewID, exist, nil
}

// 单个查询，暴露 exist 参数，方便内部模块调用自己决定对存在与否的行为
func (dvs *dataViewService) CheckDataViewExistByID(ctx context.Context, tx *sql.Tx, viewID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get data view name by ID")
	defer span.End()

	viewName, exist, err := dvs.dva.CheckDataViewExistByID(ctx, tx, viewID)
	if err != nil {
		logger.Errorf("CheckDataViewExistByID error: %s", err.Error())

		span.SetStatus(codes.Error, "Check data view exist by ID failed")
		return viewName, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return viewName, exist, nil
}

// 根据数据视图ID数组去获取ID与数据视图详情的映射关系
// 删除视图、更新视图、批量更新视图属性时校验，如果有一个不存在，则返回错误，allowNonExist 为 false
// 链路模型使用，那边如果有视图不存在，只打印 warn 日志，不报错, allowNonExist 为 true
func (dvs *dataViewService) GetSimpleDataViewsByIDs(ctx context.Context, viewIDs []string, allowNonExist bool) (map[string]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Check data views exist by IDs")
	defer span.End()

	viewMap, err := dvs.dva.GetSimpleDataViewMapByIDs(ctx, viewIDs)
	if err != nil {
		logger.Errorf("Get simple data view map by ids failed, err: %v", err)
		span.SetStatus(codes.Error, "get simple data view map by ids failed")
		return viewMap, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetSimpleDataViewMapByIDsFailed).WithErrorDetails(err.Error())
	}

	// 如果允许包含不存在的视图，则返回
	if allowNonExist {
		span.SetStatus(codes.Ok, "")
		return viewMap, nil
	}

	// 如果包含不存在的，返回错误
	for _, viewID := range viewIDs {
		if _, ok := viewMap[viewID]; !ok {
			errDetails := fmt.Sprintf("Data view '%s' does not exist!", viewID)
			logger.Error(errDetails)

			o11y.Error(ctx, errDetails)
			span.SetStatus(codes.Error, errDetails)
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
				WithErrorDetails(errDetails)
		}
	}

	// 校验查看权限
	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range viewIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data view's view_detail operation.")
		}
	}

	span.SetStatus(codes.Ok, "")
	return viewMap, nil
}

// 修改原子视图
func (dvs *dataViewService) UpdateAtomicDataViews(ctx context.Context, attrs *interfaces.AtomicViewUpdateReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update atomic data views")
	defer span.End()

	// 判断userid是否有修改数据视图的权限（策略决策）
	err := dvs.ps.CheckPermission(ctx,
		interfaces.Resource{
			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
			ID:   attrs.ViewID,
		},
		[]string{interfaces.OPERATION_TYPE_MODIFY},
	)
	if err != nil {
		return err
	}

	// 从数据库查询旧的视图信息
	oldViews, err := dvs.dva.GetDataViews(ctx, []string{attrs.ViewID})
	if err != nil {
		logger.Errorf("GetDataViews error: %s", err.Error())
		span.SetStatus(codes.Error, "GetDataViews failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsFailed).WithErrorDetails(err.Error())
	}

	if len(oldViews) == 0 {
		return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
			WithErrorDetails(fmt.Sprintf("Data view '%s' does not exist", attrs.ViewID))
	}
	oldView := oldViews[0]

	oldViewName := oldView.ViewName
	newViewName := attrs.ViewName

	// 校验视图名称在分组内是否已存在
	if newViewName != oldViewName {
		// 校验名称是否和分组内名称重复
		_, exist, httpErr := dvs.CheckDataViewExistByName(ctx, nil, attrs.ViewName, oldView.GroupName)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Check data view exist by name failed")
			return httpErr
		}

		if exist {
			errDetails := fmt.Sprintf("Data view '%s' already exists in group '%s'", attrs.ViewName, oldView.GroupName)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
				WithDescription(map[string]any{"ViewName": attrs.ViewName, "GroupName": oldView.GroupName}).
				WithErrorDetails(errDetails)
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	attrs.Updater = accountInfo
	attrs.UpdateTime = time.Now().UnixMilli()

	err = dvs.dva.UpdateDataViewsAttrs(ctx, attrs)
	if err != nil {
		logger.Errorf("UpdateDataViewsAttrs error: %s", err.Error())
		span.SetStatus(codes.Error, "Update data views attrs failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = dvs.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   attrs.ViewID,
		Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
		Name: common.ProcessUngroupedName(ctx, oldView.GroupName, attrs.ViewName),
	})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 批量移动视图的分组
func (dvs *dataViewService) UpdateDataViewsGroup(ctx context.Context, viewsMap map[string]*interfaces.DataView, viewGroupReq *interfaces.ViewGroupReq) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update the group of data views")
	defer span.End()

	// 使用数据库事务
	tx, err := dvs.db.Begin()
	if err != nil {
		logger.Errorf("UpdateDataViewsGroup begin DB transaction failed: %s", err.Error())
		span.SetStatus(codes.Error, "UpdateDataViewsGroup begin DB transaction failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_BeginDbTransactionFailed).
			WithErrorDetails(err.Error())
	}

	needRollback := false
	defer func() {
		if !needRollback {
			err = tx.Commit()
			if err != nil {
				span.SetStatus(codes.Error, "UpdateDataViewsGroup commit DB transaction failed")
				logger.Errorf("UpdateDataViewsGroup commit DB transaction failed: %s", err.Error())
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				span.SetStatus(codes.Error, "UpdateDataViewsGroup rollback DB transaction failed")
				logger.Errorf("UpdateDataViewsGroup rollback DB transaction failed: %s", err.Error())
			}
		}
	}()

	// 根据分组名称获取分组 ID，如果不存在，则创建分组
	groupID, isBuilitinGroup, httpErr := dvs.RetriveGroupIDByGroupName(ctx, tx, viewGroupReq)
	if httpErr != nil {
		needRollback = true
		span.SetStatus(codes.Error, "Retrive group ID by group name failed")
		return httpErr
	}

	// 校验组下的模型的修改权限
	// 先获取资源列表
	vIDs := make([]string, 0, len(viewsMap))
	for _, v := range viewsMap {
		vIDs = append(vIDs, v.ViewID)
	}

	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
		vIDs, []string{interfaces.OPERATION_TYPE_MODIFY}, false)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources failed")
		return err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能修改
	if len(matchResouces) != len(vIDs) {
		span.SetStatus(codes.Error, "Access denied: insufficient permissions for data view's update operation")
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for data view's update operation.")
	}

	//获取分组内的数据视图
	existViews, err := dvs.dva.GetSimpleDataViewsByGroupID(ctx, tx, groupID)
	if err != nil {
		needRollback = true
		logger.Errorf("GetDataViewsByGroupID error: %s", err.Error())
		span.SetStatus(codes.Error, "Get data views by group ID failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDataViewsByGroupIDFailed).WithErrorDetails(err.Error())
	}

	viewIDs := make([]string, 0, len(viewsMap))
	for _, newView := range viewsMap {
		// 内置视图使用内置分组，非内置视图使用非内置分组
		if newView.Builtin != isBuilitinGroup {
			needRollback = true
			errDetails := "Built-in views must use built-in groups, non-built-in views must use non-built-in groups"
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusForbidden,
				derrors.DataModel_DataView_InvalidBuiltinGroupMatch).WithErrorDetails(errDetails)
		}

		// 如果移入的视图和分组内名称重复，则操作不成功，给出提示
		viewIDs = append(viewIDs, newView.ViewID)
		for _, oldView := range existViews {
			if newView.ViewName == oldView.ViewName {
				needRollback = true
				errDetails := fmt.Sprintf("Data view '%s' already exsited in group '%s'", newView.ViewName, viewGroupReq.GroupName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
					WithDescription(map[string]any{"ViewName": newView.ViewName, "GroupName": viewGroupReq.GroupName}).
					WithErrorDetails(errDetails)
			}
		}
	}

	err = dvs.dva.UpdateDataViewsGroup(ctx, tx, viewIDs, groupID)
	if err != nil {
		needRollback = true
		logger.Errorf("UpdateDataViewsGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "Update data views group failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_UpdateDataViewsGroupFailed).WithErrorDetails(err.Error())
	}

	// 批量改模型的分组，不改模型的名称，不用调更新资源名称的接口

	span.SetStatus(codes.Ok, "")
	return nil
}

// 根据分组名称获取分组ID
// 1. 如果分组存在则获取已有ID
// 2. 如果分组不存在则创建新分组并获取ID
func (dvs *dataViewService) RetriveGroupIDByGroupName(ctx context.Context, tx *sql.Tx, viewGroupReq *interfaces.ViewGroupReq) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("retrive data view group id by group name '%s'", viewGroupReq.GroupName))
	defer span.End()

	span.SetAttributes(attr.Key("group_name").String(viewGroupReq.GroupName))

	//查询groupName在指定范围内是否存在，f_builtin 和 f_group_name 联合唯一
	groupInfo, exist, err := dvs.dvga.CheckDataViewGroupExistByName(ctx, tx, viewGroupReq.GroupName, viewGroupReq.Builtin)
	if err != nil {
		logger.Errorf("Check data view group exist by group name error: %s", err.Error())
		span.SetStatus(codes.Error, "Check data view group exist by group name failed")
		return "", false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed).WithErrorDetails(err.Error())
	}

	if exist {
		span.SetStatus(codes.Ok, "Data view group exist, return group info")
		return groupInfo.GroupID, groupInfo.Builtin, nil
	}

	// 如果未指定groupID, 则生成一个
	if viewGroupReq.GroupID == "" {
		viewGroupReq.GroupID = xid.New().String()
	}

	currentTime := time.Now().UnixMilli()
	group := &interfaces.DataViewGroup{
		GroupName:  viewGroupReq.GroupName,
		GroupID:    viewGroupReq.GroupID,
		CreateTime: currentTime,
		UpdateTime: currentTime,
		Builtin:    viewGroupReq.Builtin,
	}

	err = dvs.dvga.CreateDataViewGroup(ctx, tx, group)
	if err != nil {
		logger.Errorf("CreateDataViewGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "Create data view group failed")

		return "", false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_CreateGroupFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return viewGroupReq.GroupID, viewGroupReq.Builtin, nil
}

// 创建和更新视图的一些通用操作，反序列化数据源，检查索引库是否存在、字段类型是否冲突
func (dvs *dataViewService) commonForCreateAndUpdate(ctx context.Context, view *interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Common operation for creating and updating views")
	defer span.End()

	// 索引库建的视图补齐type, query_type, data_source_id, data_source_type
	// TODO 后续切换为扫描索引库后删掉
	// DataSourceID 使用内置的 data source id
	if view.GroupName == interfaces.GroupName_IndexBase {
		view.Builtin = true
		view.DataSourceID = interfaces.DataSourceID_IndexBase
		view.DataSourceType = interfaces.DataSourceType_IndexBase
		view.QueryType = interfaces.QueryType_IndexBase
		view.Type = interfaces.ViewType_Atomic
		// 索引库创建的视图名称为 baseType
		view.TechnicalName = view.ViewName
	} else if view.Type == interfaces.ViewType_Atomic { //原子视图
		if view.DataSourceID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("Data source id is empty")
		}

		// 检查数据源id是否存在
		dataSource, err := dvs.dsa.GetDataSourceByID(ctx, view.DataSourceID)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("Get data source by id failed, %v", err))
		}

		// 这两种类型不支持创建视图
		if dataSource.Type == interfaces.DataSourceType_TingYun || dataSource.Type == interfaces.DataSourceType_AS7 {
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("data source type %s does not support creating data view, id is %s",
					dataSource.Type, dataSource.ID))
		}

		// 原子视图都是内置视图
		view.Builtin = true
		// 补全数据源类型
		view.DataSourceType = dataSource.Type
		// 补全分组名称和分组id
		view.GroupID = view.DataSourceID
		view.GroupName = dataSource.Name
		// 补全查询类型
		if dataSource.Type == interfaces.DataSourceType_IndexBase {
			view.QueryType = interfaces.QueryType_IndexBase
		} else if dataSource.Type == interfaces.DataSourceType_OpenSearch {
			view.QueryType = interfaces.QueryType_DSL
		} else {
			view.QueryType = interfaces.QueryType_SQL
		}
		// 补全sqlStr和DataSourceCatalog
		if view.QueryType == interfaces.QueryType_SQL {
			catalogName := dataSource.BinData.CatalogName
			schemaName := dataSource.BinData.Schema
			// 先用schema，没有再用database，
			if schemaName == "" {
				schemaName = dataSource.BinData.DataBaseName
			}
			// database也没有使用默认值 default
			if schemaName == "" {
				schemaName = interfaces.DefaultSchema
			}

			metaTableName := fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName),
				common.QuotationMark(view.TechnicalName))
			if view.MetaTableName == "" {
				// 补全 metatable name
				view.MetaTableName = metaTableName
			}
			if view.SQLStr == "" {
				// 补齐 sqlstr
				view.SQLStr = fmt.Sprintf(`SELECT * FROM %s`, view.MetaTableName)
			}

			view.VegaDataSource = dataSource
		}
	} else {
		// 自定义视图
		if view.DataScope == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("Data scope is empty")
		}

		nodeMap := make(map[string]struct{})
		for _, ds := range view.DataScope {
			nodeMap[ds.ID] = struct{}{}
		}

		dataScopeViewMap := make(map[string]*interfaces.DataView)

		for _, node := range view.DataScope {
			switch node.Type {
			case interfaces.DataScopeNodeType_View:
				// 校验视图节点
				err := validateViewNode(ctx, dvs, node, dataScopeViewMap)
				if err != nil {
					return err
				}
			case interfaces.DataScopeNodeType_Join:
				err := validateJoinNode(ctx, node, nodeMap)
				if err != nil {
					return err
				}
			case interfaces.DataScopeNodeType_Union:
				err := validateUnionNode(ctx, view.QueryType, node, nodeMap)
				if err != nil {
					return err
				}
			case interfaces.DataScopeNodeType_Sql:
				if view.QueryType != interfaces.QueryType_SQL {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
						WithErrorDetails("The sql node is only supported in sql query type")
				}

				err := validateSqlNode(ctx, node, nodeMap)
				if err != nil {
					return err
				}
			case interfaces.DataScopeNodeType_Output:
				err := validateOutputNode(ctx, node, nodeMap)
				if err != nil {
					return err
				}

				if len(view.Fields) == 0 {
					view.Fields = node.OutputFields
				}
			default:
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
					WithErrorDetails("The data scope node type is invalid")
			}
		}

		dataScopeViewQueryType := make(map[string]struct{})
		dataScopeViewDataSourceID := make(map[string]struct{})
		techNames := make([]string, 0)
		for _, dsView := range dataScopeViewMap {
			dataScopeViewQueryType[dsView.QueryType] = struct{}{}
			dataScopeViewDataSourceID[dsView.DataSourceID] = struct{}{}
			techNames = append(techNames, dsView.TechnicalName)
		}

		if len(dataScopeViewQueryType) != 1 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The source view of the custom view must have the same query type")
		}

		// 如果数据源类型是opensearch，则不能跨opensearch数据源选择
		if view.QueryType == interfaces.QueryType_DSL && len(dataScopeViewDataSourceID) > 1 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The source view of query type DSL must have the same data source when create custom view")
		}

		var queryType string
		for key := range dataScopeViewQueryType {
			queryType = key
			break
		}

		// 自定义视图技术名称为空
		view.TechnicalName = ""
		// 补全查询类型
		view.QueryType = queryType

		allFieldsMap := make(map[string]*interfaces.ViewField)
		for _, vf := range view.Fields {
			allFieldsMap[vf.OriginalName] = vf
		}

		// 校验 primary_keys 里的字段是否在视图字段列表里
		for _, key := range view.PrimaryKeys {
			if _, ok := allFieldsMap[key]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
					WithErrorDetails(fmt.Sprintf("The primary key '%s' is not in the view '%s' fields", key, view.ViewName))
			}
		}

		// 如果是IndexBase类型的视图，校验视图字段类型是否和索引库里的字段类型一样
		err := dvs.checkFieldTypeConflict(ctx, view, techNames)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 校验字段类型是否冲突
func (dvs *dataViewService) checkFieldTypeConflict(ctx context.Context, view *interfaces.DataView, baseTypes []string) error {
	if view.QueryType != interfaces.QueryType_IndexBase {
		return nil
	}

	baseInfos, err := dvs.iba.GetIndexBasesByTypes(ctx, baseTypes)
	if err != nil {
		return fmt.Errorf("view '%s' get index bases failed, %s", view.ViewName, err.Error())
	}

	fieldsMap := make(map[string]string)
	for _, viewField := range view.Fields {
		if existType, ok := fieldsMap[viewField.Name]; ok {
			if viewField.Type != existType {
				errDetails := fmt.Sprintf("View '%s' field '%s' has two different types: '%s' and '%s'",
					view.ViewName, viewField.Name, viewField.Type, existType)
				logger.Errorf(errDetails)
				return errors.New(errDetails)
			}
		} else {
			fieldsMap[viewField.Name] = viewField.Type
		}
	}

	for _, base := range baseInfos {
		baseFields := mergeIndexBaseFields(base.Mappings)

		for _, baseField := range baseFields {
			// 校验视图字段类型是否和索引库里的字段类型一样
			if viewFieldType, ok := fieldsMap[baseField.Field]; ok {
				// 特殊处理一下date类型，因为索引库里没有datetime类型，只有date类型
				baseFieldType, ok := dtype.IndexBase_DataType_Map[baseField.Type]
				if !ok {
					baseFieldType = baseField.Type
				}

				if viewFieldType != baseFieldType {
					errDetails := fmt.Sprintf("View '%s' field '%s' of type '%s' is different from index base field type '%s'",
						view.ViewName, baseField.Field, viewFieldType, baseFieldType)
					return errors.New(errDetails)

				}
			}
		}
	}

	// 导入时如果缺少元字段，自动将元字段添加到字段列表中
	// 指针类型，可直接修改 view 对象
	for metaField, metaFieldType := range interfaces.META_FIELDS {
		if _, ok := fieldsMap[metaField]; !ok {
			view.Fields = append(view.Fields, &interfaces.ViewField{
				Name:         metaField,
				Type:         metaFieldType,
				DisplayName:  metaField,
				OriginalName: metaField,
			})
			// 元字段补充到 fieldsMap 里
			fieldsMap[metaField] = metaFieldType
		}
	}

	return nil
}

// 获取索引库信息
func (dvs *dataViewService) GetIndexBases(ctx context.Context, view *interfaces.SimpleDataView) ([]interfaces.IndexBase, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get index bases")
	defer span.End()

	// switch view.DataSource["type"].(string) {
	// case interfaces.INDEX_BASE:
	// 	var bases []interfaces.SimpleIndexBase
	// 	err := mapstructure.Decode(view.DataSource[interfaces.INDEX_BASE], &bases)
	// 	if err != nil {
	// 		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataSource).
	// 			WithErrorDetails(fmt.Sprintf("mapstructure decode dataSource failed, %s", err.Error()))
	// 	}

	baseTypes := []string{view.TechnicalName}
	// for _, base := range bases {
	// 	baseTypes = append(baseTypes, base.BaseType)
	// }

	// 根据索引库类型获取索引库信息
	baseInfos, err := dvs.iba.GetIndexBasesByTypes(ctx, baseTypes)
	if err != nil {
		// 索引库不存在时不报错
		logger.Errorf("get index bases failed, %s", err.Error())
		span.SetStatus(codes.Error, "Get index bases by base_types failed")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetIndexBaseByTypeFailed).WithErrorDetails(err.Error())
	}

	// 获取索引库名称和备注
	// for index, baseInfo := range baseInfos {
	// 	bases[index].Name = baseInfo.Name
	// 	bases[index].Comment = baseInfo.Comment
	// }

	// view.DataSource[interfaces.INDEX_BASE] = bases

	span.SetStatus(codes.Ok, "")
	return baseInfos, nil
	// default:
	// 	errDetails := "Unsupported dataSource type, only 'index_base' is supported currently"

	// 	span.SetStatus(codes.Error, errDetails)
	// 	return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_UnsupportDataSourceType).
	// 		WithErrorDetails(errDetails)
	// }
}

// 根据数据视图 ID 数组去获取 ID 与数据视图详情的映射关系
func (dvs *dataViewService) GetDetailedDataViewMapByIDs(ctx context.Context, viewIDs []string) (viewMap map[string]*interfaces.DataView, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get detailed data view map by IDs")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 调用driven层, 获取 ID 与数据视图详情的映射关系
	viewMap, err = dvs.dva.GetDetailedDataViewMapByIDs(ctx, viewIDs)
	if err != nil {
		logger.Errorf("Get detailed data view map by IDs failed, err: %v", err.Error())
		return viewMap, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_GetDetailedDataViewMapByIDsFailed).
			WithErrorDetails(err.Error())
	}

	for viewID, view := range viewMap {
		switch view.QueryType {
		case interfaces.QueryType_IndexBase:
			// 暂时支持元数据视图
			if view.Type == interfaces.ViewType_Atomic {
				baseInfos, err := dvs.GetIndexBases(ctx, &view.SimpleDataView)
				if err != nil {
					span.SetStatus(codes.Error, "Get index bases info failed")
					return nil, err
				}

				viewFields := []*interfaces.ViewField{}

				for _, base := range baseInfos {
					allBaseFields := mergeIndexBaseFields(base.Mappings)

					for _, field := range allBaseFields {
						displayName := field.DisplayName
						if displayName == "" {
							displayName = field.Field
						}

						fieldType, ok := dtype.IndexBase_DataType_Map[field.Type]
						if !ok {
							fieldType = field.Type
						}

						viewFields = append(viewFields, &interfaces.ViewField{
							Name:         field.Field,
							Type:         fieldType,
							DisplayName:  displayName,
							OriginalName: field.Field,
						})
					}
				}

				view.Fields = viewFields
			}
		case interfaces.QueryType_SQL:
			// SQL类存储了字段列表
		case interfaces.QueryType_DSL:
			// DSL类存储了字段列表
		}

		viewMap[viewID] = view
	}

	return viewMap, nil
}

func (dvs *dataViewService) ListDataViewSrcs(ctx context.Context, params *interfaces.ListViewQueryParams) ([]*interfaces.Resource, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "logic layer: List data view resources")
	listSpan.End()

	views, err := dvs.dva.ListDataViews(listCtx, params)
	if err != nil {
		logger.Errorf("ListDataViews error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List data views error")
		listSpan.End()
		return nil, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_DataView_InternalError_ListDataViewsFailed).WithErrorDetails(err.Error())
	}
	if len(views) == 0 {
		return []*interfaces.Resource{}, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, v := range views {
		resMids = append(resMids, v.ViewID)
	}

	// 分批处理，每批1万个resmids, fix权限接口报错prepared statement contains too many placeholders
	batchSize := 10000
	// 所有有权限的视图id
	matchResouceIDMap := make(map[string]bool)

	for i := 0; i < len(resMids); i += batchSize {
		end := i + batchSize
		if end > len(resMids) {
			end = len(resMids)
		}
		batchResMids := resMids[i:end]

		var batchMatchResources map[string]interfaces.ResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
			batchResMids, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
		if err != nil {
			return nil, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResouceIDMap[resourceOps.ResourceID] = true
		}
	}

	// 遍历对象
	results := make([]*interfaces.Resource, 0)
	for _, view := range views {
		if matchResouceIDMap[view.ViewID] {
			results = append(results, &interfaces.Resource{
				ID:   view.ViewID,
				Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
				Name: common.ProcessUngroupedName(ctx, view.GroupName, view.ViewName),
			})
		}
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []*interfaces.Resource{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[params.Offset:end], len(results), nil
}

// 批量标记删除视图，内部接口，不校验权限
func (dvs *dataViewService) MarkDataViewsDeleted(ctx context.Context, tx *sql.Tx, viewIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Mark data view deleted")
	defer span.End()

	span.SetAttributes(attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)))

	param := &interfaces.MarkViewDeletedParams{
		ViewIDs:    viewIDs,
		DeleteTime: time.Now().UnixMilli(),
		ViewStatus: interfaces.ViewScanStatus_Delete,
	}

	if err := dvs.dva.MarkDataViewsDeleted(ctx, tx, param); err != nil {
		span.SetStatus(codes.Error, "mark data view deleted failed")
		o11y.Error(ctx, "mark data view deleted failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("mark data view deleted failed, %v", err))
	}

	span.SetStatus(codes.Ok, "mark data view deleted success")
	return nil
}
