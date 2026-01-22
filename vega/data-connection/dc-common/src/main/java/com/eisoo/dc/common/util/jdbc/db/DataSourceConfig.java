package com.eisoo.dc.common.util.jdbc.db;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * 数据源配置类：封装不同数据源的连接信息
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class DataSourceConfig {
    /**
     * 数据源类型（mysql/oracle/postgresql）
     */
    private String dbType;
    /**
     * 驱动类名（如com.mysql.cj.jdbc.Driver）
     */
    private String driverClass;
    /**
     * 数据库连接URL
     */
    private String url;
    /**
     * 用户名
     */
    private String username;
    /**
     * 密码
     */
    private String password;
    /**
     * token
     */
    private String token;
}
