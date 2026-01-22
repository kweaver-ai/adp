package com.eisoo.dc.common.metadata.entity;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;
import java.util.Date;
import java.util.List;

/**
 * @author Tian.lan
 */
@Data
@TableName("t_task_scan")
@AllArgsConstructor
@NoArgsConstructor
public class TaskScanEntity implements Serializable {
    /**
     * 扫描任务唯一id
     */
    @TableId(value = "id")
    private String id;
    /**
     * 扫描任务:0 :数据源1 :及时2: 定时-数据源3: 定时-及时
     */
    @TableField(value = "type")
    private Integer type = 0;
    /**
     * 任务名称
     */
    @TableField(value = "name")
    private String name;
//    /**
//     * 数据源是否最新一次扫描：0是1不是
//     */
//    @TableField(value = "ds_scan_latest")
//    private Integer dsScanLatest = 0;
    /**
     * 数据源id
     */
    @TableField(value = "ds_id")
    private String dsId;
    /**
     * 扫描任务:0 wait 1 running 2 success 3 fail
     */
    @TableField(value = "scan_status")
    private Integer scanStatus = 0;
    /**
     * 创建时间
     */
    @TableField(value = "start_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date startTime;
    /**
     * 更新时间
     */
    @TableField(value = "end_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date endTime;
    /**
     * 创建用户（ID），默认空字符串
     */
    @TableField(value = "create_user")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private String createUser;

    /**
     * 任务执行参数信息
     */
    @TableField(value = "task_params_info")
    private String taskParamsInfo;
    /**
     * 任务执行参数信息
     */
    @TableField(value = "task_process_info")
    private String taskProcessInfo;
    //
    /**
     * 任务执行结果信息
     */
    @TableField(value = "task_result_info")
    private String taskResultInfo;

    /**
     * 定时任务id
     */
    @TableField(value = "schedule_id")
    private String scheduleId;

    @Data
    @AllArgsConstructor
    public static class TaskParamsInfo {
        @JsonProperty(required = true, value = "cron_expression")
        private String cronExpression;
        @JsonProperty(required = true, value = "tables_count")
        private Integer tablesCount = 0;
        @JsonProperty(required = true, value = "tables")
        private List<String> tables;
        @JsonProperty(required = true, value = "scan_strategy")
        private List<String> scanStrategy;
    }
}
