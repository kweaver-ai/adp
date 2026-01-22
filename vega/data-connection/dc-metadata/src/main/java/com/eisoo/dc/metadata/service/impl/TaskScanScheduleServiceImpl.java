package com.eisoo.dc.metadata.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.eisoo.dc.common.enums.ScheduleJobStatusEnum;
import com.eisoo.dc.common.metadata.entity.TaskScanScheduleEntity;
import com.eisoo.dc.common.metadata.mapper.TaskScanScheduleMapper;
import com.eisoo.dc.metadata.service.ITaskScanScheduleService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class TaskScanScheduleServiceImpl extends ServiceImpl<TaskScanScheduleMapper, TaskScanScheduleEntity> implements ITaskScanScheduleService {
    @Autowired
    private TaskScanScheduleMapper taskScanScheduleMapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean insert(TaskScanScheduleEntity entity) {
        int insert = taskScanScheduleMapper.insert(entity);
        return insert > 0;
    }

    @Override
    public List<TaskScanScheduleEntity> getActiveTaskScanSchedule() {
        LambdaQueryWrapper<TaskScanScheduleEntity> queryWrapper = new LambdaQueryWrapper<TaskScanScheduleEntity>()
                .eq(TaskScanScheduleEntity::getTaskStatus, ScheduleJobStatusEnum.OPEN.getCode());
        return taskScanScheduleMapper.selectList(queryWrapper);
    }

    @Override
    public List<TaskScanScheduleEntity> getTaskScanSchedule(String dsId) {
        LambdaQueryWrapper<TaskScanScheduleEntity> queryWrapper = new LambdaQueryWrapper<TaskScanScheduleEntity>()
                .eq(TaskScanScheduleEntity::getDsId, dsId);
        return taskScanScheduleMapper.selectList(queryWrapper);
    }

    @Override
    public void updateEntityById(TaskScanScheduleEntity taskScanScheduleEntity) {
        taskScanScheduleMapper.updateEntityById(taskScanScheduleEntity);
    }
}
