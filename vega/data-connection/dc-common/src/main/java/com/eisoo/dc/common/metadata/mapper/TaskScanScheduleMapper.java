package com.eisoo.dc.common.metadata.mapper;

import com.eisoo.dc.common.metadata.entity.TaskScanScheduleEntity;
import com.github.yulichang.base.MPJBaseMapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * @author Tian.lan
 */
public interface TaskScanScheduleMapper extends MPJBaseMapper<TaskScanScheduleEntity> {
    List<TaskScanScheduleEntity> selectTaskScans(String dsId, String keyword);

    void updateEntityById(TaskScanScheduleEntity taskScanScheduleEntity);

    TaskScanScheduleEntity selectByDsId(@Param("dsId") String dsId);
}
