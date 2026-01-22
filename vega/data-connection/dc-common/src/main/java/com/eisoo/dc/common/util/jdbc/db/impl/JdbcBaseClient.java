package com.eisoo.dc.common.util.jdbc.db.impl;

import com.alibaba.fastjson2.JSON;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.connector.TypeConfig;
import com.eisoo.dc.common.constant.CatalogConstant;
import com.eisoo.dc.common.enums.ConnectorEnums;
import com.eisoo.dc.common.metadata.entity.AdvancedParamsDTO;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.util.RSAUtil;
import com.eisoo.dc.common.util.jdbc.db.DataSourceConfig;
import com.eisoo.dc.common.util.jdbc.db.DbClientInterface;
import com.eisoo.dc.common.util.jdbc.db.DbConnectionStrategyFactory;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;

import java.sql.*;
import java.util.*;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Slf4j
public abstract class JdbcBaseClient implements DbClientInterface {
    @Override
    public Connection getConnection(DataSourceConfig config) throws Exception {
        Class.forName(config.getDriverClass());
        Properties props = new Properties();
        String url = config.getUrl();
        String token = config.getToken();
        if (CatalogConstant.INCEPTOR_JDBC_CATALOG.equals(config.getDbType()) && StringUtils.isNotEmpty(token)) {
            url = url + ";guardianToken=" + token;
            log.info("inceptor-jdbc:url:{}", url);
        } else {
            props.setProperty("user", config.getUsername());
            props.setProperty("password", config.getPassword());
        }
        // 2. 获取连接
        return DriverManager.getConnection(
                url,
                props
        );
    }

    @Override
    public Map<String, TableScanEntity> getTables(DataSourceEntity dataSourceEntity, List<String> scanStrategy) throws Exception {
        String fType = dataSourceEntity.getFType();
        String fSchema = dataSourceEntity.getFSchema();
        if (StringUtils.isEmpty(fSchema)) {
            fSchema = dataSourceEntity.getFDatabase();
        }
        String fCatalog = dataSourceEntity.getFCatalog();
        String fId = dataSourceEntity.getFId();

        Map<String, TableScanEntity> currentTables = new HashMap<>();
        String fToken = dataSourceEntity.getFToken();
        String fPassword = dataSourceEntity.getFPassword();
        if (StringUtils.isNotEmpty(fPassword)) {
            fPassword = RSAUtil.decrypt(fPassword);
        }
        DataSourceConfig dataSourceConfig = new DataSourceConfig(
                fType,
                DbConnectionStrategyFactory.DRIVER_CLASS_MAP.get(fType),
                DbConnectionStrategyFactory.getDriverURL(dataSourceEntity),
                dataSourceEntity.getFAccount(),
                fPassword,
                fToken);
        Connection connection = null;
        Statement statement = null;
        ResultSet resultSet = null;
        try {
            connection = this.getConnection(dataSourceConfig);
            statement = connection.createStatement();
            long offset = 0;
            String sql = DbConnectionStrategyFactory.TABLE_METADATA_SQL_TEMPLATE_MAP.get(fType);
            while (true) {
                int currentBatchSize = 0;
                if ("oracle".equals(fType)) {
                    sql = String.format(sql, offset + 1000, fSchema);
                } else {
                    sql = String.format(sql, fSchema, offset);
                }
                log.info("【{}采集table元数据】:dsId:{};sql:\n {}", fType, fId, sql);
                // 高级参数
                List<AdvancedParamsDTO> advancedParamsDTOList = new ArrayList<>();
                AdvancedParamsDTO vCatalogNameParam = new AdvancedParamsDTO("vCatalogName", "fCatalog");
                advancedParamsDTOList.add(vCatalogNameParam);
//                String advancedParams = new JSONArray(priStoreSize).toJSONString();
                resultSet = statement.executeQuery(sql);
                long startTimeScan = System.currentTimeMillis();
                while (resultSet.next()) {
                    TableScanEntity tableScanEntity = new TableScanEntity();
                    tableScanEntity.setFId(UUID.randomUUID().toString());
                    tableScanEntity.setFName(resultSet.getString("table_name"));
                    tableScanEntity.setFDescription(resultSet.getString("remarks"));
                    // 高级参数
                    if (ConnectorEnums.MYSQL.getConnector().equals(fType) || ConnectorEnums.MARIA.getConnector().equals(fType)) {
                        advancedParamsDTOList.add(new AdvancedParamsDTO("engine", resultSet.getString("ENGINE")));
                        advancedParamsDTOList.add(new AdvancedParamsDTO("table_rows", String.valueOf(resultSet.getInt("TABLE_ROWS"))));
                        advancedParamsDTOList.add(new AdvancedParamsDTO("create_time", resultSet.getString("CREATE_TIME")));
                        advancedParamsDTOList.add(new AdvancedParamsDTO("update_time", resultSet.getString("UPDATE_TIME")));
                        advancedParamsDTOList.add(new AdvancedParamsDTO("data_length", String.valueOf(resultSet.getDouble("DATA_LENGTH"))));
                        advancedParamsDTOList.add(new AdvancedParamsDTO("index_length", String.valueOf(resultSet.getDouble("INDEX_LENGTH"))));
                        if ("VIEW".equals(tableScanEntity.getFDescription())) {
                            tableScanEntity.setFDescription("");
                        }
                    }
                    tableScanEntity.setFAdvancedParams(JSON.toJSONString(advancedParamsDTOList, WriteMapNullValue));
                    String tableType = resultSet.getString("table_type");
                    Integer type = null;
                    if (StringUtils.isNotEmpty(tableType)) {
                        if ("table".equalsIgnoreCase(tableType)) {
                            type = 0;
                        } else if ("view".equalsIgnoreCase(tableType)) {
                            type = 1;
                        }
                    }
                    tableScanEntity.setFScanSource(type);
                    tableScanEntity.setFDataSourceName(dataSourceEntity.getFName());
                    ++currentBatchSize;
                    currentTables.put(tableScanEntity.getFName(), tableScanEntity);
                }
                long endTime = System.currentTimeMillis();
                log.info("采集table元数据:[{}}] schema [{}}] startTime [{}}] endTime [{}}] currentBatchSize [{}}] totalSize [{}}]",
                        fCatalog,
                        fSchema,
                        startTimeScan,
                        endTime,
                        currentBatchSize,
                        currentTables.size());
                if (currentTables.size() == 0) {
                    break;
                }
                if (currentTables.size() < 1000) {
                    break;
                }
                offset += 1000;
            }
            log.info("-----------------------【采集table元数据】成功！ catalog [{}] schema [{}] totalSize [{}] -----------------------",
                    fCatalog,
                    fSchema,
                    currentTables.size());
        } catch (Exception e) {
            log.error("【{}采集table元数据】:dsId:{}", fType, fId, e);
            throw e;
        } finally {
            try {
                resultSet.close();
                statement.close();
                connection.close();
            } catch (Exception e) {
            }
        }
        return currentTables;
    }


    @Override
    public Map<String, FieldScanEntity> getFields(String tableName, DataSourceEntity dataSourceEntity, ConnectorConfig connectorConfig) throws Exception {
        return getFieldsInner(tableName, dataSourceEntity, connectorConfig);
    }

    // 模板方法：定义固定流程（用final防止子类修改执行顺序）
    public final Map<String, FieldScanEntity> getFieldsInner(String tableName, DataSourceEntity dataSourceEntity, ConnectorConfig connectorConfig) throws Exception {
        List<FieldScanEntity> fieldScanEntities = mainJob(tableName, dataSourceEntity, connectorConfig);
        processLengthAndPrecision(fieldScanEntities);
        Map<String, FieldScanEntity> result = new HashMap<>();
        for (FieldScanEntity fieldScanEntity : fieldScanEntities) {
            result.put(fieldScanEntity.getFFieldName(), fieldScanEntity);
        }
        return result;
    }

    private List<FieldScanEntity> mainJob(String tableName,
                                          DataSourceEntity dataSourceEntity,
                                          ConnectorConfig connectorConfig) throws Exception {
        List<TypeConfig> typeList = connectorConfig.getType();
        HashMap<String, String> typeMap = new HashMap<>();
        for (TypeConfig typeConfig : typeList) {
            typeMap.put(typeConfig.getSourceType(), typeConfig.getVegaType());
        }
        String fType = dataSourceEntity.getFType();
        String fToken = dataSourceEntity.getFToken();
        String fPassword = dataSourceEntity.getFPassword();
        if (StringUtils.isNotEmpty(fPassword)) {
            fPassword = RSAUtil.decrypt(fPassword);
        }
        DataSourceConfig dataSourceConfig = new DataSourceConfig(
                fType,
                DbConnectionStrategyFactory.DRIVER_CLASS_MAP.get(fType),
                DbConnectionStrategyFactory.getDriverURL(dataSourceEntity),
                dataSourceEntity.getFAccount(),
                fPassword,
                fToken);
        Connection connection = null;
        ResultSet columnSet = null;
        try {
            connection = this.getConnection(dataSourceConfig);
            DatabaseMetaData metadata = connection.getMetaData();
            String fDatabase = dataSourceEntity.getFSchema();
            if (StringUtils.isBlank(fDatabase)) {
                fDatabase = dataSourceEntity.getFDatabase();
            }
            columnSet = metadata.getColumns(
                    fDatabase,
                    null,
                    tableName,
                    null);
            List<FieldScanEntity> list = new ArrayList<>();
            while (columnSet.next()) {
                FieldScanEntity fieldScanEntity = new FieldScanEntity();
                fieldScanEntity.setFId(UUID.randomUUID().toString());
                fieldScanEntity.setFTableName(columnSet.getString("TABLE_NAME"));
                String columnDef = columnSet.getString("COLUMN_DEF");
                fieldScanEntity.setFFieldName(columnSet.getString("COLUMN_NAME"));
                String type = columnSet.getString("TYPE_NAME");
                String[] strings = StringUtils.split(type, '(');
                type = (strings == null || strings.length == 0) ? null : strings[0];
                fieldScanEntity.setFFieldType(type);
                fieldScanEntity.setFFieldComment(columnSet.getString("REMARKS"));
                Integer length = columnSet.getInt("COLUMN_SIZE");
                // 不同数据源实现有差异，因此子类实现
                String decimalDigits = processDecimalDigits(columnSet);
                Integer precision;
                try {
                    precision = Integer.valueOf(decimalDigits);
                } catch (Exception e) {
                    precision = null;
                }
                fieldScanEntity.setFFieldLength(length);
                fieldScanEntity.setFFieldPrecision(precision);
                //--------------------上面基本的处理完成，下面是不同数据源单独的处理--------------------
                // 钩子方法：需要对上述进行额外的处理
                if (needProcessColumnSet()) {
                    processColumnSet(columnSet, fieldScanEntity);
                }
                // 处理高级参数
                List<AdvancedParamsDTO> advancedParamsDTOList = new ArrayList<>();
                AdvancedParamsDTO isPrimaryKeyParam = new AdvancedParamsDTO("checkPrimaryKey", "true");
                advancedParamsDTOList.add(isPrimaryKeyParam);
                AdvancedParamsDTO columnDefParam = new AdvancedParamsDTO("COLUMN_DEF", columnDef == null ? "" : columnDef);
                advancedParamsDTOList.add(columnDefParam);
                AdvancedParamsDTO originFieldTypeParam = new AdvancedParamsDTO("originFieldType", StringUtils.lowerCase(type));
                advancedParamsDTOList.add(originFieldTypeParam);
                AdvancedParamsDTO virtualFieldTypeParam = new AdvancedParamsDTO("virtualFieldType", typeMap.get(StringUtils.lowerCase(type)));
                advancedParamsDTOList.add(virtualFieldTypeParam);
                fieldScanEntity.setFAdvancedParams(JSON.toJSONString(advancedParamsDTOList));
                list.add(fieldScanEntity);
            }
            return list;
        } catch (Exception e) {
            log.error("获取字段信息异常", e);
            throw e;
        } finally {
            try {
                if (columnSet != null) {
                    columnSet.close();
                }
                if (connection != null) {
                    connection.close();
                }
            } catch (Exception e) {
            }
        }
    }

    protected abstract String processDecimalDigits(ResultSet columnSet) throws SQLException;

    protected void processLengthAndPrecision(List<FieldScanEntity> fieldScanEntities) {
        for (FieldScanEntity fieldScanEntity : fieldScanEntities) {
            String type = fieldScanEntity.getFFieldType();
            Integer length = fieldScanEntity.getFFieldLength();
            Integer precision = fieldScanEntity.getFFieldPrecision();
            if (type.equalsIgnoreCase("decimal") || type.equalsIgnoreCase("numeric") || type.equalsIgnoreCase("number")) {
                //按BUG单709240特殊要求限制带精度字段类型字段精度为18
                //按BUG单709240特殊要求限制带精度字段类型字段长度为38
                // 特殊情况处理：当length为0且precision为-127时
                if (length != null && length == 0 && precision != null && precision == -127) {
                    length = 38;// 设置默认长度为38
                    precision = 0;// 设置默认精度为0
                } else {
                    // 一般情况处理
                    if (length == null) {
                        length = null;
                    } else {
                        length = Math.min(length, 38);
                    }
                    if (precision == null) {
                        precision = null;
                    } else {
                        precision = Math.min(precision, 18);
                    }
                }
            } else {
                if (length == null) {
                    length = null;
                }
                if (precision == null) {
                    precision = null;
                }
            }
            fieldScanEntity.setFFieldLength(length);
            fieldScanEntity.setFFieldPrecision(precision);
        }
    }

    protected boolean needProcessColumnSet() {
        return false;
    }

    protected void processColumnSet(ResultSet columnSet, FieldScanEntity fieldScanEntity) throws Exception {
        log.info("======需要额外处理columnSet======");
    }

    @Override
    public boolean judgeTwoFiledIsChange(FieldScanEntity newScanEntity, FieldScanEntity oldScanEntity) {
        String fFieldCommentNew = newScanEntity.getFFieldComment();
        String fFieldCommentOld = oldScanEntity.getFFieldComment();

        Integer fFieldLengthNew = newScanEntity.getFFieldLength();
        Integer fFieldLengthOld = oldScanEntity.getFFieldLength();

        Integer fFieldPrecisionNew = newScanEntity.getFFieldPrecision();
        Integer fFieldPrecisionOld = oldScanEntity.getFFieldPrecision();

        String fAdvancedParamsNew = newScanEntity.getFAdvancedParams();
        String fAdvancedParamsOld = oldScanEntity.getFAdvancedParams();

        String fFieldTypeNew = newScanEntity.getFFieldType();
        String fFieldTypeOld = oldScanEntity.getFFieldType();

        if (!equalsWithEmptyAsNull(fFieldCommentNew, fFieldCommentOld)) {
            return true;
        } else if (!equalsWithEmptyAsNull(fFieldLengthNew, fFieldLengthOld)) {
            return true;
        } else if (!equalsWithEmptyAsNull(fFieldPrecisionNew, fFieldPrecisionOld)) {
            return true;
        } else if (!equalsWithEmptyAsNull(fAdvancedParamsNew, fAdvancedParamsOld)) {
            return true;
        } else if (!equalsWithEmptyAsNull(fFieldTypeNew, fFieldTypeOld)) {
            return true;
        }
        return false;
    }

    private static boolean equalsWithEmptyAsNull(Integer i1, Integer i2) {
        // 步骤1：统一处理null和""，转换为同一个标识（比如空字符串）
        i1 = (i1 == null) ? 0 : i1;
        i2 = (i2 == null) ? 0 : i2;
        // 步骤2：安全比较（此时s1和s2都非null）
        return i1.equals(i2);
    }

    private static boolean equalsWithEmptyAsNull(String str1, String str2) {
        // 步骤1：统一处理null和""，转换为同一个标识（比如空字符串）
        String s1 = (str1 == null) ? "" : str1;
        String s2 = (str2 == null) ? "" : str2;
        // 步骤2：安全比较（此时s1和s2都非null）
        return s1.equals(s2);
    }
}
