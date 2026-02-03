package factory

import (
	"vega-manager/logics/connectors/local/index/opensearch"
	"vega-manager/logics/connectors/local/table/mysql"
)

// InitLocalConnectors 初始化本地 connector
func (cf *ConnectorFactory) InitLocalConnectors() {
	cf.connectors["mysql"] = mysql.NewMySQLConnector()
	cf.connectors["opensearch"] = opensearch.NewOpenSearchConnector()
}
