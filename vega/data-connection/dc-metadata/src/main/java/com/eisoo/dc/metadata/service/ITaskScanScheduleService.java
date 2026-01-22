package com.eisoo.dc.metadata.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.eisoo.dc.common.metadata.entity.TaskScanScheduleEntity;

import java.util.List;

/**
 * @author Tian.lan
 */
public interface ITaskScanScheduleService extends IService<TaskScanScheduleEntity> {
    boolean insert(TaskScanScheduleEntity entity);

    List<TaskScanScheduleEntity> getActiveTaskScanSchedule();

    List<TaskScanScheduleEntity> getTaskScanSchedule(String dsId);

    void updateEntityById(TaskScanScheduleEntity taskScanScheduleEntity);

}
