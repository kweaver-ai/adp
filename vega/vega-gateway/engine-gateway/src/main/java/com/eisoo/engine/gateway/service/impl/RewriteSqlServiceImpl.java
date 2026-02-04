package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONUtil;
import com.aishu.af.vega.sql.extract.SqlExtractUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.service.ClientIdService;
import com.eisoo.engine.gateway.service.RewriteSqlService;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.entity.ClientIdEntity;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.AFUtil;
import com.eisoo.engine.utils.util.StringUtils;
import com.eisoo.engine.utils.vo.AuthTokenInfo;
import com.eisoo.engine.utils.vo.RowColumnRuleVo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.TimeUnit;

/**
 * @Author Xiaoxiang.er
 *
 **/
@Service
public class RewriteSqlServiceImpl implements RewriteSqlService {
    @Autowired
    ClientIdService clientIdService;

    @Value(value = "${af-auth.clientid-token}")
    private String clientIdTokenUrl;

    @Value(value = "${af-auth.row-column-rule}")
    private String rowColumnRuleUrl;

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    private String clientIdToken;

    private AFUtil afUtil = new AFUtil();

    private static final Logger log = LoggerFactory.getLogger(RewriteSqlServiceImpl.class);

    private void refreshClientIdToken() {
        ClientIdEntity clientIdEntity = clientIdService.getClientIdAndSecret();
        AuthTokenInfo authTokenInfo = afUtil.login(clientIdTokenUrl, clientIdEntity.getClientId(), clientIdEntity.getClientSecret());
        if (authTokenInfo.getStatusCode() == 400 && authTokenInfo.getErrorMsg().contains("not allowed to use authorization grant 'client_credentials'")) {
            clientIdEntity = clientIdService.reRegistClient();
            log.info("flush client_id with client_credentials clientId:{}", clientIdEntity);
            authTokenInfo = afUtil.login(clientIdTokenUrl, clientIdEntity.getClientId(), clientIdEntity.getClientSecret());
        }
        clientIdToken = authTokenInfo.getAccessToken();
    }

    private RowColumnRuleVo getRowColumnRule(String tableName, String userId, String action, String token) {
        RowColumnRuleVo rowColumnRuleVo;
        // userId为空说明是jdbc请求，token为用户类型
        if (userId == null) {
            rowColumnRuleVo = afUtil.getRowColumnRule(rowColumnRuleUrl, tableName, userId, action, token);
        } else {
            if (clientIdToken == null) {
                refreshClientIdToken();
            }
            rowColumnRuleVo = afUtil.getRowColumnRule(rowColumnRuleUrl, tableName, userId, action, clientIdToken);
            if (rowColumnRuleVo.getStatusCode() == 401) {
                refreshClientIdToken();
                rowColumnRuleVo = afUtil.getRowColumnRule(rowColumnRuleUrl, tableName, userId, action, clientIdToken);
            }
        }
        return rowColumnRuleVo;
    }

    @Override
    public RowColumnRuleVo getSqlByTable(SqlExtractUtil.TableName tableName, String userId, Set<String> allColumnSet, String action, String token, String rowRule) {
        RowColumnRuleVo rowColumnRuleVo;
        String[] arr = tableName.getTableName().split("\\.");
        // 如果不是查询逻辑视图，则不增加行列权限
        if (arr.length != 0) {
            String catalogStr = showCatalogInfo(openlookengUrl,arr[0]);
            String connectorName = "";
            if (catalogStr.contains(CatalogConstant.CONNECTOR_NAME)){
                connectorName = JSONUtil.parseObj(catalogStr).get(CatalogConstant.CONNECTOR_NAME).toString();
            }
            if (StringUtils.isNotEmpty(connectorName) && !connectorName.equals("vdm")) {
                rowColumnRuleVo = new RowColumnRuleVo();
                rowColumnRuleVo.setStatusCode(200);
                rowColumnRuleVo.setRuleSql(tableName.getTableName());
                return rowColumnRuleVo;
            }
        }
        rowColumnRuleVo = getRowColumnRule(tableName.getTableName(), userId, action, token);
        log.info("tableName:{} user_id:{} rule:{} action:{}", tableName, userId, rowColumnRuleVo.getEntries(), action);
        if (allColumnSet == null) {
            allColumnSet = new TreeSet<>();
        }
        if (rowColumnRuleVo == null || rowColumnRuleVo.getEntries() == null) {
            // 如果没有查到行列权限则认为没有任何权限
            rowColumnRuleVo.setRuleSql(" (select * from " + tableName.getTableName() + " where 1=2) ");
            return rowColumnRuleVo;
        }
        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
            if (StringUtils.isNotEmpty(entry.getColumns())) {
                for (String c : entry.getColumns()) {
                    allColumnSet.add(c);
                }
            }
        }
        // rowColumnRuleVo.setRuleSql(createUnionSqlByRule(allColumnSet, rowColumnRuleVo, tableName));
        rowColumnRuleVo.setRuleSql(createIfSqlByRule(allColumnSet, rowColumnRuleVo, tableName, rowRule));
        return rowColumnRuleVo;
    }

    /**
     * 使用if判断拼接多个行列规则生成sql
     * @param allColumnSet
     * @param rowColumnRuleVo
     * @param tableName
     * @return
     */
    private String createIfSqlByRule(Set<String> allColumnSet, RowColumnRuleVo rowColumnRuleVo, SqlExtractUtil.TableName tableName, String rowRule) {
        StringBuffer sql = new StringBuffer();
        sql.append(" (select ");
        if (allColumnSet.size() > 0) {
            for (String column : allColumnSet) {
                if (needAddColumn(column)) {
                    sql.append("\"").append(column.replaceAll("\"", "")).append("\"").append(",");
                }
            }
            sql = sql.deleteCharAt(sql.length() - 1);
        } else {
            sql.append("*");
        }
        if (sql.toString().equals(" (select")) {
            sql.append("*");
        }
        sql.append(" from ( select ");
        if (havaColumn(rowColumnRuleVo)) {
            Set<String> columnSet = getAllColumnByRule(rowColumnRuleVo);
            Map<String, String> columnRowMap = getColumnRowMap(rowColumnRuleVo);
            for (String column : allColumnSet) {
                if (!column.equals("*")) {
                    if (columnSet.contains(column)) {
                        if (columnRowMap.containsKey(column)) {
                            sql.append("if(")
                                    .append(columnRowMap.get(column)).append(",")
                                    .append("\"")
                                    .append(column.replaceAll("\"", ""))
                                    .append("\"")
                                    .append(",null) as ")
                                    .append("\"")
                                    .append(column.replaceAll("\"", ""))
                                    .append("\"")
                                    .append(",");
                        } else {
                            sql.append(column).append(",");
                        }
                    } else {
                        sql.append("null as ").append("\"").append(column.replaceAll("\"", "")).append("\"").append(",");
                    }
                }
            }
            sql = sql.deleteCharAt(sql.length() - 1);
            sql.append(" from ").append(tableName.getTableName());
            if (havaRowRule(rowColumnRuleVo)) {
                sql.append(" where (");
                for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
                    if (StringUtils.isNotEmpty(entry.getRowRule())) {
                        sql.append("(").append(entry.getRowRule()).append(") or ");
                    } else {
                        sql.append("(1=1) or ");
                    }
                }
                sql = sql.delete(sql.length() - 3, sql.length());
                sql.append(")");
                if (StringUtils.isNotEmpty(rowRule)) {
                    sql.append(" and (").append(rowRule).append(")");
                }
            } else {
                if (StringUtils.isNotEmpty(rowRule)) {
                    sql.append(" where ").append(rowRule);
                }
            }
        } else {
            // 若没有授权任何一列，则什么都查不到
            sql.append(" * from ").append(tableName.getTableName()).append(" where 1=2 ");
        }
        sql.append(")a) ");
        // 若表没有别名则默认表名为别名
        if (!tableName.haveTableAliasName()) {
            sql.append(parseTableName(tableName.getTableName()));
        }
        return sql.toString();
    }

    private String parseTableName(String fullName) {
        String[] arr = fullName.split("\\.");
        return arr[arr.length - 1];
    }

    private Map<String, String> getColumnRowMap(RowColumnRuleVo rowColumnRuleVo) {
        Map<String, String> columnRowMap = new HashMap<>();
        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
            for (String column : entry.getColumns()) {
                if (StringUtils.isNotEmpty(entry.getRowRule())) {
                    if (columnRowMap.containsKey(column)) {
                        columnRowMap.put(column, columnRowMap.get(column) + " or " + "(" + entry.getRowRule() + ")");
                    } else {
                        columnRowMap.put(column, "(" + entry.getRowRule() + ")");
                    }
                } else {
                    if (columnRowMap.containsKey(column)) {
                        columnRowMap.put(column, "(1=1) or " + columnRowMap.get(column));
                    } else {
                        columnRowMap.put(column, "(1=1)");
                    }
                }
            }
        }
        return columnRowMap;
    }

    private Set<String> getAllColumnByRule(RowColumnRuleVo rowColumnRuleVo) {
        Set<String> set = new HashSet<>();
        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
            set.addAll(entry.getColumns());
        }
        return set;
    }

    private boolean havaColumn(RowColumnRuleVo rowColumnRuleVo) {
        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
            if (entry.getColumns() != null && entry.getColumns().size() > 0) {
                return true;
            }
        }
        return false;
    }

    private boolean havaRowRule(RowColumnRuleVo rowColumnRuleVo) {
        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
            if (StringUtils.isNotEmpty(entry.getRowRule())) {
                return true;
            }
        }
        return false;
    }

//    /**
//     * 使用union拼接多个行列规则生成sql
//     * @param allColumnSet
//     * @param rowColumnRuleVo
//     * @param tableName
//     * @return
//     */
//    private String createUnionSqlByRule(Set<String> allColumnSet, RowColumnRuleVo rowColumnRuleVo, String tableName) {
//        StringBuffer sql = new StringBuffer();
//        sql.append(" (select ");
//        if (allColumnSet.size() > 0) {
//            for (String column : allColumnSet) {
//                if (needAddColumn(column)) {
//                    sql.append(column).append(",");
//                }
//            }
//            sql = sql.deleteCharAt(sql.length() - 1);
//        } else {
//            sql.append("*");
//        }
//        sql.append(" from (");
//        for (RowColumnRuleVo.Entry entry : rowColumnRuleVo.getEntries()) {
//            sql.append(" select ");
//            if (entry.getColumns() != null && entry.getColumns().size() > 0) {
//                for (String column : allColumnSet) {
//                    if (entry.getColumns().contains(column)) {
//                        sql.append(column).append(",");
//                    } else {
//                        if (needAddColumn(column)) {
//                            sql.append("null as ").append(column).append(",");
//                        }
//                    }
//                }
//                sql = sql.deleteCharAt(sql.length() - 1);
//                sql.append(" from ");
//                sql.append(tableName);
//                // 若有行限制则拼接where条件，没有则认为拥有所有行权限
//                if (StringUtils.isNotEmpty(entry.getRowRule())) {
//                    sql.append(" where ").append(entry.getRowRule()).append(" ");
//                }
//            } else {
//                // 若没有授权任何一列，则什么都查不到
//                sql.append("* from ").append(tableName).append(" where 1=2");
//            }
//            sql.append(" union");
//        }
//        sql = sql.delete(sql.length() - 5, sql.length());
//        sql.append(")a) ");
//        return sql.toString();
//    }

    private boolean needAddColumn(String column) {
        if (column.startsWith("count(") ||
                column.equals("*")) {
            return false;
        }
        return true;
    }

    /**
     * 改写sql，从而实现行列权限控制
     * @param sourceSql
     * @param userId
     * @return
     */
    @Override
    public RowColumnRuleVo rewriteSql(String sourceSql, String userId, String action, String token) {
        RowColumnRuleVo rowColumnRuleVo = new RowColumnRuleVo();
        if (!sourceSql.toLowerCase().contains("from")) {
            rowColumnRuleVo.setStatusCode(200);
            rowColumnRuleVo.setTargetSql(sourceSql);
            return rowColumnRuleVo;
        }
        Map<SqlExtractUtil.TableName, Set<String>> tableColumnMap;
        try {
            // 解析sql获取表名和字段名
            tableColumnMap = SqlExtractUtil.extractTableAndColumnRelationFromSqlNew(sourceSql);
        } catch (Exception e) {
            log.error("SQL语句解析异常,detail:{}", sourceSql);
            throw new AiShuException(ErrorCodeEnum.SqlSyntaxError, e.getMessage(), Message.MESSAGE_QUERY_SQL_SOLUTION);
        }
        // 根据表名和userId获取行列规则并生成带权限的sql
        Map<String, RowColumnRuleVo> tableSqlMap = new HashMap<>();
        for (SqlExtractUtil.TableName table : tableColumnMap.keySet()) {
            RowColumnRuleVo rc = getSqlByTable(table, userId, tableColumnMap.get(table), action, token, null);
            if (rc.getStatusCode() != 200) {
                return rc;
            }
            tableSqlMap.put(table.getTableName(), rc);
        }
        String newSql = sourceSql;
        // 把catalog.schema.table.column改成table.column
        for (SqlExtractUtil.TableName tableName : tableColumnMap.keySet()) {
            for (String column : tableColumnMap.get(tableName)) {
                String fullColumn = tableName.getTableName() + "." + column;
                if (newSql.contains(fullColumn)) {
                    newSql = newSql.replaceAll(fullColumn, parseTableName(tableName.getTableName()) + "." + column);
                }
            }
        }
        // 根据行列规则生成的sql改写用户输入的sql
        for (String table : tableSqlMap.keySet()) {
            if (newSql.contains(table + " ")) {
                newSql = newSql.replaceAll(table + " ", tableSqlMap.get(table).getRuleSql() + " ");
            } else if (newSql.contains(table + ")")) {
                newSql = newSql.replaceAll(table + "\\)", tableSqlMap.get(table).getRuleSql() + ")");
            } else {
                if (newSql.contains(table)) {
                    newSql = newSql.replaceAll(table, tableSqlMap.get(table).getRuleSql());
                } else {
                    String[] arr = table.split("\\.");
                    String tableName = arr[0] + "." + arr[1] + "." + "\"" + arr[2] + "\"";
                    if (newSql.contains(tableName)) {
                        newSql = newSql.replaceAll(tableName, tableSqlMap.get(table).getRuleSql());
                    }
                }
            }
        }
        rowColumnRuleVo.setStatusCode(200);
        rowColumnRuleVo.setTargetSql(newSql);
        return rowColumnRuleVo;
    }

    /**
     * 查询数据源详情GET请求
     *
     * @param url
     * @return
     */
    public String showCatalogInfo(String url,String catalogName) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_SHOW_CATALOG+"/"+catalogName;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGet(urlOpen, Constants.DEFAULT_AD_HOC_USER);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError, e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http showCatalogInfo 请求失败: httpStatus={}, result={}, 耗时={}ms", 
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http showCatalogInfo 请求成功: httpStatus={}, result={}, 耗时={}ms", 
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));

        return result.getResult();
    }
}
