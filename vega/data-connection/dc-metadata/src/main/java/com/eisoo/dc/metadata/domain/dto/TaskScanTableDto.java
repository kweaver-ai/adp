package com.eisoo.dc.metadata.domain.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import io.swagger.annotations.ApiModel;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;
import java.util.Date;

/**
 * @author Tian.lan
 */
@Data
@JsonInclude(JsonInclude.Include.ALWAYS)
@ApiModel
@AllArgsConstructor
@NoArgsConstructor
public class TaskScanTableDto implements Serializable {
    @JsonProperty(value = "task_id")
    private String taskId;
    @JsonProperty(value = "table_id")
    private String tableId;
    @JsonProperty(value = "table_name")
    private String tableName;
    @JsonProperty(value = "scan_status")
    private String scanStatus;
    @JsonProperty(value = "start_time")
    private Date startTime;
    @JsonProperty(value = "end_time")
    private Date endTime;
}
