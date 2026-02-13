// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	derrors "data-model/errors"
	"data-model/interfaces"
)

// 数据字典校验函数(1): 数据字典校验和补充
func ValidateDict(ctx context.Context, dict interfaces.DataDict) error {
	// 名称非空校验
	if dict.DictName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_NullParameter_DictName).
			WithErrorDetails("Dictionary name is null")
	}

	// 名称长度校验
	if utf8.RuneCountInString(dict.DictName) > interfaces.DATA_DICT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictName).
			WithErrorDetails("The length of the dictionary name exceeds " + fmt.Sprint(interfaces.DATA_DICT_NAME_MAX_LENGTH))
	}

	// 维度个数
	if len(dict.Dimension.Keys)+len(dict.Dimension.Values) > interfaces.DATA_DICT_DIMENSION_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
			WithErrorDetails("The dictionary dimensions more than limit " + fmt.Sprint(interfaces.DATA_DICT_DIMENSION_MAX_LENGTH))
	}

	// 名称不能为 id 或者 comment
	for _, k := range dict.Dimension.Keys {
		if k.Name == interfaces.DATA_DICT_DIMENSION_NAME_ID ||
			k.Name == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {

			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails("Dictionary dimension name cannot be 'id' or 'comment' ")
		}
	}
	for _, v := range dict.Dimension.Values {
		if v.Name == interfaces.DATA_DICT_DIMENSION_NAME_ID ||
			v.Name == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {

			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails("Dictionary dimension name cannot be 'id' or 'comment' ")
		}
	}

	return nil
}

// 数据字典校验函数(2): 字符串数组是否存在相同字符串
func ValidateDuplicate(ctx context.Context, strs []string) error {
	tmpMap := make(map[string]interface{})
	for _, val := range strs {
		//判断主键为val的map是否存在
		_, ok := tmpMap[val]
		if ok {
			return fmt.Errorf("objects are duplicated: %s", val)
		} else {
			tmpMap[val] = nil
		}
	}
	return nil
}

// 校验单个 维度字典项
// keys 是key维度名称数组
// values 是values维度名称数组
func ValidateItemKeyAndValue(ctx context.Context, item map[string]string,
	keys map[string]any, values map[string]any) *rest.HTTPError {

	matchKeyCount := 0
	for k, v := range item {
		if k == interfaces.DATA_DICT_DIMENSION_NAME_ID {
			continue
		}

		if k == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {
			if utf8.RuneCountInString(v) > interfaces.COMMENT_MAX_LENGTH {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictItemComment).
					WithErrorDetails("The length of the comment exceeds " + fmt.Sprint(interfaces.COMMENT_MAX_LENGTH))
			}
			continue
		}

		if _, ok := keys[k]; ok {
			if utf8.RuneCountInString(v) > interfaces.DATA_DICT_ITEM_MAX_LENGTH {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictItemKey).
					WithErrorDetails("The length of the key exceeds " + fmt.Sprint(interfaces.DATA_DICT_ITEM_MAX_LENGTH))
			}
			matchKeyCount++
			continue
		}

		if _, ok := values[k]; ok {
			if utf8.RuneCountInString(v) > interfaces.DATA_DICT_ITEM_MAX_LENGTH {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictItemValue).
					WithErrorDetails("The length of the value exceeds " + fmt.Sprint(interfaces.DATA_DICT_ITEM_MAX_LENGTH))
			}
			continue
		}

		// item中的字段在字典项中不存在
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
			WithErrorDetails(fmt.Sprintf("Dictionary item is invalid or lacks dimensions：%s", k))
	}

	// 字典项key维度值不能全为空
	if matchKeyCount == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_NullParameter_DictItemKey).
			WithErrorDetails("Dictionary item keys cannot be all empty")
	}

	return nil
}

// 数据字典校验函数(3)：导入请求体的key value 的合法性校验
func ValidateKeyValue(ctx context.Context, key string, value string) *rest.HTTPError {
	if key == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_NullParameter_DictItemKey).
			WithErrorDetails("Dictionary item key is null")
	}
	if value == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_NullParameter_DictItemValue).
			WithErrorDetails("Dictionary item value is null")
	}
	if utf8.RuneCountInString(key) > interfaces.DATA_DICT_ITEM_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictItemKey).
			WithErrorDetails("The length of the key exceeds " + fmt.Sprint(interfaces.DATA_DICT_ITEM_MAX_LENGTH))
	}
	if utf8.RuneCountInString(value) > interfaces.DATA_DICT_ITEM_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_LengthExceeded_DictItemValue).
			WithErrorDetails("The length of the value exceeds " + fmt.Sprint(interfaces.DATA_DICT_ITEM_MAX_LENGTH))
	}
	return nil
}

// 校验请求体中维度字典的字典项合法性
func validateDimensionDictItems(ctx context.Context, dict interfaces.DataDict) *rest.HTTPError {
	// 请求体维度名称
	dimensionNames := []string{}
	dimensionKeyNames := []string{}
	dimensionKeys := map[string]any{}
	dimensionValues := map[string]any{}
	for _, k := range dict.Dimension.Keys {
		if k.Name == interfaces.DATA_DICT_DIMENSION_NAME_ID ||
			k.Name == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {

			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails("Dictionary dimension name cannot be id or comment")
			return httpErr
		}
		dimensionNames = append(dimensionNames, k.Name)
		dimensionKeyNames = append(dimensionKeyNames, k.Name)
		dimensionKeys[k.Name] = nil
	}
	for _, v := range dict.Dimension.Values {
		if v.Name == interfaces.DATA_DICT_DIMENSION_NAME_ID ||
			v.Name == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {

			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictDimension).
				WithErrorDetails("Dictionary dimension name cannot be id or comment")
			return httpErr
		}
		dimensionNames = append(dimensionNames, v.Name)
		dimensionValues[v.Name] = nil
	}

	err := ValidateDuplicate(ctx, dimensionNames)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictDimension).
			WithErrorDetails("Dictionary dimension duplicated: " + err.Error())
		return httpErr
	}

	// dimension 字典项合法性校验 comment合法性校验
	if len(dict.DictItems) == 0 {
		return nil
	}

	keySet := map[string]any{}
	for _, item := range dict.DictItems {
		// 检查字典项值是否长度超限制 key维度值是否全为空
		// 检查map中的key是否都存在于dimension字段维度中
		httpErr := ValidateItemKeyAndValue(ctx, item, dimensionKeys, dimensionValues)
		if httpErr != nil {
			return httpErr
		}

		// 校验字典项comment合法性
		if comment, ok := item["comment"]; ok {
			err = validateObjectComment(ctx, comment)
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				return rest.NewHTTPError(ctx, http.StatusBadRequest, httpErr.BaseError.ErrorCode)
			}
		}

		if dict.UniqueKey {
			keys := []string{}
			for _, k := range dimensionKeyNames {
				keys = append(keys, item[k])
			}
			itemKey := strings.Join(keys, interfaces.ITEM_KEY_SEPARATOR)
			if _, ok := keySet[itemKey]; ok {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictItemKeyInFile).
					WithDescription(map[string]any{"ItemKey": strings.Join(keys, " ")}).
					WithErrorDetails(fmt.Sprintf("Dictionary item is duplicated in the file: '%v'", item))
				return httpErr
			}
			keySet[itemKey] = nil
		}
	}

	return nil
}

// 校验请求体中KV字典的字典项合法性
func validateKVDictItems(ctx context.Context, dict interfaces.DataDict) *rest.HTTPError {
	// KV 字典项合法性校验 字典项comment合法性校验
	if len(dict.DictItems) == 0 {
		return nil
	}

	keySet := map[string]any{}
	for _, v := range dict.DictItems {
		// map -> json -> struct
		item := interfaces.KvDictItem{}
		itemByte, err := sonic.Marshal(v)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictItems)
		}
		err = sonic.Unmarshal(itemByte, &item)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictItems)
		}

		// 校验 K V 合法性
		httpErr := ValidateKeyValue(ctx, item.Key, item.Value)
		if httpErr != nil {
			return httpErr
		}

		// 校验字典项comment合法性
		err = validateObjectComment(ctx, item.Comment)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			return httpErr
		}

		// 字典项校验key是否重复
		if dict.UniqueKey {
			if _, ok := keySet[item.Key]; ok {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictItemKeyInFile).
					WithDescription(map[string]any{"ItemKey": item.Key}).
					WithErrorDetails(fmt.Sprintf("Dictionary item is duplicated in the file: '%s'", item.Key))
				return httpErr
			}
			keySet[item.Key] = nil
		}
	}

	return nil
}

func validateKVItem(ctx context.Context, dataDictItem map[string]string, dict interfaces.DataDict) error {
	itemKey := ""
	itemValue := ""
	if dataDictItem["key"] != "" {
		itemKey = dataDictItem["key"]
	}
	if dataDictItem["value"] != "" {
		itemValue = dataDictItem["value"]
	}
	// 校验key是否合法
	httpErr := ValidateKeyValue(ctx, itemKey, itemValue)
	if httpErr != nil {
		return httpErr
	}
	// 赋值给维度项结构体
	// {"key":"字典项key值","value":"字典项value值"}
	for ik, dik := range dict.Dimension.Keys {
		// name key
		// id item_key
		// value 字典项key值
		if dik.Name == "key" {
			dict.Dimension.Keys[ik].Value = dataDictItem["key"]
		}
	}
	for iv, div := range dict.Dimension.Values {
		// name value
		// id item_value
		// value 字典项key值
		if div.Name == "value" {
			dict.Dimension.Values[iv].Value = dataDictItem["value"]
		}
	}
	return nil
}

func validateDimensionItem(ctx context.Context, dataDictItem map[string]string, dict interfaces.DataDict) error {
	dimensionKeys := map[string]any{}
	dimensionValues := map[string]any{}
	for _, k := range dict.Dimension.Keys {
		dimensionKeys[k.Name] = nil
	}
	for _, v := range dict.Dimension.Values {
		dimensionValues[v.Name] = nil
	}
	// 检查字典项值是否长度超限制 key维度值是否全为空
	// 检查字典项map中的key（维度名称）是否都存在于dimension字段维度中
	httpErr := ValidateItemKeyAndValue(ctx, dataDictItem, dimensionKeys, dimensionValues)
	if httpErr != nil {
		return httpErr
	}
	// 赋值给维度项结构体
	// {"k1":"字典项k1值","k2":"字典项k2值","v1":"字典项v1值","v2":"字典项v2值"}
	for k, v := range dataDictItem {
		for ik, dik := range dict.Dimension.Keys {
			// name k 即 维度名称
			// id 具体列名
			// value 字典项k1值
			if k == dik.Name {
				dict.Dimension.Keys[ik].Value = v
			}
		}
		for iv, div := range dict.Dimension.Values {
			if k == div.Name {
				dict.Dimension.Values[iv].Value = v
			}
		}
	}
	return nil
}
