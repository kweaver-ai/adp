package com.eisoo.dc.common.metadata.mapper;

import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.github.yulichang.base.MPJBaseMapper;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Update;

import java.util.Date;
import java.util.List;
import java.util.Set;

/**
 * @author Tian.lan
 */
public interface TableScanMapper extends MPJBaseMapper<TableScanEntity> {
    List<TableScanEntity> selectByDsId(@Param("id") String id);

    @Update("UPDATE  t_table_scan SET f_status=#{status} WHERE f_id = #{id}")
    int updateScanStatusById(@Param("id") String id, @Param("status") int status);

    @Update("UPDATE  t_table_scan SET f_status=#{status},f_task_id=#{taskId},f_operation_time=NOW() WHERE f_id = #{id}")
    int updateScanStatusAndOperationTimeById(@Param("id") String id, @Param("taskId") String taskId, @Param("status") int status);

    List<TableScanEntity> getTableListByDsId(@Param("dsId") String dsId,
                                             @Param("keyword") String keyword,
                                             @Param("offset") int offset,
                                             @Param("limit") int limit,
                                             @Param("sortOrder") String sortOrder,
                                             @Param("direction") String direction);

    List<TableScanEntity> selectPage(
            @Param("includeIds") Set<String> includeIds,
            @Param("keyword") String keyword,
            @Param("offset") int offset,
            @Param("limit") int limit,
            @Param("sortOrder") String sortOrder,
            @Param("direction") String direction
    );

    long selectCount(String dsId, String keyword);

    @Update("UPDATE t_table_scan SET f_operation_type=1, f_operation_time=#{operationTime} WHERE f_id = #{id}")
    int deleteById(@Param("id") String id, @Param("operationTime") Date operationTime);

    @Update("UPDATE t_table_scan SET f_operation_type=1, f_operation_time=#{operationTime} WHERE f_data_source_id = #{dataSourceId}")
    int deleteByDataSourceId(@Param("dataSourceId") String dataSourceId, @Param("operationTime") Date operationTime);

    List<TableScanEntity> getTableListByDsIdsBatch(List<String> dsIds, String updateTime, String keyword, int offset, int limit, String sortOrder, String direction);

    long selectCountByDsIdsBatch(List<String> dsIds, String updateTime, String keyword);

    List<TableScanEntity> selectPageBatch(Set<String> includeIds, String keyword, int offset, int limit, String sortOrder, String direction);

    List<TableScanEntity> selectPageBatchIds(Set<String> includeIds);

    @Delete("DELETE FROM t_table_scan WHERE f_data_source_id = #{dsId}")
    int deleteBysId(String dsId);
}
