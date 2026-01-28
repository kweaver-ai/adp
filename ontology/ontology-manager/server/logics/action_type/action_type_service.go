package action_type

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"ontology-manager/common"
	cond "ontology-manager/common/condition"
	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
	"ontology-manager/logics"
	"ontology-manager/logics/object_type"
	"ontology-manager/logics/permission"
)

var (
	atServiceOnce sync.Once
	atService     interfaces.ActionTypeService
)

type actionTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ata        interfaces.ActionTypeAccess
	cga        interfaces.ConceptGroupAccess
	mfa        interfaces.ModelFactoryAccess
	osa        interfaces.OpenSearchAccess
	ots        interfaces.ObjectTypeService
	ps         interfaces.PermissionService
	uma        interfaces.UserMgmtAccess
}

func NewActionTypeService(appSetting *common.AppSetting) interfaces.ActionTypeService {
	atServiceOnce.Do(func() {
		atService = &actionTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			ata:        logics.ATA,
			cga:        logics.CGA,
			mfa:        logics.MFA,
			osa:        logics.OSA,
			ots:        object_type.NewObjectTypeService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			uma:        logics.UMA,
		}
	})
	return atService
}

func (ats *actionTypeService) CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckActionTypeExistByID")
	defer span.End()

	atName, exist, err := ats.ata.CheckActionTypeExistByID(ctx, knID, branch, atID)
	if err != nil {
		logger.Errorf("CheckActionTypeExistByID error: %s", err.Error())
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%v]获取行动类失败: %v", atID, err))
		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%v]获取行动类失败", atID))
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_CheckActionTypeIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atName, exist, nil
}

func (ats *actionTypeService) CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckActionTypeExistByName")
	defer span.End()

	actionTypeID, exist, err := ats.ata.CheckActionTypeExistByName(ctx, knID, branch, atName)
	if err != nil {
		logger.Errorf("CheckActionTypeExistByName error: %s", err.Error())
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按名称[%s]获取行动类失败: %v", atName, err))
		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s]获取行动类失败", atName))
		return actionTypeID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_CheckActionTypeIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return actionTypeID, exist, nil
}

func (ats *actionTypeService) CreateActionTypes(ctx context.Context, tx *sql.Tx, actionTypes []*interfaces.ActionType, mode string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateActionTypes")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   actionTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return []string{}, err
	}

	currentTime := time.Now().UnixMilli()
	for _, actionType := range actionTypes {
		// 若提交的模型id为空，生成分布式ID
		if actionType.ATID == "" {
			actionType.ATID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}

		actionType.Creator = accountInfo
		actionType.Updater = accountInfo

		actionType.CreateTime = currentTime
		actionType.UpdateTime = currentTime

		// 绑定的对象类非空时，需校验对象类存在性
		if actionType.ObjectTypeID != "" {
			// 导入时未提交，在一个事务中get
			_, err = ats.ots.GetObjectTypeByID(ctx, tx, actionType.KNID, actionType.Branch, actionType.ObjectTypeID)
			if err != nil {
				return []string{}, err
			}
		}
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateActionType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("CreateActionType Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("CreateActionType Transaction Commit Success")
				o11y.Debug(ctx, "CreateActionType Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateActionType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("CreateActionType Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	createActionTypes, updateActionTypes, err := ats.handleActionTypeImportMode(ctx, mode, actionTypes)
	if err != nil {
		return []string{}, err
	}

	// 创建
	atIDs := []string{}
	for _, actionType := range createActionTypes {
		atIDs = append(atIDs, actionType.ATID)
		err = ats.ata.CreateActionType(ctx, tx, actionType)
		if err != nil {
			logger.Errorf("CreateActionType error: %s", err.Error())
			span.SetStatus(codes.Error, "创建行动类失败")
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ActionType_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	// 更新
	for _, actionType := range updateActionTypes {
		// 提交的已存在，需要更新
		err = ats.UpdateActionType(ctx, tx, actionType)
		if err != nil {
			return []string{}, err
		}
	}

	insetActionTypes := createActionTypes
	insetActionTypes = append(insetActionTypes, updateActionTypes...)
	err = ats.InsertOpenSearchData(ctx, insetActionTypes)
	if err != nil {
		logger.Errorf("InsertOpenSearchData error: %s", err.Error())
		span.SetStatus(codes.Error, "行动类索引写入失败")
		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

func (ats *actionTypeService) ListActionTypes(ctx context.Context, query interfaces.ActionTypesQueryParams) ([]*interfaces.ActionType, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListActionTypes")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ActionType{}, 0, err
	}

	//获取行动类列表
	actionTypes, err := ats.ata.ListActionTypes(ctx, query)
	if err != nil {
		logger.Errorf("ListActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "List action types error")
		return []*interfaces.ActionType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(err.Error())
	}
	if len(actionTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return actionTypes, 0, nil
	}

	// 获取绑定对象类的名称拿到
	for _, actionType := range actionTypes {
		objectTypeMap, err := ats.ots.GetObjectTypesMapByIDs(ctx, query.KNID,
			query.Branch, []string{actionType.ObjectTypeID}, false)
		if err != nil {
			return []*interfaces.ActionType{}, 0, err
		}

		if objectTypeMap[actionType.ObjectTypeID] != nil {
			actionType.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[actionType.ObjectTypeID].OTID,
				OTName: objectTypeMap[actionType.ObjectTypeID].OTName,
				Icon:   objectTypeMap[actionType.ObjectTypeID].Icon,
				Color:  objectTypeMap[actionType.ObjectTypeID].Color,
			}
		}
	}
	total := len(actionTypes)

	// limit = -1,则返回所有
	if query.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(actionTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ActionType{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(actionTypes) {
			end = len(actionTypes)
		}

		actionTypes = actionTypes[query.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(actionTypes)*2)
	for _, at := range actionTypes {
		accountInfos = append(accountInfos, &at.Creator, &at.Updater)
	}

	err = ats.uma.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")
		return []*interfaces.ActionType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, total, nil
}

func (ats *actionTypeService) GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*interfaces.ActionType, error) {
	// 获取行动类
	ctx, span := ar_trace.Tracer.Start(ctx, "GetActionTypesByIDs")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ActionType{}, err
	}

	// id去重后再查
	atIDs = common.DuplicateSlice(atIDs)

	// 获取模型基本信息
	actionTypes, err := ats.ata.GetActionTypesByIDs(ctx, knID, branch, atIDs)
	if err != nil {
		logger.Errorf("GetActionTypesByATIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get action type[%v] error: %v", atIDs, err))
		return []*interfaces.ActionType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_GetActionTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	if len(actionTypes) != len(atIDs) {
		errStr := fmt.Sprintf("Exists any action types not found, expect action types nums is [%d], actual action types num is [%d]", len(atIDs), len(actionTypes))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		return []*interfaces.ActionType{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			oerrors.OntologyManager_ActionType_ActionTypeNotFound).WithErrorDetails(errStr)
	}

	// todo:翻译绑定的对象类、影响对象类、和对应的api文档
	// 获取绑定对象类和影响对象类的名称拿到
	for _, actionType := range actionTypes {
		affectObjectTypeID := ""
		if actionType.Affect != nil && actionType.Affect.ObjectTypeID != "" {
			affectObjectTypeID = actionType.Affect.ObjectTypeID
		}

		objectTypeMap, err := ats.ots.GetObjectTypesMapByIDs(ctx, knID, branch,
			[]string{actionType.ObjectTypeID, affectObjectTypeID}, false)
		if err != nil {
			return []*interfaces.ActionType{}, err
		}

		if objectTypeMap[actionType.ObjectTypeID] != nil {
			actionType.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[actionType.ObjectTypeID].OTID,
				OTName: objectTypeMap[actionType.ObjectTypeID].OTName,
				Icon:   objectTypeMap[actionType.ObjectTypeID].Icon,
				Color:  objectTypeMap[actionType.ObjectTypeID].Color,
			}
		}

		if objectTypeMap[affectObjectTypeID] != nil {
			actionType.Affect.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[affectObjectTypeID].OTID,
				OTName: objectTypeMap[affectObjectTypeID].OTName,
				Icon:   objectTypeMap[affectObjectTypeID].Icon,
				Color:  objectTypeMap[affectObjectTypeID].Color,
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, nil
}

// 更新行动类
func (ats *actionTypeService) UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *interfaces.ActionType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateActionType")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   actionType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	actionType.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 行动类的update_time是int类型
	actionType.UpdateTime = currentTime

	if tx == nil {
		// 0. 开始事务
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateActionType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateActionType Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("UpdateActionType Transaction Commit Success:%v", actionType.ATName)
				o11y.Debug(ctx, fmt.Sprintf("UpdateActionType Transaction Commit Success: %s", actionType.ATName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateActionType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateActionType Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 更新模型信息
	err = ats.ata.UpdateActionType(ctx, tx, actionType)
	if err != nil {
		logger.Errorf("UpdateActionType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	err = ats.InsertOpenSearchData(ctx, []*interfaces.ActionType{actionType})
	if err != nil {
		logger.Errorf("InsertOpenSearchData error: %s", err.Error())
		span.SetStatus(codes.Error, "行动类索引写入失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ats *actionTypeService) DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteActionTypesByIDs")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteActionTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteActionTypes Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("DeleteActionTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, atIDs)
				o11y.Debug(ctx, fmt.Sprintf("DeleteActionTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, atIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteActionTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteActionTypes Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 删除行动类
	rowsAffect, err := ats.ata.DeleteActionTypesByIDs(ctx, tx, knID, branch, atIDs)
	if err != nil {
		logger.Errorf("DeleteActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteActionTypes: Rows affected is %v, request delete ATIDs is %v!", rowsAffect, len(atIDs))
	if rowsAffect != int64(len(atIDs)) {
		logger.Warnf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(atIDs))
		o11y.Warn(ctx, fmt.Sprintf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(atIDs)))
	}

	for _, atID := range atIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_ACTION_TYPE, atID, branch)
		err = ats.osa.DeleteData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid)
		if err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部接口，不校验权限， tx必须传
func (ats *actionTypeService) DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteActionTypesByKnID")
	defer span.End()

	if tx == nil {
		logger.Errorf("missing transaction")
		o11y.Error(ctx, "missing transaction")
		span.SetStatus(codes.Error, "缺少事务")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_MissingTransaction).
			WithErrorDetails("missing transaction")
	}

	// 删除行动类
	rowsAffect, err := ats.ata.DeleteActionTypesByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteActionTypesByKnID success, the kn_id is [%s], branch is [%s], rowsAffect is [%d]",
		knID, branch, rowsAffect)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ats *actionTypeService) handleActionTypeImportMode(ctx context.Context, mode string,
	actionTypes []*interfaces.ActionType) ([]*interfaces.ActionType, []*interfaces.ActionType, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "handleActionTypeImportMode")
	defer span.End()

	creates := []*interfaces.ActionType{}
	updates := []*interfaces.ActionType{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for _, actionType := range actionTypes {
		creates = append(creates, actionType)
		idExist := false
		_, idExist, err := ats.CheckActionTypeExistByID(ctx, actionType.KNID, actionType.Branch, actionType.ATID)
		if err != nil {
			return creates, updates, err
		}

		// 校验 请求体与现有模型名称的重复性
		existID, nameExist, err := ats.CheckActionTypeExistByName(ctx, actionType.KNID, actionType.Branch, actionType.ATName)
		if err != nil {
			return creates, updates, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The action type with id [%s] already exists!", actionType.ATID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusBadRequest,
						oerrors.OntologyManager_ActionType_ActionTypeIDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("action type name '%s' already exists", actionType.ATName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						oerrors.OntologyManager_ActionType_ActionTypeNameExisted).
						WithDescription(map[string]any{"name": actionType.ATName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				creates = creates[:len(creates)-1]
			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的行动类 id 和当前行动类 id 不一样，则报错
					if existID != actionType.ATID {
						errDetails := fmt.Sprintf("ActionType ID '%s' and name '%s' already exist, but the exist action type id is '%s'",
							actionType.ATID, actionType.ATName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
							oerrors.OntologyManager_ActionType_ActionTypeNameExisted).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						creates = creates[:len(creates)-1]
						updates = append(updates, actionType)
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					// 从create数组中删除, 放到更新数组中
					creates = creates[:len(creates)-1]
					updates = append(updates, actionType)
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("ActionType ID '%s' does not exist, but name '%s' already exists",
						actionType.ATID, actionType.ATName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						oerrors.OntologyManager_ActionType_ActionTypeNameExisted).
						WithErrorDetails(errDetails)
				}

				// 如果 id 不存在，name不存在，度量名称不存在，不需要做什么，创建
				// if !idExist && !nameExist {}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return creates, updates, nil
}

func (ats *actionTypeService) InsertOpenSearchData(ctx context.Context, actionTypes []*interfaces.ActionType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "InsertOpenSearchData")
	defer span.End()

	if len(actionTypes) == 0 {
		return nil
	}

	if ats.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, actionType := range actionTypes {
			arr := []string{actionType.ATName}
			arr = append(arr, actionType.Tags...)
			arr = append(arr, actionType.Comment, actionType.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := ats.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := ats.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取行动类向量失败")
			return err
		}

		if len(vectors) != len(actionTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
			span.SetStatus(codes.Error, "获取行动类向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
		}

		for i, actionType := range actionTypes {
			actionType.Vector = vectors[i].Vector
		}
	}

	for _, actionType := range actionTypes {
		docid := interfaces.GenerateConceptDocuemtnID(actionType.KNID, interfaces.MODULE_TYPE_ACTION_TYPE,
			actionType.ATID, actionType.Branch)
		actionType.ModuleType = interfaces.MODULE_TYPE_ACTION_TYPE

		err := ats.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, actionType)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			span.SetStatus(codes.Error, "行动类概念索引写入失败")
			return err
		}
	}
	return nil
}

func (ats *actionTypeService) SearchActionTypes(ctx context.Context, query *interfaces.ConceptsQuery) (interfaces.ActionTypes, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "SearchActionTypes")
	defer span.End()

	response := interfaces.ActionTypes{}

	// 构造 DSL 过滤条件
	condtion, err := cond.NewCondition(ctx, query.ActualCondition, 1, interfaces.CONCPET_QUERY_FIELD)
	if err != nil {
		return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyManager_ActionType_InvalidParameter_ConceptCondition).
			WithErrorDetails(fmt.Sprintf("failed to new condition, %s", err.Error()))
	}

	// 转换到dsl
	conditionDslStr := "{}"
	if condtion != nil {
		conditionDslStr, err = condtion.Convert(ctx, func(ctx context.Context, words []string) ([]*cond.VectorResp, error) {
			if !ats.appSetting.ServerSetting.DefaultSmallModelEnabled {
				err = errors.New("DefaultSmallModelEnabled is false, does not support knn condition")
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}
			dftModel, err := ats.mfa.GetDefaultModel(ctx)
			if err != nil {
				logger.Errorf("GetDefaultModel error: %s", err.Error())
				span.SetStatus(codes.Error, "获取默认模型失败")
				return nil, err
			}
			return ats.mfa.GetVector(ctx, dftModel, words)
		})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyManager_ActionType_InvalidParameter_ConceptCondition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", err.Error()))
		}
	}

	// 1. 获取组下的关系类
	atIDMap := map[string]bool{} // 分组下的对象类id
	atIDs := []string{}          // 不同组下的对象类可以重叠，所以需要对对象类id的数组去重
	if len(query.ConceptGroups) > 0 {
		// 校验分组是否都存在，按分组id获取分组
		cgCnt, err := ats.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
			KNID:   query.KNID,
			Branch: query.Branch,
			CGIDs:  query.ConceptGroups,
		})
		if err != nil {
			logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", query.KNID, err.Error())
			span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", query.KNID, err))

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(err.Error())
		}
		if cgCnt == 0 {
			errStr := fmt.Sprintf("all concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				cgCnt, len(query.ConceptGroups))
			logger.Errorf(errStr)

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ActionType_InternalError).
				WithErrorDetails(errStr)
		}

		// 在当前业务知识网络下查找属于请求的分组范围内的行动类ID
		atIDArr, err := ats.cga.GetActionTypeIDsFromConceptGroupRelation(ctx, interfaces.ConceptGroupRelationsQueryParams{
			KNID:        query.KNID,
			Branch:      query.Branch,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE, // 概念与分组关系中的概念类型
			CGIDs:       query.ConceptGroups,
		})
		if err != nil {
			errStr := fmt.Sprintf("GetActionTypeIDsFromConceptGroupRelation failed, kn_id:[%s],branch:[%s],cg_ids:[%v], error: %v",
				query.KNID, query.Branch, query.ConceptGroups, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)
			span.End()

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ActionType_InternalError).WithErrorDetails(errStr)
		}
		if len(atIDArr) == 0 {
			// 概念分组下没有对象类,返回空
			if len(atIDArr) == 0 {
				return response, nil
			}
		}

		for _, atID := range atIDArr {
			if !atIDMap[atID] {
				atIDMap[atID] = true
				atIDs = append(atIDs, atID)
			}
		}
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		if len(query.ConceptGroups) == 0 {
			// 后面分批查询会改变query，所以先查总数
			dsl, err := logics.BuildDslQuery(ctx, conditionDslStr, query)
			if err != nil {
				return response, err
			}
			total, err := ats.GetTotal(ctx, dsl)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		} else {
			// 指定了分组，需要查询分组内且符合条件的总数
			// 方法1：在OpenSearch中通过ID列表过滤查询总数
			total, err := ats.GetTotalWithLargeATIDs(ctx, conditionDslStr, atIDs)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		}
	}

	// 4. 迭代查询直到获取足够数量或没有更多数据
	actionTypes := []*interfaces.ActionType{}
	var totalFilteredCount int64 = 0
	for {
		// 构建当前分页的DSL
		dsl, err := logics.BuildDslQuery(ctx, conditionDslStr, query)
		if err != nil {
			return response, err
		}

		// 请求opensearch
		result, err := ats.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, dsl)
		if err != nil {
			logger.Errorf("SearchData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络行动类检索查询失败")
			return response, err
		}

		// 如果没有数据了，跳出循环
		if len(result) == 0 {
			break
		}
		// 更新searchAfter用于下一次查询
		if len(result) > 0 {
			query.SearchAfter = result[len(result)-1].Sort
		} else {
			query.SearchAfter = nil
		}

		// 5. 过滤属于分组的结果
		for _, concept := range result {
			// 转成 action type 的 struct
			jsonByte, err := json.Marshal(concept.Source)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal opensearch hit _source, %s", err.Error()))
			}
			var actionType interfaces.ActionType
			err = json.Unmarshal(jsonByte, &actionType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal opensearch hit _source to Object Type, %s", err.Error()))
			}

			// 如果没有指定分组，或者对象类属于分组，则添加
			if len(query.ConceptGroups) == 0 || atIDMap[actionType.ATID] {
				actionType.Score = &concept.Score
				actionType.Vector = nil
				actionTypes = append(actionTypes, &actionType)
				totalFilteredCount++

				// 如果已经收集到足够的数量，跳出循环
				if len(actionTypes) >= query.Limit {
					break
				}
			}
		}
		// 如果已经收集到足够的数量或者没有更多数据了，跳出循环
		if len(actionTypes) >= query.Limit || len(result) < query.Limit {
			break
		}
	}

	response.Entries = actionTypes
	response.SearchAfter = query.SearchAfter
	return response, nil
}

func (ats *actionTypeService) GetTotal(ctx context.Context, dsl map[string]any) (total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetTotal")
	defer span.End()

	// delete(dsl, "pit")
	delete(dsl, "from")
	delete(dsl, "size")
	delete(dsl, "sort")
	totalBytes, err := ats.osa.Count(ctx, interfaces.KN_CONCEPT_INDEX_NAME, dsl)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	totalNode, err := sonic.Get(totalBytes, "count")
	if err != nil {
		span.SetStatus(codes.Error, "Get total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	total, err = totalNode.Int64()
	if err != nil {
		span.SetStatus(codes.Error, "Convert total documents count to type int64 failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 内部调用，不加权限校验
func (ats *actionTypeService) GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	// 获取行动类
	ctx, span := ar_trace.Tracer.Start(ctx, "GetActionTypeIDsByKnID")
	defer span.End()

	// 获取模型基本信息
	atIDs, err := ats.ata.GetActionTypeIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetActionTypeIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get action type[%v] error: %v", atIDs, err))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ActionType_InternalError_GetActionTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

// 分批查询
func (ats *actionTypeService) GetTotalWithLargeATIDs(ctx context.Context, conditionDslStr string, atIDs []string) (int64, error) {

	total := int64(0)
	for i := 0; i < len(atIDs); i += interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE {
		end := i + interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE
		if end > len(atIDs) {
			end = len(atIDs)
		}

		batchIDs := atIDs[i:end]
		batchTotal, err := ats.GetTotalWithATIDs(ctx, conditionDslStr, batchIDs)
		if err != nil {
			return 0, err
		}

		total += batchTotal
	}

	return total, nil
}

// 查询指定对象类ID列表的对象类总数
func (ats *actionTypeService) GetTotalWithATIDs(ctx context.Context, conditionDslStr string, atIDs []string) (int64, error) {

	var dslMap map[string]any
	err := json.Unmarshal([]byte(conditionDslStr), &dslMap)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyManager_InternalError_UnMarshalDataFailed).
			WithErrorDetails(fmt.Sprintf("failed to unMarshal dslStr to map, %s", err.Error()))
	}

	// 构建包含ATID过滤的DSL
	dsl := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					// 原有的查询条件
					dslMap,
					// ATID过滤条件
					map[string]any{
						"terms": map[string]any{
							"id": atIDs,
						},
					},
				},
			},
		},
	}

	queryJSON, err := sonic.Marshal(dsl)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal query: %w", err)
	}
	logger.Debug(string(queryJSON))

	// 执行计数查询
	total, err := ats.GetTotal(ctx, dsl)
	if err != nil {
		return total, err
	}

	return total, nil
}
