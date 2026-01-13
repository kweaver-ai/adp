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
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        String fType = dataSourceEntity.getFType();

        Map<String, TableScanEntity> currentTables = new HashMap<>();
        int maxCount = 3;
        for (int i = 0; i < maxCount; i++) {
            try {
                currentTables = DbConnectionStrategyFactory
                        .getStrategy(dataSourceEntity.getFType())
                        .getTables(dataSourceEntity);
                break;
            } catch (Exception e) {
                log.error("【{}采集table元数据】失败：taskId:{};dsId:{}",
                        fType,
                        taskId,
                        dsId,
                        e);
                if (i == maxCount - 1) {
                    saveFail(taskScanEntity, 1, e.getMessage());
                    throw new Exception(e);
                }
            }
        }
        // 对要操作的表要加锁
        List<String> lockIds = new ArrayList<>();
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
        List<TableScanEntity> saveList = new ArrayList<>();
        List<TableScanEntity> updateList = new ArrayList<>();
        List<TableScanEntity> deletes = new ArrayList<>();
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        // 3. 格式化输出
        String now = LocalDateTime.now().format(formatter);
        try {
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
                        deletes.add(tableEntity);
                    }
                }
            }
            //2,update
            // 获取当前表级元数据并判断修改/新增
            for (String tableName : currentTables.keySet()) {
                TableScanEntity currentTable = currentTables.get(tableName);
                TableScanEntity oldTable = oldTables.get(tableName);
                // 取出id,update
                if (CommonUtil.isNotEmpty(oldTable)) {
                    // table的更新由field体现
//                    currentTable.setFId(oldTable.getFId());
//                    currentTable.setFTaskId(taskId);
//                    currentTable.setFVersion(1 + oldTable.getFVersion());
////                    currentTable.setFOperationTime(now);
//                    currentTable.setFOperationUser(userId);
//                    currentTable.setFOperationType(OperationTyeEnum.UPDATE.getCode());
//                    currentTable.setFStatus(ScanStatusEnum.WAIT.getCode());//初始化
//                    currentTable.setFStatusChange(1);
//                    updateList.add(currentTable);
                } else {
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
            // 更新
            tableScanService.tableScanBatch(deletes, updateList, saveList);
            log.info("【获取table元数据失败】:taskId:{};dsId:{}:成功将table元数据更新", taskId, dsId);
            // 把t_table_scan的插入到t_task_scan_table表里面
            List<TableScanEntity> tableScanEntities = tableScanMapper.selectByDsId(dsId);
            List<TaskScanTableEntity> data = new ArrayList<>(tableScanEntities.size());
            for (TableScanEntity table : tableScanEntities) {
                // 删除的不要
                if (1 == table.getFOperationType()) {
                    continue;
                }
                TaskScanTableEntity taskScanTableEntity = new TaskScanTableEntity(
                        UUID.randomUUID().toString(),
                        taskId,
                        dsId,
                        table.getFDataSourceName(),
                        table.getFId(),
                        table.getFName(),
                        table.getFSchemaName(),
                        ScanStatusEnum.WAIT.getCode(),
                        startTime,
                        new Date(),
                        userId,
                        null,
                        null,
                        null
                );
                data.add(taskScanTableEntity);
            }
            // 首先删除掉冗余文件
            if (!data.isEmpty()) {
                int delCount = taskScanTableService.deleteBatchByTaskIdAndTableId(data);
                taskScanTableService.saveBatchTaskScanTable(data, 100);
                log.info("更新t_task_scan_table成功");
            } else {
                log.warn("不需要更新t_task_scan_table");
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
            String idsEnd = String.join("\n", lockIds);
            log.info("【获取table元数据失败】：taskId:{};dsId:{}:获取元数据对如下的table释放了锁!\n{}",
                    taskId,
                    dsId,
                    idsEnd);
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
