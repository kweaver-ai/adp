package dsl_connectors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"vega-gateway-pro/common"
	"vega-gateway-pro/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// OpenSearchClient OpenSearch 客户端
type OpenSearchClient struct {
	connInfo *interfaces.DataSource
}

// NewOpenSearchClient 创建 OpenSearch 客户端
func NewOpenSearchClient(dataSource *interfaces.DataSource) (*OpenSearchClient, error) {
	return &OpenSearchClient{
		connInfo: dataSource,
	}, nil
}

// QueryStatement 执行 DSL 查询
func (c *OpenSearchClient) QueryStatement(indexes []string, dsl map[string]any) (any, error) {
	// 构建请求 URL
	index := ""
	if len(indexes) > 0 {
		index = "/" + strings.Join(indexes, ",")
	}
	url := fmt.Sprintf("%s://%s:%d%s/_search", c.connInfo.BinData.ConnectProtocol, c.connInfo.BinData.Host, c.connInfo.BinData.Port, index)

	// 创建请求体
	dslBytes, err := json.Marshal(dsl)
	if err != nil {
		logger.Errorf("OpenSearch query marshal request body failed: %s", err.Error())
		return "", err
	}

	// 设置请求头
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	// 添加 Basic 认证
	if c.connInfo.BinData.Account != "" && c.connInfo.BinData.Password != "" {
		headers[interfaces.HTTP_HEADER_AUTHORIZATION] = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.connInfo.BinData.Account, c.connInfo.BinData.Password))))
	}

	// 发送请求
	respCode, respData, err := common.NewHTTPClient().PostNoUnmarshal(context.Background(), url, headers, dslBytes)
	if err != nil {
		logger.Errorf("OpenSearch query do request failed: %s", err.Error())
		return "", err
	}

	// 解析响应体
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(respData, &bodyMap); err != nil {
		logger.Errorf("OpenSearch query unmarshal response body failed: %s", err.Error())
		return "", err
	}

	// 检查响应状态
	if respCode != http.StatusOK {
		logger.Errorf("OpenSearch query failed: httpStatus=%d, response=%s", respCode, string(respData))

		// 尝试解析错误信息
		if errObj, ok := bodyMap["error"].(map[string]interface{}); ok {
			if errMsg, err := json.Marshal(errObj); err == nil {
				return "", errors.New(string(errMsg))
			}
		}
		return "", errors.New(string(respData))
	}

	return bodyMap, nil
}
