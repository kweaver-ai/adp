package com.eisoo.dc.common.metadata.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.eisoo.dc.common.metadata.entity.TableOldEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Update;

import java.util.Date;
import java.util.List;
import java.util.Set;

public interface TableOldMapper extends BaseMapper<TableOldEntity> {

    List<TableOldEntity> getTableListByDsId(@Param("dsId") String dsId,
                                            @Param("keyword") String keyword,
                                            @Param("offset") int offset,
                                            @Param("limit") int limit,
                                            @Param("sortOrder") String sortOrder,
                                            @Param("direction") String direction);

    long selectCount(String dsId, String keyword);

    List<TableOldEntity> selectPage(
            @Param("includeIds") Set<Long> includeIds,
            @Param("keyword") String keyword,
            @Param("offset") int offset,
            @Param("limit") int limit,
            @Param("sortOrder") String sortOrder,
            @Param("direction") String direction
    );

    @Update("UPDATE t_table SET f_delete_flag=1, f_delete_time=#{operationTime} WHERE f_data_source_id = #{dataSourceId}")
    int deleteByDataSourceId(@Param("dataSourceId") String dataSourceId, @Param("operationTime") Date operationTime);
    @Delete("DELETE FROM t_table WHERE f_data_source_id = #{dsId}")
    int deleteBysId(String dsId);
}
