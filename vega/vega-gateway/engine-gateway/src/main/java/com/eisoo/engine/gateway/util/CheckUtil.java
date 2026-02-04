package com.eisoo.engine.gateway.util;

import cn.hutool.core.date.StopWatch;
import cn.hutool.core.util.ArrayUtil;
import cn.hutool.json.JSONUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.common.QueryConstant;
import com.eisoo.engine.gateway.domain.dto.CatalogRuleDto;
import com.eisoo.engine.gateway.domain.dto.ExcelTableConfigDto;
import com.eisoo.engine.gateway.domain.dto.RuleDto;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.service.CatalogService;
import com.eisoo.engine.metadata.entity.CatalogRuleEntity;
import com.eisoo.engine.metadata.mapper.CatalogRuleMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.StringUtils;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.regex.Pattern;
import java.util.stream.Collectors;


/**
 * @Author zdh
 **/
public class CheckUtil {
    private static final Logger log = LoggerFactory.getLogger(CheckUtil.class);

    public static void catalogCheck(String catalogName, String connectorName) {
        String[] connectors = {"vdm"};

        if (!ArrayUtil.contains(connectors, connectorName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_TYPE_UNSUPPORTED, String.format(Message.MESSAGE_CATALOG_TYPES_SOLUTION, Arrays.toString(connectors)));
        }
        if (!checkcatalogName(catalogName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_NAME_ERROR, Message.MESSAGE_CATALOG_NAME_SOLUTION);
        }
    }

    public static String checkCatalog(CatalogService catalogService, String catalog) {
        String catalogStr = catalogService.showCatalogInfo(catalog);
        if (catalogStr!= null && !catalogStr.startsWith("{")) {
            log.error("cactalog不存在:{}", catalog);
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalog);
        }
        return catalogStr;
    }

    public static void checkVdmCatalog(CatalogService catalogService, String vdmCatalog) {
        if (org.apache.commons.lang3.StringUtils.isEmpty(vdmCatalog)) {
            log.error("vdm_catalog不能为空");
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.VDM_CATALOG_NOT_NULL);
        } else {
            String catalogStr = catalogService.showCatalogInfo(vdmCatalog);
            if (!catalogStr.startsWith("{")) {
                log.error("vdm_catalog不存在:{}", vdmCatalog);
                throw new AiShuException(ErrorCodeEnum.CatalogNotExist, vdmCatalog, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
            } else {
                com.alibaba.fastjson2.JSONObject catalogJson = com.alibaba.fastjson2.JSONObject.parseObject(catalogStr);
                String connectorName = catalogJson.getString("connector.name");
                if (!connectorName.equals(CatalogConstant.VDM_CATALOG)) {
                    log.error("vdm_catalog错误:{}:{}", vdmCatalog, connectorName);
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_TYPE_INCONSISTENT);
                }
            }
        }
    }

    public static void catalogCheck(CatalogRuleDto catalogRuleDto, CatalogRuleMapper catalogRuleMapper){
        String[] rules = {"FilterNode","ProjectNode","AggregationNode","TopNNode","LimitNode","JoinNode","GroupIdNode","MarkDistinctNode","UnionNode"};
        String[] bool = {"true","false"};
        for (String name : catalogRuleDto.getRule().stream().map(rule -> rule.getName()).collect(Collectors.toList())) {
            if (!ArrayUtil.contains(rules, name)) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.ENUM_UNSUPPORTED,  String.format(Message.MESSAGE_ENUMS_SOLUTION, Arrays.toString(rules)));
            }
        }
        for (String name : catalogRuleDto.getRule().stream().map(rule -> rule.getIsEnable()).collect(Collectors.toList())) {
            if (!ArrayUtil.contains(bool, name)) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.ENUM_UNSUPPORTED, String.format(Message.MESSAGE_ENUMS_SOLUTION, Arrays.toString(bool)));
            }
        }
        List<CatalogRuleEntity> catalogRuleEntityList = catalogRuleMapper.selectRuleInfo(catalogRuleDto.getCatalogName());
        if (catalogRuleDto.getRule().stream().filter(x -> x.getName().equals("ProjectNode")).map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("false")) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.PROJECT_NODE_RULE_NOT_CANCELED);
        }

        if (catalogRuleEntityList != null && catalogRuleEntityList.size() > 0) {
            for (CatalogRuleEntity catalogRuleEntity : catalogRuleEntityList) {
                if ((catalogRuleEntity.getPushdownRule().equals("FilterNode") || catalogRuleEntity.getPushdownRule().equals("ProjectNode")) &&
                        catalogRuleEntity.getIsEnabled().equals("false")) {
                    if ((catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("AggregationNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("TopNNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("LimitNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("JoinNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("GroupIdNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("JoinNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("UnionNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("MarkDistinctNode")) &&
                            catalogRuleDto.getRule().stream().map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("true")) {
                        List<RuleDto> ruleDtoList1 = catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("FilterNode")).collect(Collectors.toList());
                        List<RuleDto> ruleDtoList2 = catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("ProjectNode")).collect(Collectors.toList());
                        if (ruleDtoList1.size() > 0 && ruleDtoList2.size() > 0) {
                            String enabled1 = ruleDtoList1.get(0).getIsEnable();
                            String enabled2 = ruleDtoList2.get(0).getIsEnable();
                            if (!StringUtils.equalsIgnoreCase(enabled1, "true") && !StringUtils.equalsIgnoreCase(enabled2, "true")) {
                                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.MISSING_DEPENDENCY_RULES_FILTER_OR_PROJECT);
                            }
                        }
                    }
                } else {
                    if (catalogRuleDto.getRule().stream().filter(x -> x.getName().equals("FilterNode")).map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("false")
                            && (catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("AggregationNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("TopNNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("LimitNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("JoinNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("GroupIdNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("TopNNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("MarkDistinctNode")) &&
                            catalogRuleDto.getRule().stream().filter(entity -> !entity.getName().equals("ProjectNode")).map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("true")) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.MISSING_DEPENDENCY_RULES_FILTER_OR_PROJECT);
                    }
                }
                if (catalogRuleEntity.getPushdownRule().equals("AggregationNode") && catalogRuleEntity.getIsEnabled().equals("false")) {
                    if ((catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("GroupIdNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("MarkDistinctNode")) &&
                            catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("GroupIdNode") || entity.getName().equals("MarkDistinctNode"))
                                    .map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("true") &&
                            catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("AggregationNode"))
                                    .map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("false")) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.MISSING_DEPENDENCY_RULES_AGGREGATION);

                    }
                } else {
                    // when aggregation is ture
                    if ((catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("GroupIdNode") ||
                            catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("MarkDistinctNode")) &&
                            catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("GroupIdNode") || entity.getName().equals("MarkDistinctNode"))
                                    .map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("true") &&
                            catalogRuleDto.getRule().stream().filter(entity -> entity.getName().equals("AggregationNode")).map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("false")) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.MISSING_DEPENDENCY_RULES_AGGREGATION);

                    }
                }
            }
        } else {
            if ((catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("TopNNode")
                    || catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("GroupIdNode")
                    || catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("MarkDistinctNode"))
                    && catalogRuleDto.getRule().stream().map(x -> x.getIsEnable()).collect(Collectors.toList()).contains("true")
                    && !catalogRuleDto.getRule().stream().map(x -> x.getName()).collect(Collectors.toList()).contains("AggregationNode")) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.MISSING_DEPENDENCY_RULES_AGGREGATION);
            }
        }
    }

    public static boolean checkcatalogName(String catalogName) {


        if (catalogName.length() > 100) {
            return false;
        }
        if (Pattern.matches("[0-9].*", catalogName)) {
            return false;
        }

        if (!Pattern.matches("^[0-9a-z_]{1,}$", catalogName)) {
            return false;
        }
        return true;
    }

    public static boolean checkViewName(String ViewName) {
        if (ViewName.length() > 100) {
            return false;
        }
        if (!Pattern.matches("^[a-z0-9_\\u4e00-\\u9fff\\s]+$", ViewName)) {
            return false; // 如果字符串不符合规定的字符集，则返回false
        }
        return true; // 如果所有条件都通过，则返回true
    }

    public static boolean checkExcelColumnName(String columnName) {
        if (columnName.length() > 100) {
            return false;
        }
        if (!Pattern.matches("^(?!.*[\\\\/:\\*?\"<>|A-Z]).*$", columnName)) {
            return false; // 如果字符串不符合规定的字符集，则返回false \/:\*?"<>|[A-Z]
        }
        return true; // 如果所有条件都通过，则返回true
    }

    public static int checkDdlOrDmlDatabase(String url, String statement, String user) {
        int first = StringUtils.indexOf(statement, ".");
        int second = StringUtils.lastIndexOf(statement, ".");
        if (first < 0 || second < 0 || first == second) {
            return 0;
        }
        String catalog = StringUtils.substring(statement, 0, first);
        String database = StringUtils.substring(statement, first + 1, statement.length());
        String catalogEnd = StringUtils.substring(catalog, StringUtils.lastIndexOf(catalog, " ") + 1, catalog.length());
        String databaseEnd = StringUtils.substring(database, 0, StringUtils.indexOf(database, "."));
        log.info("创建表、写入数据、清空数据检验database和schema,catalogName:{},databaseName:{}", catalogEnd, databaseEnd);

        String catalogNameList = getCatalogNameList(url);
        ObjectMapper mapper = new ObjectMapper();
        List<String> result;
        try {
            result = mapper.readValue(catalogNameList, new TypeReference<List<String>>() {});
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        if (!result.contains(catalogEnd)) {
            return 1;
        }

        String res = validationCatalog(url,catalogEnd,user);
        List<Object> schemasList = (List<Object>) JSONUtil.parseObj(res).get("schemas");
        if (schemasList.contains(databaseEnd)) {
            return 0;
        }else{
            return 2;
        }
    }

    public static boolean checkDataType(String statement) {
        String sql = statement.replaceAll("\t", "").replaceAll("\n", "").replaceAll("\\s+", "");
        int index = StringUtils.indexOf(sql.toLowerCase(), "varchar(0)");
        if (index > 0) {
            return true;
        }
        return false;
    }

    public static boolean checkTableNameLen(String statement) {
        int first = StringUtils.indexOf(statement, "(");
        String ddl = StringUtils.substring(statement, 0, first);
        int second = StringUtils.lastIndexOf(ddl, ".");
        String tableName = StringUtils.substring(ddl.trim(), second+1, statement.length());
        if (tableName.length() > 128) {
            return false;
        }
        if (Pattern.matches("[0-9].*", tableName)) {
            return false;
        }

        if (!Pattern.matches("^[0-9a-zA-Z_]{1,}$", tableName)) {
            return false;
        }
        return true;
    }

    /**
     * 查询数据源名称列表GET请求
     *
     * @param url
     * @return
     */
    public static String getCatalogNameList(String url) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_CATALOG;
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
            log.error("Http getCatalogNameList 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getCatalogNameList 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));

        return result.getResult();
    }

    public static String validationCatalog(String url, String catalog, String user) {
        String urlOpen = url + QueryConstant.VIRTUAL_V1_SCHEMA + catalog;
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
            log.error("Http validationCatalog 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http validationCatalog 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

    public static void excelTableConfigDtoTrim(ExcelTableConfigDto excelTableConfigDto) {
        if (StringUtils.isNotEmpty(excelTableConfigDto.getCatalog())) {
            excelTableConfigDto.setCatalog(excelTableConfigDto.getCatalog().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getVdmCatalog())) {
            excelTableConfigDto.setVdmCatalog(excelTableConfigDto.getVdmCatalog().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getTableName())) {
            excelTableConfigDto.setTableName(excelTableConfigDto.getTableName().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getFileName())) {
            excelTableConfigDto.setFileName(excelTableConfigDto.getFileName().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getSheet())) {
            excelTableConfigDto.setSheet(excelTableConfigDto.getSheet().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getStartCell())) {
            excelTableConfigDto.setStartCell(excelTableConfigDto.getStartCell().trim());
        }
        if (StringUtils.isNotEmpty(excelTableConfigDto.getEndCell())) {
            excelTableConfigDto.setEndCell(excelTableConfigDto.getEndCell().trim());
        }
        if (excelTableConfigDto.getColumns() != null) {
            for (ExcelTableConfigDto.ColumnType columnType : excelTableConfigDto.getColumns()) {
                if (StringUtils.isNotEmpty(columnType.getColumn())) {
                    columnType.setColumn(columnType.getColumn().trim());
                }
                if (StringUtils.isNotEmpty(columnType.getType())) {
                    columnType.setType(columnType.getType().trim());
                }
            }
        }
    }
}
