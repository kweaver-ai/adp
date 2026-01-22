package com.eisoo.dc.common.enums;

import com.baomidou.mybatisplus.annotation.EnumValue;
import com.eisoo.dc.common.constant.CatalogConstant;

import java.util.*;
import java.util.stream.Collectors;

public enum ConnectorEnums {

    // 结构化数据
    MYSQL("structured", "mysql", "MySQL", "jdbc"),
    MARIA("structured", "maria", "MariaDB", "jdbc"),
    ORACLE("structured", "oracle", "Oracle", "jdbc"),
    POSTGRESQL("structured", "postgresql", "PostgreSQL", "jdbc"),
    SQLSERVER("structured", "sqlserver", "SQL Server", "jdbc"),
    DORIS("structured", "doris", "Apache Doris", "jdbc"),
    HOLOGRES("structured", "hologres","Hologres", "jdbc"),
    OPENGAUSS("structured", "opengauss","OpenGauss", "jdbc"),
    DAMENG("structured", "dameng", "Dameng", "jdbc"),
    GAUSSDB("structured", "gaussdb","GaussDB", "jdbc"),
    MONGODB("structured", "mongodb", "MongoDB", "jdbc"),
    HIVE("structured", "hive", "Apache Hive", "jdbc,thrift"),
    CLICKHOUSE("structured", "clickhouse", "ClickHouse", "jdbc"),
    INCEPTOR("structured", "inceptor-jdbc","TDH Inceptor", "jdbc"),
    MAXCOMPUTE("structured", "maxcompute","MaxCompute", "https"),

    // 非结构化数据
    EXCEL("no-structured", "excel", "Excel", "https,http"),
    ANYSHARE7("no-structured","anyshare7", "AnyShare 7.0", "https"),

    // 其他
    TINGYUN("other","tingyun", "听云", "https,http"),
    OPENSEARCH("other","opensearch", "OpenSearch", "https,http");

    //INDEXBASE("other","indexbase", "IndexBase", "https,http");

    ConnectorEnums(String type,String connector,String mapping,String connectProtocol) {
        this.type = type;
        this.connector = connector;
        this.mapping = mapping;
        this.connectProtocol = connectProtocol;
    }

    @EnumValue
    private final String type;
    private final String connector;
    private final String mapping;
    private final String connectProtocol;

    public String getType() {
        return type;
    }

    public String getConnector() {
        return connector;
    }

    public String getMapping() {
        return mapping;
    }

    public String getConnectProtocol() {
        return connectProtocol;
    }

    public static ConnectorEnums fromConnector(String connector) {
        for (ConnectorEnums connection : values()) {
            if (connection.connector.equalsIgnoreCase(connector)) {
                return connection;
            }
        }
        throw new IllegalArgumentException("No enum constant with connector: " + connector);
    }

    public static Set<String> getAllConnectors() {
        return Arrays.stream(ConnectorEnums.values())
                .map(ConnectorEnums::getConnector)
                .collect(Collectors.toSet());
    }

    /**
     * 检查连接器是否被支持
     *
     * @param connector 连接器名称
     * @return 如果支持则返回true
     * @throws IllegalArgumentException 如果不支持则抛出异常，提示安装Etrino可选包
     */
    public static boolean checkSupportedConnector(String connector) {
        ConnectorEnums connectorEnum = fromConnector(connector);
        Set<String> supportedConnectors = Collections.unmodifiableSet(new HashSet<>(getNonEtrinoConnectors()));
        
        if (!supportedConnectors.contains(connectorEnum.getConnector().toLowerCase())) {
            throw new IllegalArgumentException(
                String.format("Connector '%s' is not supported in the current installation. " +
                    "Please install the Etrino optional package to enable support for this connector.", 
                    connectorEnum.getMapping())
            );
        }
        return true;
    }

    /*
     * 获取非Etrino支持的连接器
     */
    public static List<String> getNonEtrinoConnectors() {
        return Arrays.asList(
                MYSQL.connector,
                MARIA.connector,
                OPENSEARCH.connector
        );
    }
}
