// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	ibAccessOnce sync.Once
	ibAccess     interfaces.IndexBaseAccess
)

type indexBaseAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewIndexBaseAccess(appSetting *common.AppSetting) interfaces.IndexBaseAccess {
	ibAccessOnce.Do(func() {
		ibAccess = &indexBaseAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})

	return ibAccess
}

// 根据索引库类型获取索引库详情
func (iba *indexBaseAccess) GetIndexBasesByTypes(ctx context.Context,
	baseTypes []string) ([]interfaces.IndexBase, error) {

	baseTypesStr := strings.Join(baseTypes, ",")
	url := fmt.Sprintf("%s/%s", iba.appSetting.IndexBaseUrl, baseTypesStr)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] with headers[%v] finished, response code is [%d], result is [%s], error is [%v]",
		url, headers, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("Get indexbases by base types '%s' failed, %s", baseTypes, err.Error())
		logger.Error(errDetails)

		return nil, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)

			return nil, err
		}

		return nil, fmt.Errorf("get indexbases '%s' failed, errDetails: %v", baseTypes, baseError.ErrorDetails)
	}

	var bases []interfaces.IndexBase
	if err := json.Unmarshal(respData, &bases); err != nil {
		errDetails := fmt.Sprintf("Unmarshal indexbase respData failed, %s", err.Error())
		logger.Error(errDetails)

		return nil, err
	}

	// 索引库接口默认批量查询，只返回存在的索引库信息
	if len(bases) < len(baseTypes) {
		nonexistentBases := make([]string, 0)
		typesMap := make(map[string]struct{})

		for _, base := range bases {
			base := base
			typesMap[base.BaseType] = struct{}{}
		}

		for _, baseType := range baseTypes {
			baseType := baseType
			if _, ok := typesMap[baseType]; !ok {
				nonexistentBases = append(nonexistentBases, baseType)
			}
		}

		errDetails := fmt.Sprintf("IndexBases %v doesn't exist", nonexistentBases)
		logger.Error(errDetails)

		return nil, errors.New(errDetails)
	}

	return bases, nil
}
