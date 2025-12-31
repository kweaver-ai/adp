package com.eisoo.dc.metadata.service.impl;

import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.config.OpenSearchClientCfg;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.OpenSearchEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.metadata.service.IFieldScanService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.util.CommonUtil;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.opensearch.OpenSearchClient;
import org.opensearch.client.opensearch._types.mapping.*;
import org.opensearch.client.opensearch.indices.GetIndexRequest;
import org.opensearch.client.opensearch.indices.GetIndexResponse;
import org.opensearch.client.opensearch.indices.IndexState;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.Callable;
import java.util.stream.Collectors;

/**
 * @author Tian.lan
 */
@Slf4j
public class OpenSearchFieldFetchTask implements Callable<String> {
    private final ITableScanService tableScanService;
    private final IFieldScanService fieldScanService;
    private final ITaskScanTableService taskScanTableService;
    private final TaskScanTableEntity taskScanTableEntity;
    private final OpenSearchClientCfg openSearchClientCfg;
    private final String userId;
    private final ConnectorConfig connectorConfig;

    public OpenSearchFieldFetchTask(TaskScanTableEntity taskScanTableEntity,
                                    ITableScanService tableScanService,
                                    IFieldScanService fieldScanService,
                                    ITaskScanTableService taskScanTableService,
                                    OpenSearchClientCfg openSearchClientCfg,
                                    String userId,
                                    ConnectorConfig connectorConfig) {
        this.taskScanTableEntity = taskScanTableEntity;
        this.tableScanService = tableScanService;
        this.fieldScanService = fieldScanService;
        this.taskScanTableService = taskScanTableService;
        this.openSearchClientCfg = openSearchClientCfg;
        this.userId = userId;
        this.connectorConfig = connectorConfig;
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

    @Override
    public String call() {
        OpenSearchClient openSearchClient = CommonUtil.getOpenSearchClient(openSearchClientCfg);
        String indexName = taskScanTableEntity.getTableName();
        String taskId = taskScanTableEntity.getTaskId();
        String tableId = taskScanTableEntity.getTableId();
        // 记录开始时刻
        // 2. 定义格式化器（线程安全，可全局复用）
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        // 3. 格式化输出
        String now =  LocalDateTime.now().format(formatter);


        // 上一次的taskId
        String preTaskId = "";
        // filed变化的标记
        boolean fieldChanged = false;
        try {
            // 更新 t_task_scan_table : 这里仅仅更新状态
            taskScanTableService.updateScanStatusById(taskScanTableEntity.getId(), ScanStatusEnum.RUNNING.getCode());
            // 更新 t_table_scan : 任务成功,更新task_id和status;否则不更新,因此记录一个之前的taskId
            TableScanEntity table = tableScanService.getById(tableId);
            preTaskId = table.getFTaskId();
            tableScanService.updateScanStatusById(tableId, ScanStatusEnum.RUNNING.getCode());
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
            log.info("taskId:{};indexName:{};tableId:{}:open search获取field元数据成功：count:{}", taskId, indexName, tableId, list.size());
            Map<String, FieldScanEntity> currentFields = new HashMap<>();
            for (OpenSearchEntity.OpenSearchField field : list) {
                FieldScanEntity fieldScanEntity = new FieldScanEntity();
                fieldScanEntity.setFId(UUID.randomUUID().toString());
                fieldScanEntity.setFFieldName(field.getName());
                fieldScanEntity.setFFieldType(field.getType());
                fieldScanEntity.setFAdvancedParams(CommonUtil.getOpenSearchFieldParam(connectorConfig, field));
                currentFields.put(fieldScanEntity.getFFieldName(), fieldScanEntity);
            }
            // 添加_id字段&_index字段
            FieldScanEntity id = CommonUtil.makeFieldScanEntity("_id");
            currentFields.put("_id", id);
            FieldScanEntity index = CommonUtil.makeFieldScanEntity("_index");
            currentFields.put("_index", index);

            List<FieldScanEntity> fieldScanEntities = fieldScanService.selectByTableId(taskScanTableEntity.getTableId());
            Map<String, FieldScanEntity> oldFields = new HashMap<>();
            for (FieldScanEntity old : fieldScanEntities) {
                oldFields.put(old.getFFieldName(), old);
            }
            //1,获取待删除列表并删除
            List<FieldScanEntity> deletes = oldFields.keySet().stream().filter(fieldName -> !currentFields.containsKey(fieldName)).map(fieldName -> {
                FieldScanEntity old = oldFields.get(fieldName);
                if (!old.getFOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                    old.setFOperationType(OperationTyeEnum.DELETE.getCode());
                    old.setFStatusChange(1);
                    old.setFVersion(old.getFVersion() + 1);
                    old.setFOperationUser(userId);
                    old.setFOperationTime(now);
                }
                return old;
            }).collect(Collectors.toList());
            if (CommonUtil.isNotEmpty(deletes)) {
                // 标记删除
                fieldScanService.updateBatchById(deletes);
                fieldChanged = true;
            }
            // 2,
            List<FieldScanEntity> saveList = new ArrayList<>();
            List<FieldScanEntity> updateList = new ArrayList<>();

            currentFields.keySet().forEach(fieldName -> {
                FieldScanEntity currentField = currentFields.get(fieldName);
                FieldScanEntity oldField = oldFields.get(fieldName);
                // 取出id,update
                if (CommonUtil.isNotEmpty(oldField)) {
                    // 这里判断update的标准
                    boolean change = CommonUtil.judgeTwoFiledIsChane(currentField, oldField);
                    if (change) {
                        // 3. 格式化输出
                        currentField.setFId(oldField.getFId());
                        currentField.setFOperationType(OperationTyeEnum.UPDATE.getCode());
                        currentField.setFStatusChange(1);
                        currentField.setFVersion(oldField.getFVersion() + 1);
                        currentField.setFOperationTime(LocalDateTime.now().format(formatter));
                        currentField.setFOperationUser(userId);
                        updateList.add(currentField);
                    }
                } else {
                    // 这里判断insert的标准
                    currentField.setFTableId(tableId);
                    currentField.setFTableName(indexName);
                    currentField.setFOperationType(OperationTyeEnum.INSERT.getCode());
                    currentField.setFStatusChange(1);
                    currentField.setFVersion(1);
                    currentField.setFCreatTime(now);
                    currentField.setFCreatUser(userId);
                    currentField.setFOperationTime(LocalDateTime.now().format(formatter));
                    currentField.setFOperationUser(userId);
                    saveList.add(currentField);
                }
            });
            if (!saveList.isEmpty()) {
                fieldScanService.saveBatch(saveList, 1000);
                fieldChanged = true;
            }
            if (!updateList.isEmpty()) {
                fieldScanService.updateBatchById(updateList, 1000);
                fieldChanged = true;
            }
            taskScanTableEntity.setScanStatus(ScanStatusEnum.SUCCESS.getCode());
            taskScanTableEntity.setEndTime(new Date());
            taskScanTableService.updateById(taskScanTableEntity);
            // 更新 t_table_scan
            if (fieldChanged) {
                tableScanService.updateScanStatusAndOperationTimeById(tableId, taskId, ScanStatusEnum.SUCCESS.getCode());
            } else {
                tableScanService.updateScanStatusById(tableId, ScanStatusEnum.SUCCESS.getCode());
            }
            log.info("taskId:{};indexName:{};tableId:{}:openSearch获取field元数据成功结束", taskId, indexName, tableId);
            return ScanStatusEnum.fromCode(ScanStatusEnum.SUCCESS.getCode());
        } catch (Exception e) {
            log.error("open search获取field元数据失败!taskId:{};indexName:{};tableId:{}", taskId, indexName, tableId, e);
            taskScanTableEntity.setScanStatus(ScanStatusEnum.FAIL.getCode());
            taskScanTableEntity.setEndTime(new Date());
            taskScanTableEntity.setErrorStack(e.toString());
            taskScanTableService.updateByIdNewRequires(taskScanTableEntity);
            if (CommonUtil.isNotEmpty(preTaskId)) {
                // 更新 t_table_scan : 这里仅仅更新状态和taskId,也就是说fail不记录
                tableScanService.updateScanStatusByIdNewRequires(tableId, preTaskId, ScanStatusEnum.SUCCESS.getCode());
            }
            return ScanStatusEnum.fromCode(ScanStatusEnum.FAIL.getCode());
        }
    }
}
