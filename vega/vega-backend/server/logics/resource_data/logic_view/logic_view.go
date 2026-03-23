package logic_view

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

func QueryLogicView(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, []any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query logic view")
	defer span.End()

	logger.Debugf("Query logic view, resourceID: %s, params: %v",
		resource.ID, params)

	// 递归解析所有资源节点，并从第一个资源节点获取 CatalogID
	var firstCatalogID string
	for _, node := range resource.LogicDefinition {
		if node.Type != interfaces.LogicDefinitionNodeType_Resource {
			continue
		}

		var nodeCfg interfaces.ResourceNodeCfg
		if err := mapstructure.Decode(node.Config, &nodeCfg); err != nil {
			span.SetStatus(codes.Error, "Decode resource node config failed")
			return nil, 0, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to decode resource node config: %v", err))
		}

		sourceResource, err := rds.rs.GetByID(ctx, nodeCfg.ResourceID)
		if err != nil {
			span.SetStatus(codes.Error, "Get source resource failed")
			return nil, 0, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to get source resource %s: %v", nodeCfg.ResourceID, err))
		}
		if sourceResource == nil {
			span.SetStatus(codes.Error, "Source resource not found")
			return nil, 0, nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
				WithErrorDetails(fmt.Sprintf("source resource %s not found", nodeCfg.ResourceID))
		}

		// 将解析后的资源填回 Config
		node.Config["resource"] = sourceResource

		// 取第一个资源节点的 CatalogID 作为逻辑视图的 CatalogID
		if firstCatalogID == "" {
			firstCatalogID = sourceResource.CatalogID
		}
	}

	// 逻辑视图本身没有 CatalogID，需要从底层资源继承
	// if resource.CatalogID == "" && firstCatalogID != "" {
	// 	resource.CatalogID = firstCatalogID
	// }

	if firstCatalogID != "" {
		resource.CatalogID = firstCatalogID
	}

	// 交给 QueryData，由底层 Connector 处理 SQL push-down
	return rds.QueryData(ctx, resource, params)
}
