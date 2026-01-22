package com.eisoo.dc.common.util.jdbc.db;


import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;

import java.sql.Connection;
import java.sql.SQLException;
import java.util.List;
import java.util.Map;

/**
 * 数据库连接策略接口：封装不同数据源的Connection获取逻辑
 */
public interface DbClientInterface {
    /**
     * 根据数据源配置获取Connection
     *
     * @param config 数据源配置
     * @return 数据库连接
     * @throws ClassNotFoundException 驱动类未找到
     * @throws SQLException           连接异常
     */
    Connection getConnection(DataSourceConfig config)  throws Exception;

    Map<String, TableScanEntity> getTables(DataSourceEntity dataSourceEntity,List<String> scanStrategy) throws Exception;

    Map<String, FieldScanEntity> getFields(String tableName,DataSourceEntity dataSourceEntity, ConnectorConfig connectorConfig) throws Exception;

    boolean judgeTwoFiledIsChange(FieldScanEntity newScanEntity, FieldScanEntity oldScanEntity1);
}
