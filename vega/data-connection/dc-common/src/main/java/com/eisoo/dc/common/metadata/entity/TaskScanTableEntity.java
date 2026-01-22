package com.eisoo.dc.common.metadata.entity;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.io.Serializable;
import java.util.Date;

/**
 * @author Tian.lan
 */
@Data
@TableName("t_task_scan_table")
@AllArgsConstructor
public class TaskScanTableEntity implements Serializable {
    /**
     * 扫描任务唯一id
     */
    @TableId(value = "id")
    private String id;
    /**
     * 关联任务id
     */
    @TableField(value = "task_id")
    private String taskId;
    /**
     * 数据源唯一标识
     */
    @TableField(value = "ds_id")
    private String dsId;
    /**
     * 数据源名称
     */
    @TableField(value = "ds_name")
    private String dsName;
    /**
     * table的唯一id
     */
    @TableField(value = "table_id")
    private String tableId;
    /**
     * table的name
     */
    @TableField(value = "table_name")
    private String tableName;
    /**
     * schema的name
     */
    @TableField(value = "schema_name")
    private String schemaName;
    /**
     * 任务状态：0 成功;1 失败;2 进行中;3 初始化;4 等待
     */
    @TableField(value = "scan_status")
    private Integer scanStatus = 0;
    /**
     * 任务开始时间
     */
    @TableField(value = "start_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date startTime;

    /**
     * 任务结束时间
     */
    @TableField(value = "end_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date endTime;
    /**
     * 创建用户
     */
    @TableField(value = "create_user")
    private String creatUser;
    /**
     * 任务执行参数信息
     */
    @TableField(value = "scan_params")
    private String scanParams;
    /**
     * 任务执行结果
     */
    @TableField(value = "scan_result_info")
    private String scanResultInfo;
    /**
     * 异常堆栈信息
     */
    @TableField(value = "error_stack")
    private String errorStack;
    /**
     * 操作类型：0 insert;1 delete;2 update;3 unknown
     */
    @TableField(value = "operation_type")
    private Integer operationType;

}
