package com.eisoo.dc.metadata.domain.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * @author Tian.lan
 */
@Data
@AllArgsConstructor
@NoArgsConstructor
@JsonInclude(JsonInclude.Include.ALWAYS)
public class ScheduleTaskInfoDto {
    @JsonProperty(value = "task_id", required = true)
    private String taskId;
    @JsonProperty(value = "scan_status", required = true)
    private String scanStatus;
    private String duration;
    @JsonProperty(value = "start_time")
    private String startTime;
    @JsonProperty(value = "end_time")
    private String endTime;
    @JsonProperty(value = "task_process_info")
    private String taskProcessInfo;
    @JsonProperty(value = "task_result_info")
    private String taskResultInfo;
}
