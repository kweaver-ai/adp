package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONUtil;
import com.alibaba.fastjson2.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.connector.ConnectorConfigCache;
import com.eisoo.engine.gateway.connector.OperatorConfig;
import com.eisoo.engine.gateway.connector.mapping.TypeMappingFactory;
import com.eisoo.engine.gateway.domain.dto.CatalogRuleDto;
import com.eisoo.engine.gateway.domain.dto.CatalogTypeDto;
import com.eisoo.engine.gateway.domain.dto.RuleDto;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.domain.vo.RuleNameVo;
import com.eisoo.engine.gateway.service.CatalogRuleService;
import com.eisoo.engine.gateway.util.CheckUtil;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.entity.CatalogRuleEntity;
import com.eisoo.engine.metadata.mapper.CatalogRuleMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.StringUtils;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Service
public class CatalogRuleServiceImpl implements CatalogRuleService {
    private static final Logger log = LoggerFactory.getLogger(CatalogRuleServiceImpl.class);

    @Autowired
    ConnectorConfigCache connectorConfigCache;

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;

    @Autowired
    TypeMappingFactory mappingTypeFactory;

    @Autowired(required = false)
    CatalogRuleMapper catalogRuleMapper;
    @Override
    public ResponseEntity<?> OperatorList(String operator) {
        OperatorConfig operatorConfig =connectorConfigCache.getRuleConfig(operator);
        if(operatorConfig==null){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, String.format(Detail.OPERATOR_UNSUPPORTED, operator));
        }
        RuleNameVo connectorNameVos = new RuleNameVo();
        connectorNameVos.setOperators(operatorConfig.getOperators());
        return ResponseEntity.ok(connectorNameVos);
    }

    @Override
    public ResponseEntity<?> configRule(CatalogRuleDto catalogRuleDto, String user) {
        if (StringUtils.isNull(catalogRuleDto)) {
            // 未知的数据源类型
            throw new AiShuException(ErrorCodeEnum.InvalidParameter);
        }
        if (StringUtils.isNull(user)) {
            // 用户不能为空
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.X_PRESTO_USER_MISSING);
        }
        if (catalogRuleDto.getRule()==null || catalogRuleDto.getRule().size() == 0) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.RULE_NOT_NULL);
        }
        for(RuleDto ruleDto:catalogRuleDto.getRule()){
            String name=ruleDto.getName();
            String operator=ruleDto.getIsEnable();
            if(name==null){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.NAME_MISSING);
            }
            if(operator==null){
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.IS_ENABLE_MISSING);
            }
        }
        if(StringUtils.isEmpty(catalogRuleDto.getDatasourceType())){
            throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.CATALOG_TYPE_NOT_NULL);
        }

        if(StringUtils.endsWith(CatalogConstant.HOLOGRES_CATALOG,catalogRuleDto.getDatasourceType())){
            catalogRuleDto.setDatasourceType(CatalogConstant.POSTGRESQL_CATALOG);
        }
        if(StringUtils.endsWith(CatalogConstant.GAUSSDB_CATALOG,catalogRuleDto.getDatasourceType())){
            catalogRuleDto.setDatasourceType(CatalogConstant.OPENGAUSS_CATALOG);
        }

        if(StringUtils.isNotEmpty(catalogRuleDto.getCatalogName())){
            String result=showCatalogInfo(openlookengUrl,catalogRuleDto.getCatalogName());
            if (!result.contains(catalogRuleDto.getDatasourceType())) {
                log.error("数据源名称与类型不匹配:{}", catalogRuleDto.getCatalogName());
                throw new AiShuException(ErrorCodeEnum.InvalidParameter,Detail.CATALOG_TYPE_INCONSISTENT);
            }
        }

        CheckUtil.catalogCheck(catalogRuleDto,catalogRuleMapper);

        String catalogListStr = showCatalogList(openlookengUrl);
        int count=0;
        cn.hutool.json.JSONObject jsonObject = JSONUtil.createObj();
        if(StringUtils.isEmpty(catalogRuleDto.getCatalogName()) && StringUtils.isNotEmpty(catalogRuleDto.getDatasourceType())){
            List<CatalogTypeDto> list;
            try {
                list=this.convertJsonToList(catalogListStr);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
            List<String> str=list.stream().filter(catalogTypeDto -> catalogTypeDto.getConnectorName().equals(catalogRuleDto.getDatasourceType())).map(x->x.getCatalogName()).collect(Collectors.toList());

            for (String name : str){
                catalogRuleDto.setCatalogName(name);
                postCatalogRule(openlookengUrl, catalogRuleDto, user);
                count++;
            }
        }else{
            postCatalogRule(openlookengUrl, catalogRuleDto, user);
            count++;
        }
        jsonObject.set("count", count);
        jsonObject.set("msg", "successful");
        return ResponseEntity.ok(jsonObject.toStringPretty());
    }

    public List<CatalogTypeDto> convertJsonToList(String json) throws IOException {
        ObjectMapper objectMapper = new ObjectMapper();
        TypeReference<List<CatalogTypeDto>> typeReference = new TypeReference<List<CatalogTypeDto>>() {};

        return objectMapper.readValue(json, typeReference);
    }

    @Override
    public ResponseEntity<?> QueryOperatorList() {
        Map<String, List<JSONObject>> catalogMap = new HashMap<>();
        List<CatalogRuleEntity> catalogRuleEntities=catalogRuleMapper.selectAll();
        if(catalogRuleEntities!=null &&catalogRuleEntities.size()>0){
            for(CatalogRuleEntity catalogRuleEntity:catalogRuleEntities){
                String catalogName = catalogRuleEntity.getCatalogName();
                // 检查Map中是否已有该catalogName的列表
                List<JSONObject> jsonList = catalogMap.getOrDefault(catalogName, new ArrayList<>());

                // 创建一个表示节点的JSONObject
                JSONObject jsonObject = new JSONObject();
                jsonObject.put("name", catalogRuleEntity.getPushdownRule());
                jsonObject.put("is_enabled", catalogRuleEntity.getIsEnabled());

                // 将JSONObject添加到列表中
                jsonList.add(jsonObject);

                // 将更新后的列表放回到Map中
                catalogMap.put(catalogName, jsonList);
            }
        }

        // 创建一个最终的JSONObject来包含所有的catalogs
        JSONObject finalJson = new JSONObject();
        // 遍历Map，并将每个catalog的JSONArray添加到finalJson中
        for (Map.Entry<String, List<JSONObject>> entry : catalogMap.entrySet()) {
            String catalogName = entry.getKey();
            JSONArray jsonArray = new JSONArray(entry.getValue());
            finalJson.put(catalogName, jsonArray);
        }
        return ResponseEntity.ok(finalJson.toString());
    }

    @Override
    public ResponseEntity<?> RuleList() {
        Set<String> connectorName=connectorConfigCache.getRuleNames();
        return ResponseEntity.ok(connectorName.stream().sorted());
    }

    /**
     * 查询数据源详情GET请求
     *
     * @param url
     * @return
     */
    public String showCatalogInfo(String url, String catalogName) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_SHOW_CATALOG+"/"+catalogName;
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

    public void postCatalogRule(String url, CatalogRuleDto params, String user) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_CATALOG_RULE;
        String ruleInfo = JSONUtil.parseObj(params).toString();
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        log.info("开始添加数据源:{},数据源类型:{}", params.getCatalogName(), params.getDatasourceType());
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendPost(urlOpen, ruleInfo, user);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http postCatalogRule 请求失败: httpStatus={}, result={}, 耗时={}ms", 
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http postCatalogRule 请求成功: httpStatus={}, result={}, 耗时={}ms", 
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
    }

    public String showCatalogList(String url) {
        String urlOpen = url + CatalogConstant.VIRTUAL_V1_SHOW_CATALOG;
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
            log.error("Http showCatalogList 请求失败: httpStatus={}, result={}, 耗时={}ms", 
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http showCatalogList 请求成功: httpStatus={}, result={}, 耗时={}ms", 
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

}
