package factory

import (
	"vega-manager/logics/connectors/local/index/opensearch"
	"vega-manager/logics/connectors/local/table/mysql"
)

func (cf *ConnectorFactory) InitLocalConnectors() {
	cf.connectors["mysql"] = mysql.NewMySQLConnector()
	cf.connectors["opensearch"] = opensearch.NewOpenSearchConnector()
}
