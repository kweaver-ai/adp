// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics"
	"data-model/logics/permission"
)

var (
	dvgServiceOnce sync.Once
	dvgService     interfaces.DataViewGroupService
)

type dataViewGroupService struct {
	appSetting *common.AppSetting
	dvga       interfaces.DataViewGroupAccess
	dva        interfaces.DataViewAccess
	dvs        interfaces.DataViewService
	ps         interfaces.PermissionService
	db         *sql.DB
}

func NewDataViewGroupService(appSetting *common.AppSetting) interfaces.DataViewGroupService {
	dvgServiceOnce.Do(func() {
		dvgService = &dataViewGroupService{
			appSetting: appSetting,
			db:         logics.DB,
			dva:        logics.DVA,
			dvga:       logics.DVGA,
			dvs:        NewDataViewService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return dvgService
}

// 创建数据视图分组
func (dvgs *dataViewGroupService) CreateDataViewGroup(ctx context.Context, tx *sql.Tx, dataViewGroup *interfaces.DataViewGroup) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Create data view group")
	defer span.End()

	// 生成分布式 ID
	// 如果未指定groupID, 则生成一个
	if dataViewGroup.GroupID == "" {
		dataViewGroup.GroupID = xid.New().String()
	}
	currentTime := time.Now().UnixMilli()

	dataViewGroup.CreateTime = currentTime
	dataViewGroup.UpdateTime = currentTime

	span.SetAttributes(
		attr.Key("data_view_group_id").String(dataViewGroup.GroupID),
		attr.Key("data_view_group_name").String(dataViewGroup.GroupName),
	)

	err := dvgs.dvga.CreateDataViewGroup(ctx, tx, dataViewGroup)
	if err != nil {
		logger.Errorf("Create data view group error: %s", err.Error())
		span.SetStatus(codes.Error, "create data view group failed")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_CreateGroupFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return dataViewGroup.GroupID, nil
}

// 删除数据视图分组
func (dvgs *dataViewGroupService) DeleteDataViewGroup(ctx context.Context, groupID string, includeViews bool) ([]*interfaces.SimpleDataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Delete data view groups")
	defer span.End()

	dataViews, err := dvgs.dva.GetSimpleDataViewsByGroupID(ctx, nil, groupID)
	if err != nil {
		logger.Errorf("Get data views by group id error: %s", err.Error())
		span.SetStatus(codes.Error, "get data views by group id failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_GetViewsByGroupIDFailed).WithErrorDetails(err.Error())
	}

	viewIDs := make([]string, 0)
	for _, view := range dataViews {
		viewIDs = append(viewIDs, view.ViewID)
	}

	// 组内包含视图，但未指定删除视图，不允许删除分组
	if !includeViews && len(viewIDs) != 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataViewGroup_GroupNotEmpty).
			WithErrorDetails("delete group failed, cause the group is not empty")
	}

	// 组内包含视图，校验视图是否具有删除权限，如果没有，分组不允许删除
	if len(viewIDs) != 0 {
		matchResouces, err := dvgs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
			[]string{interfaces.OPERATION_TYPE_DELETE}, false)
		if err != nil {
			return nil, err
		}
		// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
		for _, vID := range viewIDs {
			if _, exist := matchResouces[vID]; !exist {
				return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
					WithErrorDetails("Access denied: insufficient permissions for data view's delete operation.")
			}
		}
	}

	// 组内包含视图，指定删除视图，使用事务将分组和视图一起删除
	// 组内不包含视图，指定删除视图，直接删除分组
	// 组内不包含视图，未指定删除视图，直接删除分组
	// 这三种情况可以统一用下面的逻辑处理

	// 使用数据库事务
	tx, err := dvgs.db.Begin()
	if err != nil {
		logger.Errorf("Begin DB transaction failed: %s", err.Error())
		span.SetStatus(codes.Error, "Begin DB transaction failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_BeginDBTransactionFailed).WithErrorDetails(err.Error())
	}

	needRollback := false
	defer func() {
		if !needRollback {
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteDataViewGroup commit DB transaction failed: %s", err.Error())
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				logger.Errorf("DeleteDataViewGroup rollback DB transaction failed: %s", err.Error())
			}
		}
	}()

	// 删除分组
	err = dvgs.dvga.DeleteDataViewGroup(ctx, tx, groupID)
	if err != nil {
		needRollback = true
		logger.Errorf("Delete data view group error: %s", err.Error())
		span.SetStatus(codes.Error, "delete data view group failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_DeleteGroupFailed).WithErrorDetails(err.Error())
	}

	// 删除分组下的视图
	if len(viewIDs) > 0 {
		err = dvgs.dva.DeleteDataViews(ctx, tx, viewIDs)
		if err != nil {
			needRollback = true
			logger.Errorf("Delete data views error: %s", err.Error())
			span.SetStatus(codes.Error, "delete data views failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataViewGroup_InternalError_DeleteDataViewsInGroupFailed).WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return dataViews, nil
}

// 修改数据视图分组
func (dvgs *dataViewGroupService) UpdateDataViewGroup(ctx context.Context, dataViewGroup *interfaces.DataViewGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update data view group")
	defer span.End()

	span.SetAttributes(
		attr.Key("data_view_group_id").String(dataViewGroup.GroupID),
		attr.Key("data_view_group_name").String(dataViewGroup.GroupName),
	)

	dataViewGroup.UpdateTime = time.Now().UnixMilli()
	err := dvgs.dvga.UpdateDataViewGroup(ctx, dataViewGroup)
	if err != nil {
		logger.Errorf("update data view group error, %v", err)
		span.SetStatus(codes.Error, "update data view group failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_UpdateGroupFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询数据视图分组列表
func (dvgs *dataViewGroupService) ListDataViewGroups(ctx context.Context,
	params *interfaces.ListViewGroupQueryParams, includeViews bool) ([]*interfaces.DataViewGroup, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: List data view groups")
	defer span.End()

	groups, err := dvgs.dvga.ListDataViewGroups(ctx, params)
	if err != nil {
		logger.Errorf("List data view groups error: %s", err.Error())
		span.SetStatus(codes.Error, "list data view groups failed")

		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_ListGroupsFailed).WithErrorDetails(err.Error())
	}

	total, err := dvgs.dvga.GetDataViewGroupsTotal(ctx, params)
	if err != nil {
		logger.Errorf("GetDataViewGroupsTotal error: %s", err.Error())

		span.SetStatus(codes.Error, "Get data view groups total failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_GetGroupsTotalFailed).WithErrorDetails(err.Error())
	}

	if includeViews {
		// list所有视图，按权限过滤，对过滤完成的数据按组进行分组计数
		views, _, err := dvgs.dvs.ListDataViews(ctx, &interfaces.ListViewQueryParams{
			GroupID:   interfaces.GroupID_All,
			GroupName: interfaces.GroupName_All,
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
				Sort:  interfaces.DATA_VIEW_GROUP_SORT["name"],
			},
		})
		if err != nil {
			return nil, 0, err
		}

		groupViewCnt := map[string]int{}
		for _, view := range views {
			if cnt, exist := groupViewCnt[view.GroupID]; !exist {
				groupViewCnt[view.GroupID] = 1
			} else {
				groupViewCnt[view.GroupID] = cnt + 1
			}
		}

		// 赋值
		for i := range groups {
			if cnt, exist := groupViewCnt[groups[i].GroupID]; exist {
				groups[i].DataViewCount = cnt
			} else {
				groups[i].DataViewCount = 0
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return groups, total, nil
}

// 按ID获取数据视图分组信息
func (dvgs *dataViewGroupService) GetDataViewGroupByID(ctx context.Context, groupID string) (*interfaces.DataViewGroup, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Get data view group '%s' info", groupID))
	defer span.End()

	span.SetAttributes(attr.Key("group_id").String(groupID))

	var dvg *interfaces.DataViewGroup
	group, exist, err := dvgs.dvga.GetDataViewGroupByID(ctx, groupID)
	if err != nil {
		logger.Errorf("Get data view group by id error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get data view group '%s' failed", groupID))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_GetGroupByIDFailed).WithErrorDetails(err.Error())
	}

	if !exist {
		logger.Debugf("Data view group '%s' not found", groupID)
		span.SetStatus(codes.Error, fmt.Sprintf("Data view group '%s' not found", groupID))
		return dvg, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataViewGroup_GroupNotFound)
	}

	span.SetStatus(codes.Ok, "")
	return group, nil
}

// 根据groupName检查分组是否存在
func (dvgs *dataViewGroupService) CheckDataViewGroupExistByName(ctx context.Context, tx *sql.Tx, groupName string, builtin bool) (*interfaces.DataViewGroup, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Check data view group '%s' existence", groupName))
	defer span.End()

	span.SetAttributes(attr.Key("group_name").String(groupName))

	group, exist, err := dvgs.dvga.CheckDataViewGroupExistByName(ctx, tx, groupName, builtin)
	if err != nil {
		logger.Errorf("Check data view group existence by name error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to check data view group '%s' existence", groupName))
		return group, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_CheckGroupExistByNameFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return group, exist, nil
}

// 标记删除分组，内部接口，不校验权限
func (dvgs *dataViewGroupService) MarkDataViewGroupDeleted(ctx context.Context, groupID string, includeViews bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Mark data view group '%s' deleted", groupID))
	defer span.End()

	span.SetAttributes(attr.Key("group_id").String(groupID))

	var tx *sql.Tx
	var err error
	needRollback := false
	if includeViews {
		// 使用数据库事务
		tx, err = dvgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin DB transaction failed: %s", err.Error())
			span.SetStatus(codes.Error, "Begin DB transaction failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataViewGroup_InternalError_BeginDBTransactionFailed).WithErrorDetails(err.Error())
		}

		defer func() {
			if !needRollback {
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteDataViewGroup commit DB transaction failed: %s", err.Error())
				}
			} else {
				err = tx.Rollback()
				if err != nil {
					logger.Errorf("DeleteDataViewGroup rollback DB transaction failed: %s", err.Error())
				}
			}
		}()
	}

	param := &interfaces.MarkViewGroupDeletedParams{
		GroupID:    groupID,
		DeleteTime: time.Now().UnixMilli(),
	}
	// 标记删除分组
	if err := dvgs.dvga.MarkDataViewGroupDeleted(ctx, tx, param); err != nil {
		needRollback = true
		logger.Errorf("Mark data view group deleted error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Mark data view group '%s' deleted failed", groupID))
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	if !includeViews {
		return nil
	}

	dataViews, err := dvgs.dva.GetSimpleDataViewsByGroupID(ctx, tx, groupID)
	if err != nil {
		needRollback = true
		logger.Errorf("Get data views by group id error: %s", err.Error())
		span.SetStatus(codes.Error, "get data views by group id failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataViewGroup_InternalError_GetViewsByGroupIDFailed).WithErrorDetails(err.Error())
	}

	viewIDs := make([]string, 0)
	for _, view := range dataViews {
		viewIDs = append(viewIDs, view.ViewID)
	}

	// 标记删除数据视图
	if err := dvgs.dvs.MarkDataViewsDeleted(ctx, tx, viewIDs); err != nil {
		needRollback = true
		logger.Errorf("Mark data views deleted error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Mark data views in group '%s' deleted failed", groupID))
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
