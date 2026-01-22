package com.eisoo.dc.common.metadata.mapper;

import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.github.yulichang.base.MPJBaseMapper;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;
import java.util.Set;

/**
 * @author Tian.lan
 */
public interface TaskScanMapper extends MPJBaseMapper<TaskScanEntity> {
    @Delete("Delete  FROM t_task_scan WHERE ds_id = #{dsId}")
    int delByDsId(@Param("dsId") String dsId);

    @Select("SELECT count(*) FROM t_task_scan WHERE ds_id = #{dsId} and scan_status=1 and type=0")
    int getRunningDs(@Param("dsId") String dsId);

    List<TaskScanEntity> selectTaskStatusByDsIds(@Param("dsIds") Set<String> dsIds);

    @Update("UPDATE  t_task_scan SET scan_status=#{status},start_time=now() WHERE id = #{id}")
    int updateScanStatusStart(@Param("id") String id, @Param("status") int status);

    @Update("UPDATE  t_task_scan SET scan_status=#{status},end_time=now() WHERE id = #{id}")
    int updateScanStatusEnd(@Param("id") String id, @Param("status") int status);

    List<TaskScanEntity> selectTaskScans(
            @Param("id") String id,
            @Param("dsId") String dsId,
            @Param("taskType") List<Integer> taskType,
            @Param("statusList") List<Integer> statusList,
            @Param("keyword") String keyword
    );

    long selectCount(@Param("includeIds") Set<String> includeIds,
                     @Param("dsId") String dsId,
                     @Param("taskType") List<Integer> taskType,
                     @Param("statusList") List<Integer> statusList,
                     @Param("keyword") String keyword
    );

    List<TaskScanEntity> selectPage(
            @Param("includeIds") Set<String> includeIds,
            @Param("keyword") String keyword,
            @Param("offset") int offset,
            @Param("limit") int limit,
            @Param("sortOrder") String sortOrder,
            @Param("direction") String direction
    );

    @Select("SELECT count(*) FROM t_task_scan WHERE ds_id = #{dsId} and scan_status = #{scanStatus}")
    int getTaskCountByDsIdAndScanStatus(@Param("dsId") String dsId, @Param("scanStatus") int scanStatus);

    TaskScanEntity selectLastScheduleScanTask(@Param("scheduleId") String scheduleId);

    List<TaskScanEntity> selectScheduleScanExecList(String scheduleId, int limit, int offset);

    @Select("SELECT count(*) FROM t_task_scan WHERE  schedule_id = #{scheduleId}")
    int selectScheduleScanExecCount(String scheduleId);

    @Select("SELECT * FROM t_task_scan WHERE schedule_id = #{jobId} limit 2")
    List<TaskScanEntity> getWaitTaskByName(String jobId);
}
