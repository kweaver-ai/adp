package resource_data

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func (rds *resourceDataService) QueryLogicalView(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query logical view")
	defer span.End()

	logger.Debugf("Query logical view, resourceID: %s, params: %v",
		resource.ID, params)

	var outputNode *interfaces.DataScopeNode
	var inputNodes []*interfaces.DataScopeNode
	for _, logicNode := range resource.LogicDefinition {
		switch logicNode.Type {
		case interfaces.DataScopeNodeType_Resource:
			inputNodes = append(inputNodes, logicNode)
		case interfaces.DataScopeNodeType_Output:
			outputNode = logicNode
		}
	}

	type inputNode struct {
		Node     *interfaces.DataScopeNode
		Config   *interfaces.ResourceNodeCfg
		Resource *interfaces.Resource
	}

	// 简单的衍生表查询
	inputResources := make([]*inputNode, 0, len(inputNodes))
	for _, sourceNode := range inputNodes {
		var sourceNodeConfig interfaces.ResourceNodeCfg
		err := mapstructure.Decode(sourceNode.Config, &sourceNodeConfig)
		if err != nil {
			span.SetStatus(codes.Error, "Decode source node config failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to decode source node config: %v", err))
		}

		sourceResource, err := rds.rs.GetByID(ctx, sourceNodeConfig.ResourceID)
		if err != nil {
			span.SetStatus(codes.Error, "Get source resource failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to get source resource: %v", err))
		}
		if sourceResource == nil {
			span.SetStatus(codes.Error, "Source resource not found")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
				WithErrorDetails(fmt.Sprintf("source resource %s not found", sourceNodeConfig.ResourceID))
		}
		inputResources = append(inputResources, &inputNode{
			Node:     sourceNode,
			Config:   &sourceNodeConfig,
			Resource: sourceResource,
		})
	}

	outputFields := make([]string, 0, len(outputNode.OutputFields))
	for _, f := range outputNode.OutputFields {
		outputFields = append(outputFields, f.Name)
	}

	newParams := &interfaces.ResourceDataQueryParams{
		Offset:        params.Offset,
		Limit:         params.Limit,
		Sort:          params.Sort,
		FilterCondCfg: inputResources[0].Config.Filters,
		OutputFields:  outputFields,
		NeedTotal:     params.NeedTotal,
		Format:        params.Format,
		Timeout:       params.Timeout,
	}

	return rds.Query(ctx, inputResources[0].Resource, newParams)
}
