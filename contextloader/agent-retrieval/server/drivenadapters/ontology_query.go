package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/common"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
	infraErr "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/errors"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/rest"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/utils"
)

type ontologyQueryClient struct {
	logger     interfaces.Logger
	baseURL    string
	httpClient interfaces.HTTPClient
}

var (
	ontologyQueryOnce sync.Once
	ontologyQuery     interfaces.DrivenOntologyQuery
)

const (
	// https://{host}:{port}/api/ontology-query/in/v1/knowledge-networks/:kn_id/object-types/:ot_id?include_type_info=true
	queryObjectInstancesURI = "/in/v1/knowledge-networks/%s/object-types/%s?include_type_info=%v&include_logic_params=%v"
	// https://{host}:{port}/api/ontology-query/in/v1/knowledge-networks/:kn_id/object-types/:ot_id/properties
	queryLogicPropertiesURI = "/in/v1/knowledge-networks/%s/object-types/%s/properties"
	// https://{host}:{port}/api/ontology-query/v1/knowledge-networks/:kn_id/action-types/:at_id
	queryActionsURI = "/in/v1/knowledge-networks/%s/action-types/%s"
	// https://{host}:{port}/api/ontology-query/in/v1/knowledge-networks/:kn_id/subgraph
	queryInstanceSubgraphURI = "/in/v1/knowledge-networks/%s/subgraph"
)

// NewOntologyQueryAccess 创建OntologyQueryAccess
func NewOntologyQueryAccess() interfaces.DrivenOntologyQuery {
	ontologyQueryOnce.Do(func() {
		configLoader := config.NewConfigLoader()
		ontologyQuery = &ontologyQueryClient{
			logger: configLoader.GetLogger(),
			baseURL: fmt.Sprintf("%s://%s:%d/api/ontology-query",
				configLoader.OntologyQuery.PrivateProtocol,
				configLoader.OntologyQuery.PrivateHost,
				configLoader.OntologyQuery.PrivatePort),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return ontologyQuery
}

// QueryObjectInstances 检索指定对象类的对象的详细数据
func (o *ontologyQueryClient) QueryObjectInstances(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (resp *interfaces.QueryObjectInstancesResp, err error) {
	uri := fmt.Sprintf(queryObjectInstancesURI, req.KnID, req.OtID, req.IncludeTypeInfo, req.IncludeLogicParams)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	_, respBody, err := o.httpClient.Post(ctx, url, header, req)
	if err != nil {
		o.logger.WithContext(ctx).Warnf("[OntologyQuery#QueryObjectInstances] QueryObjectInstances request failed, err: %v", err)
		return
	}
	resp = &interfaces.QueryObjectInstancesResp{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OntologyQuery#QueryObjectInstances] Unmarshal %s err:%v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	return
}

// QueryLogicProperties 查询逻辑属性值
func (o *ontologyQueryClient) QueryLogicProperties(ctx context.Context, req *interfaces.QueryLogicPropertiesReq) (resp *interfaces.QueryLogicPropertiesResp, err error) {
	uri := fmt.Sprintf(queryLogicPropertiesURI, req.KnID, req.OtID)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 构建请求体
	body := map[string]interface{}{
		"unique_identities": req.UniqueIdentities,
		"properties":        req.Properties,
		"dynamic_params":    req.DynamicParams,
	}

	// 📤 记录调用 ontology-query 的完整入参
	bodyJSON, _ := json.Marshal(body)
	o.logger.WithContext(ctx).Debugf("  ├─ [ontology-query 调用] URL: %s", url)
	o.logger.WithContext(ctx).Debugf("  ├─ [ontology-query 请求] Body: %s", string(bodyJSON))

	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"

	_, respBody, err := o.httpClient.Post(ctx, url, header, body)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("  └─ [ontology-query 响应] ❌ 请求失败: %v", err)
		return nil, err
	}

	resp = &interfaces.QueryLogicPropertiesResp{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("  └─ [ontology-query 响应] ❌ JSON 解析失败: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	// 📥 记录 ontology-query 的完整出参
	respJSON, _ := json.Marshal(resp)
	o.logger.WithContext(ctx).Debugf("  └─ [ontology-query 响应] ✅ 成功 (%d 条数据): %s", len(resp.Datas), string(respJSON))
	return resp, nil
}

// QueryActions 查询行动
func (o *ontologyQueryClient) QueryActions(ctx context.Context, req *interfaces.QueryActionsRequest) (resp *interfaces.QueryActionsResponse, err error) {
	uri := fmt.Sprintf(queryActionsURI, req.KnID, req.AtID)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 构建请求体
	body := map[string]interface{}{
		"unique_identities": req.UniqueIdentities,
	}

	// 记录请求日志
	bodyJSON, _ := json.Marshal(body)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryActions] URL: %s", url)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryActions] Request Body: %s", string(bodyJSON))

	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"

	_, respBody, err := o.httpClient.Post(ctx, url, header, body)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OntologyQuery#QueryActions] Request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, fmt.Sprintf("行动查询接口调用失败: %v", err))
	}

	resp = &interfaces.QueryActionsResponse{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OntologyQuery#QueryActions] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析行动查询响应失败: %v", err))
		return nil, err
	}

	// 记录响应日志
	respJSON, _ := json.Marshal(resp)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryActions] Response: %s", string(respJSON))

	return resp, nil
}

// QueryInstanceSubgraph 查询对象子图
func (o *ontologyQueryClient) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (resp *interfaces.QueryInstanceSubgraphResp, err error) {
	// 构建查询参数 - QueryType 固定为 "relation_path"
	queryParams := []string{}
	if req.IncludeLogicParams {
		queryParams = append(queryParams, fmt.Sprintf("include_logic_params=%v", req.IncludeLogicParams))
	}
	// 固定 query_type 为 relation_path
	queryParams = append(queryParams, "query_type=relation_path")

	queryStr := ""
	if len(queryParams) > 0 {
		queryStr = "?" + queryParams[0]
		for i := 1; i < len(queryParams); i++ {
			queryStr += "&" + queryParams[i]
		}
	}

	uri := fmt.Sprintf(queryInstanceSubgraphURI, req.KnID) + queryStr
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 构建请求体 - 直接透传 RelationTypePaths (interface{})
	body := map[string]interface{}{
		"relation_type_paths": req.RelationTypePaths,
	}

	// 记录请求日志
	bodyJSON, _ := json.Marshal(body)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryInstanceSubgraph] URL: %s", url)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryInstanceSubgraph] Request Body: %s", string(bodyJSON))

	// 构建请求头
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"

	// 发送请求
	_, respBody, err := o.httpClient.Post(ctx, url, header, body)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OntologyQuery#QueryInstanceSubgraph] Request failed, err: %v", err)
		return nil, err
	}

	// 解析响应 - 直接解析到 interface{}
	resp = &interfaces.QueryInstanceSubgraphResp{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OntologyQuery#QueryInstanceSubgraph] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析子图查询响应失败: %v", err))
		return nil, err
	}

	// 记录响应日志
	respJSON, _ := json.Marshal(resp)
	o.logger.WithContext(ctx).Debugf("[OntologyQuery#QueryInstanceSubgraph] Response: %s", string(respJSON))

	return resp, nil
}
