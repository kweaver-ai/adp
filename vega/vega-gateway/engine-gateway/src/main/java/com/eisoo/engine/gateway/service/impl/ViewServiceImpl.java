package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.aishu.af.vega.sql.extract.SqlExtractUtil;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.domain.dto.CatalogDto;
import com.eisoo.engine.gateway.domain.dto.ViewDto;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.domain.vo.ViewListVo;
import com.eisoo.engine.gateway.service.SchemaService;
import com.eisoo.engine.gateway.service.ViewService;
import com.eisoo.engine.gateway.util.CheckUtil;
import com.eisoo.engine.gateway.util.ErrorParseUtils;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.entity.ExcelTableConfigEntity;
import com.eisoo.engine.metadata.entity.TblsEntity;
import com.eisoo.engine.metadata.mapper.ExcelTableConfigMapper;
import com.eisoo.engine.metadata.mapper.TblsMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.StringUtils;
import com.eisoo.engine.utils.vo.ViewTableVo;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.concurrent.TimeUnit;

/**
 * @Author zdh
 **/
@Service
public class ViewServiceImpl implements ViewService {

    private static final Logger log = LoggerFactory.getLogger(ViewServiceImpl.class);
    @Autowired
    OlkCollectionServiceImpl olkCollection;

    @Autowired
    CatalogServiceImpl catalogService;

    @Autowired
    SchemaService schemaService;

    @Value(value = "${olk.view.catalog}")
    private String defaultCatalogName;

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    @Autowired(required = false)
    TblsMapper tblsMapper;

    @Autowired(required = false)
    ExcelTableConfigMapper excelTableConfigMapper;


    @Override
    public ResponseEntity<?> createView(ViewDto params, String user, boolean allowCreateExcelView) {
        checkParams(params, user);
        String query = params.getQuery();
        String viewName = params.getViewName();
        String viewCatalogName = params.getCatalogName();
        String catalogName = parseCatalog(viewCatalogName);

        if (StringUtils.isNotEmpty(viewCatalogName) && !CheckUtil.checkcatalogName(catalogName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_NAME_ERROR, Message.MESSAGE_CATALOG_NAME_SOLUTION);
        }
        boolean notExist = checkCatalogExist(catalogName);
        if (!notExist) {
            log.error("数据源VDM不存在catalogName:{}", catalogName);
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalogName, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        if (!CheckUtil.checkViewName(viewName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NAME_ERROR, Message.MESSAGE_VIEW_NAME_SOLUTION);
        }
        boolean viewNotExist = checkExistsView(viewCatalogName, params.getViewName());
        if (!viewNotExist) {
            throw new AiShuException(ErrorCodeEnum.ViewExist, params.getViewName(), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        // 不允许通过通用创建视图接口来创建excel视图
        if (!allowCreateExcelView) {
            Map<String, Set<String>> tableColumnMap = SqlExtractUtil.extractTableAndColumnRelationFromSql(query);
            if (tableColumnMap != null && tableColumnMap.size() > 0) {
                for (String fullTableName : tableColumnMap.keySet()) {
                    String catalog = fullTableName.split("\\.")[0];
                    checkCatalogType(catalog);
                }
            }
        }

        if (StringUtils.isBlank(viewCatalogName) || StringUtils.containsAnyIgnoreCase(viewCatalogName, CatalogConstant.OLK_VIEW_VDM) && !checkCatalogDefaultExist(viewCatalogName)) {
            viewCatalogName = checkAndCreateCatalog(defaultCatalogName, CatalogConstant.VIEW_DEFAULT_SCHEMA);
        }
        viewName = viewCatalogName + "." + "\"" + viewName + "\"";
        query = String.format("CREATE VIEW %s as %s", viewName, query);
        String jsonStr = olkCollection.getStatementURL(query, user);
        if (StringUtils.isBlank(jsonStr)) {
            throw new AiShuException(ErrorCodeEnum.InternalError, jsonStr, Message.MESSAGE_INTERNAL_ERROR);
        }
        String msg=olkCollection.checkState(jsonStr);
        if (StringUtils.endsWithIgnoreCase(msg, "FAILED") ||StringUtils.isEmpty(msg)) {
            log.error("SQL语句执行异常,detail:{}", jsonStr);
            throw ErrorParseUtils.finerErrorInfo(jsonStr,"");
        } else if (StringUtils.endsWithIgnoreCase(olkCollection.checkState(jsonStr), "FINISHED")) {
            log.info("创建视图成功");
            JSONArray array = JSONUtil.createArray();
            JSONObject obj = JSONUtil.createObj();
            obj.putOpt("name", viewName);
            array.add(obj);
            return ResponseEntity.ok(array);
        } else {
            log.error("未知错误,请求结果:{}", jsonStr);
            throw new AiShuException(ErrorCodeEnum.InternalError, jsonStr, Message.MESSAGE_INTERNAL_ERROR);
        }
    }

    public void checkCatalogType(String catalog) {
        String catalogStr = catalogService.showCatalogInfo(catalog);
        if (catalogStr != null ) {
            if (!catalogStr.startsWith("{")) {
                log.error("cactalog不存在:{}", catalog);
                throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalog, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
            } else {
                com.alibaba.fastjson2.JSONObject catalogJson = com.alibaba.fastjson2.JSONObject.parseObject(catalogStr);
                if (CatalogConstant.EXCEL_CATALOG.equals(catalogJson.getString("connector.name"))) {
                    log.error("不支持该操作:{}", catalog);
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.EXCEL_VIEW_OPERATE_UNSUPPORTED);
                }
            }
        }
    }

    /*public String decodeStr(String sql) {
        Pattern chineseTableNamePattern = Pattern.compile("(?<!\")([\\u4e00-\\u9fff\\w_]+)(?![\\w_]*(?=\"))");
        Matcher matcher = chineseTableNamePattern.matcher(sql);
        StringBuffer modifiedSql = new StringBuffer();
        while (matcher.find()) {
            String tableName = matcher.group();
            // 检查表名和字段名称是否包含中文字符且未被双引号包围
            if (containsChineseCharacters(tableName) && !isAlreadyQuoted(tableName, sql)) {
                // 用双引号包围中文表名
                String quotedTableName = "\"" + tableName + "\"";
                // 替换原始表名为带双引号的表名
                matcher.appendReplacement(modifiedSql, quotedTableName);
            } else {
                // 如果表名已经被双引号包围或者不包含中文字符，则不替换
                matcher.appendReplacement(modifiedSql, matcher.group());
            }
        }
        matcher.appendTail(modifiedSql);

        // 使用修改后的SQL语句进行解析
        String modifiedSqlString = modifiedSql.toString();
        return modifiedSqlString;
    }
    private boolean containsChineseCharacters(String input) {
        // 之前的正则表达式检查中文字符的方法
        Pattern chineseCharactersPattern = Pattern.compile("[\\u4e00-\\u9fff]+");
        Matcher matcher = chineseCharactersPattern.matcher(input);
        return matcher.find();
    }

    private boolean isAlreadyQuoted(String tableName, String sql) {
        // 检查表名是否已经被双引号包围
        Pattern quotedPattern = Pattern.compile("\"(" + tableName + ")\"");
        Matcher quotedMatcher = quotedPattern.matcher(sql);
        return quotedMatcher.find();
    }*/

    @Override
    public ResponseEntity<?> viewList(Long pageNum, Long pageSize, String catalogName,String schemaName,String viewName) {


        ViewListVo viewListVo = new ViewListVo();
        IPage<ViewTableVo> viewTableVoIPage;
        if (pageNum != null && pageSize != null) {
            if (pageNum <= 0 || pageSize <= 0) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.PAGE_OR_SIZE_ERROR);
            }
            Page<TblsEntity> page = new Page<TblsEntity>(pageNum, pageSize);
            log.info("分页查询视图,pageNum:{},pageSize:{},catalogName:{},schemaName:{},viewName:{}",
                    pageNum, pageSize, catalogName,schemaName,viewName);
            viewTableVoIPage = tblsMapper.queryList(page, catalogName,schemaName,viewName);
            viewListVo.setTotal(viewTableVoIPage.getTotal());
            viewListVo.setPages(viewTableVoIPage.getPages());
        } else {
            // 全量查询
            log.info("全量查询视图,catalogName:{},schemaName:{},viewName:{}", catalogName,schemaName,viewName);
            List<ViewTableVo> allRecords = tblsMapper.queryAll(catalogName,schemaName,viewName);
            viewTableVoIPage = new Page<>();
            viewTableVoIPage.setRecords(allRecords);
            viewListVo.setTotal(allRecords.size());
            viewListVo.setPages(1);
        }

        ArrayList<Object> records = new ArrayList<>();
        for (ViewTableVo record : viewTableVoIPage.getRecords()) {
            HashMap<String, String> r = new HashMap<>();
            r.put("catalogName", record.getCatalogName());
            r.put("viewName", record.getViewName());
            r.put("schema", record.getSchema());
            records.add(r);
        }
        viewListVo.setEntries(records);
        return ResponseEntity.ok(viewListVo);
    }

    @Override
    public ResponseEntity<?> replaceView(ViewDto params, String user) {
        checkParams(params, user);
        String query = params.getQuery();
        String viewName = params.getViewName();
        String viewCatalogName = params.getCatalogName();
        String catalogName = parseCatalog(viewCatalogName);

        if (StringUtils.isNotEmpty(viewCatalogName) && !CheckUtil.checkcatalogName(catalogName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CATALOG_NAME_ERROR, Message.MESSAGE_CATALOG_NAME_SOLUTION);
        }
        boolean notExist = checkCatalogExist(catalogName);
        if (!notExist) {
            log.info("数据源VDM不存在catalogName:{}", catalogName);
            throw new AiShuException(ErrorCodeEnum.CatalogNotExist, catalogName, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        if (StringUtils.isBlank(viewCatalogName)) {
            viewCatalogName = defaultCatalogName + "." + CatalogConstant.VIEW_DEFAULT_SCHEMA;
            catalogName = defaultCatalogName;
        }

        if (!CheckUtil.checkViewName(viewName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NAME_ERROR, Message.MESSAGE_VIEW_NAME_SOLUTION);
        }
        String name = tblsMapper.existView(catalogName, viewName);
        log.info("判断视图是否存在,视图名:{}", name);
        boolean viewNotExist = StringUtils.isBlank(name);
        if (viewNotExist) {
            throw new AiShuException(ErrorCodeEnum.ViewNotExist, name, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        viewName = viewCatalogName + "." + "\"" + viewName + "\"";
        query = String.format("CREATE OR REPLACE VIEW %s as %s", viewName, query);

        String statementResponse = olkCollection.getStatement(query, user);
        String nextUri = olkCollection.getnextUri(statementResponse);
        if (StringUtils.isBlank(nextUri)) {
            throw new AiShuException(ErrorCodeEnum.InternalError, nextUri, Message.MESSAGE_INTERNAL_ERROR);
        }
        String queryId = olkCollection.getQueryId(statementResponse);
        log.info("replaceView虚拟化引擎底层执行QueryId:{}\n" +
                "ViewServiceImpl.replaceView SQL statement:\n{}", queryId, query);
        //多次调用nextUri接口
        String jsonStr = olkCollection.execute(nextUri, user);
        if (StringUtils.isBlank(jsonStr)) {
            throw new AiShuException(ErrorCodeEnum.InternalError, jsonStr, Message.MESSAGE_INTERNAL_ERROR);
        }
        if (StringUtils.endsWithIgnoreCase(olkCollection.checkState(jsonStr), "FAILED")) {
            String errorMessage = olkCollection.getErrorMessage(jsonStr);
            String errorName = olkCollection.getErrorName(jsonStr);
            log.error("SQL语句执行异常,detail:{}", errorMessage);
            throw ErrorParseUtils.finerErrorInfo(errorMessage,errorName);
        } else if (StringUtils.endsWithIgnoreCase(olkCollection.checkState(jsonStr), "FINISHED")) {
            log.info("更新视图成功，开始更新元数据持久化");
            updateView(viewCatalogName, params.getViewName(), user);
            JSONArray array = JSONUtil.createArray();
            JSONObject obj = JSONUtil.createObj();
            obj.putOpt("name", viewName);
            array.add(obj);
            return ResponseEntity.ok(array);
        } else {
            log.error("未知错误,请求结果:{}", jsonStr);
            throw new AiShuException(ErrorCodeEnum.InternalError, jsonStr, Message.MESSAGE_INTERNAL_ERROR);
        }
    }


    private boolean checkExistsView(String catalogName, String viewName) {
        if (StringUtils.isEmpty(catalogName)) {
            catalogName = defaultCatalogName;
        }else{
            catalogName = parseCatalog(catalogName);
        }
        String name = tblsMapper.existView(catalogName, viewName);
        log.info("判断视图是否存在,视图名:{}", name);
        return StringUtils.isBlank(name);
    }


    @Override
    public ResponseEntity<?> deleteView(ViewDto params, String user, boolean allowCreateExcelView) {
        if (StringUtils.isNull(params)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter);
        }
        if (StringUtils.isBlank(user)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        String viewName = params.getViewName();
        String viewCatalogName = params.getCatalogName();
        if (StringUtils.isBlank(viewCatalogName)) {
            viewCatalogName = defaultCatalogName + "." + CatalogConstant.VIEW_DEFAULT_SCHEMA;
        }

        if (!allowCreateExcelView) {
            String vdmCatalog = viewCatalogName.split("\\.")[0];
            String schemaName = viewCatalogName.split("\\.")[1];
            QueryWrapper<ExcelTableConfigEntity> wrapper = new QueryWrapper<>();
            wrapper.eq("vdm_catalog", vdmCatalog)
                    .eq("schema_name", schemaName)
                    .eq("table_name", viewName);
            List<ExcelTableConfigEntity> tableConfigList = excelTableConfigMapper.selectList(wrapper);
            if (tableConfigList.size() > 0) {
                log.error("不支持该操作");
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.EXCEL_VIEW_OPERATE_UNSUPPORTED);
            }
        }

        viewName = viewCatalogName + "." + "\"" + viewName + "\"";
        String query = String.format("DROP VIEW %s", viewName);

        if (StringUtils.equalsIgnoreCase(CatalogConstant.OLK_VIEW_VDM, params.getViewName())) {
            log.error("默认视图无法删除,viewCatalogName:{},viewName:{}", viewCatalogName, params.getViewName());
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.DEFAULT_VIEW_DEL_UNSUPPORTED);
        }
        boolean viewNotExist = checkExistsView(viewCatalogName, params.getViewName());
        if (viewNotExist) {
            log.error("数据视图不存在无法删除,viewCatalogName:{},viewName:{}", viewCatalogName, params.getViewName());
            throw new AiShuException(ErrorCodeEnum.ViewNotExist, params.getViewName(), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        String statementResponse = olkCollection.getStatement(query, user);
        String nextUri = olkCollection.getnextUri(statementResponse);
        if (StringUtils.isBlank(nextUri)) {
            throw new AiShuException(ErrorCodeEnum.InternalError, nextUri, Message.MESSAGE_INTERNAL_ERROR);
        }
        log.info("deleteView虚拟化引擎底层执行QueryId:{}\n" +
                "ViewServiceImpl.deleteView SQL statement:\n{}", olkCollection.getQueryId(statementResponse), query);
        //多次调用nextUri接口
        String jsonStr = olkCollection.execute(nextUri, user);
        if (StringUtils.endsWithIgnoreCase(olkCollection.checkState(jsonStr), "FAILED")) {
            String errorMessage = olkCollection.getErrorMessage(jsonStr);
            String errorName = olkCollection.getErrorName(jsonStr);
            log.error("SQL语句执行异常,detail:{}", errorMessage);
            throw ErrorParseUtils.finerErrorInfo(errorMessage,errorName);
        } else if (StringUtils.endsWithIgnoreCase(olkCollection.checkState(jsonStr), "FINISHED")) {
            log.info("删除视图成功,开始删除元数据");
            deleteView(viewCatalogName, params.getViewName());
            //catalogService.deleteCatalogName(viewCatalogName,"vdm");
            JSONArray array = JSONUtil.createArray();
            JSONObject obj = JSONUtil.createObj();
            obj.putOpt("name", viewName);
            array.add(obj);
            return ResponseEntity.ok(array);
        } else {
            log.error("未知错误,请求结果:{}", jsonStr);
            throw new AiShuException(ErrorCodeEnum.InternalError, jsonStr, Message.MESSAGE_INTERNAL_ERROR);
        }
    }


    private void checkParams(ViewDto params, String user) {
        if (StringUtils.isNull(params)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter);
        }
        if (StringUtils.isBlank(user)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.X_PRESTO_USER_MISSING);
        }
        if (StringUtils.isBlank(params.getQuery())) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.SQL_NULL);
        }
    }


    private String checkAndCreateCatalog(String catalogName, String schemaName) {
        String catalogView = catalogName + "." + schemaName;
        String catalogNameList = getCatalogNameList(openlookengUrl);
        ObjectMapper mapper = new ObjectMapper();
        List<String> result;
        try {
            result = mapper.readValue(catalogNameList, new TypeReference<List<String>>() {});
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        if (!result.contains(catalogName)) {
            log.info("create default view catalog and schema");
            CatalogDto catalogDto = new CatalogDto();
            catalogDto.setCatalogName(catalogName);
            catalogDto.setConnectorName("vdm");
            ResponseEntity<?> responseEntity = catalogService.create(catalogDto);
            if (responseEntity.getStatusCode().value() != 200) {
                log.error("View 创建内置视图源失败,catalogName:{}", schemaName);
                throw new AiShuException(ErrorCodeEnum.InternalError, Detail.BUILT_IN_CATALOG_CREATE_ERROR, Message.MESSAGE_INTERNAL_ERROR);
            }
            log.info("create default view catalog and schema successfully");
        } else {
            ViewTableVo entity=tblsMapper.existViewByCatalog(catalogName);
            catalogView = entity.getCatalogName() + "." + entity.getSchema();
        }
        return catalogView;
    }

    private String parseCatalog(String catalogName) {
        if (StringUtils.isNotBlank(catalogName)) {
            String[] catalog = StringUtils.split(catalogName, ".");
            if (catalog.length > 1) {
                catalogName = catalog[0];
            }
        }
        return catalogName;
    }

    private boolean checkCatalogExist(String catalogName) {
        if (StringUtils.isNotEmpty(catalogName) && !StringUtils.equalsIgnoreCase(catalogName, CatalogConstant.OLK_VIEW_VDM) && tblsMapper.existViewByCatalog(catalogName) == null) {
            return false;
        }
        return true;
    }

    private boolean checkCatalogDefaultExist(String viewCatalogName) {
        return  tblsMapper.existViewByCatalog(viewCatalogName) != null;
    }

    @Transactional(rollbackFor = Exception.class)
    public void updateView(String viewCatalogName, String viewName, String user) {
        String catalogName = parseCatalog(viewCatalogName);
        tblsMapper.updateView(catalogName, viewName, user);
    }

    @Transactional(rollbackFor = Exception.class)
    public void deleteView(String catalogName, String viewName) {
        String[] catalog = StringUtils.split(catalogName, ".");
        if (catalog.length > 1) {
            catalogName = catalog[0];
        }
        tblsMapper.deleteByView(catalogName, viewName);
    }

    /**
     * 查询数据源名称列表GET请求
     *
     * @param url
     * @return
     */
    public String getCatalogNameList(String url) {
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


}
