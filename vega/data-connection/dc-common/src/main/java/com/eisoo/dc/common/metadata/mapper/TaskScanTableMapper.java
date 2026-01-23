package com.eisoo.dc.common.metadata.mapper;

import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.github.yulichang.base.MPJBaseMapper;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Set;

/**
 * @author Tian.lan
 */
public interface TaskScanTableMapper extends MPJBaseMapper<TaskScanTableEntity> {
    @Delete("Delete  FROM t_task_scan_table WHERE ds_id = #{dsId}")
    int delByDsId(@Param("dsId") String dsId);
    @Select("SELECT * FROM t_task_scan_table WHERE task_id = #{id}")
    List<TaskScanTableEntity> selectByTaskId(@Param("id") String taskId);

    int deleteBatchByTaskIdAndTableId(@Param("taskId") String taskId, @Param("list") List<String> ids);

    List<TaskScanTableEntity> selectByTaskIdAndIds(@Param("taskId") String taskId, @Param("list") List<String> ids);

    void updateScanStatusById(@Param("id") String id, @Param("status") int status);

    List<TaskScanTableEntity> selectTaskScanTables(@Param("taskId") String taskId, @Param("statusList") Set<Integer> statusList, String keyword);

    long selectCountTaskScanTables(@Param("includeIds") Set<String> includeIds, @Param("taskId") String taskId, @Param("statusList") Set<Integer> statusList, String keyword);

    List<TaskScanTableEntity> selectPageTaskScanTables(@Param("includeIds") Set<String> includeIds, String keyword, int offset, int limit, String sortOrder, String direction);
}
