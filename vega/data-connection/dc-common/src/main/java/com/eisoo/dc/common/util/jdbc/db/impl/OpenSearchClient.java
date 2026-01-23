package com.eisoo.dc.common.util.jdbc.db.impl;


import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.config.OpenSearchClientCfg;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.RSAUtil;
import com.eisoo.dc.common.util.StringUtils;
import com.eisoo.dc.common.util.jdbc.db.DataSourceConfig;
import com.eisoo.dc.common.util.jdbc.db.DbClientInterface;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.opensearch.client.opensearch.cat.IndicesRequest;
import org.opensearch.client.opensearch.cat.IndicesResponse;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.sql.Connection;
import java.util.*;

import static com.eisoo.dc.common.util.CommonUtil.isEmpty;
import static com.eisoo.dc.common.util.CommonUtil.isNotEmpty;

/**
 * @author Tian.lan
 */
@Slf4j
public class OpenSearchClient implements DbClientInterface {
    @Override
    public Connection getConnection(DataSourceConfig config) throws Exception {
        return null;
    }

    @Override
    public Map<String, TableScanEntity> getTables(DataSourceEntity dataSourceEntity, List<String> scanStrategy) throws Exception {
        IndicesResponse response = null;
        Map<String, TableScanEntity> currentTables = new HashMap<>();
        String dsId = dataSourceEntity.getFId();
        org.opensearch.client.opensearch.OpenSearchClient openSearchClient = null;
        try {
            OpenSearchClientCfg openSearchClientCfg = new OpenSearchClientCfg(dataSourceEntity.getFConnectProtocol(),
                    dataSourceEntity.getFHost(),
                    dataSourceEntity.getFPort(),
                    dataSourceEntity.getFAccount(),
                    RSAUtil.decrypt(dataSourceEntity.getFPassword()));
            openSearchClient = CommonUtil.getOpenSearchClient(openSearchClientCfg);
            // 1. 构建cat indices请求
            IndicesRequest request = new IndicesRequest.Builder().build();
            // 2. 执行请求
            response = openSearchClient.cat().indices(request);
            // 封装index
            for (IndicesRecord record : response.valueBody()) {
                // "."开头的index不处理
                String indexName = record.index();
                assert indexName != null;
                if (indexName.startsWith(".")) {
                    continue;
                }
                TableScanEntity tableScanEntity = new TableScanEntity();
                tableScanEntity.setFId(UUID.randomUUID().toString());
                tableScanEntity.setFName(indexName);
                tableScanEntity.setFAdvancedParams(CommonUtil.getOpenSearchParam(record));
                currentTables.put(tableScanEntity.getFName(), tableScanEntity);
            }
        } catch (Exception e) {
            log.error("opensearch获取index元数据失败：dsId:{};response:{}",
                    dsId,
                    response,
                    e);
            throw new Exception(e);
        } finally {
            if (openSearchClient != null) {
                try {
                    openSearchClient._transport().close();
                    log.info("opensearch:dsId:{}:关闭client成功", dsId);
                } catch (IOException e) {
                    log.error("opensearch:dsId:{}:关闭client失败", dsId, e);
                }
            }
        }
        return currentTables;
    }

    @Override
    public Map<String, FieldScanEntity> getFields(String indexName, DataSourceEntity dataSourceEntity, ConnectorConfig connectorConfig) throws Exception {
        String url = dataSourceEntity.getFConnectProtocol() + "://" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + "/" + indexName;
        JSONObject result = JSONObject.parseObject(getIndexMetadata(url, dataSourceEntity.getFAccount(), RSAUtil.decrypt(dataSourceEntity.getFPassword())));
        Map<String, FieldScanEntity> currentFields = new HashMap<>();
        JSONObject properties = result.getJSONObject(indexName).getJSONObject("mappings").getJSONObject("properties");
        extractFields(null, currentFields, properties, connectorConfig);
        return currentFields;
    }

    private void extractFields(String parentName, Map<String, FieldScanEntity> currentFields, JSONObject properties, ConnectorConfig connectorConfig) {
        for (String fieldName : properties.keySet()) {
            String concatFieldName;
            if (StringUtils.isEmpty(parentName)) {
                concatFieldName = fieldName;
            } else {
                concatFieldName = parentName + "." + fieldName;
            }
            // 包含properties说明属性为object类型
            if (properties.getJSONObject(fieldName).containsKey("properties")) {
                extractFields(concatFieldName, currentFields, properties.getJSONObject(fieldName).getJSONObject("properties"), connectorConfig);
            } else {
                FieldScanEntity fieldScanEntity = new FieldScanEntity();
                fieldScanEntity.setFId(UUID.randomUUID().toString());
                fieldScanEntity.setFFieldName(concatFieldName);
                fieldScanEntity.setFFieldType(properties.getJSONObject(fieldName).getString("type"));
                fieldScanEntity.setFAdvancedParams(CommonUtil.getOpenSearchFieldParamByJson(connectorConfig, properties.getJSONObject(fieldName)));
                currentFields.put(fieldScanEntity.getFFieldName(), fieldScanEntity);
            }
        }
    }

    public String getIndexMetadata(String url, String username, String password) throws Exception {
        // 创建凭证提供者
        CredentialsProvider credentialsProvider = new BasicCredentialsProvider();
        credentialsProvider.setCredentials(
                AuthScope.ANY,
                new UsernamePasswordCredentials(username, password)
        );

        // 创建 HttpClient
        try (CloseableHttpClient httpClient = HttpClients.custom()
                .setDefaultCredentialsProvider(credentialsProvider)
                .build()) {

            // 创建 GET 请求
            HttpGet httpGet = new HttpGet(url);
            httpGet.setHeader("Accept", "application/json");
            httpGet.setHeader("Content-Type", "application/json");

            // 执行请求
            HttpResponse response = httpClient.execute(httpGet);

            // 检查响应状态
            int statusCode = response.getStatusLine().getStatusCode();
            if (statusCode != 200) {
                throw new RuntimeException("HTTP Error: " + statusCode);
            }

            // 读取响应内容
            HttpEntity entity = response.getEntity();
            if (entity != null) {
                return EntityUtils.toString(entity, StandardCharsets.UTF_8);
            }

            return "{}";
        }
    }

    @Override
    public boolean judgeTwoFiledIsChange(FieldScanEntity newFieldScanEntity, FieldScanEntity oldFieldScanEntity) {
        // opensearch 没有f_field_length  f_field_precision  f_field_comment
        String fieldTypeNew = newFieldScanEntity.getFFieldType();
        String fieldTypeOld = oldFieldScanEntity.getFFieldType();
        // 类型变化
        if (!fieldTypeNew.equals(fieldTypeOld)) {
            return true;
        }
        if ("text".equalsIgnoreCase(fieldTypeNew)) {
            String paramsNew = newFieldScanEntity.getFAdvancedParams();
            String paramsOld = oldFieldScanEntity.getFAdvancedParams();
            if (isNotEmpty(paramsNew) && !paramsNew.equals(paramsOld)) {
                return true;
            }
            if (isEmpty(paramsNew) && isNotEmpty(paramsOld)) {
                return true;
            }
        }
        return false;
    }
}
