package com.eisoo.dc.common.util.jdbc.db.impl;


import com.eisoo.dc.common.config.OpenSearchClientCfg;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.OpenSearchEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.RSAUtil;
import com.eisoo.dc.common.util.jdbc.db.DataSourceConfig;
import com.eisoo.dc.common.util.jdbc.db.DbClientInterface;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.opensearch._types.mapping.*;
import org.opensearch.client.opensearch.cat.IndicesRequest;
import org.opensearch.client.opensearch.cat.IndicesResponse;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.opensearch.client.opensearch.indices.GetIndexRequest;
import org.opensearch.client.opensearch.indices.GetIndexResponse;
import org.opensearch.client.opensearch.indices.IndexState;

import java.io.IOException;
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
        OpenSearchClientCfg openSearchClientCfg = new OpenSearchClientCfg(dataSourceEntity.getFConnectProtocol(),
                dataSourceEntity.getFHost(),
                dataSourceEntity.getFPort(),
                dataSourceEntity.getFAccount(),
                RSAUtil.decrypt(dataSourceEntity.getFPassword()));
        org.opensearch.client.opensearch.OpenSearchClient openSearchClient = null;
        Map<String, FieldScanEntity> currentFields = new HashMap<>();
        try {
            openSearchClient = CommonUtil.getOpenSearchClient(openSearchClientCfg);
            // 创建获取索引请求
            GetIndexRequest request = new GetIndexRequest.Builder().index(indexName)
                    .flatSettings(true)
                    .build();
            GetIndexResponse getIndexResponse = openSearchClient.indices().get(request);
            Map<String, IndexState> result = getIndexResponse.result();
            IndexState indexState = result.get(indexName);
            TypeMapping mappings = indexState.mappings();
            assert mappings != null;
            Map<String, Property> properties = mappings.properties();
            ArrayList<OpenSearchEntity.OpenSearchField> list = new ArrayList<>();
            extractFields("", properties, list);
            log.info("indexName:{}:open search获取field元数据成功：count:{}",
                    indexName,
                    list.size()
            );
            for (OpenSearchEntity.OpenSearchField field : list) {
                FieldScanEntity fieldScanEntity = new FieldScanEntity();
                fieldScanEntity.setFId(UUID.randomUUID().toString());
                fieldScanEntity.setFFieldName(field.getName());
                fieldScanEntity.setFFieldType(field.getType());
                fieldScanEntity.setFAdvancedParams(CommonUtil.getOpenSearchFieldParam(connectorConfig, field));
                currentFields.put(fieldScanEntity.getFFieldName(), fieldScanEntity);
            }
        } catch (Exception e) {
            log.error("opensearch:dsId:{}:获取index元数据失败",
                    dataSourceEntity.getFId(),
                    e);
            throw new Exception(e);
        } finally {
            if (openSearchClient != null) {
                try {
                    openSearchClient._transport().close();
                    log.info("opensearch:dsId:{}:关闭client成功", dataSourceEntity.getFId());
                } catch (IOException e) {
                    log.error("opensearch:dsId:{}:关闭client失败", dataSourceEntity.getFId(), e);
                }
            }
        }
        return currentFields;
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

    private static void extractFields(String parentName, Map<String, Property> properties, ArrayList<OpenSearchEntity.OpenSearchField> list) {
        // fields.keyword.type
        // fields.keyword.ignore_above
        // norms
        // analyzer
        Set<Map.Entry<String, Property>> entries = properties.entrySet();
        for (Map.Entry<String, Property> entry : entries) {
            OpenSearchEntity.OpenSearchField openSearchField = new OpenSearchEntity.OpenSearchField();

            String fieldName = entry.getKey();
            Property property = entry.getValue();
            String typeName = property._kind().name();
            if (parentName.length() != 0) {
                fieldName = parentName + "." + fieldName;
            }
            openSearchField.setName(fieldName);
            openSearchField.setType(typeName);
            PropertyVariant propertyVariant = property._get();
            if (property.isObject()) {
                ObjectProperty o = (ObjectProperty) propertyVariant;
                Map<String, Property> fields = o.properties();
                extractFields(fieldName, fields, list);
            } else {
                PropertyBase base = (PropertyBase) propertyVariant;
                if (property.isText()) {
                    TextProperty t = (TextProperty) propertyVariant;
                    String analyzer = t.analyzer();//ik_max_word
                    Boolean norms = t.norms();
                    openSearchField.setNorms(norms);
                    openSearchField.setAnalyzer(analyzer);
                }
                // 判断是否有fields子字段
                Map<String, Property> fields = base.fields();
                // 判断是否有keyword这个field
                if (fields.containsKey("keyword")) {
                    Property keywordProperty = fields.get("keyword");
                    String type = keywordProperty._kind().name();
                    PropertyVariant variant = keywordProperty._get();
                    if (variant instanceof KeywordProperty) {
                        KeywordProperty p = (KeywordProperty) variant;
                        Integer ignoreAbove = p.ignoreAbove();
                        openSearchField.setKeywordType(type);
                        openSearchField.setIgnoreAbove(ignoreAbove);
                    }
                }
                list.add(openSearchField);
            }
        }
    }

}
