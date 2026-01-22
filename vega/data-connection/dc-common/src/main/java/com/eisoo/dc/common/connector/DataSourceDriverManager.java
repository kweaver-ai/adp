package com.eisoo.dc.common.connector;

import com.eisoo.dc.common.vo.BinDataVo;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import java.sql.SQLException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 数据源驱动管理器
 * 负责管理和提供各种数据源的连接测试驱动
 */
@Slf4j
@Component
public class DataSourceDriverManager {

    @Autowired
    private List<DataSourceDriver> drivers;

    private Map<String, DataSourceDriver> driverMap = new HashMap<>();

    /**
     * 初始化方法，将所有驱动实现注册到驱动映射中
     */
    @PostConstruct
    public void init() {
        if (drivers != null) {
            for (DataSourceDriver driver : drivers) {
                String type = driver.getSupportedType();
                driverMap.put(type, driver);
                log.info("注册数据源驱动: {}", type);
            }
        }
    }

    /**
     * 根据数据源类型获取对应的驱动实现
     * @param type 数据源类型，如 "mysql", "postgresql" 等
     * @return 对应的驱动实现，如果不存在则返回null
     */
    public DataSourceDriver getDriver(String type) {
        return driverMap.get(type);
    }

    /**
     * 测试数据源连接
     * @param type 数据源类型
     * @param binData 数据源连接配置信息
     * @return 连接测试结果，true表示连接成功，false表示连接失败
     * @throws IllegalArgumentException 如果不支持该类型的数据源
     */
    public boolean testConnection(String type, BinDataVo binData) throws SQLException {
        DataSourceDriver driver = getDriver(type);
        if (driver == null) {
            throw new IllegalArgumentException("不支持的数据源类型: " + type);
        }

        return driver.testConnection(binData);
    }

    /**
     * 验证数据源连接参数是否合法
     * @param type 数据源类型
     * @param binData 数据源连接配置信息
     * @throws IllegalArgumentException 如果不支持该类型的数据源或参数不合法
     */
    public void validateConnectionParams(String type, BinDataVo binData) {
        DataSourceDriver driver = getDriver(type);
        if (driver == null) {
            throw new IllegalArgumentException("不支持的数据源类型: " + type);
        }

        driver.validateConnectionParams(binData);
    }

    /**
     * 获取数据源连接URL模板
     * @param type 数据源类型
     * @return 连接URL模板，如果不支持该类型则返回null
     */
    public String getConnectionUrlTemplate(String type) {
        DataSourceDriver driver = getDriver(type);
        if (driver == null) {
            return null;
        }

        return driver.getConnectionUrlTemplate();
    }
}
