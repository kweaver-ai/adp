package com.eisoo.dc.metadata.service.impl;

import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.jdbc.db.DbClientInterface;
import com.eisoo.dc.common.util.jdbc.db.DbConnectionStrategyFactory;
import com.eisoo.dc.metadata.service.IFieldScanService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import lombok.extern.slf4j.Slf4j;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.Callable;

/**
 * @author Tian.lan
 */
@Slf4j
public class FieldFetchCallable implements Callable<String> {
    private final ITableScanService tableScanService;
    private final IFieldScanService fieldScanService;
    private final ITaskScanTableService taskScanTableService;
    private final TaskScanTableEntity taskScanTableEntity;
    private final DataSourceEntity dataSourceEntity;
    private final String userId;
    private final ConnectorConfig connectorConfig;

    public FieldFetchCallable(TaskScanTableEntity taskScanTableEntity,
                              ITableScanService tableScanService,
                              IFieldScanService fieldScanService,
                              ITaskScanTableService taskScanTableService,
                              DataSourceEntity dataSourceEntity,
                              String userId,
                              ConnectorConfig connectorConfig) {
        this.taskScanTableEntity = taskScanTableEntity;
        this.tableScanService = tableScanService;
        this.fieldScanService = fieldScanService;
        this.taskScanTableService = taskScanTableService;
        this.dataSourceEntity = dataSourceEntity;
        this.userId = userId;
        this.connectorConfig = connectorConfig;
    }

    @Override
    public String call() {
        String tableName = taskScanTableEntity.getTableName();
        String taskId = taskScanTableEntity.getTaskId();
        String tableId = taskScanTableEntity.getTableId();
        String fType = dataSourceEntity.getFType();
        // 记录开始时刻
        // 2. 定义格式化器（线程安全，可全局复用）
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        // 3. 格式化输出
        String now = LocalDateTime.now().format(formatter);
        // 上一次的taskId
        String preTaskId = "";
        // filed变化的标记
        boolean fieldChanged = false;

        try {
            TableScanEntity tableScanEntity = tableScanService.getById(tableId);

            // 更新 t_task_scan_table和t_table_scan的状态：RUNNING
            taskScanTableEntity.setStartTime(new Date());
            taskScanTableEntity.setScanStatus(ScanStatusEnum.RUNNING.getCode());
            tableScanEntity.setFStatus(ScanStatusEnum.RUNNING.getCode());

            taskScanTableService.updateScanStatusBothTable(taskScanTableEntity, tableScanEntity);

            // 更新 t_table_scan : 任务成功,更新task_id和status;否则不更新,因此记录一个之前的taskId
            preTaskId = tableScanEntity.getFTaskId();

            DbClientInterface dbClientInterface = DbConnectionStrategyFactory.getStrategy(fType);
            Map<String, FieldScanEntity> currentFields = new HashMap<>();
            int maxCount = 3;
            for (int i = 0; i < maxCount; i++) {
                try {
                    currentFields = dbClientInterface.getFields(tableName,
                            dataSourceEntity,
                            connectorConfig);
                } catch (Exception e) {
                    log.error("【{}采集field元数据】失败：tableName:{};taskId:{}",
                            fType,
                            tableName,
                            taskId,
                            e);
                    if (i == maxCount - 1) {
                        log.warn("---已经第{}次重试，采集任务失败退出---", i);
                        throw new Exception(e);
                    }
                    log.warn("---开始第{}次重试---", i + 1);
                }
            }
            List<FieldScanEntity> fieldScanEntities = fieldScanService.selectByTableId(tableId);
            Map<String, FieldScanEntity> oldFields = new HashMap<>();

            List<FieldScanEntity> saveList = new ArrayList<>();
            List<FieldScanEntity> updateList = new ArrayList<>();
            List<FieldScanEntity> deletes = new ArrayList<>();

            for (FieldScanEntity old : fieldScanEntities) {
                oldFields.put(old.getFFieldName(), old);
            }
            //1,获取待删除列表并删除
            for (String fieldName : oldFields.keySet()) {
                if (!currentFields.containsKey(fieldName)) {
                    FieldScanEntity old = oldFields.get(fieldName);
                    if (!old.getFOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                        old.setFOperationType(OperationTyeEnum.DELETE.getCode());
                        old.setFStatusChange(1);
                        old.setFVersion(old.getFVersion() + 1);
                        old.setFOperationUser(userId);
                        old.setFOperationTime(now);
                        deletes.add(old);
                    }
                }
            }

            for (String fieldName : currentFields.keySet()) {
                FieldScanEntity currentField = currentFields.get(fieldName);
                FieldScanEntity oldField = oldFields.get(fieldName);
                // 取出id,update
                if (CommonUtil.isNotEmpty(oldField)) {
                    // 这里判断update的标准
                    boolean change = dbClientInterface.judgeTwoFiledIsChange(currentField, oldField);
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
                    currentField.setFTableName(tableName);
                    currentField.setFOperationType(OperationTyeEnum.INSERT.getCode());
                    currentField.setFStatusChange(1);
                    currentField.setFVersion(1);
                    currentField.setFCreatTime(now);
                    currentField.setFCreatUser(userId);
                    currentField.setFOperationTime(LocalDateTime.now().format(formatter));
                    currentField.setFOperationUser(userId);
                    saveList.add(currentField);
                }
            }
            // 持久化field元数据信息
            fieldScanService.fieldScanBatch(deletes, updateList, saveList);
            if (!updateList.isEmpty() || !deletes.isEmpty() || !saveList.isEmpty()) {
                fieldChanged = true;
            }
            // 更新状态
            taskScanTableEntity.setScanStatus(ScanStatusEnum.SUCCESS.getCode());
            taskScanTableEntity.setEndTime(new Date());
            tableScanEntity.setFStatus(ScanStatusEnum.SUCCESS.getCode());
            // 更新 t_table_scan
            if (fieldChanged) {
                tableScanEntity.setFTaskId(taskId);
                if (OperationTyeEnum.UNKNOWN.getCode() == taskScanTableEntity.getOperationType()) {
                    taskScanTableEntity.setOperationType(OperationTyeEnum.UNKNOWN.getCode());
                }
            }
            taskScanTableService.updateScanStatusBothTable(taskScanTableEntity, tableScanEntity);
            log.info("taskId:{};tableName:{}:获取field元数据成功结束",
                    taskId,
                    tableName);
            return ScanStatusEnum.fromCode(ScanStatusEnum.SUCCESS.getCode());
        } catch (Exception e) {
            log.error("taskId:{};tableName:{}:获取field元数据失败!",
                    taskId,
                    tableName,
                    e
            );

            taskScanTableEntity.setScanStatus(ScanStatusEnum.FAIL.getCode());
            taskScanTableEntity.setEndTime(new Date());
            taskScanTableEntity.setErrorStack(e.toString());

            taskScanTableService.updateByIdNewRequires(taskScanTableEntity);
            if (CommonUtil.isNotEmpty(preTaskId)) {
                // 更新 t_table_scan : 这里仅仅更新状态和taskId,也就是说fail不记录
                tableScanService.updateScanStatusByIdNewRequires(tableId,
                        preTaskId,
                        ScanStatusEnum.SUCCESS.getCode()
                );
            }
            return ScanStatusEnum.fromCode(ScanStatusEnum.FAIL.getCode());
        }
    }
}
