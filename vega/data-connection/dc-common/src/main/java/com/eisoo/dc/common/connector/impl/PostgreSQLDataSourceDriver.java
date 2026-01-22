package com.eisoo.dc.common.connector.impl;

import com.eisoo.dc.common.connector.DataSourceDriver;
import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.vo.BinDataVo;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;

/**
 * PostgreSQL数据源连接测试驱动实现
 */
@Slf4j
@Component
public class PostgreSQLDataSourceDriver implements DataSourceDriver {

    private static final String POSTGRESQL_TYPE = "postgresql";
    private static final String JDBC_URL_TEMPLATE = "jdbc:postgresql://%s:%d/%s";

    @Override
    public String getSupportedType() {
        return POSTGRESQL_TYPE;
    }

    @Override
    public boolean testConnection(BinDataVo binData) throws SQLException {
        // 验证参数
        validateConnectionParams(binData);

        // 构建连接URL
        String url = String.format(JDBC_URL_TEMPLATE, 
            binData.getHost(), 
            binData.getPort(), 
            binData.getDatabaseName());

        Connection connection = null;
        try {
            // 尝试建立连接
            connection = DriverManager.getConnection(url, binData.getAccount(), binData.getPassword());
            log.info("PostgreSQL连接测试成功: {}", url);
            return true;
        } catch (SQLException e) {
            log.error("PostgreSQL连接测试失败: {}, 错误信息: {}", url, e.getMessage());
            throw e;
        } finally {
            // 关闭连接
            if (connection != null) {
                try {
                    connection.close();
                } catch (SQLException e) {
                    log.warn("关闭PostgreSQL连接时发生错误: {}", e.getMessage());
                }
            }
        }
    }

    @Override
    public void validateConnectionParams(BinDataVo binData) {
        if (binData == null) {
            throw new IllegalArgumentException("数据源配置不能为空");
        }

        if (StringUtils.isBlank(binData.getHost())) {
            throw new IllegalArgumentException("主机地址" + Message.MESSAGE_INPUT_NOT_EMPTY);
        }

        if (binData.getPort() <= 0 || binData.getPort() > 65535) {
            throw new IllegalArgumentException("端口号必须在1-65535之间");
        }

        if (StringUtils.isBlank(binData.getAccount())) {
            throw new IllegalArgumentException("用户名" + Message.MESSAGE_INPUT_NOT_EMPTY);
        }

        if (StringUtils.isBlank(binData.getPassword())) {
            throw new IllegalArgumentException("密码" + Message.MESSAGE_INPUT_NOT_EMPTY);
        }

        if (StringUtils.isBlank(binData.getDatabaseName())) {
            throw new IllegalArgumentException("数据库名称" + Message.MESSAGE_INPUT_NOT_EMPTY);
        }
    }

    @Override
    public String getConnectionUrlTemplate() {
        return JDBC_URL_TEMPLATE;
    }
}
