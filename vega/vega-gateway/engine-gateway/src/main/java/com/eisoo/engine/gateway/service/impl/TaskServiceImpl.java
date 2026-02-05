package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.common.QueryConstant;
import com.eisoo.engine.gateway.common.SqlTemplateConstant;
import com.eisoo.engine.gateway.domain.dto.ResponseMessage;
import com.eisoo.engine.gateway.domain.vo.AsyncQuery;
import com.eisoo.engine.gateway.domain.vo.ParamVo;
import com.eisoo.engine.gateway.domain.vo.TaskInfo;
import com.eisoo.engine.gateway.domain.vo.FieldInfoVo;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.domain.vo.QueryParam;
import com.eisoo.engine.gateway.service.TaskService;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.entity.TaskEntity;
import com.eisoo.engine.metadata.entity.TaskEntityQuery;
import com.eisoo.engine.metadata.mapper.QueryMapper;
import com.eisoo.engine.metadata.mapper.TaskMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.JsonUtil;
import com.eisoo.engine.utils.util.StringUtils;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import io.prestosql.sql.parser.ParsingOptions;
import io.prestosql.sql.parser.SqlParser;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CompletionException;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

/**
 * paul.yan
 */

@Service
public class TaskServiceImpl implements TaskService {

    private static final Logger log = LoggerFactory.getLogger(TaskServiceImpl.class);

    @Autowired(required = false)
    TaskMapper taskMapper;
    @Autowired(required = false)
    QueryMapper queryMapper;

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    @Value(value = "${sql.enabled}")
    private boolean enabled;

    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyyMMdd_HHmmss");
    private static final Random RANDOM = new Random();
    private static final String LETTERS = "abcdefghijklmnopqrstuvwxyz";

    public ResponseEntity<?> Task(String statement, String user, String type) {
        ObjectMapper mapper = new ObjectMapper();
        AsyncQuery asyncQuery;
        //type 1 filed data exploration
        if(StringUtils.endsWith(type,"1")){
            return scan(statement,user,type);
        }
        try {
            asyncQuery = mapper.readValue(statement, AsyncQuery.class);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.JSON_ANALYZE_ERROR, Message.MESSAGE_QUERY_SQL_SOLUTION);
        }
        //sql validate
        sqlValidate(enabled, asyncQuery.statement);

        String message;
        try {
             message = getStatement(statement, user, type);
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError, e.getMessage(), Message.MESSAGE_OPENLOOKENG_ERROR);
        }

        String response;
        try {
            response = printResponseJson(message);
        } catch (Exception e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.JSON_ANALYZE_ERROR, Message.MESSAGE_JSON_SOLUTION);
        }
        return ResponseEntity.ok(response);
    }

    public ResponseEntity<?> scan(String statement, String user, String type) {
        ObjectMapper mapper = new ObjectMapper();
        QueryParam queryParam = null;
        try {
            queryParam = mapper.readValue(statement, QueryParam.class);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            if(e.getMessage().contains("groupLimit must be an String")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, "groupLimit传入非数字字符串", "", "请检查参数的值");
            }else if(e.getMessage().contains("Limit must be an String")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, "limit传入非数字字符串", "", "请检查参数的值");
            }
            throw new RuntimeException(e);
        }
        if (!queryParam.getLimit().matches("^\\d*$")){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, "limit传入非数字字符串", "", "请检查参数的值");
        }
        if(!queryParam.getGroupLimit().matches("^\\d*$")){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, "groupLimit传入非数字字符串", "", "请检查参数的值");
        }
        StringBuilder sql = new StringBuilder("SELECT ");
        StringBuilder sqlAvg = new StringBuilder("");
        StringBuilder sqlUnion = new StringBuilder("");
        StringBuilder sqlGroup = new StringBuilder("");

        String limit=queryParam.getLimit();
        String groupLimit=queryParam.getGroupLimit();
        String error=checkLimit(limit,groupLimit);
        String catalogName = "\"" + queryParam.getCatalogName() + "\"";
        String schema = "\"" + queryParam.getSchema() + "\"";
        String table = "\"" + queryParam.getTable() + "\"";
        if(StringUtils.isNotEmpty(error)){
            return ResponseEntity.ok(error);
        }

        String resultCatalog = validationCatalog(queryParam.getCatalogName(), user);
        Map<String, List<Object>> stringListMap=null;
        try {
            stringListMap = parseCatalog(resultCatalog);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }

        if (queryParam.getCatalogName()==null||stringListMap.get(queryParam.getCatalogName()) == null) {
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist);
        }

        if (!stringListMap.get(queryParam.getCatalogName()).contains(queryParam.getSchema())) {
            throw new AiShuException(ErrorCodeEnum.SchemaNotExist, queryParam.getSchema(), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        //校验字段类型
        validateFieldTypes(queryParam.getFields());

        if (queryParam.getTopic().matches("\\d+|.*[\u4e00-\u9fff]+.*")) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, "参数异常: topic 必须为字符串类型", "", "请检查topic参数");
        }

        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_COLUMNS + queryParam.getCatalogName() + "/" + schema + "/" + queryParam.getTable();

        JsonNode rootNode = getTableJsonNode(urlOpen, user);
        if (rootNode.isArray() && rootNode.isEmpty()) {
            throw new AiShuException(ErrorCodeEnum.TableNotExist, null, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        QueryParam finalQueryParam = queryParam;
        List<String> sqlFragments = queryParam.getFields().stream()
                .map(fieldParam -> generateSqlFragments(fieldParam, finalQueryParam))
                .collect(Collectors.toList());
        if (sqlFragments.size() == 1) {
            String[] args = sqlFragments.get(0).split("&");
            for (int i = 0; i < args.length; i++) {
                String str = args[i];
                if (str.contains("group by ") || str.contains("DAY") || str.contains("MONTH") ||str.contains("YEAR")) {
                    sqlGroup.append(str).append(" union all ");
                }else if(StringUtils.isNotEmpty(str)) {
                    sqlAvg.append(str).append(",");
                }
            }
        } else {
            for (String sqlFragment : sqlFragments) {
                if(sqlFragment.contains("&")){
                    String[] args = sqlFragment.split("&");
                    for (int i = 0; i < args.length; i++) {
                        String str = args[i];
                        if (str.contains("group by ") || str.contains("DAY") || str.contains("MONTH") ||str.contains("YEAR")) {
                            sqlGroup.append(str).append(" union all ");
                        } else if(StringUtils.isNotEmpty(str)) {
                            sqlAvg.append(str).append(",");
                        }
                    }
                }
            }
        }
        String sqlNotNull = buildNotNullSql(queryParam);
        //sqlAvg.setLength(sqlAvg.length()-2);
        if(StringUtils.isNotEmpty(sqlAvg)){
            sqlAvg.append("  COUNT(1) AS COUNT ");
            sql.append(sqlAvg);
            sql.append(" FROM ").append(catalogName).append(".");
            sql.append(schema).append(".");
            sql.append(table);

            if(StringUtils.isNotEmpty(queryParam.getLimit())){
                sql.append(" LIMIT ").append(queryParam.getLimit());
            }
        }else if(StringUtils.isNotEmpty(sqlUnion) || StringUtils.isNotEmpty(sqlGroup)){
            sqlAvg.append(" COUNT(1) AS COUNT ");
            sql.append(sqlAvg);
            sql.append(" FROM ").append(catalogName).append(".");
            sql.append(schema).append(".");
            sql.append(table);
        }

        if(StringUtils.isNotEmpty(sqlUnion.toString())){
            sqlUnion.setLength(sqlUnion.length()-10);
        }
        String sqlAverage=sql.toString();
        if(StringUtils.endsWithIgnoreCase("SELECT ",sqlAverage)){
            sqlAverage=null;
        }
        String sqlUnions=sqlUnion.toString();
        String groups=sqlGroup.toString();
        if(groups.endsWith("union all ")){
            sqlGroup.setLength(sqlGroup.length()-10);
        }

        String groupSQL=replaceStr(sqlGroup.toString());
        String taskId=generateRule();
        if(StringUtils.isEmpty(sqlAverage) && StringUtils.isEmpty(sqlUnions) &&StringUtils.isEmpty(sqlNotNull)
        &&StringUtils.isEmpty(groupSQL)){
            throw new AiShuException(ErrorCodeEnum.RuleError);
        }


        String jsonStr = JsonUtil.buildSqlRequest(sqlAverage, sqlUnions, sqlNotNull, groupSQL, taskId, queryParam.getTopic());
        getCreateModelAsync(jsonStr,user);
        return ResponseEntity.ok(toJsonTaskId(taskId));
    }

    private void validateFieldTypes(List<FieldInfoVo> fields) {
        String[] commonTypes = {"varchar", "char", "int", "bigint", "smallint", "tinyint", "decimal", "float", "double", "date", "datetime", "timestamp"};
        String[] commonFunction = {"NotNull", "NullCount","CountTable", "BlankCount", "Max", "Min", "Zero", "Avg", "StddevPop", "VarPop", "True", "False", "Day", "Month", "Year", "Quantile", "Group"};

        List<String> commonTypeList = Arrays.asList(commonTypes);
        List<String> commonFunctionList = Arrays.asList(commonFunction);
        fields.stream()
                .filter(field -> commonTypeList.stream().noneMatch(field.getType().toLowerCase()::startsWith))
                .findFirst()
                .ifPresent(field -> {
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter,
                            "fields.type不支持输入【key="+field.getKey() +"】，可支持输入type包括【varchar, char, int, bigint, smallint, tinyint, decimal, float, double, date, datetime, timestamp】",
                            Message.MESSAGE_PARAM_ERROR_SOLUTION);
                });

        for (FieldInfoVo fieldInfoVo : fields) {
            String name = fieldInfoVo.getKey();
            List<String> fieldInfoVoValue = fieldInfoVo.getValue();
            if (fieldInfoVoValue.size() > 0) {
                for (String function : fieldInfoVoValue) {
                    if (!commonFunctionList.contains(function)) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter,
                                "fields.value不支持输入【key="+name +"】，可支持输入value包括【NotNull, NullCount, BlankCount, Max, Min, Zero, Avg, StddevPop, VarPop, True, False, Day, Month, Year, Quantile, Group】",
                                Message.MESSAGE_PARAM_ERROR_SOLUTION);

                    }
                }
            }
        }
    }
    public static String replaceStr(String input) {
        // 编译正则表达式模式，(?s)允许dot匹配任何字符，包括换行符
        String regex = "(?s)union\\s+all\\s+with\\s+t1\\s+as\\s+\\(.*?\\)";
        Pattern pattern = Pattern.compile(regex);

        // 执行替换
        Matcher matcher = pattern.matcher(input);
        String result = matcher.replaceAll(" union all ");

        return result;
    }

    private String checkLimit(String limit,String limitGroup){
        if(StringUtils.isNotEmpty(limit)){
            int limitCount=Integer.parseInt(limit);
            if(limitCount<0){
                String msg = "{\"code\":1,\"message\":\"设置limit必须大于0，请检查\"}";
                return msg;
            }
        }
        if(StringUtils.isNotEmpty(limitGroup)){
            int limitCount=Integer.parseInt(limitGroup);
            if(limitCount<0){
                String msg = "{\"code\":1,\"message\":\"设置groupLimit必须大于0，请检查\"}";
                return msg;
            }
        }
        return null;
    }

    public String validationCatalog(String catalog, String user) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA + catalog;
        String result = getTableRequest(urlOpen, user);
        return result;
    }
    public String getTableRequest(String urlOpen, String user) {
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGetOlk(urlOpen, user);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http getTableRequest 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getTableRequest 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }
    public JsonNode getTableJsonNode(String urlOpen, String user){
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGet(urlOpen, user);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }

        if(result==null){
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,"计算引擎异常",Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http getTableJsonNode 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getTableJsonNode 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));

        ObjectMapper mapper = new ObjectMapper();
        JsonNode rootNode;
        try {
            rootNode = mapper.readTree(result.getResult());
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        stopWatch.stop();
        return rootNode;
    }
    public Map<String, List<Object>> parseCatalog(String result) throws JsonProcessingException {
        Map<String, Object> catalogName = new ObjectMapper().readValue(result, Map.class);
        Map<String, List<Object>> map = new ConcurrentHashMap<>();
        List<Object> schemasList = (List<Object>) catalogName.get("schemas");
        map.put(String.valueOf(catalogName.get("catalogName")), schemasList);
        return map;
    }
    private String generateSqlFragments(FieldInfoVo fieldParam, QueryParam queryParam) {
        StringBuilder sqlBuilder = new StringBuilder();

        List<String> sqlFragments = new ArrayList<>();
        List<String> sqlUnionFragments = new ArrayList<>();
        List<String> sqlGroupFragments = new ArrayList<>();

        String catalog = queryParam.getCatalogName();
        String schema = queryParam.getSchema();
        String table = queryParam.getTable();
        String limit = queryParam.getLimit();
        String groupLimit = queryParam.getGroupLimit();
        String alias = "\"" + catalog + "\"" + "." + "\"" + schema + "\"" + "." + "\"" + table + "\"";

        for (String func : fieldParam.getValue()) {
            String columnName = fieldParam.getKey();
            String type = fieldParam.getType();
            String sqlFragment = "";
            String sqlUnionFragment = "";
            String sqlGroupFragment = "";

            switch (func.trim()) {
                case "Max":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.MAX, columnName);
                    break;
                case "Min":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.MIN, columnName);
                    break;
                case "NullCount":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.NULL_COUNT, columnName);
                    break;
                case "BlankCount":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.BLANK_COUNT, columnName, columnName, columnName);
                    break;
                case "CountTable":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.COUNT_TABLE, columnName, columnName);
                    break;
                case "Zero":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.ZERO, columnName);
                    break;
                case "Avg":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.AVG, columnName);
                    break;
                case "StddevPop":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.STD, columnName);
                    break;
                case "VarPop":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.VAR, columnName);
                    break;
                case "True":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.TRUE, columnName);
                    break;
                case "False":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.FALSE, columnName);
                    break;
                case "Quantile":
                    sqlFragment = replaceTemplate(SqlTemplateConstant.APPROX, columnName, columnName, columnName, columnName, columnName, columnName);
                    break;
                case "Day":
                    if (StringUtils.isEmpty(limit)) {
                        sqlGroupFragment=replaceTemplate(SqlTemplateConstant.DAY_LIMIT,columnName,columnName,alias,columnName);
                    } else {
                        sqlGroupFragment = replaceTemplate(SqlTemplateConstant.DAY_LIMIT_SQL, columnName,columnName,alias,columnName,queryParam.getLimit());
                    }

                    break;
                case "Month":
                    if (StringUtils.isEmpty(limit)) {
                        sqlGroupFragment=replaceTemplate(SqlTemplateConstant.MONTH_LIMIT, columnName,columnName,alias,columnName);
                    } else {
                        sqlGroupFragment = replaceTemplate(SqlTemplateConstant.MONTH_LIMIT_SQL, columnName,columnName,alias,columnName,queryParam.getLimit());
                    }
                    break;
                case "Year":
                    if (StringUtils.isEmpty(limit)) {
                        sqlGroupFragment=replaceTemplate(SqlTemplateConstant.YEAR_LIMIT, columnName,columnName,alias,columnName);
                    } else {
                        sqlGroupFragment = replaceTemplate(SqlTemplateConstant.YEAR_LIMIT_SQL, columnName,columnName,alias,columnName,queryParam.getLimit());
                    }
                    break;
                case "Group":
                    if (StringUtils.isEmpty(groupLimit)) {

                        if(StringUtils.endsWith("datetime",type) || StringUtils.endsWith("date",type)|| StringUtils.endsWith("timestamp",type)||
                                StringUtils.endsWith("int",type)||StringUtils.endsWith("bigint",type)){
                            sqlGroupFragment = replaceTemplate(SqlTemplateConstant.GROUP_SINGLE, columnName,columnName,alias,columnName);
                            break;
                        }
                        sqlGroupFragment=replaceTemplate(SqlTemplateConstant.GROUP_SINGLE, columnName,columnName,alias,columnName);
                    } else {
                        if(StringUtils.endsWith("datetime",type) || StringUtils.endsWith("date",type)|| StringUtils.endsWith("timestamp",type)||
                                StringUtils.endsWith("int",type)||StringUtils.endsWith("bigint",type)){
                            sqlGroupFragment = replaceTemplate(SqlTemplateConstant.GROUP_Template, columnName,columnName,alias,columnName,queryParam.getGroupLimit());
                            break;
                        }
                        sqlGroupFragment = replaceTemplate(SqlTemplateConstant.GROUP_Template, columnName,columnName,alias,columnName,queryParam.getGroupLimit());
                    }
                    break;
                default:
                    // Handle other cases if needed
                    break;
            }

            if (!sqlFragment.isEmpty()) {
                sqlFragments.add(sqlFragment);
            }
            if (!sqlUnionFragment.isEmpty()) {
                sqlUnionFragments.add(sqlUnionFragment);
            }
            if (!sqlGroupFragment.isEmpty()) {
                sqlGroupFragments.add(sqlGroupFragment);
            }
        }

        // 将所有 SQL 片段连接成一个字符串
        for (String fragment : sqlFragments) {
            sqlBuilder.append(fragment).append(" & ");
        }
        for (String fragment : sqlUnionFragments) {
            sqlBuilder.append(fragment).append(" union all ");
        }
        for (String fragment : sqlGroupFragments) {
            sqlBuilder.append(fragment).append(" & ");
        }
        return sqlBuilder.toString();
    }

    private static String buildNotNullSql(QueryParam queryParam) {
        String tableName = "\"" + queryParam.getCatalogName() +"\"" + "." + "\""+ queryParam.getSchema() +"\"" + "." + "\""+ queryParam.getTable() +"\"";
        StringBuilder wherePart = new StringBuilder();
        boolean hasWhereConditions = false;
        int fieldCount = queryParam.getFields().size();

        List<String> selectPartList = new ArrayList<>(fieldCount);

        for (FieldInfoVo fieldInfoVo : queryParam.getFields()) {
            String key = fieldInfoVo.getKey();
            List<String> value = fieldInfoVo.getValue();
            for (String rule : value) {
            String type = fieldInfoVo.getType();
            if (!rule.equalsIgnoreCase("NotNull")) {
                continue;
            }

            if ("NotNull".equalsIgnoreCase(rule)) {
                selectPartList.add("max(\"" + key + "\") AS " + "\"" +key + "\"");

                if (hasWhereConditions) {
                    wherePart.append(" OR ");
                } else {
                    hasWhereConditions = true;
                    wherePart.append(SqlTemplateConstant.WHERE_TEMPLATE);
                }

                // 根据字段类型判断是否添加非空字符串检查
                if ("VARCHAR".equalsIgnoreCase(type)) {
                    wherePart.append(String.format(SqlTemplateConstant.NOTNULL_AND_NONEMPTY_CONDITION, key, key));
                } else {
                    wherePart.append(String.format(SqlTemplateConstant.NOTNULL_CONDITION, key));
                }
            }
          }
        }

        String selectPartStr = String.join(",", selectPartList);

        if (hasWhereConditions) {
            return String.format(SqlTemplateConstant.SELECT_TEMPLATE + " " + wherePart, selectPartStr, tableName).replace("${table}", tableName) + SqlTemplateConstant.LIMIT_CONDITION;
        } else {
            if (StringUtils.isNotEmpty(selectPartStr)) {
                return String.format(SqlTemplateConstant.SELECT_TEMPLATE, selectPartStr).replace("${table}", tableName) + SqlTemplateConstant.LIMIT_CONDITION;
            } else {
                return null;
            }
        }
    }


    public static String generateRule() {
        String dateTimePart = LocalDateTime.now().format(DATE_FORMATTER);
        String randomPart = String.format("%05d", RANDOM.nextInt(100000));
        String letterPart = generateRandomLetters(5);

        return dateTimePart + "_" + randomPart + "_" + letterPart;
    }

    private static String generateRandomLetters(int length) {
        StringBuilder letterPart = new StringBuilder();
        for (int i = 0; i < length; i++) {
            int index = RANDOM.nextInt(LETTERS.length());
            letterPart.append(LETTERS.charAt(index));
        }
        return letterPart.toString();
    }
    // 替换模板中的占位符的方法
    private static String replaceTemplate(String template, String columnName) {
        return String.format(template, columnName, columnName);
    }

    private static String replaceTemplate(String template, String... replacements) {
        return String.format(template, (Object[]) replacements);
    }

    @Override
    public ResponseEntity<?> statementTask(String statement, String user, String type) {
        checkParameter(statement, user, type);
        return Task(statement,user,type);
    }

    public ResponseEntity<?> checkParameter(String statement, String user, String type){
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        if(StringUtils.isNull(type)){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.TYPE_NOT_NULL,Message.MESSAGE_INPUT_NOT_EMPTY);
        }
        if (!type.equals("0") && !type.equals("1")) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TYPE_ENUM_MESSAGE, Message.MESSAGE_PARAM_ERROR_SOLUTION);
        }
        if(type.equals("1") && statement.contains("statement")){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.BODY_MESSAGE, Message.MESSAGE_PARAM_ERROR_SOLUTION);
        }

       /* if(statement.replaceAll("\\s+", "").equals("{}")){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.REQUEST_NOT_NULL, Message.MESSAGE_INPUT_NOT_EMPTY);
        }*/
        ObjectMapper objectMapper = new ObjectMapper();
        AsyncQuery jsonData;
        if(StringUtils.endsWith(type,"0")){
            try {
                jsonData = objectMapper.readValue(statement, AsyncQuery.class);
            } catch (JsonProcessingException e) {
                log.error(e.getMessage());
                throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.BODY_MESSAGE,Message.MESSAGE_PARAM_ERROR_SOLUTION);
            }
            // 判断statement是否为空
            if (jsonData.getStatement().size() == 0 || jsonData.getStatement().get(0).length() == 0) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.STATEMENT_NOT_NULL,Message.MESSAGE_INPUT_NOT_EMPTY);
            }
            if (jsonData.getTopic().matches("\\d+|.*[\u4e00-\u9fff]+.*")) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TOPIC_ERROR, Message.MESSAGE_PARAM_ERROR_SOLUTION);
            }

            if (!(jsonData.getTopic() instanceof String)) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TOPIC_ERROR, Message.MESSAGE_PARAM_ERROR_SOLUTION);
            }
        }else{
            if(!statement.contains("fields")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.FIELD_NOT_NULL,Message.MESSAGE_INPUT_NOT_EMPTY);
            }
            if(!statement.contains("catalogName")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_NOT_NULL, Message.MESSAGE_INPUT_NOT_EMPTY);
            }
            if(!statement.contains("schema")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.SCHEMA_NOT_NULL, Message.MESSAGE_INPUT_NOT_EMPTY);
            }
            if(!statement.contains("table")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NOT_NULL, Message.MESSAGE_INPUT_NOT_EMPTY);
            }
        }
        return ResponseEntity.ok().build();
    }

    public ResponseEntity<?> check(String statement,String user){
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }

        if(StringUtils.isNull(statement)){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.STATEMENT_MISSING);
        }
        long startTime = System.currentTimeMillis();

        try {
             new SqlParser().createStatement(statement, new ParsingOptions(ParsingOptions.DecimalLiteralTreatment.AS_DOUBLE));
        } catch (Exception e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.SqlSyntaxError, statement, e.getMessage(), Message.MESSAGE_QUERY_SQL_SOLUTION);
        }
        long endTime = System.currentTimeMillis();
        long executionTime = endTime - startTime; // 计算执行时间

        log.debug("函数执行时间: " + executionTime + " 毫秒");

        return ResponseEntity.ok().build();
    }

    public ResponseEntity<?> cancelAllTask(String taskIdList, String user){
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }

        if(StringUtils.isNull(taskIdList)){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TASKID_MISSING);
        }

        ObjectMapper mapper = new ObjectMapper();
        try {
             mapper.readValue(taskIdList, ParamVo.class);
        } catch (JsonProcessingException e) {
            log.error("Error parsing JSON: " + e.getMessage());
            if(e.getMessage().contains("task_id cannot be empty")||e.getMessage().contains("Cannot coerce empty String")){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.TASKID_MISSING,Message.MESSAGE_INPUT_NOT_EMPTY);
            } else if (e.getMessage().contains("Cannot construct instance of `java.util.ArrayList`") ||e.getMessage().contains("Cannot deserialize value of type")) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.TASKID_FORMAT_ERROR,Message.MESSAGE_PARAM_ERROR_SOLUTION);
            }

        }
        /*if ((Pattern.compile("\"task_id\"\\s*:\\s*\\{[^{}]*\\}").matcher(taskIdList).find())) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.TASKID_MISSING,Message.MESSAGE_INPUT_NOT_EMPTY);
        }*/
        List<String> list=convertJsonToList(taskIdList);
        //update db state cancel
        try {
            String updateTime = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
            List<TaskEntityQuery> taskList=taskMapper.batchTaskIds1(list);
            if (taskList != null && taskList.size()>0) {
                Optional<TaskEntityQuery> deletedTask = taskList.stream()
                        .filter(task -> "CANCEL".equals(task.getState()))
                        .findFirst();
                if(deletedTask.isPresent()){
                    throw new AiShuException(ErrorCodeEnum.TaskCanceled,Detail.CANCELED_ERROR,Message.MESSAGE_PARAM_ERROR_SOLUTION);
                }

                /*Optional<TaskEntityQuery> runningOrPlannedTask = taskList.stream()
                        .filter(task -> "FINISHED".equals(task.getState())|| "FAILED".equals(task.getState()))
                        .findFirst();

                if (runningOrPlannedTask.isPresent()) {
                    throw new AiShuException(ErrorCodeEnum.TaskCompleted,Detail.CANCELED_FINISHED_ERROR,Message.MESSAGE_PARAM_ERROR_SOLUTION);
                }*/
                Optional<TaskEntityQuery> deletedStateTask = taskList.stream()
                        .filter(task -> "DELETED".equals(task.getState()))
                        .findFirst();

                if (deletedStateTask.isPresent()) {
                    throw new AiShuException(ErrorCodeEnum.TaskDeleted,Detail.TASK_DELETED,Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
                }
            }else{
                throw new AiShuException(ErrorCodeEnum.TaskNotExist,Detail.TASKID_MISSING,Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
            }
            // 使用 Stream 实现并异步处理
            List<CompletableFuture<Void>> futures = list.stream()
                    .map(taskId -> CompletableFuture.runAsync(() -> {
                        try {
                            deleteTaskReq(taskId);
                            log.debug("任务 " + taskId + " 已删除");
                        } catch (Exception ex) {
                            log.error("任务 " + taskId + " 删除时发生异常: " + ex.getMessage(), ex);
                        }
                    }))
                    .collect(Collectors.toList());

            // 等待所有任务完成
            CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).join();

            int updatedCount = taskMapper.updateTaskStateBatch(list,"CANCEL",updateTime);
            if(updatedCount>0){
                taskMapper.updateQueryStateBatch(list, "CANCEL", updateTime);
            }
            log.debug("更新了 " + updatedCount + " 条记录");
        } catch (Exception e) {
            // 打印异常信息
            log.error("更新任务状态时发生异常: " + e.getMessage());
            throw new AiShuException(ErrorCodeEnum.DBError,e.getMessage(),Message.MESSAGE_DATABASE_ERROR_SOLUTION);
        }



        return ResponseEntity.ok(new ResponseMessage("取消成功"));
    }

    public static List<String> convertJsonToList(String json) {
        JSONObject jsonObject = JSONUtil.parseObj(json);
        JSONArray taskIds = jsonObject.getJSONArray("task_id");
        return taskIds.toList(String.class);
    }
    @Override
    public ResponseEntity<?> getTask(String taskId, String user) {
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.X_PRESTO_USER_MISSING);
        }
        String strResponse;
        try {
            strResponse = this.getTaskReq(taskId);
        } catch (Exception e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.JSON_ANALYZE_ERROR, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        if(StringUtils.isEmpty(strResponse)){
            throw new AiShuException(ErrorCodeEnum.TaskNotExist,"",Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        return ResponseEntity.ok(strResponse);
    }

    @Override
    public ResponseEntity<?> deleteTask(String taskId, String user) {
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.X_PRESTO_USER_MISSING);
        }
        List<TaskEntity> taskEntityList=taskMapper.selectJoinOne(taskId);
        if(taskEntityList==null ||taskEntityList.size()==0){
            throw new AiShuException(ErrorCodeEnum.TaskNotExist, "", Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }else{
            TaskEntity taskEntity=taskEntityList.get(0);
            if(taskEntity!=null && (taskEntity.getState().equals("RUNNING") || taskEntity.getState().equals("PLANNED"))){
                throw new AiShuException(ErrorCodeEnum.RunningError,Detail.TASK_RUNNING_ERROR,Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
            }
            if(taskEntity!=null && (taskEntity.getState().equals("CANCEL") || taskEntity.getState().equals("FINISHED")
                    || taskEntity.getState().equals("FAILED"))){
                taskMapper.deleteTaskById(taskId);
                taskMapper.deleteQueryById(taskId);
            }
        }
        return ResponseEntity.ok(new ResponseMessage("删除成功"));
    }

    @Override
    public ResponseEntity<?> deleteOrigTask(String taskId, String user) {
         if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.X_PRESTO_USER_MISSING);
        }
        String strResponse = this.getTaskOrigReq(taskId);
        if (StringUtils.isEmpty(strResponse)) {
            throw new AiShuException(ErrorCodeEnum.TaskNotExist,"","请检查任务id是否正确");
        }
        ObjectMapper mapper = new ObjectMapper();
        try {
            List<TaskInfo> tasks = mapper.readValue(strResponse, new TypeReference<List<TaskInfo>>() {});
            for (TaskInfo task : tasks) {
                String subtaskId=task.getSubTaskId();
                deleteTaskReq(subtaskId);
            }
        } catch (Exception e) {
            e.printStackTrace();

        }
        return ResponseEntity.ok(200);
    }


    public String getTaskOrigReq(String taskId) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_QUERY_TASK + taskId;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGet(urlOpen, Constants.UTF8);
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();
        log.info("Http get 请求时长为：{}ms", stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }
    public CompletableFuture<String> getCreateModelAsync(String statement,String user) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_DATA_EXPLORATION;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        AtomicReference<HttpResInfo> result = null;
        com.alibaba.fastjson2.JSONObject clientInfo = new com.alibaba.fastjson2.JSONObject();
        clientInfo.put(CatalogConstant.CANCEL_MAX_RUN, true);
        clientInfo.put(CatalogConstant.CANCEL_MAX_EXECUTION, true);
        CompletableFuture<String> futureResult = CompletableFuture.supplyAsync(() -> {
            try {
                 result.set(HttpOpenUtils.sendPost(urlOpen, statement, user, clientInfo, null));
                return result.get().getResult();
            } catch (Exception e) {
                throw new CompletionException(new AiShuException(ErrorCodeEnum.OpenLooKengError, e.getMessage(), Message.MESSAGE_OPENLOOKENG_ERROR));
            }
        });

        if(futureResult==null){
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,"",Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        // 在异步操作完成后执行的动作，这里可以是记录日志或其他操作
        futureResult.thenAccept(msg -> {
            stopWatch.stop();
            log.info("Http get 请求时长为：{}ms", stopWatch.getTotal(TimeUnit.MILLISECONDS));
        });

        return futureResult;
    }


    public String getStatement(String statement, String user, String type) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_STATEMENT_TASK;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendPostWithParams(urlOpen, type, statement, user);
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus() >= HttpStatus.BAD_REQUEST) {
            log.error("Http getStatement 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            if (result.getResult().contains("Error message: Hetu metastore operation failed")) {
                throw new AiShuException(ErrorCodeEnum.DBError, result.getResult(), Message.MESSAGE_DATABASE_ERROR_SOLUTION);
            } else {
                throw new AiShuException(ErrorCodeEnum.OpenLooKengError, result.getResult(), Message.MESSAGE_OPENLOOKENG_ERROR);
            }

        }
        log.info("Http getStatement 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

    public String getTaskReq(String taskId) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_QUERY + taskId;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGet(urlOpen, Constants.DEFAULT_AD_HOC_USER);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http getTaskReq 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getTaskReq 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

    public String deleteTaskReq(String taskId) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_QUERY + taskId;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendDelete(urlOpen, Constants.DEFAULT_AD_HOC_USER);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http DeleteTaskReq 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http DeleteTaskReq 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }
    public String toJsonTaskId(String taskId) {
        JsonObject jsonObject = new JsonObject();
        List<String> stringList = new ArrayList<>();
        stringList.add(taskId);
        JsonArray jsonArray = new JsonArray();
        for (String item : stringList) {
            jsonArray.add(item);
        }
        jsonObject.add("task_id", jsonArray);
        return jsonObject.toString();
    }
    public String printResponseJson(String taskId) throws Exception {
        ObjectMapper objectMapper = new ObjectMapper();
        List<String> stringList;
        try {
            stringList = objectMapper.readValue(taskId, new TypeReference<List<String>>(){});
            Map<String, List<String>> jsonMap = new HashMap<>();
            jsonMap.put("task_id", stringList);
            return objectMapper.writeValueAsString(jsonMap);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.JSON_ANALYZE_ERROR, Message.MESSAGE_PARAM_ERROR_SOLUTION);
        }
    }

    public void sqlValidate(boolean enabled, List<String> stringList) {
        if (enabled) {
            for (String sql : stringList) {
                try {
                    new SqlParser().createStatement(sql, new ParsingOptions(ParsingOptions.DecimalLiteralTreatment.AS_DOUBLE));
                } catch (Exception e) {
                    log.error(e.getMessage());
                    throw new AiShuException(ErrorCodeEnum.SqlSyntaxError, sql, e.getMessage(), Message.MESSAGE_QUERY_SQL_SOLUTION);
                }
            }
        }

    }
}
