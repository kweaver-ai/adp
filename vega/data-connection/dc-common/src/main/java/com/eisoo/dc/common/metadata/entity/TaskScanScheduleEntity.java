package com.eisoo.dc.common.metadata.entity;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;
import java.util.Date;

/**
 * @author Tian.lan
 */
@Data
@TableName("t_task_scan_schedule")
@AllArgsConstructor
@NoArgsConstructor
public class TaskScanScheduleEntity implements Serializable {
    /**
     * 定时任务唯一id
     */
    @TableId(value = "id")
    private String id;
    /**
     * 扫描任务:3: 定时-数据源；4: 定时快速-数据源
     */
    @TableField(value = "type")
    private Integer type;
    /**
     * 任务名称
     */
    @TableField(value = "name")
    private String name;
    /**
     * 数据源id
     */
    @TableField(value = "ds_id")
    private String dsId;
    /**
     * 任务cron表达式
     */
    @TableField(value = "cron_expression")
    private String cronExpression;
    /**
     * 扫描策略
     */
    @TableField(value = "scan_strategy")
    private String scanStrategy;
    /**
     * 扫描任务:0 暂停 -1 删除 1 启用
     */
    @TableField(value = "task_status")
    private Integer taskStatus = 0;
    /**
     * 创建时间
     */
    @TableField(value = "create_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date createTime;
    /**
     * 创建用户（ID），默认空字符串
     */
    @TableField(value = "create_user")
    private String createUser;
    /**
     * 修改时间
     */
    @TableField(value = "operation_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date operationTime;
    /**
     * 修改用户
     */
    @TableField(value = "operation_user")
    private String operationUser;

    /**
     * 状态：0新增1删除2更新
     */
    @TableField(value = "operation_type")
    private Integer operationType = 0;
}
