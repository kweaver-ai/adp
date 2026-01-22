package com.eisoo.dc.metadata.domain.dto;

import com.eisoo.dc.metadata.domain.vo.TaskScanVO;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * @author Tian.lan
 */
@Data
@AllArgsConstructor
@NoArgsConstructor
@JsonInclude(JsonInclude.Include.ALWAYS)
public class ScheduleTaskStatusInfoDto {
    @JsonProperty(value = "last_scan_task_id", required = true)
    private String lastScanTaskId;
    @JsonProperty(value = "scan_strategy", required = true)
    private List<String> scanStrategy;
    @JsonProperty(value = "cron_expression", required = true)
    private TaskScanVO.CronExpressionObj cronExpression;
    @JsonProperty(value = "task_status", required = true)
    private String taskStatus;
    @JsonProperty(value = "scan_status", required = true)
    private String scanStatus;
    private String duration;
    @JsonProperty(value = "start_time")
    private String startTime;
    @JsonProperty(value = "end_time")
    private String endTime;
}
