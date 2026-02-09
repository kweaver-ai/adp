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
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
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
	ddiServiceOnce sync.Once
	ddiService     interfaces.DataDictItemsService
)

// 字典项管理的service
type dataDictItemService struct {
	appSetting *common.AppSetting
	dda        interfaces.DataDictAccess
	ddia       interfaces.DataDictItemAccess
	db         *sql.DB
	ps         interfaces.PermissionService
}

func NewDataDictItemService(appSetting *common.AppSetting) interfaces.DataDictItemsService {
	ddiServiceOnce.Do(func() {
		ddiService = &dataDictItemService{
			appSetting: appSetting,
			dda:        logics.DDA,
			ddia:       logics.DDIA,
			db:         logics.DB,
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return ddiService
}

// 获取KV字典的所有字典项
func (ddis *dataDictItemService) GetKVDictItems(ctx context.Context, dictID string) ([]map[string]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "查询键值数据字典项")
	defer span.End()

	// 校验字典的查看权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dictID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []map[string]string{}, err
	}

	dictItemArr, err := ddis.ddia.GetKVDictItems(ctx, dictID)
	if err != nil {
		logger.Errorf("GetDataDictItems error: %s", err.Error())
		span.SetStatus(codes.Error, "Get KV Dict Items error")
		return []map[string]string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return dictItemArr, nil
}

// 获取dimension字典的所有字典项
func (ddis *dataDictItemService) GetDimensionDictItems(ctx context.Context, dictID string, dictStore string,
	dimension interfaces.Dimension) ([]map[string]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询维度数据字典项")
	defer span.End()

	// 校验字典的查看权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dictID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []map[string]string{}, err
	}

	items, err := ddis.ddia.GetDimensionDictItems(ctx, dictStore, dimension)
	if err != nil {
		logger.Errorf("GetDataDictItems error: %s", err.Error())
		span.SetStatus(codes.Error, "Get Dimension Dict Items error")
		return []map[string]string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return items, nil
}

// 更新维度信息
func (ddis *dataDictItemService) UpdateDimension(ctx context.Context, dictStore string, prefix string, newDimItem []interfaces.DimensionItem, oldDimItem []interfaces.DimensionItem) ([]interfaces.DimensionItem, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update dimension")
	span.SetAttributes(attr.Key("dict_store").String(dictStore))
	defer span.End()

	// 是否添加或者删除列
	flag := false
	for j := range newDimItem {
		// 需要新增的列
		if newDimItem[j].ID == "" {
			flag = true
			id := xid.New().String()
			newDimItem[j].ID = prefix + "_" + id
			err := ddis.ddia.AddDimensionColumn(ctx, dictStore, newDimItem[j])
			if err != nil {
				logger.Errorf("Add Dimension Column error: %v \n", err)
				span.SetStatus(codes.Error, "新增维度列失败")
				return newDimItem, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails("Add Dimension Column error!")
			}
		}
		for i, old := range oldDimItem {
			// 无需新增的列 相同从旧map删除 最后剩下删除的列
			if newDimItem[j].ID == old.ID {
				oldDimItem = append(oldDimItem[:i], oldDimItem[i+1:]...)
			}
		}
	}
	// 需要删的列
	if len(oldDimItem) > 0 {
		flag = true
		for _, old := range oldDimItem {
			// drop 删除的列
			// 达梦好像不支持删除多列
			err := ddis.ddia.DropDimensionColumn(ctx, dictStore, old)
			if err != nil {
				logger.Errorf("Drop Dimension Column error: %v \n", err)
				span.SetStatus(codes.Error, "删除维度列失败")
				return newDimItem, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails("Drop Dimension Column error!")
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return newDimItem, flag, nil
}

// 根据字典id 清空删除数据字典项 或 删除整个维度表
func (ddis *dataDictItemService) DeleteDataDictItems(ctx context.Context, dict interfaces.DataDict) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict items")
	defer span.End()

	// 校验字典的编辑权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 根据表名称dictStore
	// 字典项表 清空字典项
	// t_开头的维度表 直接删除表
	if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
		err := ddis.ddia.DeleteDataDictItems(ctx, dict.DictID)
		if err != nil {
			logger.Errorf("DeleteDataDictItems error: %s", err.Error())
			span.SetStatus(codes.Error, "清空数据字典项失败")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails(err.Error())
		}
	} else {
		err := ddis.ddia.DeleteDimensionTable(ctx, dict.DictStore)
		if err != nil {
			logger.Errorf("DeleteDimensionTable error: %s", err.Error())
			span.SetStatus(codes.Error, "清空数据字典项失败")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 分页查询数据字典项
func (ddis *dataDictItemService) ListDataDictItems(ctx context.Context, dict interfaces.DataDict,
	listDictItemsQuery interfaces.DataDictItemQueryParams) ([]map[string]string, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询数据字典项列表和总数")
	defer span.End()

	// 校验字典的查看权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []map[string]string{}, 0, err
	}

	// 调用driven层
	dictItemArr, err := ddis.ddia.ListDataDictItems(ctx, dict, listDictItemsQuery)
	if err != nil {
		logger.Errorf("ListDictItems error: %s", err.Error())
		span.SetStatus(codes.Error, "List Dict Items error")
		return dictItemArr, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_DictItemNotFound).
			WithErrorDetails(err.Error())
	}

	// 调用driven层，获取字典项总数
	total, err := ddis.ddia.GetDictItemTotal(ctx, dict, listDictItemsQuery)
	if err != nil {
		logger.Errorf("GetDictItemTotal error: %s", err.Error())
		span.SetStatus(codes.Error, "Get Dict Item Total error")
		return dictItemArr, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_DictItemNotFound).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return dictItemArr, total, nil
}

// 创建单个数据字典项
func (ddis *dataDictItemService) CreateDataDictItem(ctx context.Context,
	dict interfaces.DataDict, dimension interfaces.Dimension) (string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create data dict")
	span.SetAttributes(
		attr.Key("dict_id").String(dict.DictID),
		attr.Key("dict_name").String(dict.DictName))
	defer span.End()

	// 校验字典的编辑权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return "", err
	}

	if dict.UniqueKey {
		// 查询key维度值是否已经存在
		// cnt重复个数
		cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dimension.Keys)
		if err != nil {
			logger.Errorf("CountDictItemByKey is error: %s", err.Error())
			o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey error: %v.", err))
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails(err.Error())
		}
		// 存在重复的
		if cnt > 0 {
			o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey exist: %v.", err))
			return "", rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataDict_Duplicated_DictItemKey).
				WithErrorDetails("Dictionary item key already exist!")
		}
	}

	// 生成分布式ID itemid
	itemID := xid.New().String()

	//调用driven层
	err = ddis.ddia.CreateDataDictItem(ctx, dict.DictID, itemID, dict.DictStore, dimension)
	if err != nil {
		logger.Errorf("CreateDataDictItem error: %s", err.Error())
		span.SetStatus(codes.Error, "创建数据字典项失败")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	// 更新数据字典的update_time
	updateTime := time.Now().UnixMilli()
	err = ddis.dda.UpdateDictUpdateTime(ctx, dict.DictID, updateTime)
	if err != nil {
		logger.Errorf("UpdateDataDictItem UpdateDictUpdateTime error: %s", err.Error())
		span.SetStatus(codes.Error, "创建数据字典项，修改更新时间失败")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails("Updata dictionary time failed!")
	}

	span.SetStatus(codes.Ok, "")
	return itemID, nil
}

func (ddis *dataDictItemService) UpdateDataDictItem(ctx context.Context,
	dict interfaces.DataDict, itemID string, dimension interfaces.Dimension) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Update metric model")
	span.SetAttributes(
		attr.Key("dict_id").String(dict.DictID),
		attr.Key("item_id").String(itemID),
		attr.Key("dict_name").String(dict.DictName))
	defer span.End()

	// 校验字典的编辑权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 校验item是否存在
	oldDictItem, err := ddis.ddia.GetDictItemByItemID(ctx, dict.DictStore, itemID)
	if err != nil {
		logger.Errorf("GetDictItemByItemID is error: %s", err.Error())
		span.SetStatus(codes.Error, "修改数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_DictItemNotFound).
			WithErrorDetails(err.Error())
	}

	if dict.UniqueKey {
		// 校验修改后的 是否与已经存在的重复
		cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dimension.Keys)
		if err != nil {
			logger.Errorf("CountDictItemByKey is error: %s", err.Error())
			o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey error: %v.", err))
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
				WithErrorDetails(err.Error())
		}
		// 手动校验数据字典项重复性
		// flag 新旧字典项key是否重复
		// cnt 新字典项重复个数
		// 新名称与原来不同 cnt=0 是不重复
		// 新名称与原来相同 cnt=1 是不重复
		flag := true
		if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
			if oldDictItem["f_item_key"] != dimension.Keys[0].Value {
				flag = false
			}
		} else {
			for k, v := range oldDictItem {
				for _, di := range dimension.Keys {
					if k == di.ID {
						if v != di.Value {
							flag = false
						}
					}
				}
			}
		}
		// 相同时 减去自身
		if flag {
			cnt = cnt - 1
		}
		// cnt还大于0即为重复
		if cnt > 0 {
			span.SetStatus(codes.Error, "修改数据字典项失败")
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataDict_Duplicated_DictItemKey).
				WithErrorDetails("Dictionary item key already exist!")
		}
	}

	// 调用driven层
	err = ddis.ddia.UpdateDataDictItem(ctx, dict.DictID, itemID, dict.DictStore, dimension)
	if err != nil {
		logger.Errorf("UpdateDataDictItem error: %s", err.Error())
		span.SetStatus(codes.Error, "修改数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	// 更新数据字典的update_time
	updateTime := time.Now().UnixMilli()
	err = ddis.dda.UpdateDictUpdateTime(ctx, dict.DictID, updateTime)
	if err != nil {
		logger.Errorf("UpdateDataDictItem UpdateDictUpdateTime error: %s", err.Error())
		span.SetStatus(codes.Error, "修改数据字典项，修改更新时间失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails("Updata dictionary time failed!")
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除单个数据字典项
func (ddis *dataDictItemService) DeleteDataDictItem(ctx context.Context,
	dict interfaces.DataDict, itemID string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Delete data dict item")
	defer span.End()

	// 校验字典的编辑权限
	err := ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	rowsAffect, err := ddis.ddia.DeleteDataDictItem(ctx, dict.DictID, itemID, dict.DictStore)
	if err != nil {
		logger.Errorf("DeleteDataDictItem error: %s", err.Error())
		span.SetStatus(codes.Error, "删除数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	if rowsAffect == 0 {
		logger.Errorf("DeleteDataDictItem error: RowsAffected 0!")
		span.SetStatus(codes.Error, "删除数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictItemNotFound).
			WithErrorDetails("Dictionary item not found!")
	}

	// 更新数据字典的update_time
	updateTime := time.Now().UnixMilli()
	err = ddis.dda.UpdateDictUpdateTime(ctx, dict.DictID, updateTime)
	if err != nil {
		logger.Errorf("UpdateDataDictItem UpdateDictUpdateTime error: %s", err.Error())
		span.SetStatus(codes.Error, "删除数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails("Updata dictionary time failed!")
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 根据itemids获取[]item
func (ddis *dataDictItemService) GetDictItemsByItemIDs(ctx context.Context, dictStore string, itemIDs []string) ([]map[string]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get Dict Items By Item IDs")
	defer span.End()

	items := make([]map[string]string, 0)
	strErr := ""
	for i := 0; i < len(itemIDs); i++ {
		item, err := ddis.ddia.GetDictItemByItemID(ctx, dictStore, itemIDs[i])
		if err != nil || len(item) == 0 {
			logger.Errorf("GetDictItemsByItemIDs Failed! ")
			strErr = strErr + " " + itemIDs[i]
		} else {
			items = append(items, item)
		}
	}
	if len(items) != len(itemIDs) {
		span.SetStatus(codes.Error, "获取数据字典项失败")
		return []map[string]string{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictItemNotFound).
			WithErrorDetails("Dictionary items <" + strErr + " > not found!")
	}

	span.SetStatus(codes.Ok, "")
	return items, nil
}

// 导入数据字典项
func (ddis *dataDictItemService) ImportDataDictItems(ctx context.Context,
	dict *interfaces.DataDict, items []map[string]string, mode string) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "导入数据字典项")
	defer span.End()

	// 校验字典的编辑权限
	err = ddis.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_DATA_DICT,
		ID:   dict.DictID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_BadRequest_ItemsAreEmpty).
			WithErrorDetails(err.Error())
	}

	// 开始事务
	tx, err := ddis.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataDict_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}

	needRollback := false
	// 异常时
	defer func() {
		if !needRollback {
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("ImportDataDictItems Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "导入数据字典项事务提交失败")
			}
			logger.Infof("ImportDataDictItems Transaction Commit Success:%v", dict.DictName)
		} else {
			err := tx.Rollback()
			if err != nil {
				logger.Errorf("ImportDataDictItems Transaction Rollback Error:%v", err)
				span.SetStatus(codes.Error, "导入数据字典项事务回滚失败")
			}
		}
	}()

	var loopCount int
	if len(items)%interfaces.CREATE_DATA_DICT_ITEM_SIZE != 0 {
		loopCount = len(items)/interfaces.CREATE_DATA_DICT_ITEM_SIZE + 1
	} else {
		loopCount = len(items) / interfaces.CREATE_DATA_DICT_ITEM_SIZE
	}

	for i := 0; i < loopCount; i++ {
		var createItems []map[string]string
		if i == loopCount-1 {
			createItems = items[i*interfaces.CREATE_DATA_DICT_ITEM_SIZE : len(dict.DictItems)]
		} else {
			createItems = items[i*interfaces.CREATE_DATA_DICT_ITEM_SIZE : (i+1)*interfaces.CREATE_DATA_DICT_ITEM_SIZE]
		}

		if dict.DictType == interfaces.DATA_DICT_TYPE_KV {
			// KV手动校验与已存在的重复性，这里传来的 items 数组内不会包含重复的元素
			if dict.UniqueKey {
				var httpErr error
				createItems, needRollback, httpErr = ddis.handleKVDictImportMode(ctx, mode, needRollback, tx, dict, createItems)
				if httpErr != nil {
					return httpErr
				}
			}

			err = ddis.CreateKVDictItems(ctx, tx, dict.DictID, createItems)
			if err != nil {
				needRollback = true
				logger.Errorf("Create Dictionary Items error: %s", err.Error())
				span.SetStatus(codes.Error, "create kv dict items failed")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails(err.Error())
			}

		} else {
			// 手动校验与已存在的重复性, 这里传来的 items 数组内不会包含重复的元素
			if dict.UniqueKey {
				var httpErr error
				createItems, needRollback, httpErr = ddis.handleDimensionDictImportMode(ctx, mode, needRollback, tx, dict, createItems)
				if httpErr != nil {
					return httpErr
				}
			}

			err = ddis.CreateDimensionDictItems(ctx, tx, dict.DictID, dict.DictStore, dict.Dimension, createItems)
			if err != nil {
				needRollback = true
				logger.Errorf("Create Dictionary Items error: %s", err.Error())
				span.SetStatus(codes.Error, "create dimension dict items failed")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 更新数据字典的update_time
	updateTime := time.Now().UnixMilli()
	err = ddis.dda.UpdateDictUpdateTime(ctx, dict.DictID, updateTime)
	if err != nil {
		needRollback = true
		logger.Errorf("UpdateDataDictItem UpdateDictUpdateTime error: %s", err.Error())
		span.SetStatus(codes.Error, "导入数据字典项，修改更新时间失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails("Update dictionary time failed!")
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 对和已有重复的 kv 数据字典项做处理
func (ddis *dataDictItemService) handleKVDictImportMode(ctx context.Context, mode string, needRollback bool, tx *sql.Tx,
	dict *interfaces.DataDict, createItems []map[string]string) ([]map[string]string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "kv dict items import mode logic")
	defer span.End()

	switch mode {
	case interfaces.ImportMode_Normal:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			dict.Dimension.Keys[0].Value = createItem["key"]
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey failed: %v", err)
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey failed: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				needRollback = true
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey exist: %v", err))
				span.SetStatus(codes.Error, "dict item exists")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusForbidden,
					derrors.DataModel_DataDict_Duplicated_DictItemKey).WithErrorDetails(
					fmt.Sprintf("Dictionary item key already exist: %v", createItem))

			}
		}
	case interfaces.ImportMode_Ignore:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			dict.Dimension.Keys[0].Value = createItem["key"]
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey failed: %v", err)
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey failed: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				createItems = append(createItems[:i], createItems[i+1:]...) // 删除元素
				i--                                                         // 调整索引
			}
		}
	case interfaces.ImportMode_Overwrite:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			dict.Dimension.Keys[0].Value = createItem["key"]
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey failed: %v", err)
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey failed: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				// 通过唯一 key 获取字典项 ID
				itemIDs, err := ddis.ddia.GetDictItemIDByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
				if err != nil {
					needRollback = true
					logger.Errorf("GetDictItemIDByKey failed: %v", err)
					o11y.Error(ctx, fmt.Sprintf("GetDictItemIDByKey failed: %v", err))
					span.SetStatus(codes.Error, "get dict item ids by key failed")
					return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
				}

				// 使用事务删除数据库中的记录
				err = ddis.ddia.DeleteDataDictItemsByItemIDs(ctx, tx, dict.DictID, itemIDs, dict.DictStore)
				if err != nil {
					needRollback = true
					logger.Errorf("Delete dict items by item ids error: %v", err)
					o11y.Error(ctx, fmt.Sprintf("delete dict items by item ids failed: %v", err))
					span.SetStatus(codes.Error, "delete dict items by item ids failed")
					return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
				}
			}
		}

	default:
		return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_InvalidParameter_ImportMode).WithErrorDetails(fmt.Sprintf("unsupport import_mode %s", mode))
	}

	span.SetStatus(codes.Ok, "")
	return createItems, needRollback, nil
}

// 对和已有重复的维度数据字典项做处理
func (ddis *dataDictItemService) handleDimensionDictImportMode(ctx context.Context, mode string, needRollback bool, tx *sql.Tx,
	dict *interfaces.DataDict, createItems []map[string]string) ([]map[string]string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "dimension dict items import mode logic")
	defer span.End()

	switch mode {
	case interfaces.ImportMode_Normal:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			for j, k := range dict.Dimension.Keys {
				dict.Dimension.Keys[j].Value = createItem[k.Name]
			}
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey is error: %s", err.Error())
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey error: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				needRollback = true
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey exist: %v", err))
				span.SetStatus(codes.Error, "dict item exists")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusForbidden,
					derrors.DataModel_DataDict_Duplicated_DictItemKey).WithErrorDetails(
					fmt.Sprintf("Dictionary item key already exist: %v", createItem))

			}
		}
	case interfaces.ImportMode_Ignore:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			for j, k := range dict.Dimension.Keys {
				dict.Dimension.Keys[j].Value = createItem[k.Name]
			}
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey is error: %s", err.Error())
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey error: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				createItems = append(createItems[:i], createItems[i+1:]...)
				i--
			}
		}
	case interfaces.ImportMode_Overwrite:
		for i := 0; i < len(createItems); i++ {
			createItem := createItems[i]
			for j, k := range dict.Dimension.Keys {
				dict.Dimension.Keys[j].Value = createItem[k.Name]
			}
			cnt, err := ddis.ddia.CountDictItemByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
			if err != nil {
				needRollback = true
				logger.Errorf("CountDictItemByKey is error: %s", err.Error())
				o11y.Error(ctx, fmt.Sprintf("CountDictItemByKey error: %v", err))
				span.SetStatus(codes.Error, "count dict items by key failed")
				return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
			}
			// 存在重复的
			if cnt > 0 {
				// 通过唯一 key 获取字典项 ID
				itemIDs, err := ddis.ddia.GetDictItemIDByKey(ctx, dict.DictID, dict.DictStore, dict.Dimension.Keys)
				if err != nil {
					needRollback = true
					logger.Errorf("GetDictItemIDByKey failed: %v", err)
					o11y.Error(ctx, fmt.Sprintf("GetDictItemIDByKey failed: %v", err))
					span.SetStatus(codes.Error, "get dict item ids by key failed")
					return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
				}

				// 使用事务删除数据库中的记录
				err = ddis.ddia.DeleteDataDictItemsByItemIDs(ctx, tx, dict.DictID, itemIDs, dict.DictStore)
				if err != nil {
					needRollback = true
					logger.Errorf("Delete dict items by item ids error: %v", err)
					o11y.Error(ctx, fmt.Sprintf("delete dict items by item ids failed: %v", err))
					span.SetStatus(codes.Error, "delete dict items by item ids failed")
					return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_DataDict_InternalError).WithErrorDetails(err.Error())
				}

			}
		}
	default:
		return createItems, needRollback, rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_InvalidParameter_ImportMode).WithErrorDetails(fmt.Sprintf("unsupport import_mode %s", mode))
	}

	span.SetStatus(codes.Ok, "")
	return createItems, needRollback, nil
}

// 批量创建KV数据字典项
func (ddis *dataDictItemService) CreateKVDictItems(ctx context.Context, tx *sql.Tx, dictID string, items []map[string]string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "批量创建键值数据字典项")
	defer span.End()

	// 导入模式为忽略时，过滤之后的items可能为空，此时则不用插入数据库
	if len(items) == 0 {
		return nil
	}

	itemStructs := make([]interfaces.KvDictItem, len(items))
	for i := 0; i < len(items); i++ {
		// 生成分布式ID itemid
		itemID := xid.New().String()
		itemStructs[i] = interfaces.KvDictItem{
			ItemID:  itemID,
			Key:     items[i]["key"],
			Value:   items[i]["value"],
			Comment: items[i][interfaces.DATA_DICT_DIMENSION_NAME_COMMENT],
			DictID:  dictID,
		}
	}

	//调用driven层
	err := ddis.ddia.CreateKVDictItems(ctx, tx, dictID, itemStructs)
	if err != nil {
		logger.Errorf("CreateDataDictItem error: %s", err.Error())
		span.SetStatus(codes.Error, "批量创建键值数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 批量创建维度字典项
func (ddis *dataDictItemService) CreateDimensionDictItems(ctx context.Context, tx *sql.Tx, dictID string, dictStore string, dimension interfaces.Dimension, items []map[string]string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "批量创建维度数据字典项")
	defer span.End()

	// 导入模式为忽略时，过滤之后的items可能为空，此时则不用插入数据库
	if len(items) == 0 {
		return nil
	}

	// map 数组
	// 每个map包含各个维度键值、comment与comment值
	dimensions := []interfaces.Dimension{}
	for i := 0; i < len(items); i++ {
		// 生成分布式ID itemid
		itemID := xid.New().String()

		cpk := make([]interfaces.DimensionItem, len(dimension.Keys))
		cpv := make([]interfaces.DimensionItem, len(dimension.Values))
		copy(cpk, dimension.Keys)
		copy(cpv, dimension.Values)

		// 赋值给维度项结构体
		// {"k1":"字典项k1值","k2":"字典项k2值","v1":"字典项v1值","v2":"字典项v2值"}
		for ik, dik := range dimension.Keys {
			cpk[ik].Value = items[i][dik.Name]
		}
		for iv, div := range dimension.Values {
			cpv[iv].Value = items[i][div.Name]
		}

		tem := interfaces.Dimension{
			Keys:    cpk,
			Values:  cpv,
			ItemID:  itemID,
			Comment: items[i][interfaces.DATA_DICT_DIMENSION_NAME_COMMENT],
		}
		dimensions = append(dimensions, tem)
	}

	//调用driven层
	err := ddis.ddia.CreateDimensionDictItems(ctx, tx, dictID, dictStore, dimensions)
	if err != nil {
		logger.Errorf("CreateDataDictItem error: %s", err.Error())
		span.SetStatus(codes.Error, "批量创建维度数据字典项失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError).
			WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return nil
}
