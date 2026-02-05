package object_type

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
	"ontology-manager/logics/permission"
)

var (
	otServiceOnce sync.Once
	otService     interfaces.ObjectTypeService
)

type objectTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	cga        interfaces.ConceptGroupAccess
	dda        interfaces.DataModelAccess
	dva        interfaces.DataViewAccess
	mfa        interfaces.ModelFactoryAccess
	osa        interfaces.OpenSearchAccess
	ota        interfaces.ObjectTypeAccess
	uma        interfaces.UserMgmtAccess
	ps         interfaces.PermissionService
}

func NewObjectTypeService(appSetting *common.AppSetting) interfaces.ObjectTypeService {
	otServiceOnce.Do(func() {
		otService = &objectTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			cga:        logics.CGA,
			dda:        logics.DDA,
			dva:        logics.DVA,
			mfa:        logics.MFA,
			osa:        logics.OSA,
			ota:        logics.OTA,
			uma:        logics.UMA,
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return otService
}

func (ots *objectTypeService) CheckObjectTypeExistByID(ctx context.Context,
	knID string, branch string, otID string) (string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验对象类[%s]的存在性", otID))
	defer span.End()

	otName, exist, err := ots.ota.CheckObjectTypeExistByID(ctx, knID, branch, otID)
	if err != nil {
		logger.Errorf("CheckObjectTypeExistByID error: %s", err.Error())
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("在业务知识网络[%s]下按ID[%s]获取对象类失败: %v", knID, otID, err))
		span.SetStatus(codes.Error, fmt.Sprintf("在业务知识网络[%s]下按ID[%s]获取对象类失败", knID, otID))
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_CheckObjectTypeIfExistFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otName, exist, nil
}

func (ots *objectTypeService) CheckObjectTypeExistByName(ctx context.Context,
	knID string, branch string, otName string) (string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验对象类[%s]的存在性", otName))
	defer span.End()

	otID, exist, err := ots.ota.CheckObjectTypeExistByName(ctx, knID, branch, otName)
	if err != nil {
		logger.Errorf("CheckObjectTypeExistByName error: %s", err.Error())
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("在业务知识网络[%s]下按名称[%s]获取对象类失败: %v", knID, otName, err))
		span.SetStatus(codes.Error, fmt.Sprintf("在业务知识网络[%s]下按名称[%s]获取对象类失败", knID, otName))
		return otID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_CheckObjectTypeIfExistFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otID, exist, nil
}

func (ots *objectTypeService) CreateObjectTypes(ctx context.Context, tx *sql.Tx,
	objectTypes []*interfaces.ObjectType, mode string, needCreateConceptGroupRelation bool) ([]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return []string{}, err
	}

	currentTime := time.Now().UnixMilli()
	for _, objectType := range objectTypes {
		// 若提交的模型id为空，生成分布式ID
		if objectType.OTID == "" {
			objectType.OTID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		objectType.Creator = accountInfo
		objectType.Updater = accountInfo

		objectType.CreateTime = currentTime
		objectType.UpdateTime = currentTime

		// todo: 处理版本
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateObjectType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("CreateObjectType Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("CreateObjectType Transaction Commit Success")
				o11y.Debug(ctx, "CreateObjectType Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateObjectType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("CreateObjectType Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	createObjectTypes, updateObjectTypes, err := ots.handleObjectTypeImportMode(ctx, mode, objectTypes)
	if err != nil {
		return []string{}, err
	}

	// 创建
	otIDs := []string{}
	for _, objectType := range createObjectTypes {
		otIDs = append(otIDs, objectType.OTID)
		err = ots.ota.CreateObjectType(ctx, tx, objectType)
		if err != nil {
			logger.Errorf("CreateObjectType error: %s", err.Error())
			span.SetStatus(codes.Error, "创建对象类失败")

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).
				WithErrorDetails(err.Error())
		}

		err = ots.ota.CreateObjectTypeStatus(ctx, tx, objectType)
		if err != nil {
			logger.Errorf("CreateObjectTypeStatus error: %s", err.Error())
			span.SetStatus(codes.Error, "创建对象类状态失败")

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).
				WithErrorDetails(err.Error())
		}

		// 按需建立对象类到各个组的关系
		if needCreateConceptGroupRelation {
			// 建立对象类到各个组的关系，已经存在的关系就不需要建立，需要先获取一下对象类与组的关系
			if len(objectType.ConceptGroups) > 0 {
				err = ots.handleGroupRelations(ctx, tx, objectType, currentTime)
				if err != nil {
					span.SetStatus(codes.Error, "处理对象类与分组的关系失败")
					return []string{}, err
				}
			}
		}
	}

	// 更新
	for _, objectType := range updateObjectTypes {
		// todo: 提交的已存在，需要更新，则版本号+1
		err = ots.UpdateObjectType(ctx, tx, objectType)
		if err != nil {
			return []string{}, err
		}
	}

	insetObjectTypes := createObjectTypes
	insetObjectTypes = append(insetObjectTypes, updateObjectTypes...)
	err = ots.InsertOpenSearchData(ctx, insetObjectTypes)
	if err != nil {
		logger.Errorf("InsertOpenSearchData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otIDs, nil
}

func (ots *objectTypeService) ListObjectTypes(ctx context.Context, tx *sql.Tx,
	query interfaces.ObjectTypesQueryParams) ([]*interfaces.ObjectType, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询对象类列表")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ObjectType{}, 0, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("ListObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("ListObjectTypes Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("ListObjectTypes Transaction Commit Success")
				o11y.Debug(ctx, "ListObjectTypes Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("ListObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("ListObjectTypes Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	//获取对象类列表
	objectTypes, err := ots.ota.ListObjectTypes(ctx, tx, query)
	if err != nil {
		logger.Errorf("ListObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "List object types error")

		return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}
	if len(objectTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return objectTypes, 0, nil
	}

	total := len(objectTypes)
	// limit = -1,则返回所有
	if query.Limit != -1 {

		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(objectTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ObjectType{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(objectTypes) {
			end = len(objectTypes)
		}
		objectTypes = objectTypes[query.Offset:end]
	}

	otIDs := []string{}
	accountInfos := make([]*interfaces.AccountInfo, 0, len(objectTypes)*2)
	for _, objectType := range objectTypes {
		accountInfos = append(accountInfos, &objectType.Creator, &objectType.Updater)
		otIDs = append(otIDs, objectType.OTID)
	}

	err = ots.uma.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 获取对象类所属的分组
	otGroups, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:   query.KNID,
		Branch: query.Branch,
		OTIDs:  otIDs,
	})
	if err != nil {
		span.SetStatus(codes.Error, "GetConceptGroupsByOTIDs error")

		return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	for _, objectType := range objectTypes {
		// 获取视图字段的显示名
		if objectType.DataSource != nil && objectType.DataSource.ID != "" {
			dataView, err := ots.dva.GetDataViewByID(ctx, objectType.DataSource.ID)
			if err != nil {
				return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					oerrors.OntologyManager_ObjectType_InternalError_GetDataViewByIDFailed).
					WithErrorDetails(err.Error())
			}
			if dataView == nil {
				o11y.Warn(ctx, fmt.Sprintf("Object type [%s]'s Data view %s not found", objectType.OTID, objectType.DataSource.ID))
			} else {
				objectType.DataSource.Name = dataView.ViewName
				// 翻译数据属性映射的字段显示名
				for j, prop := range objectType.DataProperties {
					// 不为空时，才翻译字段显示名。为空则不翻译
					if prop.MappedField != nil {
						if field, exists := dataView.FieldsMap[prop.MappedField.Name]; exists {
							objectType.DataProperties[j].MappedField.DisplayName = field.DisplayName
							objectType.DataProperties[j].MappedField.Type = field.Type
						}
					}
					// 字符串类型的属性支持的操作符返回
					objectType.DataProperties[j].ConditionOperations = ots.processConditionOperations(objectType, prop, dataView)
				}
			}
		}

		// 给对象类加上分组信息
		objectType.ConceptGroups = otGroups[objectType.OTID]
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, total, nil
}

func (ots *objectTypeService) GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx,
	knID string, branch string, otIDs []string) ([]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询对象类[%s]信息", otIDs))
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ObjectType{}, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("GetObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("GetObjectTypes Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("GetObjectTypes Transaction Commit Success")
				o11y.Debug(ctx, "GetObjectTypes Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("GetObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("GetObjectTypes Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// id去重后再查
	otIDs = common.DuplicateSlice(otIDs)

	// 获取对象类基本信息
	objectTypes, err := ots.ota.GetObjectTypesByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("GetObjectTypesByObjectTypeIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object types[%s] error: %v", otIDs, err))

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	if len(objectTypes) != len(otIDs) {
		errStr := fmt.Sprintf("Exists any object types not found, expect object types nums is [%d], actual object types num is [%d]", len(otIDs), len(objectTypes))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			oerrors.OntologyManager_ObjectType_ObjectTypeNotFound).WithErrorDetails(errStr)
	}

	// 获取对象类所属的分组
	otGroups, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:   knID,
		Branch: branch,
		OTIDs:  otIDs,
	})
	if err != nil {
		span.SetStatus(codes.Error, "GetConceptGroupsByOTIDs error")

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 数据视图不为空时，需要把id转成名称
	// 请求视图
	for _, objectType := range objectTypes {
		// 处理数据源和操作符
		err = ots.processObjectTypeDetails(ctx, objectType)
		if err != nil {
			return []*interfaces.ObjectType{}, err
		}
		// 给对象类加上分组信息
		objectType.ConceptGroups = otGroups[objectType.OTID]
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

// 更新对象类
func (ots *objectTypeService) UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Update object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 校验数据属性
	for _, prop := range objectType.DataProperties {
		if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled {
			model, err := ots.mfa.GetModelByID(ctx, prop.IndexConfig.VectorConfig.ModelID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					oerrors.OntologyManager_ObjectType_InternalError_GetSmallModelByIDFailed).
					WithErrorDetails(err.Error())
			}
			if model == nil {
				return rest.NewHTTPError(ctx, http.StatusNotFound,
					oerrors.OntologyManager_ObjectType_SmallModelNotFound).
					WithErrorDetails(fmt.Sprintf("model %s not found", prop.IndexConfig.VectorConfig.ModelID))
			}
			if model.ModelType != interfaces.SMALL_MODEL_TYPE_EMBEDDING {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_ObjectType_InvalidParameter_SmallModel).
					WithErrorDetails(fmt.Sprintf("model type %s is not %s model", model.ModelType, interfaces.SMALL_MODEL_TYPE_EMBEDDING))
			}
			if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_ObjectType_InvalidParameter_SmallModel).
					WithErrorDetails(fmt.Sprintf("model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
			}
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	objectType.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 对象类的update_time是int类型
	objectType.UpdateTime = currentTime

	if tx == nil {
		// 0. 开始事务
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateObjectType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateObjectType Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("UpdateObjectType Transaction Commit Success:%v", objectType.OTName)
				o11y.Debug(ctx, fmt.Sprintf("UpdateObjectType Transaction Commit Success: %s", objectType.OTName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateObjectType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateObjectType Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 更新模型信息
	err = ots.ota.UpdateObjectType(ctx, tx, objectType)
	if err != nil {
		logger.Errorf("UpdateObjectType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	// 4. 同步分组关系（全量替换）
	if err := ots.syncObjectGroups(ctx, tx, *objectType, currentTime); err != nil {
		return err
	}

	err = ots.InsertOpenSearchData(ctx, []*interfaces.ObjectType{objectType})
	if err != nil {
		logger.Errorf("InsertOpenSearchData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新对象类数据属性
func (ots *objectTypeService) UpdateDataProperties(ctx context.Context,
	objectType *interfaces.ObjectType, dataProperties []*interfaces.DataProperty) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Update object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 校验数据属性
	for _, prop := range dataProperties {
		if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled {
			model, err := ots.mfa.GetModelByID(ctx, prop.IndexConfig.VectorConfig.ModelID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					oerrors.OntologyManager_ObjectType_InternalError_GetSmallModelByIDFailed).
					WithErrorDetails(err.Error())
			}
			if model == nil {
				return rest.NewHTTPError(ctx, http.StatusNotFound,
					oerrors.OntologyManager_ObjectType_SmallModelNotFound).
					WithErrorDetails(fmt.Sprintf("model %s not found", prop.IndexConfig.VectorConfig.ModelID))
			}
			if model.ModelType != interfaces.SMALL_MODEL_TYPE_EMBEDDING {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_ObjectType_InvalidParameter_SmallModel).
					WithErrorDetails(fmt.Sprintf("model type %s is not %s model", model.ModelType, interfaces.SMALL_MODEL_TYPE_EMBEDDING))
			}
			if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyManager_ObjectType_InvalidParameter_SmallModel).
					WithErrorDetails(fmt.Sprintf("model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
			}
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	objectType.Updater = accountInfo
	currentTime := time.Now().UnixMilli() // 对象类的update_time是int类型
	objectType.UpdateTime = currentTime

	propMap := map[string]int{}
	for idx, prop := range objectType.DataProperties {
		propMap[prop.Name] = idx
	}
	for _, prop := range dataProperties {
		if idx, ok := propMap[prop.Name]; ok {
			objectType.DataProperties[idx] = prop
		} else {
			objectType.DataProperties = append(objectType.DataProperties, prop)
		}
	}

	// 更新模型信息
	err = ots.ota.UpdateDataProperties(ctx, objectType)
	if err != nil {
		logger.Errorf("UpdateObjectType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	err = ots.InsertOpenSearchData(ctx, []*interfaces.ObjectType{objectType})
	if err != nil {
		logger.Errorf("InsertOpenSearchData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ots *objectTypeService) DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete object types")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteObjectTypes Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("DeleteObjectTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, otIDs)
				o11y.Debug(ctx, fmt.Sprintf("DeleteObjectTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, otIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteObjectTypes Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 删除对象类
	rowsAffect, err := ots.ota.DeleteObjectTypesByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("DeleteObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteObjectTypes: Rows affected is %v, request delete ObjectTypeIDs is %v!", rowsAffect, len(otIDs))
	if rowsAffect != int64(len(otIDs)) {
		logger.Warnf("Delete object types number %v not equal requerst object types number %v!", rowsAffect, len(otIDs))
		o11y.Warn(ctx, fmt.Sprintf("Delete object types number %v not equal requerst object types number %v!", rowsAffect, len(otIDs)))
	}

	rowsAffect, err = ots.ota.DeleteObjectTypeStatusByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("DeleteObjectTypeStatusByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类状态失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypeStatusByIDs success, the kn_id is [%s], branch is [%s], ot_ids is [%v], rowsAffect is [%d]",
		knID, branch, otIDs, rowsAffect)

	for _, otID := range otIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_OBJECT_TYPE, otID, branch)
		err = ots.osa.DeleteData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid)
		if err != nil {
			return err
		}
	}

	// 从概念与分组的关系表中删除该对象所建立的关系
	// 删除对象类与分组的绑定关系
	rowsAffect, err = ots.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:        knID,
		Branch:      branch,
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		OTIDs:       otIDs,
	})
	if err != nil {
		errStr := fmt.Sprintf("DeleteObjectTypesFromGroup failed, the kn_id is [%s], branch is [%s], ot_ids is [%v], error is [%s]",
			knID, "branch", otIDs, err.Error())
		logger.Errorf(errStr)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(errStr)
	}
	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypesFromGroup success, the kn_id is [%s], branch is [%s], ot_ids is [%v], rowsAffect is [%d]",
		knID, branch, otIDs, rowsAffect)

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部方法，删除对象类与状态，不检查权限，tx必须传入
func (ots *objectTypeService) DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete object types")
	defer span.End()

	if tx == nil {
		logger.Errorf("missing transaction")
		o11y.Error(ctx, "missing transaction")
		span.SetStatus(codes.Error, "缺少事务")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
			WithErrorDetails("missing transaction")
	}

	// 删除对象类
	rowsAffect, err := ots.ota.DeleteObjectTypesByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteObjectTypes: Rows affected is %v!", rowsAffect)
	rowsAffect, err = ots.ota.DeleteObjectTypeStatusByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteObjectTypeStatusByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类状态失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypesByKnID success, the kn_id is [%s], branch is [%s], rowsAffect is [%d]",
		knID, branch, rowsAffect)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ots *objectTypeService) handleObjectTypeImportMode(ctx context.Context, mode string,
	objectTypes []*interfaces.ObjectType) ([]*interfaces.ObjectType, []*interfaces.ObjectType, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "object type import mode logic")
	defer span.End()

	creates := []*interfaces.ObjectType{}
	updates := []*interfaces.ObjectType{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for _, objectType := range objectTypes {
		creates = append(creates, objectType)
		idExist := false
		_, idExist, err := ots.CheckObjectTypeExistByID(ctx, objectType.KNID, objectType.Branch, objectType.OTID)
		if err != nil {
			return creates, updates, err
		}

		// 校验 请求体与现有模型名称的重复性
		existID, nameExist, err := ots.CheckObjectTypeExistByName(ctx, objectType.KNID, objectType.Branch, objectType.OTName)
		if err != nil {
			return creates, updates, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The object type with id [%s] already exists!", objectType.OTID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusBadRequest,
						oerrors.OntologyManager_ObjectType_ObjectTypeIDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("object type name '%s' already exists", objectType.OTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						oerrors.OntologyManager_ObjectType_ObjectTypeNameExisted).
						WithDescription(map[string]any{"name": objectType.OTName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				creates = creates[:len(creates)-1]
			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
					if existID != objectType.OTID {
						errDetails := fmt.Sprintf("ObjectType ID '%s' and name '%s' already exist, but the exist object type id is '%s'",
							objectType.OTID, objectType.OTName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
							oerrors.OntologyManager_ObjectType_ObjectTypeNameExisted).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						creates = creates[:len(creates)-1]
						updates = append(updates, objectType)
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					// 从create数组中删除, 放到更新数组中
					creates = creates[:len(creates)-1]
					updates = append(updates, objectType)
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("ObjectType ID '%s' does not exist, but name '%s' already exists",
						objectType.OTID, objectType.OTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						oerrors.OntologyManager_ObjectType_ObjectTypeNameExisted).
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

// 内部使用，无需校验权限
func (ots *objectTypeService) GetObjectTypesMapByIDs(ctx context.Context, knID string,
	branch string, otIDs []string, needPropMap bool) (map[string]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询对象类[%v]信息", otIDs))
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return map[string]*interfaces.ObjectType{}, err
	}

	// id去重后再查
	otIDs = common.DuplicateSlice(otIDs)

	// 获取模型基本信息
	objectTypeArr, err := ots.ota.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("GetObjectTypesByObjectTypeIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type[%v] error: %v", otIDs, err))
		return map[string]*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	objectTypeMap := map[string]*interfaces.ObjectType{}
	for _, object := range objectTypeArr {
		if needPropMap {
			propMap := map[string]string{}
			for _, prop := range object.DataProperties {
				propMap[prop.Name] = prop.DisplayName
			}
			object.PropertyMap = propMap
		}
		objectTypeMap[object.OTID] = object
	}

	span.SetStatus(codes.Ok, "")
	return objectTypeMap, nil
}

func (ots *objectTypeService) InsertOpenSearchData(ctx context.Context, objectTypes []*interfaces.ObjectType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "对象类索引写入")
	defer span.End()

	if len(objectTypes) == 0 {
		return nil
	}

	if ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, objectType := range objectTypes {
			arr := []string{objectType.OTName}
			arr = append(arr, objectType.Tags...)
			arr = append(arr, objectType.Comment, objectType.Detail)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := ots.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := ots.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取业务知识网络向量失败")
			return err
		}

		if len(vectors) != len(objectTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
			span.SetStatus(codes.Error, "获取业务知识网络向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
		}

		for i, objectType := range objectTypes {
			objectType.Vector = vectors[i].Vector
		}
	}

	for _, objectType := range objectTypes {
		docid := interfaces.GenerateConceptDocuemtnID(objectType.KNID, interfaces.MODULE_TYPE_OBJECT_TYPE,
			objectType.OTID, objectType.Branch)
		objectType.ModuleType = interfaces.MODULE_TYPE_OBJECT_TYPE

		err := ots.osa.InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, docid, objectType)
		if err != nil {
			logger.Errorf("InsertData error: %s", err.Error())
			span.SetStatus(codes.Error, "对象类概念索引写入失败")
			return err
		}
	}
	return nil
}

// type vectorFunc func(ctx context.Context, words []string) ([]cond.VectorResp, error)

func (ots *objectTypeService) SearchObjectTypes(ctx context.Context,
	query *interfaces.ConceptsQuery) (interfaces.ObjectTypes, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "业务知识网络对象类检索")
	defer span.End()

	response := interfaces.ObjectTypes{}

	// 2. 构造 DSL 过滤条件
	condtion, err := cond.NewCondition(ctx, query.ActualCondition, 1, interfaces.CONCPET_QUERY_FIELD)
	if err != nil {
		return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyManager_ObjectType_InvalidParameter_ConceptCondition).
			WithErrorDetails(fmt.Sprintf("failed to new condition, %s", err.Error()))
	}

	// 转换到dsl
	conditionDslStr := "{}"
	if condtion != nil {
		conditionDslStr, err = condtion.Convert(ctx, func(ctx context.Context, words []string) ([]*cond.VectorResp, error) {
			if !ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
				err = errors.New("DefaultSmallModelEnabled is false, does not support knn condition")
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}
			dftModel, err := ots.mfa.GetDefaultModel(ctx)
			if err != nil {
				logger.Errorf("GetDefaultModel error: %s", err.Error())
				span.SetStatus(codes.Error, "获取默认模型失败")
				return nil, err
			}
			return ots.mfa.GetVector(ctx, dftModel, words)
		})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyManager_ObjectType_InvalidParameter_ConceptCondition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", err.Error()))
		}
	}

	// 1. 获取组下的对象类
	otIDMap := map[string]bool{} // 分组下的对象类id
	otIDs := []string{}          // 不同组下的对象类可以重叠，所以需要对对象类id的数组去重
	if len(query.ConceptGroups) > 0 {

		// 校验分组是否都存在，按分组id获取分组
		cgCnt, err := ots.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
			KNID:   query.KNID,
			Branch: query.Branch,
			CGIDs:  query.ConceptGroups,
		})
		if err != nil {
			logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", query.KNID, err.Error())
			span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", query.KNID, err))

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
		}
		if cgCnt == 0 {
			errStr := fmt.Sprintf("all concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				cgCnt, len(query.ConceptGroups))
			logger.Errorf(errStr)

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}

		// 在当前业务知识网络下查找属于请求的分组范围内的对象类ID
		otIDArr, err := ots.cga.GetConceptIDsByConceptGroupIDs(ctx, query.KNID,
			query.Branch, query.ConceptGroups, interfaces.MODULE_TYPE_OBJECT_TYPE)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%v], error: %v",
				query.KNID, query.Branch, query.ConceptGroups, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)
			span.End()

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
		}

		// 概念分组下没有对象类,返回空
		if len(otIDArr) == 0 {
			return response, nil
		}

		for _, otID := range otIDArr {
			if !otIDMap[otID] {
				otIDMap[otID] = true
				otIDs = append(otIDs, otID)
			}
		}
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		if len(otIDMap) == 0 {
			// 后面分批查询会改变query，所以先查总数
			dsl, err := logics.BuildDslQuery(ctx, conditionDslStr, query)
			if err != nil {
				return response, err
			}
			total, err := ots.GetTotal(ctx, dsl)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		} else {
			// 指定了分组，需要查询分组内且符合条件的总数
			// 方法1：在OpenSearch中通过ID列表过滤查询总数
			total, err := ots.GetTotalWithLargeOTIDs(ctx, conditionDslStr, otIDs)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		}
	}

	// 4. 迭代查询直到获取足够数量或没有更多数据
	objectTypes := []*interfaces.ObjectType{}
	var totalFilteredCount int64 = 0
	for {
		// 构建当前分页的DSL
		dsl, err := logics.BuildDslQuery(ctx, conditionDslStr, query)
		if err != nil {
			return response, err
		}
		// 请求opensearch
		result, err := ots.osa.SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, dsl)
		if err != nil {
			logger.Errorf("SearchData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络对象类检索失败")
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
			// 转成 object type 的 struct
			jsonByte, err := json.Marshal(concept.Source)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal opensearch hit _source, %s", err.Error()))
			}
			var objectType interfaces.ObjectType
			err = json.Unmarshal(jsonByte, &objectType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal opensearch hit _source to Object Type, %s", err.Error()))
			}
			// 如果没有指定分组，或者对象类属于分组，则添加
			if len(otIDMap) == 0 || otIDMap[objectType.OTID] {
				// 处理数据源和操作符
				err = ots.processObjectTypeDetails(ctx, &objectType)
				if err != nil {
					return response, err
				}
				objectType.Score = &concept.Score
				objectType.Vector = nil

				objectTypes = append(objectTypes, &objectType)
				totalFilteredCount++

				// 如果已经收集到足够的数量，跳出循环
				if len(objectTypes) >= query.Limit {
					break
				}
			}
		}
		// 如果已经收集到足够的数量或者没有更多数据了，跳出循环
		if len(objectTypes) >= query.Limit || len(result) < query.Limit {
			break
		}
	}

	response.Entries = objectTypes
	response.SearchAfter = query.SearchAfter
	return response, nil
}

// 提取出来的处理对象类型详情的函数
func (ots *objectTypeService) processObjectTypeDetails(ctx context.Context, objectType *interfaces.ObjectType) error {

	// 查视图组装 ops. 不需要组装,因为保存的时候会保存进去
	if objectType.DataSource != nil && objectType.DataSource.ID != "" {
		dataView, err := ots.dva.GetDataViewByID(ctx, objectType.DataSource.ID)
		if err != nil || dataView == nil {
			o11y.Warn(ctx, fmt.Sprintf("Object type [%s]'s Data view %s not found, error: %v",
				objectType.OTID, objectType.DataSource.ID, err))
		} else {
			objectType.DataSource.Name = dataView.ViewName
			// 视图不为空，则把支持的操作符返回
			for j, prop := range objectType.DataProperties {
				// 不为空时，才翻译字段显示名。为空则不翻译
				if prop.MappedField != nil {
					if field, exists := dataView.FieldsMap[prop.MappedField.Name]; exists {
						objectType.DataProperties[j].MappedField.DisplayName = field.DisplayName
						objectType.DataProperties[j].MappedField.Type = field.Type
					}
				}
				// 字符串类型的属性支持的操作符返回
				objectType.DataProperties[j].ConditionOperations = ots.processConditionOperations(objectType, prop, dataView)
			}
		}

		// 逻辑属性，资源id转名称
		for j, logicProp := range objectType.LogicProperties {
			if logicProp.DataSource != nil {
				switch logicProp.DataSource.Type {
				case interfaces.LOGIC_PROPERTY_TYPE_METRIC:
					if logicProp.DataSource.ID != "" {
						// 获取指标模型名称
						model, err := ots.dda.GetMetricModelByID(ctx, logicProp.DataSource.ID)
						if err != nil || model == nil {

							// 依赖不存在或者请求报错，不报错，跳过
							o11y.Warn(ctx, fmt.Sprintf("Object type [%s]'s logic property [%s] metric model [%s] not found, error: %v",
								objectType.OTID, logicProp.Name, objectType.DataSource.ID, err))
						} else {
							// 依赖存在时才做相关操作
							objectType.LogicProperties[j].DataSource.Name = model.ModelName

							// 逻辑属性-指标，返回指标模型的分析维度
							objectType.LogicProperties[j].AnalysisDims = model.AnalysisDims

							// 对参数填充comment
							processMetricPropertyParamComment(ctx, logicProp, model, objectType, j)
						}
					}
				case interfaces.LOGIC_PROPERTY_TYPE_OPERATOR:
					//todo: 算子的名称,前端翻译
				}
				// todo: 处理动态参数,动态参数统一放在一个新字段上,供统一召回的大模型使用(检索那边也需要处理一下)
			}
		}
	}
	return nil
}

// 处理指标属性的参数的comment
func processMetricPropertyParamComment(ctx context.Context, logicProp *interfaces.LogicProperty, model *interfaces.MetricModel,
	objectType *interfaces.ObjectType, j int) {

	// 对参数填充comment
	for k, param := range logicProp.Parameters {
		// 存在则给，否则不给，不报错，记录warn日志
		if field, exist := model.FieldsMap[param.Name]; exist {
			objectType.LogicProperties[j].Parameters[k].Comment = field.Comment
		} else if param.Name == "instant" {
			comment := "是否是即时查询。可选，默认为 false。当 instant = true 时，表示即时查询；当 instant = false 时，表示范围查询。"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		} else if param.Name == "start" {
			comment := "指标查询的开始时间。 start=<unix_timestamp>，单位到毫秒。 例如: 1646360670123"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		} else if param.Name == "end" {
			comment := "指标查询的结束时间。end=<unix_timestamp>，单位到毫秒。例如: 1646471470123"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		} else if param.Name == "step" {
			comment := "范围查询的步长。当 instant 为 false 时, 必须。step=<time_durations>，用一个数字，后面跟时间单位来定义。"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		} else {
			// 字段不存在，记录warn日志
			o11y.Warn(ctx, fmt.Sprintf("Object type [%s]'s logic property [%s]'s parameter[%s] not found in metric model[%s]",
				objectType.OTID, logicProp.Name, param.Name, objectType.DataSource.ID))
		}
	}
}

func (ots *objectTypeService) GetTotal(ctx context.Context, dsl map[string]any) (total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: search object type total ")
	defer span.End()

	// delete(dsl, "pit")
	delete(dsl, "from")
	delete(dsl, "size")
	delete(dsl, "sort")
	totalBytes, err := ots.osa.Count(ctx, interfaces.KN_CONCEPT_INDEX_NAME, dsl)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	totalNode, err := sonic.Get(totalBytes, "count")
	if err != nil {
		span.SetStatus(codes.Error, "Get total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	total, err = totalNode.Int64()
	if err != nil {
		span.SetStatus(codes.Error, "Convert total documents count to type int64 failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 内部调用，不加权限校验
func (ots *objectTypeService) GetObjectTypeIDsByKnID(ctx context.Context,
	knID string, branch string) ([]string, error) {
	// 获取对象类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("按kn_id[%s]获取对象类IDs", knID))
	defer span.End()

	// 获取对象类基本信息
	otIDs, err := ots.ota.GetObjectTypeIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetObjectTypeIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type ids by kn_id[%s] error: %v", knID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otIDs, nil
}

func (ots *objectTypeService) GetAllObjectTypesByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("按kn_id[%s]获取对象类基本信息", knID))
	defer span.End()

	// 获取对象类基本信息
	objectTypes, err := ots.ota.GetAllObjectTypesByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetAllObjectTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get all object type by kn_id[%s] error: %v", knID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

// 内部接口，不检查权限
func (ots *objectTypeService) GetObjectTypeByID(ctx context.Context, tx *sql.Tx,
	knID string, branch string, otID string) (*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询对象类[%s]信息", otID))
	defer span.End()

	var err error
	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return &interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("GetObjectTypeByID Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("GetObjectTypeByID Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("GetObjectTypeByID Transaction Commit Success")
				o11y.Debug(ctx, "GetObjectTypeByID Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("GetObjectTypeByID Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("GetObjectTypeByID Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// 获取对象类基本信息
	objectType, err := ots.ota.GetObjectTypeByID(ctx, tx, knID, branch, otID)
	if err != nil {
		logger.Errorf("GetObjectTypeByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type by id[%s] error: %v", otID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypeByIDFailed).WithErrorDetails(err.Error())
	}

	propMap := map[string]string{}
	for _, prop := range objectType.DataProperties {
		propMap[prop.Name] = prop.DisplayName
	}
	objectType.PropertyMap = propMap

	span.SetStatus(codes.Ok, "")
	return objectType, nil
}

// 处理字符串类型的操作符
func (ots *objectTypeService) processConditionOperations(objectType *interfaces.ObjectType, prop *interfaces.DataProperty,
	dataView *interfaces.DataView) []string {

	ops := []string{}
	if objectType.Status != nil && !objectType.Status.IndexAvailable {
		// 索引不可用时,按视图的字段来做,varchar是opensearch没有的,是数据库字段.keyword和text是opensearch独有的,所以按字段类型来分
		switch prop.Type {
		case "keyword":
			ops = interfaces.DSL_KEYWORD_OPS
		case "varchar", "string":
			// string的原始类型可以是keyword或者varchar,所以按视图类型来区别一下
			if dataView.QueryType == interfaces.VIEW_QueryType_DSL {
				ops = interfaces.DSL_KEYWORD_OPS
			} else {
				ops = interfaces.SQL_STRING_OPS
			}
		case "text":
			if dataView.QueryType == interfaces.VIEW_QueryType_DSL {
				ops = interfaces.DSL_TEXT_OPS // dsl的text有match
			} else {
				ops = interfaces.SQL_STRING_OPS
			}
		case "vector":
			// 小模型打开了才能支持knn操作
			if ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
				ops = append(ops, cond.OperationKNN)
			}
		}
	} else {
		opMap := map[string]string{}
		// 先看本类型，text 类型支持 match,其余的字符串类型可支持 == != in not_in
		if prop.Type == "text" {
			opMap[cond.OperationMatch] = cond.OperationMatch
			opMap[cond.OperationMultiMatch] = cond.OperationMultiMatch
		} else {
			opMap[cond.OperationEq] = cond.OperationEq
			opMap[cond.OperationNotEq] = cond.OperationNotEq
			opMap[cond.OperationIn] = cond.OperationIn
			opMap[cond.OperationNotIn] = cond.OperationNotIn
		}

		// 配置了keyword索引,则可以做 == != in not_in的操作
		if prop.IndexConfig != nil && prop.IndexConfig.KeywordConfig.Enabled {
			opMap[cond.OperationEq] = cond.OperationEq
			opMap[cond.OperationNotEq] = cond.OperationNotEq
			opMap[cond.OperationIn] = cond.OperationIn
			opMap[cond.OperationNotIn] = cond.OperationNotIn
		}
		// 配置了full text索引,则可以做  match 的操作
		if prop.IndexConfig != nil && prop.IndexConfig.FulltextConfig.Enabled {
			opMap[cond.OperationMatch] = cond.OperationMatch
			opMap[cond.OperationMultiMatch] = cond.OperationMultiMatch
		}
		// 配置了 vector 索引, 且向量化小模型是打开的,则可以做 knn 的操作
		if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled &&
			ots.appSetting.ServerSetting.DefaultSmallModelEnabled {

			opMap[cond.OperationKNN] = cond.OperationKNN
		}

		for k := range opMap {
			ops = append(ops, k)
		}
	}
	return ops
}

// 处理对象类与组的关系，并保存
func (ots *objectTypeService) handleGroupRelations(ctx context.Context, tx *sql.Tx,
	objectType *interfaces.ObjectType, currentTime int64) error {

	cgIDs := []string{}
	for _, cg := range objectType.ConceptGroups {
		cgIDs = append(cgIDs, cg.CGID)
	}
	// id去重后再查
	cgIDs = common.DuplicateSlice(cgIDs)

	// 校验分组是否都存在，按分组id获取分组
	conceptGroups, err := ots.cga.GetConceptGroupsByIDs(ctx, tx, objectType.KNID, objectType.Branch, cgIDs)
	if err != nil {
		errStr := fmt.Sprintf("GetConceptGroupsByIDs failed, the kn_id: [%s], branch: [%s], cg_ids: [%v], error: %s",
			objectType.KNID, objectType.Branch, cgIDs, err.Error())
		logger.Errorf(errStr)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(errStr)
	}
	if len(conceptGroups) != len(cgIDs) {
		errStr := fmt.Sprintf("Exists any concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
			len(cgIDs), len(conceptGroups))
		logger.Errorf(errStr)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).
			WithErrorDetails(errStr)
	}

	// 创建
	for _, cg := range objectType.ConceptGroups {
		cgRelationID := xid.New().String()
		err = ots.cga.CreateConceptGroupRelation(ctx, tx, &interfaces.ConceptGroupRelation{
			ID:          cgRelationID,
			KNID:        objectType.KNID,
			Branch:      objectType.Branch,
			CGID:        cg.CGID,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			ConceptID:   objectType.OTID,
			CreateTime:  currentTime,
		})
		if err != nil {
			errStr := fmt.Sprintf("CreateConceptGroupRelation failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object type is [%s]",
				cg.CGID, objectType.KNID, objectType.Branch, objectType.OTID)
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError_CreateConceptGroupRelationFailed).
				WithErrorDetails(err.Error())
		}
	}
	return nil
}

// syncObjectGroups 同步分组关系（更新时使用，全量替换）
func (ots *objectTypeService) syncObjectGroups(ctx context.Context, tx *sql.Tx,
	objectType interfaces.ObjectType, currentTime int64) error {

	cgIDs := []string{}
	for _, cg := range objectType.ConceptGroups {
		cgIDs = append(cgIDs, cg.CGID)
	}
	// id去重后再查
	cgIDs = common.DuplicateSlice(cgIDs)

	// 提交的对象类的分组为空，则需要解绑对象类与概念分组的绑定关系
	// 当提交的分组不为空时才校验分组是否存在
	if len(cgIDs) > 0 {
		// 校验分组是否都存在，按分组id获取分组
		conceptGroups, err := ots.cga.GetConceptGroupsByIDs(ctx, tx, objectType.KNID, objectType.Branch, cgIDs)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptGroupsByIDs failed, the kn_id: [%s], branch: [%s], cg_ids: [%v], error: %s",
				objectType.KNID, objectType.Branch, cgIDs, err.Error())
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
		// 当提交的分组不为空时才校验分组是否存在
		if len(conceptGroups) != len(cgIDs) {
			errStr := fmt.Sprintf("Exists any concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				len(cgIDs), len(conceptGroups))
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyManager_ObjectType_InvalidParameter).
				WithErrorDetails(errStr)
		}
	}

	// 1. 获取对象类现有的分组关系
	existingRelation, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:   objectType.KNID,
		Branch: objectType.Branch,
		OTIDs:  []string{objectType.OTID},
	})
	if err != nil {
		logger.Errorf(err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyManager_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 2. 计算需要添加和删除的分组
	existingGroupIDs := make(map[string]bool)
	if len(existingRelation) == 1 {
		// 对象类已建立的关系
		for _, rel := range existingRelation[objectType.OTID] {
			existingGroupIDs[rel.CGID] = true
		}
	}

	newGroupIDs := make(map[string]bool)
	for _, ref := range objectType.ConceptGroups {
		newGroupIDs[ref.CGID] = true
	}

	// 计算差异
	groupsToAdd := make([]string, 0)
	groupsToRemove := make([]string, 0)

	for groupID := range newGroupIDs {
		if !existingGroupIDs[groupID] {
			groupsToAdd = append(groupsToAdd, groupID)
		}
	}

	for groupID := range existingGroupIDs {
		if !newGroupIDs[groupID] {
			groupsToRemove = append(groupsToRemove, groupID)
		}
	}

	// 3. 执行添加操作
	if len(groupsToAdd) > 0 {
		// 构建新增关系记录
		for _, cgID := range groupsToAdd {
			cgRelationID := xid.New().String()
			err = ots.cga.CreateConceptGroupRelation(ctx, tx, &interfaces.ConceptGroupRelation{
				ID:          cgRelationID,
				KNID:        objectType.KNID,
				Branch:      objectType.Branch,
				CGID:        cgID,
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				ConceptID:   objectType.OTID,
				CreateTime:  currentTime,
			})
			if err != nil {
				errStr := fmt.Sprintf("CreateConceptGroupRelation failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object type is [%s], error is [%s]",
					cgID, objectType.KNID, objectType.Branch, objectType.OTID, err.Error())
				logger.Errorf(errStr)

				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					oerrors.OntologyManager_ObjectType_InternalError_CreateConceptGroupRelationFailed).
					WithErrorDetails(errStr)
			}
		}
	}

	// 4. 执行删除操作
	if len(groupsToRemove) > 0 {
		// 删除对象类与分组的绑定关系
		rowsAffect, err := ots.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
			KNID:        objectType.KNID,
			Branch:      objectType.Branch,
			CGIDs:       groupsToRemove,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			OTIDs:       []string{objectType.OTID},
		})
		if err != nil {
			errStr := fmt.Sprintf("DeleteObjectTypesFromGroup failed, the concept group is [%v], kn_id is [%s], branch is [%s], object type is [%s], error is [%s]",
				groupsToRemove, objectType.KNID, objectType.Branch, objectType.OTID, err.Error())
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyManager_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
		// 记录ingo日志，删除的条数
		logger.Infof("DeleteObjectTypesFromGroup success, the concept group is [%v], kn_id is [%s], branch is [%s], object type is [%s], rowsAffect is [%d]",
			groupsToRemove, objectType.KNID, objectType.Branch, objectType.OTID, rowsAffect)
	}

	return nil
}

// 分批查询
func (ots *objectTypeService) GetTotalWithLargeOTIDs(ctx context.Context,
	conditionDslStr string,
	otIDs []string) (int64, error) {

	total := int64(0)
	for i := 0; i < len(otIDs); i += interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE {
		end := i + interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE
		if end > len(otIDs) {
			end = len(otIDs)
		}

		batchIDs := otIDs[i:end]
		batchTotal, err := ots.GetTotalWithOTIDs(ctx, conditionDslStr, batchIDs)
		if err != nil {
			return 0, err
		}

		total += batchTotal
	}

	return total, nil
}

// 查询指定对象类ID列表的对象类总数
func (ots *objectTypeService) GetTotalWithOTIDs(ctx context.Context,
	conditionDslStr string,
	otIDs []string) (int64, error) {

	var dslMap map[string]any
	err := json.Unmarshal([]byte(conditionDslStr), &dslMap)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyManager_InternalError_UnMarshalDataFailed).
			WithErrorDetails(fmt.Sprintf("failed to unMarshal dslStr to map, %s", err.Error()))
	}

	// 构建包含OTID过滤的DSL
	dsl := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					// 原有的查询条件
					dslMap,
					// OTID过滤条件
					map[string]any{
						"terms": map[string]any{
							"id": otIDs,
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
	total, err := ots.GetTotal(ctx, dsl)
	if err != nil {
		return total, err
	}

	return total, nil
}
