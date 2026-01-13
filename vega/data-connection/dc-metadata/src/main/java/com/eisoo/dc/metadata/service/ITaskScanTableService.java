package com.eisoo.dc.metadata.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;

import java.util.Collection;
import java.util.List;

/**
 * @author Tian.lan
 */
public interface ITaskScanTableService extends IService<TaskScanTableEntity> {
    void insertBatch(List<String> tables, String dsId, String taskId, String userId);

    boolean saveBatchTaskScanTable(Collection<TaskScanTableEntity> entityList, int batch);

    int deleteBatchByTaskIdAndTableId(List<TaskScanTableEntity> entityList);

    void updateScanStatusById(String id, int status);

    void updateByIdNewRequires(TaskScanTableEntity taskScanTableEntity);
    void updateScanStatusBothTable(TaskScanTableEntity taskScanTableEntity,  TableScanEntity tableScanEntity);
}