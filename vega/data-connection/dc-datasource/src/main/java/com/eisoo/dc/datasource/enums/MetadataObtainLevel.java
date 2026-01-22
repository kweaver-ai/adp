package com.eisoo.dc.datasource.enums;

import com.eisoo.dc.common.constant.CatalogConstant;
import com.eisoo.dc.common.enums.ConnectorEnums;

public enum MetadataObtainLevel {
    /**
     * 1、支持数据源扫描和多表扫描 (opensearch、mysql、maria、oracle、inceptor-jdbc)
     */
    SCAN_MULTIPLE_TABLES(1),

    /**
     * 2、支持数据源扫描，不支持多表扫描
     */
    SCAN_SINGLE_DATA_SOURCE(2),

    /**
     * 3、不支持扫描，支持新建元数据
     */
    CAN_CREATE(3),

    /**
     * 4、不支持扫描和新建元数据
     */
    NO_SCAN_AND_NO_CREATE(4);

    private final int value;

    MetadataObtainLevel(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }

    public static int getByDsType(String dataSourceType) {

        // opensearch 支持数据源扫描和多表扫描
        if (ConnectorEnums.OPENSEARCH.getConnector().equals(dataSourceType) ||
                ConnectorEnums.MYSQL.getConnector().equals(dataSourceType) ||
                ConnectorEnums.MARIA.getConnector().equals(dataSourceType) ||
                ConnectorEnums.ORACLE.getConnector().equals(dataSourceType) ||
                ConnectorEnums.INCEPTOR.getConnector().equals(dataSourceType)) {
            return SCAN_MULTIPLE_TABLES.value;
        }

        // excel 不支持扫描, 支持新建元数据
        if (ConnectorEnums.EXCEL.getConnector().equals(dataSourceType)) {
            return CAN_CREATE.value;
        }

        // anyshare7、tingyun 不支持扫描和新建元数据
        if (ConnectorEnums.ANYSHARE7.getConnector().equals(dataSourceType) ||
                ConnectorEnums.TINGYUN.getConnector().equals(dataSourceType) ||
                CatalogConstant.INDEX_BASE_DS.equals(dataSourceType) ) {
            return NO_SCAN_AND_NO_CREATE.value;
        }

        // 其他数据源类型默认支持数据源扫描，不支持多表扫描
        return SCAN_SINGLE_DATA_SOURCE.value;
    }
}

