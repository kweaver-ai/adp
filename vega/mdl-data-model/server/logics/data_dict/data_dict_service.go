// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

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
	ddServiceOnce sync.Once
	ddService     interfaces.DataDictService
)

// 字典管理的service
type dataDictService struct {
	appSetting *common.AppSetting
	dda        interfaces.DataDictAccess
	ddis       interfaces.DataDictItemsService
	db         *sql.DB
	ps         interfaces.PermissionService
}

func NewDataDictService(appSetting *common.AppSetting) interfaces.DataDictService {
	ddServiceOnce.Do(func() {
		ddService = &dataDictService{
			appSetting: appSetting,
			dda:        logics.DDA,
			ddis:       NewDataDictItemService(appSetting),
			db:         logics.DB,
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return ddService
}

// 分页查询数据字典
func (dds *dataDictService) ListDataDicts(ctx context.Context,
	listDictsQuery interfaces.DataDictQueryParams) ([]interfaces.DataDict, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询数据字典列表和总数")
	defer span.End()

	// 调用driven层，获取字典列表
	dictArr, err := dds.dda.ListDataDicts(ctx, listDictsQuery)
	if err != nil {
		logger.Errorf("ListDicts error: %s", err.Error())
		span.SetStatus(codes.Error, "List data dicts error")
		return []interfaces.DataDict{}, 0, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails(err.Error())
	}

	if len(dictArr) == 0 {
		return dictArr, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range dictArr {
		resMids = append(resMids, m.DictID)
	}
	matchResoucesMap, err := dds.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_DICT, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return dictArr, 0, err
	}

	// 遍历对象
	results := make([]interfaces.DataDict, 0)
	for _, dict := range dictArr {
		if resrc, exist := matchResoucesMap[dict.DictID]; exist {
			dict.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, dict)
		}
	}

	// limit = -1,则返回所有
	if listDictsQuery.Limit == -1 {
		return results, int64(len(results)), nil
	}

	// 分页
	// 检查起始位置是否越界
	if listDictsQuery.Offset < 0 || listDictsQuery.Offset >= len(results) {
		return []interfaces.DataDict{}, 0, nil
	}
	// 计算结束位置
	end := listDictsQuery.Offset + listDictsQuery.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[listDictsQuery.Offset:end], int64(len(results)), nil
}

// 根据id获取/导出 数据字典
func (dds *dataDictService) GetDataDicts(ctx context.Context, dictIDs []string) ([]interfaces.DataDict, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("获取数据字典[%v]信息", dictIDs))
	span.SetAttributes(attr.Key("dict_ids").String(fmt.Sprintf("%v", dictIDs)))
	defer span.End()

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := dds.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_DICT, dictIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return []interfaces.DataDict{}, err
	}

	dictArr := []interfaces.DataDict{}
	for i := 0; i < len(dictIDs); i++ {

		// 获取字典信息 检查字典是否存在
		dict, err := dds.dda.GetDataDictByID(ctx, dictIDs[i])
		if err != nil {
			logger.Errorf("GetDicts error: %s", err.Error())
			return []interfaces.DataDict{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
				WithErrorDetails("Dictionary " + dictIDs[i] + " not found!")
		}

		// 获取字典项信息
		// 判断类型
		// kv 字典项在t_data_dict_item
		// dimension 字典项在dict_store字段
		if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
			kvItemArr, err := dds.ddis.GetKVDictItems(ctx, dictIDs[i])
			if err != nil {
				logger.Errorf("GetDicts items error: %s", err.Error())
				return []interfaces.DataDict{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictItemNotFound).
					WithErrorDetails("Get dictionary " + dictIDs[i] + " items failed!")
			}
			dict.DictItems = kvItemArr
		}
		if dict.DictType == interfaces.DATA_DICT_TYPE_DIMENSION {
			// 创建接收字典项的结构
			dimensionItemArr, err := dds.ddis.GetDimensionDictItems(ctx, dict.DictID, dict.DictStore, dict.Dimension)
			if err != nil {
				logger.Errorf("GetDicts items error: %s", err.Error())
				return []interfaces.DataDict{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictItemNotFound).
					WithErrorDetails("Get dictionary " + dictIDs[i] + " items failed!")
			}
			dict.DictItems = dimensionItemArr
		}
		// 添加信息到dictArr
		dictArr = append(dictArr, dict)
	}

	// 遍历对象
	results := make([]interfaces.DataDict, 0)
	for _, dict := range dictArr {
		if resrc, exist := matchResoucesMap[dict.DictID]; exist {
			dict.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, dict)
		} else {
			// 无查看权限
			return results, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data dict's view_detail operation.")
		}
	}

	return results, nil
}

// 创建单个数据字典
func (dds *dataDictService) CreateDataDict(ctx context.Context,
	dict interfaces.DataDict) (dictID string, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create data dict")
	defer span.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err = dds.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return "", err
	}

	// 生成分布式ID 字典ID
	dict.DictID = xid.New().String()

	span.SetAttributes(
		attr.Key("dict_id").String(dict.DictID),
		attr.Key("dict_name").String(dict.DictName),
	)

	// 对维度字典 生成表名称 t_data_dict_dimension + 字典ID
	// 生成分布式ID 维度ID
	if dict.DictType == interfaces.DATA_DICT_TYPE_DIMENSION {
		dict.DictStore = interfaces.DATA_DICT_DIMENSION_PREFIX_TABLE + "_" + dict.DictID
		// key维度 k+id
		for i := range dict.Dimension.Keys {
			id := xid.New().String()
			dict.Dimension.Keys[i].ID = interfaces.DATA_DICT_DIMENSION_PREFIX_KEY + "_" + id
		}
		// value维度 v+id
		for j := range dict.Dimension.Values {
			id := xid.New().String()
			dict.Dimension.Values[j].ID = interfaces.DATA_DICT_DIMENSION_PREFIX_VALUE + "_" + id
		}
		// 创建多维字典数据库表
		err := dds.dda.CreateDimensionDictStore(ctx, dict.DictStore, dict.Dimension)
		if err != nil {
			span.SetStatus(codes.Error, "创建维度字典存储表失败")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails("Create Dimension Table Error!")
		}

		// 根据唯一性决定是否创建索引
		if dict.UniqueKey {
			err = dds.dda.AddDimensionIndex(ctx, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				span.SetStatus(codes.Error, "添加维度字典表索引失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails("Add Dimension Index Error")
			}
		}
	} else { // 对KV字典 添加默认的 type 和 store
		dict.DictType = interfaces.DATA_DICT_TYPE_KV
		dict.DictStore = interfaces.DATA_DICT_STORE_DEFAULT
		dict.Dimension = interfaces.DATA_DICT_KV_DIMENSION
		// dict.UniqueKey = true
	}

	// 开始事务
	tx, err := dds.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}

	neeeRollback := false
	// 异常时
	defer func() {
		if !neeeRollback {
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateDataDict Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "创建数据字典事务提交失败")
			}
			logger.Infof("CreateDataDict Transaction Commit Success:%v", dict.DictName)
		} else {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateDataDict Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "创建数据字典事务回滚失败")
			}
		}
	}()

	now := time.Now().UnixMilli()
	dict.CreateTime = now
	dict.UpdateTime = now

	// 创建字典
	err = dds.dda.CreateDataDict(ctx, tx, dict)
	if err != nil {
		neeeRollback = true
		logger.Errorf("Create Dictionary error: %s", err.Error())
		span.SetStatus(codes.Error, "创建数据字典失败")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	// 创建字典项
	if len(dict.DictItems) > 0 {
		var loopCount int
		if len(dict.DictItems)%interfaces.CREATE_DATA_DICT_ITEM_SIZE != 0 {
			loopCount = len(dict.DictItems)/interfaces.CREATE_DATA_DICT_ITEM_SIZE + 1
		} else {
			loopCount = len(dict.DictItems) / interfaces.CREATE_DATA_DICT_ITEM_SIZE
		}

		for i := 0; i < loopCount; i++ {
			var createItems []map[string]string
			if i == loopCount-1 {
				createItems = dict.DictItems[i*interfaces.CREATE_DATA_DICT_ITEM_SIZE : len(dict.DictItems)]
			} else {
				createItems = dict.DictItems[i*interfaces.CREATE_DATA_DICT_ITEM_SIZE : (i+1)*interfaces.CREATE_DATA_DICT_ITEM_SIZE]
			}
			if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
				err = dds.ddis.CreateKVDictItems(ctx, tx, dict.DictID, createItems)
			} else {
				err = dds.ddis.CreateDimensionDictItems(ctx, tx, dict.DictID, dict.DictStore, dict.Dimension, createItems)
			}
			if err != nil {
				neeeRollback = true
				logger.Errorf("Create Dictionary Items error: %s", err.Error())
				span.SetStatus(codes.Error, "创建数据字典项失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 注册资源策略
	err = dds.ps.CreateResources(ctx, []interfaces.Resource{
		{
			ID:   dict.DictID,
			Type: interfaces.RESOURCE_TYPE_DATA_DICT,
			Name: dict.DictName,
		},
	}, interfaces.DICT_COMMON_OPERATIONS)
	if err != nil {
		return "", err
	}

	span.SetStatus(codes.Ok, "")

	return dict.DictID, nil
}

// 更新数据字典
func (dds *dataDictService) UpdateDataDict(ctx context.Context, dict interfaces.DataDict) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update metric model")
	span.SetAttributes(
		attr.Key("dict_id").String(dict.DictID),
		attr.Key("dict_name").String(dict.DictName))
	defer span.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err := dds.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	//根据id拿旧信息
	oldDict, err := dds.dda.GetDataDictByID(ctx, dict.DictID)
	if err != nil {
		logger.Errorf("UpdateDictionary error: %s", err.Error())
		span.SetStatus(codes.Error, "通过id获取数据字典失败")
		return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Get old dictionary failed!")
	}

	//校验新字典名称是否存在
	if oldDict.DictName != dict.DictName {
		exist, err := dds.dda.CheckDictExistByName(ctx, dict.DictName) //新名称
		if err != nil {
			logger.Errorf("CheckDictExistByName is error: %s", err.Error())
			span.SetStatus(codes.Error, "通过名称校验数据字典存在性失败")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails(err.Error())
		}
		if exist {
			span.SetStatus(codes.Error, "通过名称校验数据字典已存在")
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataDict_DictNameExisted).
				WithErrorDetails("Dictionary already exist!")
		}
	}
	if oldDict.DictType == interfaces.DATA_DICT_TYPE_DIMENSION {
		if oldDict.UniqueKey {
			// 先删除旧索引
			err = dds.dda.DropDimensionIndex(ctx, oldDict.DictStore)
			if err != nil {
				logger.Errorf("Drop Dimension Index failed: %s", err.Error())
				span.SetStatus(codes.Error, "删除维度字典索引失败")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails(err.Error())
			}
		}
		// 更新key维度列
		newDimensionKeys, _, err := dds.ddis.UpdateDimension(ctx, oldDict.DictStore, interfaces.DATA_DICT_DIMENSION_PREFIX_KEY, dict.Dimension.Keys, oldDict.Dimension.Keys)
		if err != nil {
			logger.Errorf("Update Dimension error: %s", err.Error())
			span.SetStatus(codes.Error, "修改数据字典维度失败")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails(err.Error())
		}
		dict.Dimension.Keys = newDimensionKeys
		if oldDict.UniqueKey {
			// 创建新索引
			err = dds.dda.AddDimensionIndex(ctx, oldDict.DictStore, dict.Dimension.Keys)
			if err != nil {
				logger.Errorf("Add Dimension Index failed: %s", err.Error())
				span.SetStatus(codes.Error, "添加维度字典索引失败")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails(err.Error())
			}
		}
		// 更新value维度时，忽略删除列数
		newDimensionValues, _, err := dds.ddis.UpdateDimension(ctx, oldDict.DictStore, interfaces.DATA_DICT_DIMENSION_PREFIX_VALUE, dict.Dimension.Values, oldDict.Dimension.Values)
		if err != nil {
			logger.Errorf("Update Dimension error: %s", err.Error())
			span.SetStatus(codes.Error, "修改数据字典维度失败")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails(err.Error())
		}
		dict.Dimension.Values = newDimensionValues
	} else {
		dict.Dimension = interfaces.DATA_DICT_KV_DIMENSION
	}

	// 手动设置更新时间
	dict.UpdateTime = time.Now().UnixMilli()

	// 更新字典信息
	err = dds.dda.UpdateDataDict(ctx, dict)
	if err != nil {
		logger.Errorf("Update Dictionary error: %s", err.Error())
		span.SetStatus(codes.Error, "修改数据字典失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	// 更新资源通知
	err = dds.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   dict.DictID,
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		Name: dict.DictName,
	})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除数据字典
func (dds *dataDictService) DeleteDataDict(ctx context.Context, dict interfaces.DataDict) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dicts")
	defer span.End()

	// 先获取资源序列 fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_RESOURCE_ID_PREFIX, metricModel.ModelID),
	matchResouces, err := dds.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_DICT, []string{dict.DictID},
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return 0, err
	}
	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能删除
	if len(matchResouces) != 1 {
		return 0, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for data-dict's delete operation.")
	}

	//删除字典
	rowsAffect, err := dds.dda.DeleteDataDict(ctx, dict.DictID)
	if err != nil {
		logger.Errorf("DeleteDataDict error: %s", err.Error())
		span.SetStatus(codes.Error, "删除数据字典失败")
		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}
	if rowsAffect > 1 {
		logger.Errorf("DeleteDataDict error: RowsAffected more than 1!")
		span.SetStatus(codes.Error, "删除数据字典失败")
		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails("RowsAffected more than 1")
	}
	//清空字典项
	err = dds.ddis.DeleteDataDictItems(ctx, dict)
	if err != nil {
		logger.Errorf("DeleteDataDictItems error: %s", err.Error())
		span.SetStatus(codes.Error, "清空数据字典项失败")
		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = dds.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_DATA_DICT, []string{dict.DictID})
	if err != nil {
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return rowsAffect, nil
}

// 根据名称检查 字典存在性
func (dds *dataDictService) CheckDictExistByName(ctx context.Context, dictName string) (bool, error) {
	exist, _ := dds.dda.CheckDictExistByName(ctx, dictName)
	if exist {
		logger.Errorf("CheckDictExistByName exist")
		return true, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataDict_DictNameExisted).
			WithErrorDetails("Dictionary " + dictName + " already exist!")
	}
	return false, nil
}

// 根据id获取 字典信息 内部调用，用于校验是否存在，不需要加权限
func (dds *dataDictService) GetDataDictByID(ctx context.Context, dictID string) (interfaces.DataDict, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询数据字典[%s]信息", dictID))
	span.SetAttributes(attr.Key("dict_id").String(dictID))
	defer span.End()

	dict, err := dds.dda.GetDataDictByID(ctx, dictID)
	if err != nil {
		logger.Errorf("GetDataDictByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get data dict[%s] error: %v", dictID, err))
		return interfaces.DataDict{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")
	}

	span.SetStatus(codes.Ok, "")
	return dict, nil
}

func (dds *dataDictService) ListDataDictSrcs(ctx context.Context,
	listDictsQuery interfaces.DataDictQueryParams) ([]interfaces.Resource, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型实例列表")
	span.End()

	//获取指标模型列表（不分页，获取所有的指标模型)
	dictArr, err := dds.dda.ListDataDicts(ctx, listDictsQuery)
	if err != nil {
		logger.Errorf("ListDicts error: %s", err.Error())
		span.SetStatus(codes.Error, "List data dicts error")
		return []interfaces.Resource{}, 0, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails(err.Error())
	}
	if len(dictArr) == 0 {
		return []interfaces.Resource{}, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range dictArr {
		resMids = append(resMids, m.DictID)
	}
	// 校验权限管理的操作权限
	matchResoucesMap, err := dds.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_DICT, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return []interfaces.Resource{}, 0, err
	}

	// 遍历对象
	results := make([]interfaces.Resource, 0)
	for _, model := range dictArr {
		if _, exist := matchResoucesMap[model.DictID]; exist {
			// 如果是未分组，组名是空，此时需要把其按语言翻译未分组
			results = append(results, interfaces.Resource{
				ID:   model.DictID,
				Type: interfaces.RESOURCE_TYPE_DATA_DICT,
				Name: model.DictName,
			})
		}
	}

	// limit = -1,则返回所有
	if listDictsQuery.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if listDictsQuery.Offset < 0 || listDictsQuery.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := listDictsQuery.Offset + listDictsQuery.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[listDictsQuery.Offset:end], len(results), nil
}
