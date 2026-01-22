package com.eisoo.dc.metadata.service.impl;

import cn.hutool.core.date.StopWatch;
import com.alibaba.fastjson2.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.constant.Constants;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.LockUtil;
import com.eisoo.dc.common.util.http.IndexbaseHttpUtils;
import com.eisoo.dc.common.vo.HttpResInfo;
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
import java.util.stream.Collectors;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class IndexBaseMetaDataFetchServiceImpl implements IMetaDataFetchService {
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
        DataSourceEntity dataSourceEntity;
        Integer failStage = 1;
        String errorStack = "";
        Date startTime = new Date();
        taskScanEntity.setStartTime(startTime);
        Map<String, TableScanEntity> currentTables = new HashMap<>();

        try {
            dataSourceEntity = dataSourceMapper.selectById(dsId);
            String protocol = dataSourceEntity.getFConnectProtocol(); // http 或 https
            String host = dataSourceEntity.getFHost();
            int port = dataSourceEntity.getFPort();
            String urlOpen = protocol + "://" + host + ":" + port + "/api/mdl-index-base/v1/index_bases";

            JSONObject catalogs = new JSONObject();
            StopWatch stopWatch = new StopWatch();
            stopWatch.start();

            HttpResInfo result;
            try {
                result = IndexbaseHttpUtils.sendGet(urlOpen,Constants.DEFAULT_AD_HOC_USER);
            } catch (AiShuException e) {
                throw e;
            } catch (Exception e) {
                throw new AiShuException(ErrorCodeEnum.CalculateError, e.getMessage());
            }
            stopWatch.stop();

            // 解析返回的JSON数据
            JSONObject jsonResponse = JSONObject.parseObject(JSONObject.toJSONString(result.getResult()));
            JSONArray entries = jsonResponse.getJSONArray("entries");

            // 封装index数据
            for (int i = 0; i < entries.size(); i++) {
                JSONObject entry = entries.getJSONObject(i);
                String indexName = entry.getString("name");

                // "."开头的index不处理
                if (indexName.startsWith(".")) {
                    continue;
                }

                TableScanEntity tableScanEntity = new TableScanEntity();
                tableScanEntity.setFId(UUID.randomUUID().toString());
                tableScanEntity.setFName(indexName);

                // 将entry中的信息转换为高级参数
                tableScanEntity.setFAdvancedParams(entry.toJSONString());
                currentTables.put(tableScanEntity.getFName(), tableScanEntity);
            }

        } catch (Exception e) {
            failStage = 1;
            errorStack = e.getMessage();
            log.error("index base获取index元数据失败：taskId:{};dsId:{}", taskId, dsId, e);
            saveFail(taskScanEntity, failStage, errorStack);
            throw new Exception(e);
        }

        // 后续处理逻辑保持不变...
        // 包括：
        // 1. 获取旧表数据
        // 2. 加锁处理
        // 3. 比较差异（新增/删除/更新）
        // 4. 批量保存/更新表信息
        // 5. 更新任务扫描表
        // 6. 释放锁

        // 查出old
        List<TableScanEntity> olds = tableScanMapper.selectByDsId(taskScanEntity.getDsId());
        Map<String, TableScanEntity> oldTables = new HashMap<>();
        List<String> lockIds = new ArrayList<>();

        for (TableScanEntity old : olds) {
            oldTables.put(old.getFName(), old);
            // 对要操作的表要加锁:阻塞直到加锁成功
            boolean getLock = LockUtil.GLOBAL_MULTI_TASK_LOCK.tryLock(old.getFId(), 0, TimeUnit.SECONDS, true);
            if (getLock) {
                lockIds.add(old.getFId());
            }
        }

        String ids = String.join("\n", lockIds);
        log.info("index base:taskId:{};dsId:{}:获取index元数据对如下的table加锁，请关注后面是否释放锁\n{}",
                taskId, dsId, ids);

        List<TableScanEntity> saveList = new ArrayList<>();
        List<TableScanEntity> updateList = new ArrayList<>();
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        String now = LocalDateTime.now().format(formatter);

        try {
            //1,获取待删除列表并删除
            List<TableScanEntity> deletes = oldTables.keySet().stream()
                    .filter(tableName -> !currentTables.containsKey(tableName))
                    .map(tableName -> {
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
                    })
                    .collect(Collectors.toList());

            //2,处理当前表（新增/更新）
            currentTables.keySet().forEach(tableName -> {
                TableScanEntity currentTable = currentTables.get(tableName);
                TableScanEntity oldTable = oldTables.get(tableName);

                if (CommonUtil.isNotEmpty(oldTable)) {
                    // 如果需要更新表信息，可以在这里处理
                    // 目前根据业务逻辑似乎不需要更新表级信息，由字段体现更新
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
                    currentTable.setFStatus(ScanStatusEnum.WAIT.getCode());
                    currentTable.setFStatusChange(1);
                    saveList.add(currentTable);
                }
            });

            // 更新
            tableScanService.tableScanBatch(deletes, updateList, saveList);
            log.info("index base:taskId:{};dsId:{}:成功将table元数据更新", taskId, dsId);

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
                        null,
                        OperationTyeEnum.INSERT.getCode()
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
            log.info("index base：taskId:{};dsId:{}:获取index元数据对如下的table释放了锁!\n{}",
                    taskId, dsId, idsEnd);
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
