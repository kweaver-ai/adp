package dsl_connectors

import (
	"fmt"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"vega-gateway-pro/interfaces"
)

// ConnectorHandler is an interface for handling DSL queries
type ConnectorHandler interface {
	QueryStatement(indexes []string, dsl map[string]any) (any, error)
}

// NewConnectorHandler returns a new ConnectorHandler based on the given DataSource
func NewConnectorHandler(dataSource *interfaces.DataSource) (ConnectorHandler, error) {
	switch dataSource.Type {
	case "opensearch":
		return NewOpenSearchClient(dataSource)
	default:
		logger.Errorf("unsupported data source type: %s", dataSource.Type)
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}
