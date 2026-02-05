package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.domain.dto.CatalogDto;
import com.eisoo.engine.gateway.domain.dto.SchemaDto;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.service.CatalogService;
import com.eisoo.engine.gateway.service.SchemaService;
import com.eisoo.engine.gateway.service.ViewService;
import com.eisoo.engine.gateway.util.ASUtil;
import com.eisoo.engine.gateway.util.ExcelUtil;
import com.eisoo.engine.gateway.util.CheckUtil;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.mapper.CatalogRuleMapper;
import com.eisoo.engine.metadata.mapper.DataSourceMapper;
import com.eisoo.engine.metadata.mapper.DictMapper;
import com.eisoo.engine.metadata.mapper.ExcelColumnTypeMapper;
import com.eisoo.engine.metadata.mapper.ExcelTableConfigMapper;
import com.eisoo.engine.metadata.mapper.TblsMapper;
import com.eisoo.engine.metadata.mapper.VegaDatasourceMapper;
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
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.concurrent.TimeUnit;

/**
 * @Author zdh
 **/
@Service
public class CatalogServiceImpl implements CatalogService {
    private static final Logger log = LoggerFactory.getLogger(CatalogServiceImpl.class);

    @Autowired(required = false)
    SchemaService schemaService;

    @Autowired(required = false)
    ViewService viewService;

    @Autowired(required = false)
    ExcelUtil excelUtil;

    @Autowired(required = false)
    ASUtil asUtil;

    @Autowired(required = false)
    ExcelTableConfigMapper excelTableConfigMapper;

    @Autowired(required = false)
    ExcelColumnTypeMapper excelColumnTypeMapper;

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    @Autowired(required = false)
    TblsMapper tblsMapper;

    @Autowired(required = false)
    CatalogRuleMapper catalogRuleMapper;

    @Autowired(required = false)
    VegaDatasourceMapper vegaDatasourceMapper;

    @Autowired(required = false)
    DataSourceMapper metadataDataSourceMapper;

    @Autowired(required = false)
    DictMapper metadataDictMapper;

    @Override
    public ResponseEntity<?> create(CatalogDto params) {
        CheckUtil.catalogCheck(params.getCatalogName(), params.getConnectorName());

        if (StringUtils.equalsIgnoreCase(params.getCatalogName(), CatalogConstant.OLK_VIEW_VDM) && !StringUtils.equalsIgnoreCase(params.getConnectorName(), "vdm")) {
            log.error("与内置视图源名称冲突：{}", params.getCatalogName());
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.BUILT_IN_CATALOG_CONFLICT);
        }

        String catalogNameList = getCatalogNameList(openlookengUrl);
        ObjectMapper mapper = new ObjectMapper();
        List<String> result;
        try {
            result = mapper.readValue(catalogNameList, new TypeReference<List<String>>() {});
        } catch (JsonProcessingException e) {
            log.error(e.getMessage());
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        if (result.contains(params.getCatalogName())) {
            log.error("数据源已存在catalogName:{}", params.getCatalogName());
            throw new AiShuException(ErrorCodeEnum.CatalogExist, params.getCatalogName(), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        if (params.getConnectorName().equals(CatalogConstant.VDM_CATALOG)) {
            return vdmCreate(openlookengUrl, params);
        }
        return null;
    }


    private ResponseEntity<?> vdmCreate(String openlookengUrl, CatalogDto catalogDto) {
        postCatalogInfo(openlookengUrl, catalogDto);
        try{
            createSchema(catalogDto);
        }catch(Exception e){
            deleteCatalogInfo(openlookengUrl, catalogDto.getCatalogName());
            log.info("catalogName:{},新增数据源时数据库同步失败，并删除数据源成功!", catalogDto.getCatalogName());
            throw new AiShuException(ErrorCodeEnum.InternalError,e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        }
        log.info("数据源添加成功:{}", catalogDto.getCatalogName());
        JSONArray jsonArray = JSONUtil.createArray();
        JSONObject jsonObject = JSONUtil.createObj();
        jsonObject.set("name", catalogDto.getCatalogName() + "." + CatalogConstant.VIEW_DEFAULT_SCHEMA);
        jsonArray.add(jsonObject);
        return ResponseEntity.ok(jsonArray);
    }

    /**
     * 视图创建schema
     *
     * @param catalogDto
     */
    private void createSchema(CatalogDto catalogDto) {
        String createSchema = String.format("CREATE SCHEMA IF NOT EXISTS %s.%s", catalogDto.getCatalogName(), CatalogConstant.VIEW_DEFAULT_SCHEMA);
        SchemaDto schemaDto = new SchemaDto();
        schemaDto.setCatalogName(catalogDto.getCatalogName());
        schemaDto.setSchemaName(CatalogConstant.VIEW_DEFAULT_SCHEMA);
        schemaDto.setQuery(createSchema);
        log.info("创建视图源schema,catalogName:{},schema:{}", catalogDto.getCatalogName(), CatalogConstant.VIEW_DEFAULT_SCHEMA);
        ResponseEntity<?> responseEntity = schemaService.createSchema(schemaDto, "admin");
        if (responseEntity.getStatusCode().value() != 200) {
            log.error("View创建视图源schema失败,catalogName:{},schema:{}", catalogDto.getCatalogName(), CatalogConstant.VIEW_DEFAULT_SCHEMA);
            throw new AiShuException(ErrorCodeEnum.InternalError, Detail.BUILT_IN_SCHEMA_CREATE_ERROR, Message.MESSAGE_INTERNAL_ERROR);
        }
    }

    /**
     * 新增数据源post请求
     *
     * @param url
     * @param params
     */
    public void postCatalogInfo(String url, CatalogDto params) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_CATALOG;
        String catalogInfo = JSONUtil.parseObj(params).toString();
        HashMap<String, String> catalogs = new HashMap<>();
        catalogs.put("catalogInformation", catalogInfo);
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        log.info("开始添加数据源:{},数据源类型:{}", params.getCatalogName(), params.getConnectorName());
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendPost(urlOpen, catalogs, Constants.DEFAULT_AD_HOC_USER);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>= HttpStatus.BAD_REQUEST){
            log.error("Http postCatalogInfo 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http postCatalogInfo 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
    }

    /**
     * 删除数据源 DELETE
     *
     * @param url
     */
    public void deleteCatalogInfo(String url, String catalogName) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_CATALOG + "/" + catalogName;
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
            log.error("Http deleteCatalogInfo 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http deleteCatalogInfo 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
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

    @Override
    public String showCatalogInfo(String catalogName) {

        String urlOpen = openlookengUrl + CatalogConstant.VIRTUAL_V1_SHOW_CATALOG+"/"+catalogName;
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
            log.error("Http showCatalogInfo 请求失败: httpStatus={}, result={}, 耗时={}ms",
                    result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http showCatalogInfo 请求成功: httpStatus={}, result={}, 耗时={}ms",
                result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }


}
