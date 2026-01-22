package com.eisoo.dc.common.metadata.entity;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.Data;

import java.io.Serializable;

/**
 * @author Tian.lan
 */
@TableName("t_table_scan")
@Data
public class TableScanEntity implements Serializable {
    /**
     * 表唯一id
     */
    @TableId(value = "f_id")
    private String fId;
    /**
     * 表名称
     */
    @TableField(value = "f_name")
    private String fName;
    /**
     * 高级参数
     */
    @TableField(value = "f_advanced_params")
    private String fAdvancedParams;

    /**
     * 表注释
     */
    @TableField(value = "f_description")
    private String fDescription;
    /**
     * 表数据量，默认0
     */
    @TableField(value = "f_table_rows")
    private Integer fTableRows = 0;
    /**
     * 数据源唯一标识
     */
    @TableField(value = "f_data_source_id")
    private String fDataSourceId;
    /**
     * 数据源名称
     */
    @TableField(value = "f_data_source_name")
    private String fDataSourceName;
    /**
     * schema名称
     */
    @TableField(value = "f_schema_name")
    private String fSchemaName;
    /**
     * task唯一标识
     */
    @TableField(value = "f_task_id")
    private String fTaskId;
    /**
     * version
     */
    @TableField(value = "f_version")
    private Integer fVersion;
    /**
     * 创建时间
     */
    @TableField(value = "f_create_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private String fCreateTime;
    /**
     * 创建用户
     */
    @TableField(value = "f_create_user")
    private String fCreatUser;
    /**
     * 修改时间
     */
    @TableField(value = "f_operation_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private String fOperationTime;
    /**
     * 修改用户
     */
    @TableField(value = "f_operation_user")
    private String fOperationUser;

    /**
     * 状态：0新增1删除2更新
     */
    @TableField(value = "f_operation_type")
    private Integer fOperationType = 0;
    /**
     * 任务状态：0 wait 1 running 2 success 3 fail
     */
    @TableField(value = "f_status")
    private Integer fStatus = 0;
    /**
     * 状态是否发生变化：0 否1 是
     */
    @TableField(value = "f_status_change")
    private Integer fStatusChange = 0;

    /**
     * 扫描来源
     */
    @TableField(value = "f_scan_source")
    private Integer fScanSource;

}
