package com.eisoo.dc.metadata.domain.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import io.swagger.annotations.ApiModel;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.io.Serializable;
import java.util.Date;

/**
 * @author Tian.lan
 */
@Data
@JsonInclude(JsonInclude.Include.ALWAYS)
@ApiModel
@AllArgsConstructor
public class TaskScanDto implements Serializable {
    @JsonProperty(value = "id")
    private String id;
    @JsonProperty("schedule_id")
    private String scheduleId;
    @JsonProperty(value = "type")
    private Integer type;
    @JsonProperty(value = "name")
    private String name;
    @JsonProperty(value = "ds_type")
    private String dsType;
    @JsonProperty(value = "allow_multi_table_scan")
    private Boolean allowMultiTableScan;
    @JsonProperty(value = "create_user")
    private String createUser;
    @JsonProperty(value = "scan_status")
    private String scanStatus;
    @JsonProperty(value = "task_status")
    private String taskStatus;
    @JsonProperty(value = "start_time")
    @JsonFormat(timezone = "GMT+8", pattern = "yyyy-MM-dd HH:mm:ss")
    private Date startTime;
    @JsonProperty(value = "task_process_info")
    private String taskProcessInfo;
    @JsonProperty(value = "task_result_info")
    private String taskResultInfo;

}
