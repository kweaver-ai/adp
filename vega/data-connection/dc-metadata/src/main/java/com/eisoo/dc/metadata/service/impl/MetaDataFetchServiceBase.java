package com.eisoo.dc.metadata.service.impl;

import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.LockUtil;
import com.eisoo.dc.common.util.jdbc.db.DbConnectionStrategyFactory;
import com.eisoo.dc.metadata.domain.dto.TaskStatusInfoDto;
import com.eisoo.dc.metadata.service.IMetaDataFetchService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.TimeUnit;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class MetaDataFetchServiceBase implements IMetaDataFetchService {
    @Autowired(required = false)
    private DataSourceMapper dataSourceMapper;
    @Autowired(required = false)
    private TableScanMapper tableScanMapper;
    @Autowired(required = false)
    private ITaskScanService taskScanService;
    @Autowired
    private ITableScanService tableScanService;
    @Autowired
    private ITaskScanTableService taskScanTableService;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void getTables(TaskScanEntity taskScanEntity, String userId) throws Exception {
        Date startTime = new Date();
        taskScanEntity.setStartTime(startTime);
        String taskId = taskScanEntity.getId();
        String dsId = taskScanEntity.getDsId();
        Integer type = taskScanEntity.getType();
        String fType = null;
        DataSourceEntity dataSourceEntity = null;
        List<String> scanStrategy = null;
        // 是否配置了采集策略
        boolean needScanStrategy = false;
        try {
            dataSourceEntity = dataSourceMapper.selectById(dsId);
            fType = dataSourceEntity.getFType();
            String taskParamsInfo = taskScanEntity.getTaskParamsInfo();
            if (CommonUtil.isNotEmpty(taskParamsInfo)) {
                TaskScanEntity.TaskParamsInfo info = JSONObject.parseObject(taskParamsInfo, TaskScanEntity.TaskParamsInfo.class);
                scanStrategy = info.getScanStrategy();
                if (CommonUtil.isNotEmpty(scanStrategy)) {
                    needScanStrategy = true;
                }
                log.info("【{}采集table元数据】：taskId:{};dsId:{};scanStrategy:{}", fType,
                        taskId,
                        dsId,
                        scanStrategy);
            }
        } catch (Exception e) {
            log.error("【获取table元数据失败】taskId:{};dsId:{}", taskId, dsId, e);
            saveFail(taskScanEntity, 1, e.getMessage());
            throw new Exception(e);
        }
        // 下面进行table元数据采集
        Map<String, TableScanEntity> currentTables = new HashMap<>();
        int maxCount = 3;
        for (int i = 0; i < maxCount; i++) {
            try {
                currentTables = DbConnectionStrategyFactory
                        .getStrategy(fType)
                        .getTables(dataSourceEntity, scanStrategy);
                break;
            } catch (Exception e) {
                log.error("【{}采集table元数据】失败：taskId:{};dsId:{}",
                        fType,
                        taskId,
                        dsId,
                        e);
                if (i == maxCount - 1) {
                    log.warn("---已经第{}次重试，采集任务失败退出---", i);
                    saveFail(taskScanEntity, 1, e.getMessage());
                    throw new Exception(e);
                }
                log.warn("---开始第{}次重试---", i + 1);
            }
        }
        // 对要操作的表要加锁
        List<String> lockIds = new ArrayList<>();
        List<TableScanEntity> saveList = new ArrayList<>();
        List<TableScanEntity> updateList = new ArrayList<>();
        List<TableScanEntity> deleteList = new ArrayList<>();
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        String now = LocalDateTime.now().format(formatter);
        try {
            // 查出old
            List<TableScanEntity> olds = tableScanMapper.selectByDsId(taskScanEntity.getDsId());
            Map<String, TableScanEntity> oldTables = new HashMap<>();
            for (TableScanEntity old : olds) {
                oldTables.put(old.getFName(), old);
                // 对要操作的表要加锁:阻塞直到加锁成功
                boolean getLock = LockUtil.GLOBAL_MULTI_TASK_LOCK.tryLock(old.getFId(),
                        0,
                        TimeUnit.SECONDS,
                        true);
                if (getLock) {
                    lockIds.add(old.getFId());
                }
            }
            if (!needScanStrategy || scanStrategy.contains("delete")) {
                //1,获取待删除列表并删除
                for (String tableName : oldTables.keySet()) {
                    if (!currentTables.containsKey(tableName)) {
                        TableScanEntity tableEntity = oldTables.get(tableName);
                        if (!tableEntity.getFOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                            tableEntity.setFTaskId(taskId);
                            tableEntity.setFVersion(tableEntity.getFVersion() + 1);
                            tableEntity.setFOperationTime(now);
                            tableEntity.setFOperationUser(userId);
                            tableEntity.setFOperationType(OperationTyeEnum.DELETE.getCode());
                            tableEntity.setFStatusChange(1);
                            deleteList.add(tableEntity);
                        }
                    }
                }
            }
            //2,update
            // 获取当前表级元数据并判断修改/新增
            for (String tableName : currentTables.keySet()) {
                TableScanEntity currentTable = currentTables.get(tableName);
                TableScanEntity oldTable = oldTables.get(tableName);
                // 取出update
                if (CommonUtil.isNotEmpty(oldTable)) {
                    if (!needScanStrategy || scanStrategy.contains(CommonUtil.UPDATE)) {
                        // table的更新由field体现
                        currentTable.setFId(oldTable.getFId());
                        currentTable.setFDataSourceId(dsId);
                        currentTable.setFDataSourceName(dataSourceEntity.getFName());
                        String fSchema = dataSourceEntity.getFSchema();
                        if (CommonUtil.isEmpty(fSchema)) {
                            fSchema = dataSourceEntity.getFDatabase();
                        }
                        currentTable.setFSchemaName(fSchema);
                        currentTable.setFTaskId(taskId);
                        currentTable.setFVersion(1 + oldTable.getFVersion());
                        // currentTable.setFOperationTime(now);
                        currentTable.setFOperationUser(userId);
                        currentTable.setFOperationType(OperationTyeEnum.UPDATE.getCode());
                        currentTable.setFStatus(ScanStatusEnum.WAIT.getCode());//初始化
                        currentTable.setFStatusChange(1);
                        updateList.add(currentTable);
                    }
                } else {
                    // 取出insert
                    if (!needScanStrategy || scanStrategy.contains(CommonUtil.INSERT)) {
                        // 新增
                        currentTable.setFDataSourceId(dsId);
                        currentTable.setFDataSourceName(dataSourceEntity.getFName());
                        String fSchema = dataSourceEntity.getFSchema();
                        if (CommonUtil.isEmpty(fSchema)) {
                            fSchema = dataSourceEntity.getFDatabase();
                        }
                        currentTable.setFSchemaName(fSchema);
                        currentTable.setFTaskId(taskId);
                        currentTable.setFVersion(1);
                        currentTable.setFCreateTime(now);
                        currentTable.setFCreatUser(userId);
                        currentTable.setFOperationTime(now);
                        currentTable.setFOperationUser(userId);
                        currentTable.setFOperationType(OperationTyeEnum.INSERT.getCode());
                        currentTable.setFStatus(ScanStatusEnum.WAIT.getCode());//初始化
                        currentTable.setFStatusChange(1);
                        saveList.add(currentTable);
                    }
                }
            }
            // 更新:update由field的采集任务负责更新维护
            tableScanService.tableScanBatch(deleteList, new ArrayList<>(), saveList);
            log.info("【获取table元数据成功】:taskId:{};dsId:{}:成功将table元数据更新", taskId, dsId);
            // 把t_table_scan的插入到t_task_scan_table表里面
            List<TableScanEntity> alls = new ArrayList<>(currentTables.size());
            // 如果不需要策略，直接查出来所有的table然后进行field采集任务
            if (!needScanStrategy) {
//                alls = tableScanMapper.selectByDsId(dsId);
                alls.addAll(saveList);
                alls.addAll(updateList);
                alls.addAll(deleteList);
            } else {
                // 需要策略，根据策略把需要操作的table弄出来
                if (scanStrategy.contains(CommonUtil.INSERT)) {
                    alls.addAll(saveList);
                }
                if (scanStrategy.contains(CommonUtil.UPDATE)) {
                    alls.addAll(updateList);
                }
                if (scanStrategy.contains(CommonUtil.DELETE)) {
                    alls.addAll(deleteList);
                }
            }
            List<TaskScanTableEntity> data = new ArrayList<>(alls.size());
            for (TableScanEntity table : alls) {
//                // 删除的不要
//                if (1 == table.getFOperationType()) {
//                    continue;
//                }
                int code = ScanStatusEnum.WAIT.getCode();
                int operationTye = table.getFOperationType();
                if (table.getFOperationType().equals(OperationTyeEnum.UPDATE.getCode())) {
                    operationTye = OperationTyeEnum.UNKNOWN.getCode();
                }
                if (table.getFOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                    code = ScanStatusEnum.SUCCESS.getCode();
                }
                TaskScanTableEntity taskScanTableEntity = new TaskScanTableEntity(
                        UUID.randomUUID().toString(),
                        taskId,
                        dsId,
                        table.getFDataSourceName(),
                        table.getFId(),
                        table.getFName(),
                        table.getFSchemaName(),
                        code,
                        startTime,
                        new Date(),
                        userId,
                        null,
                        null,
                        null,
                        operationTye
                );
                data.add(taskScanTableEntity);
            }
            // 首先删除掉冗余文件
            if (data.size() > 0) {
                taskScanTableService.deleteBatchByTaskIdAndTableId(data);
                taskScanTableService.saveBatchTaskScanTable(data, 100);
            }
            log.info("【获取table元数据成功】taskId:{};dsId:{}", taskId, dsId);
        } catch (Exception e) {
            log.error("【获取table元数据失败】taskId:{};dsId:{}", taskId, dsId, e);
            saveFail(taskScanEntity, 1, e.getMessage());
            throw new Exception(e);
        } finally {
            for (String id : lockIds) {
                if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                    LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
                }
            }
            log.info("【获取table元数据结束】taskId:{};dsId:{}:成功释放table锁!", taskId, dsId);
        }
    }

    @Override
    public void getFieldsByTable(String table) throws Exception {
    }

    /***
     *  使用 propagation = Propagation.REQUIRES_NEW 记录错误消息，不加入外层事务
     */
    public void saveFail(TaskScanEntity taskScanEntity, Integer failStage, String errorStack) {
        taskScanEntity.setScanStatus(ScanStatusEnum.FAIL.getCode());
        taskScanEntity.setEndTime(new Date());
        TaskStatusInfoDto.TaskResultInfo taskResultInfo = new TaskStatusInfoDto.TaskResultInfo();
        taskResultInfo.setTableCount(null);
        taskResultInfo.setSuccessCount(null);
        taskResultInfo.setFailCount(null);
        taskResultInfo.setFailStage(failStage);
        taskResultInfo.setErrorStack(errorStack);
        taskScanEntity.setTaskResultInfo(JSONObject.toJSONString(taskResultInfo, WriteMapNullValue));
        taskScanService.updateByIdNewRequires(taskScanEntity);
    }
}
