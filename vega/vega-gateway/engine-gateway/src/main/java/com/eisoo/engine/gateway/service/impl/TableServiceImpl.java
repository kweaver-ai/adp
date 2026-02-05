package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.common.QueryConstant;
import com.eisoo.engine.gateway.connector.ConnectorConfig;
import com.eisoo.engine.gateway.connector.ConnectorConfigCache;
import com.eisoo.engine.gateway.domain.vo.CatalogSchemas;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.domain.vo.SchemaColumns;
import com.eisoo.engine.gateway.domain.vo.SchemaTables;
import com.eisoo.engine.gateway.service.TableService;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
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
import com.fasterxml.jackson.databind.SerializationFeature;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class TableServiceImpl implements TableService {
    private static final Logger log = LoggerFactory.getLogger(TableServiceImpl.class);

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    @Autowired
    ConnectorConfigCache connectorConfigCache;

    @Override
    public ResponseEntity<?> CatalogSchemaList(String catalog, String user) {
        if (StringUtils.isNull(user)) {
            // 用户丝能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        //if (StringUtils.isNotEmpty(CheckUtil.checkPath())) {
        // 挂载文件路径为空
        //return Result.error(new AiShuException(ErrorCodeEnum.deployError));
        //}
        //validation datasource 接坣
        Map<String, List<Object>> listMap = null;
        String resultCatalog = validationCatalog(catalog, user);
        try {
            listMap = parseCatalog(resultCatalog);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        if (listMap == null || !listMap.containsKey(catalog) || listMap.get(catalog).isEmpty()) {
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalog, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        String jsonStr;
        try {
            jsonStr = ConvertSchemaJson(resultCatalog);
        } catch (Exception e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.JSON_ANALYZE_ERROR, Message.MESSAGE_INTERNAL_ERROR);
        }
        return ResponseEntity.ok(jsonStr);
    }

    @Override
    public ResponseEntity<?> SchemaTableList(String catalog, String schema, String user) {
        String emptyStr = "{\"data\":[],\"total_count\":0}";
        //schema=schema.toLowerCase(Locale.ROOT);
        if (StringUtils.isNull(user)) {
            // 用户丝能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }

        //if (StringUtils.isNotEmpty(CheckUtil.checkPath())) {
        // 挂载文件路径为空
        //return Result.error(new AiShuException(ErrorCodeEnum.deployError));
        //}

        //请求获坖table list 接坣
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_TABLE + catalog + "/" + schema;
        String catalogStr = showCatalogInfo(catalog);
        //Schema 校验
        Map<String, List<Object>> listMap = null;
        String resultCatalog = validationCatalog(catalog, user);
        try {
            listMap = parseCatalog(resultCatalog);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        if (!listMap.get(catalog).contains(schema)) {
            throw new AiShuException(ErrorCodeEnum.SchemaNotExist, schema, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        urlOpen = urlOpen + "?originCatalog=" + schema;
        String connectorName = JSONUtil.parseObj(catalogStr).get(CatalogConstant.CONNECTOR_NAME).toString();
        if (StringUtils.isNotBlank(connectorName)) {
            urlOpen = urlOpen + "&connectorName=" + connectorName;
        }

        String result = getTableRequest(urlOpen, user);
        //返回为空
        if (StringUtils.endsWith(result, "[]\n")) {
            try {
                return ResponseEntity.ok(convertEmptyJson(emptyStr));
            } catch (JsonProcessingException e) {
                log.error(e.getMessage());
                throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
            }
        }
        String jsonStr;
        try {
            jsonStr = convertTablesJson(result);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        return ResponseEntity.ok(jsonStr);
    }

    @Override
    public ResponseEntity<?> CollectSchemaTableList(String collector, String catalog, String schema, String taskId, String user, String datasourceId, Long schemaId) {
//        String emptyStr = "{data:[], total_count:0}";
        if (StringUtils.isNull(user)) {
            // 用户丝能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        //请求获坖table list 接坣
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_TABLE + "%s/%s/%s?datasourceId=%s&schemaId=%s";
        urlOpen = String.format(urlOpen, collector, catalog, schema, datasourceId, schemaId);
        String catalogStr = showCatalogInfo(catalog);
        String connectorName = null;
        log.info("【CollectSchemaTableList】:catalogStr={}", catalogStr);
        if (catalogStr.contains(CatalogConstant.CONNECTOR_NAME)) {
            urlOpen = urlOpen + "&originCatalog=" + schema + "&taskId=" + taskId;
            connectorName = JSONUtil.parseObj(catalogStr).get(CatalogConstant.CONNECTOR_NAME).toString();
        }
        log.info("【CollectSchemaTableList】:connectorName={}", connectorName);
        ConnectorConfig connectorConfig = connectorConfigCache.getConnectorConfig(connectorName);
        if (connectorConfig == null) {
            log.error("【CollectSchemaTableList】ConnectorConfig is null !connectorName:{}", connectorName);
            throw new AiShuException(ErrorCodeEnum.MetadataCollectError,
                    "ConnectorConfig is null !connectorName=" + connectorName,
                    Message.MESSAGE_METADATA_COLLECTION_LOG_SOLUTION);
        }
        log.info("【CollectSchemaTableList】:url:{};connectorConfig:{};user{}", urlOpen, connectorConfig, user);
        String result = postCollectTable(urlOpen, JsonUtil.obj2json(connectorConfig), user);
        //返回为空
        if (!"success".equals(result)) {
            throw new AiShuException(ErrorCodeEnum.MetadataCollectError, result, Message.MESSAGE_METADATA_COLLECTION_LOG_SOLUTION);
        }
        return ResponseEntity.ok("success");
    }

    @Override
    public ResponseEntity<?> SchemaTableColumns(String catalog, String schema, String tables, String user) {
//        String emptyStr = "{data:[], total_count:0}";
        //schema=schema.toLowerCase(Locale.ROOT);
        if (StringUtils.isNull(user)) {
            // 用户丝能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }

        //if (StringUtils.isNotEmpty(CheckUtil.checkPath())) {
        // 挂载文件路径为空
        //return Result.error(new AiShuException(ErrorCodeEnum.deployError));
        //}

        String resultCatalog = validationCatalog(catalog, user);
        Map<String, List<Object>> listMap;
        try {
            listMap = parseCatalog(resultCatalog);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            return null;
        }
        log.info(listMap.toString());
        if (!listMap.containsKey(catalog)) {
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalog, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        if (!listMap.get(catalog).contains(schema)) {
            throw new AiShuException(ErrorCodeEnum.SchemaNotExist, schema, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_COLUMNS + catalog + "/" + schema + "/" + tables;

        String catalogStr = showCatalogInfo(catalog);
        if (catalogStr.contains(CatalogConstant.CONNECTOR_NAME)) {
            String connectorName = JSONUtil.parseObj(catalogStr).get(CatalogConstant.CONNECTOR_NAME).toString();
            urlOpen = urlOpen + "?originCatalog=" + schema + "&connectorName=" + connectorName;
        }

        String result = getTableRequest(urlOpen, user);

        if (StringUtils.endsWith(result, "[]\n")) {
            throw new AiShuException(ErrorCodeEnum.TableNotExist, result, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        String jsonStr;
        try {
            jsonStr = convertColumnsJson(result);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        System.out.println(jsonStr);
        return ResponseEntity.ok(jsonStr);
    }

    @Override
    public ResponseEntity<?> SchemaTableColumnsFast(String catalog, String schema, String table, String user) {
        //schema=schema.toLowerCase(Locale.ROOT);
        if (StringUtils.isNull(user)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        String resultCatalog = validationCatalog(catalog, user);
        Map<String, List<Object>> listMap;
        try {
            listMap = parseCatalog(resultCatalog);
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            return null;
        }
        log.info(listMap.toString());
        if (!listMap.containsKey(catalog)) {
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalog, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        if (!listMap.get(catalog).contains(schema)) {
            throw new AiShuException(ErrorCodeEnum.SchemaNotExist, schema, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_COLUMNS + "fast/" + catalog + "/" + schema + "/" + table;

        String catalogStr = showCatalogInfo(catalog);
        if (catalogStr.contains(CatalogConstant.CONNECTOR_NAME)) {
            urlOpen = urlOpen + "?originCatalog=" + schema;
        }

        String result = getTableRequest(urlOpen, user);

        if (StringUtils.endsWith(result, "[]\n")) {
            throw new AiShuException(ErrorCodeEnum.TableNotExist, result, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        return ResponseEntity.ok(result);
    }

    public String convertTablesJson(String jsonStr) throws JsonProcessingException {
        // 将json1解枝为List<Map<String, Object>>
        List<Map<String, Object>> jsonList = new ObjectMapper().readValue(jsonStr, new TypeReference<List<Map<String, Object>>>() {
        });

        // 创建Json2对象
        SchemaTables schemaTables = new SchemaTables();
        schemaTables.setTotal_count(jsonList.size());

        List<SchemaTables.Data> dataList = new CopyOnWriteArrayList<>();
        jsonList.stream().forEach(map -> {
            SchemaTables.Data data = new SchemaTables.Data();
            data.setCatalog((String) map.get("connectorId"));
            data.setSchema((String) map.get("schema"));
            data.setTable((String) map.get("table"));
            data.setFqn((String) map.get("fqn"));
            String comment = (String) map.get("comment");
            data.setComment(comment == null ? "" : comment);
            Object tableType = map.get("tableType");
            if (tableType != null) {
                data.setTableType((String) tableType);
            } else {
                data.setTableType(null);
            }
            data.setParams(map.getOrDefault("params", null));
            dataList.add(data);
        });
        schemaTables.setData(dataList);

        // 将Json2对象转杢为json2
        String json2String = new ObjectMapper().writerWithDefaultPrettyPrinter().writeValueAsString(schemaTables);
        return json2String;
    }

    public String convertColumnsJson(String jsonStr) throws JsonProcessingException {
        List<Map<String, Object>> jsonList = new ObjectMapper().readValue(jsonStr, new TypeReference<List<Map<String, Object>>>() {
        });
        SchemaColumns schemaColumns = new SchemaColumns();
        schemaColumns.setTotal_count(jsonList.size());

        List<SchemaColumns.Data> dataList = new CopyOnWriteArrayList<>();
        jsonList.stream().forEach(map -> {
            String origType = (String) map.get("origType");
            Pattern pattern = Pattern.compile("\\((\\d+)\\)");
            Matcher matcher = pattern.matcher(origType);
            boolean primaryKey = Boolean.parseBoolean(String.valueOf(map.get("primaryKey")));
            boolean nullAble = Boolean.parseBoolean(String.valueOf(map.get("nullAble")));
            String columnDef = (String) map.get("columnDef");
            SchemaColumns.Data data = new SchemaColumns.Data();
            String type = (String) map.get("type");
            if (StringUtils.equalsIgnoreCase("numeric_v1", type)) {
                type = "numeric";
            }
            data.setName((String) map.get("name"));
            data.setType(type);
            data.setOrigType(origType.toLowerCase(Locale.ROOT));
            data.setPrimaryKey(primaryKey);
            data.setNullAble(nullAble);
            data.setColumnDef(columnDef);
            data.setComment((String) map.get("comment"));
            Map<String, Object> typeSignature = new HashMap<>();
            Map<String, Object> olkTypeSignature = (Map<String, Object>) map.get("typeSignature");
            String rawType = String.valueOf(olkTypeSignature.get("rawType"));
            if (StringUtils.equalsIgnoreCase("row", rawType)) {
                Object typeArguments = olkTypeSignature.get("typeArguments");
                Object literalArguments = olkTypeSignature.get("literalArguments");
                Object arguments = olkTypeSignature.get("arguments");
                typeSignature.put("rawType", rawType);
                typeSignature.put("typeArguments", typeArguments);
                typeSignature.put("arguments", typeArguments);
                typeSignature.put("literalArguments", literalArguments);
                typeSignature.put("olkArguments", arguments);
                data.setTypeSignature(typeSignature);
            } else if (StringUtils.equalsIgnoreCase("map", rawType) || StringUtils.equalsIgnoreCase("array", rawType)) {
                Object typeArguments = olkTypeSignature.get("typeArguments");
                Object literalArguments = olkTypeSignature.get("literalArguments");
                Object arguments = olkTypeSignature.get("arguments");
                typeSignature.put("rawType", rawType);
                typeSignature.put("typeArguments", typeArguments);
                JSONArray jsonArray = new JSONArray(arguments);
                String jsonString = jsonArray.toString();
                List<Map<String, Object>> mapList = null;
                try {
                    mapList = getStr(jsonString);
                } catch (JsonProcessingException e) {
                    log.error(e.getMessage());
                    throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
                }
                typeSignature.put("arguments", mapList);
                typeSignature.put("literalArguments", literalArguments);
                typeSignature.put("olkArguments", arguments);
                data.setTypeSignature(typeSignature);
            } else if (StringUtils.equalsIgnoreCase("numeric_v1", rawType)) {
                typeSignature.put("rawType", "numeric");
                Object typeArguments = olkTypeSignature.get("typeArguments");
                Object literalArguments = olkTypeSignature.get("literalArguments");
                Object arguments = olkTypeSignature.get("arguments");
                typeSignature.put("typeArguments", typeArguments);
                typeSignature.put("arguments", arguments);
                typeSignature.put("literalArguments", literalArguments);
                data.setTypeSignature(typeSignature);
            } else if (StringUtils.containsAnyIgnoreCase("bigint", rawType) && matcher.find()) {
                typeSignature.put("rawType", "bigint");
                Object typeArguments = olkTypeSignature.get("typeArguments");
                Object literalArguments = olkTypeSignature.get("literalArguments");
                Map<String, Object> argMap = new HashMap<>();
                argMap.put("kind", "LONG_LITERAL");
                argMap.put("value", Integer.parseInt(matcher.group(1)));
                List<Map<String, Object>> list = new ArrayList<>();
                list.add(argMap);
                typeSignature.put("typeArguments", typeArguments);
                typeSignature.put("arguments", list);
                typeSignature.put("literalArguments", literalArguments);
                data.setTypeSignature(typeSignature);
            } else {
                typeSignature = olkTypeSignature;
                data.setTypeSignature(typeSignature);
            }
            dataList.add(data);
        });
        schemaColumns.setData(dataList);
        // 将Json2对象转杢为json2
        String json2String = new ObjectMapper().writerWithDefaultPrettyPrinter().writeValueAsString(schemaColumns);
        return json2String;
    }

    public List<Map<String, Object>> getStr(String jsonString) throws JsonProcessingException {
        List<String> jsonStrings = extractJsonStrings(jsonString);
        List<Map<String, Object>> myObjects = parseJsonStrings(jsonStrings);
        return myObjects;
    }

    private static List<String> extractJsonStrings(String jsonString) {
        List<String> jsonStrings = new ArrayList<>();

        int startIndex = jsonString.indexOf("{");
        int endIndex = jsonString.lastIndexOf("}");

        jsonString = jsonString.substring(startIndex, endIndex + 1);

        String[] parts = jsonString.split("},\\s*\\{");

        for (String part : parts) {
            if (!part.startsWith("{")) {
                part = "{" + part;
            }
            if (!part.endsWith("}")) {
                part = part + "}";
            }
            jsonStrings.add(part);
        }

        return jsonStrings;
    }

    private static List<Map<String, Object>> parseJsonStrings(List<String> jsonStrings) throws JsonProcessingException {
        ObjectMapper objectMapper = new ObjectMapper();
        List<Map<String, Object>> list = new ArrayList<>();

        for (String jsonString : jsonStrings) {
            String str = fixMissingBrace(jsonString);
            JsonNode jsonNode = objectMapper.readTree(str);

            // 提取 arguments 数组以及后面的值
            JsonNode argumentsNode = jsonNode.get("value").get("arguments");
            // 将数据放入 List<HashMap>

            for (JsonNode argumentNode : argumentsNode) {
                String kind = argumentNode.get("kind").asText();
                Object value = getValueFromArgumentNode(argumentNode);

                HashMap<String, Object> hashMap = new HashMap<>();
                hashMap.put("kind", kind);
                hashMap.put("value", value);

                list.add(hashMap);
            }
        }
        return list;
    }


    private static Object getValueFromArgumentNode(JsonNode argumentNode) {
        String kind = argumentNode.get("kind").asText();
        JsonNode valueNode = argumentNode.get("value");

        if (kind.equals("LONG_LITERAL")) {
            return valueNode.asLong();
        } else if (kind.equals("STRING_LITERAL")) {
            return valueNode.asText();
        }

        return null;
    }

    public static String fixMissingBrace(String originalString) {
        try {
            ObjectMapper objectMapper = new ObjectMapper();
            objectMapper.readTree(originalString);

            // 如果解析成功，说明字符串已经是有效的JSON格式，直接返回
            return originalString;
        } catch (Exception e) {
            // 如果解析失败，说明字符串缺少结束符号
            // 在字符串末尾添加 '}' 符号
            return originalString + "}";
        }
    }

    public String getTableRequest(String urlOpen, String user) {
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try {
            result = HttpOpenUtils.sendGetOlk(urlOpen, user);
        } catch (AiShuException e) {
            throw e;
        } catch (Exception e) {
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus() >= HttpStatus.BAD_REQUEST) {
            log.error("Http getTableRequest 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError, result.getResult(), Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getTableRequest 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }
//    public String getSchemaTable(String urlOpen, String user) {
//
//        StopWatch stopWatch = new StopWatch();
//        stopWatch.start();
//        CompletableFuture<String> future = HttpOpenUtils.sendGetAsync(urlOpen, user);
//        String result = future.join();
//        stopWatch.stop();
//        log.info("Http get 请求时长为：{}ms", stopWatch.getTotal(TimeUnit.MILLISECONDS));
//        return result;
//    }

    public String postCollectTable(String urlOpen, String json, String user) {
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try {
            result = HttpOpenUtils.sendPost(urlOpen, json, user);
        } catch (AiShuException e) {
            throw e;
        } catch (Exception e) {
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus() >= HttpStatus.BAD_REQUEST) {
            log.error("Http postCollectTable 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError, result.getResult(), Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http postCollectTable 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

    public String validationCatalog(String catalog, String user) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA + catalog;
        return getTableRequest(urlOpen, user);
    }

    public Map<String, List<Object>> parseCatalog(String result) throws JsonProcessingException {
        Map<String, Object> catalogName = new ObjectMapper().readValue(result, Map.class);
        Map<String, List<Object>> map = new ConcurrentHashMap<>();
        List<Object> schemasList = (List<Object>) catalogName.get("schemas");
        map.put(String.valueOf(catalogName.get("catalogName")), schemasList);
        return map;
    }

    /**
     * format emptyStr
     *
     * @param emptyStr
     * @return
     * @throws JsonProcessingException
     */
    public String convertEmptyJson(String emptyStr) throws JsonProcessingException {
        // 创建ObjectMapper对象
        ObjectMapper objectMapper = new ObjectMapper();

        // 将字符串转杢为JSON对象
        log.info("", emptyStr);
        Object json = objectMapper.readValue(emptyStr, Object.class);

        // 将JSON对象转杢为格弝化的JSON字符串
        String jsonString = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(json);
        return jsonString;
    }

//    public List<SchemaTables.Data> validationTable(String user, String catalog, String schema) throws JsonProcessingException {
//        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_SCHEMA_TABLE + catalog + "/" + schema;
//        String result = getTableRequest(urlOpen, user);
//        if (result==null){
//            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,urlOpen,Message.MESSAGE_OPENLOOKENG_ERROR);
//        }
//
//        // 将json1解枝为List<Map<String, Object>>
//        List<Map<String, Object>> json1List = new ObjectMapper().readValue(result, new TypeReference<List<Map<String, Object>>>() {});
//
//        // 将json1List中的毝个Map转杢为Json2.Data对象，并添加到Json2.data中
//        List<SchemaTables.Data> dataList = new ArrayList<>();
//        for (Map<String, Object> map : json1List) {
//            SchemaTables.Data data = new SchemaTables.Data(); // 创建新的对象实例
//            data.setTable((String) map.get("table"));
//            dataList.add(data);
//        }
//
//        return dataList;
//    }

    public String ConvertSchemaJson(String jsonStr) throws JsonProcessingException {
        ObjectMapper mapper = new ObjectMapper();
        mapper.enable(SerializationFeature.INDENT_OUTPUT);
        Object obj = mapper.readValue(jsonStr, CatalogSchemas.class).schemas;
        String formattedJson = mapper.writerWithDefaultPrettyPrinter().writeValueAsString(obj);
        return formattedJson;
    }


    /**
     * 查询数据源详情GET请求
     *
     * @param
     * @return
     */
    public String showCatalogInfo(String catalogName) {
        String urlOpen = openlookengUrl + CatalogConstant.VIRTUAL_V1_SHOW_CATALOG + "/" + catalogName;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try {
            result = HttpOpenUtils.sendGet(urlOpen, Constants.DEFAULT_AD_HOC_USER);
        } catch (AiShuException e) {
            throw e;
        } catch (Exception e) {
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus() >= HttpStatus.BAD_REQUEST) {
            log.error("Http showCatalogInfo 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError, result.getResult(), Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http showCatalogInfo 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));

        return result.getResult();
    }

}
