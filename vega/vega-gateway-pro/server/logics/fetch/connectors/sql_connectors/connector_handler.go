package sql_connectors

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL驱动
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"vega-gateway-pro/interfaces"
)

// ConnectorHandler is an interface for handling SQL queries
type ConnectorHandler interface {
	GetResultSet(sql string) (any, error)
	GetColumns(resultSet any) ([]*interfaces.Column, error)
	GetData(resultSet any, columnSize int, batchSize int) (any, []*[]any, error)
	Close() error
}

// NewConnectorHandler returns a new ConnectorHandler based on the given DataSource
func NewConnectorHandler(dataSource *interfaces.DataSource) (ConnectorHandler, error) {
	switch dataSource.Type {
	case "mysql":
		return NewMySQLConnector(dataSource)
	case "maria":
		return NewMySQLConnector(dataSource)
	default:
		logger.Errorf("unsupported data source type: %s", dataSource.Type)
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}
