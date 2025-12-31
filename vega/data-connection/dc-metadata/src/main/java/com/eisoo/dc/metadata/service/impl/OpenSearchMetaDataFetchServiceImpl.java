package com.eisoo.dc.metadata.service.impl;

import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.RSAUtil;
import com.eisoo.dc.common.config.OpenSearchClientCfg;
import com.eisoo.dc.metadata.domain.dto.TaskStatusInfoDto;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.metadata.service.IMetaDataFetchService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import com.eisoo.dc.common.util.LockUtil;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.opensearch.OpenSearchClient;
import org.opensearch.client.opensearch.cat.IndicesRequest;
import org.opensearch.client.opensearch.cat.IndicesResponse;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class OpenSearchMetaDataFetchServiceImpl implements IMetaDataFetchService {
    @Autowired(required = false)
    private ITaskScanService taskScanService;
    @Autowired
    private ITableScanService tableScanService;
    @Autowired
    private ITaskScanTableService taskScanTableService;
    @Autowired(required = false)
    private TableScanMapper tableScanMapper;
    @Autowired(required = false)
    DataSourceMapper dataSourceMapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void getTables(TaskScanEntity taskScanEntity, String userId) throws Exception {
        String dsId = taskScanEntity.getDsId();
        String taskId = taskScanEntity.getId();
        IndicesResponse response = null;
        DataSourceEntity dataSourceEntity;
        Integer failStage = 1;
        String errorStack = "";
        Date startTime = new Date();
        taskScanEntity.setStartTime(startTime);
        Map<String, TableScanEntity> currentTables = new HashMap<>();
        try {
            dataSourceEntity = dataSourceMapper.selectById(dsId);
            OpenSearchClientCfg openSearchClientCfg = new OpenSearchClientCfg(dataSourceEntity.getFConnectProtocol(),
                    dataSourceEntity.getFHost(),
                    dataSourceEntity.getFPort(),
                    dataSourceEntity.getFAccount(),
                    RSAUtil.decrypt(dataSourceEntity.getFPassword()));
            final OpenSearchClient client = CommonUtil.getOpenSearchClient(openSearchClientCfg);
            // 1. 构建cat indices请求
            IndicesRequest request = new IndicesRequest.Builder().build();
            // 2. 执行请求
            response = client.cat().indices(request);
        } catch (Exception e) {
            failStage = 1;
            errorStack = e.getMessage();
            log.error("opensearch获取index元数据失败：taskId:{};dsId:{};response:{}",
                    taskId,
                    dsId,
                    response,
                    e);
            saveFail(taskScanEntity, failStage, errorStack);
            throw new Exception(e);
        }
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
        // 对要操作的表要加锁
        List<String> lockIds = new ArrayList<>();
        // 查出old
        List<TableScanEntity> olds = tableScanMapper.selectByDsId(taskScanEntity.getDsId());
        Map<String, TableScanEntity> oldTables = new HashMap<>();
        for (TableScanEntity old : olds) {
            oldTables.put(old.getFName(), old);
            // 对要操作的表要加锁:阻塞直到加锁成功
            boolean getLock = LockUtil.GLOBAL_MULTI_TASK_LOCK.tryLock(old.getFId(), 0, TimeUnit.SECONDS, true);
            if (getLock) {
                lockIds.add(old.getFId());
            }
        }
        String ids = String.join("\n", lockIds);
        log.info("open search:taskId:{};dsId:{}:获取index元数据对如下的table加锁，请关注后面是否释放锁\n{}",
                taskId,
                dsId,
                ids);
        List<TableScanEntity> saveList = new ArrayList<>();
        List<TableScanEntity> updateList = new ArrayList<>();
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        // 3. 格式化输出
        String now = LocalDateTime.now().format(formatter);
        try {
            //1,获取待删除列表并删除
            List<TableScanEntity> deletes = oldTables.keySet().stream().filter(tableName -> !currentTables.containsKey(tableName)).map(tableName -> {
                TableScanEntity tableEntity = oldTables.get(tableName);
                if (!tableEntity.getFOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                    tableEntity.setFTaskId(taskId);
                    tableEntity.setFVersion(tableEntity.getFVersion() + 1);
                    tableEntity.setFOperationTime(now);
                    tableEntity.setFOperationUser(userId);
                    tableEntity.setFOperationType(OperationTyeEnum.DELETE.getCode());
                    tableEntity.setFStatusChange(1);
                }
                return tableEntity;
            }).collect(Collectors.toList());

            //2,update
            // 获取当前表级元数据并判断修改/新增
            currentTables.keySet().forEach(tableName -> {
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
                    currentTable.setFSchemaName(dataSourceEntity.getFSchema());
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
            });
            // 更新
            tableScanService.tableScanBatch(deletes, updateList, saveList);
            log.info("open search:taskId:{};dsId:{}:成功将table元数据更新", taskId, dsId);
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
            int delCount = taskScanTableService.deleteBatchByTaskIdAndTableId(data);
            taskScanTableService.saveBatchTaskScanTable(data, 100);
            log.info("【获取index元数据成功】taskId:{};dsId:{}", taskId, dsId);
        } catch (Exception e) {
            failStage = 1;
            errorStack = e.getMessage();
            log.error("【获取index元数据失败】taskId:{};dsId:{}", taskId, dsId, e);
            saveFail(taskScanEntity, failStage, errorStack);
            throw new Exception(e);
        } finally {
            for (String id : lockIds) {
                if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                    LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
                }
            }
            String idsEnd = String.join("\n", lockIds);
            log.info("open search：taskId:{};dsId:{}:获取index元数据对如下的table释放了锁!\n{}",
                    taskId,
                    dsId,
                    idsEnd);
        }
    }

    @Override
    public void getFieldsByTable(String indexName) {
        //TODO:异步实现，在OpenSearchFieldFetchTask里面，同步实现暂时不需要
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
