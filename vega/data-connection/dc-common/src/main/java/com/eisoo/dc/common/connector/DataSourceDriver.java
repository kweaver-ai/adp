package com.eisoo.dc.common.connector;

import com.eisoo.dc.common.vo.BinDataVo;

import java.sql.SQLException;

/**
 * 数据源连接测试驱动接口
 * 用于抽象各种数据源的连接测试功能
 */
public interface DataSourceDriver {

    /**
     * 获取驱动支持的数据源类型
     * @return 数据源类型，如 "mysql", "postgresql" 等
     */
    String getSupportedType();

    /**
     * 测试数据源连接
     * @param binData 数据源连接配置信息
     * @return 连接测试结果，true表示连接成功，false表示连接失败
     */
    boolean testConnection(BinDataVo binData) throws SQLException;

    /**
     * 验证数据源连接参数是否合法
     * @param binData 数据源连接配置信息
     * @throws IllegalArgumentException 如果参数不合法
     */
    void validateConnectionParams(BinDataVo binData);

    /**
     * 获取数据源连接URL模板
     * @return 连接URL模板，用于日志记录或调试
     */
    String getConnectionUrlTemplate();
}
