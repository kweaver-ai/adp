package com.eisoo.dc.metadata.service.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.eisoo.dc.common.enums.ConnectorEnums;
import com.eisoo.dc.common.enums.OperationTyeEnum;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.metadata.mapper.TaskScanTableMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;

@Service
@Slf4j
public class TaskScanTableServiceImpl extends ServiceImpl<TaskScanTableMapper, TaskScanTableEntity> implements ITaskScanTableService {
    @Autowired(required = false)
    private TaskScanTableMapper taskScanTableMapper;
    @Autowired(required = false)
    DataSourceMapper dataSourceMapper;
    @Autowired(required = false)
    private TableScanMapper tableScanMapper;
    @Autowired(required = false)
    private ITableScanService tableScanService;


    @Transactional(rollbackFor = Exception.class)
    @Override
    public void insertBatch(List<String> tables, String dsId, String taskId, String userId) {
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        String fSchema = dataSourceEntity.getFSchema();
        String fType = dataSourceEntity.getFType();
        if (ConnectorEnums.OPENSEARCH.getConnector().equalsIgnoreCase(fType)) {
            fSchema = "default";
        }else if (CommonUtil.isEmpty(fSchema)) {
            fSchema = dataSourceEntity.getFDatabase();
        }
        List<TaskScanTableEntity> data = new ArrayList<>(tables.size());
        for (String tableId : tables) {
            TableScanEntity tableScanEntity = tableScanMapper.selectById(tableId);
            if (tableScanEntity == null) {
                log.error("tableId:{}不存在", tableId);
                throw new RuntimeException();
            }
            TaskScanTableEntity taskScanTableEntity = new TaskScanTableEntity(
                    UUID.randomUUID().toString(),
                    taskId,
                    dsId,
                    dataSourceEntity.getFName(),
                    tableId,
                    tableScanEntity.getFName(),
                    fSchema,
                    ScanStatusEnum.WAIT.getCode(),
                    new Date(),
                    null,
                    userId,
                    null,
                    null,
                    null,
                    OperationTyeEnum.INSERT.getCode()
            );
            data.add(taskScanTableEntity);
        }
        int delCount = deleteBatchByTaskIdAndTableId(data);
        log.info("成功删除了{}条", delCount);
        saveBatchTaskScanTable(data, 100);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public boolean saveBatchTaskScanTable(Collection<TaskScanTableEntity> entityList, int batch) {
        return saveBatch(entityList, batch);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public int deleteBatchByTaskIdAndTableId(List<TaskScanTableEntity> entityList) {
        return taskScanTableMapper.deleteBatchByTaskIdAndTableId(entityList);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public void updateScanStatusById(String id, int status) {
        taskScanTableMapper.updateScanStatusById(id, status);
    }

    @Override
    @Transactional(rollbackFor = Exception.class, propagation = Propagation.REQUIRES_NEW)
    public void updateByIdNewRequires(TaskScanTableEntity taskScanTableEntity) {
        updateById(taskScanTableEntity);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public void updateScanStatusBothTable(TaskScanTableEntity taskScanTableEntity, TableScanEntity tableScanEntity) {
        this.updateById(taskScanTableEntity);
        tableScanService.updateById(tableScanEntity);
    }
}
