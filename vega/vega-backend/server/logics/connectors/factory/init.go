package factory

import (
	"vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/connectors/local/table/mysql"
)

// InitLocalConnectors 初始化本地 connector
func (cf *ConnectorFactory) InitLocalConnectors() {
	cf.connectors["mysql"] = mysql.NewMySQLConnector()
	cf.connectors["opensearch"] = opensearch.NewOpenSearchConnector()
}
