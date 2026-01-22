package com.eisoo.dc.metadata.domain.dto;

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
public class TaskStatusInfoDto {
    private String id;
    private String status;
    @JsonProperty(value = "data", required = true)
    private TaskData data;


    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class TaskData {
        @JsonProperty(value = "task_process_info", required = true)
        private TaskProcessInfo taskProcessInfo;
        @JsonProperty(value = "task_result_info", required = true)
        private TaskResultInfo taskResultInfo;
    }

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class TaskProcessInfo {
        @JsonProperty(value = "table_count", required = true)
        private Integer tableCount = 0;
        @JsonProperty(value = "success_count", required = true)
        private Integer successCount = 0;
        @JsonProperty(value = "fail_count", required = true)
        private Integer failCount = 0;
    }

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class TaskResultInfo {
        @JsonProperty(value = "table_count", required = true)
        private Integer tableCount = 0;
        @JsonProperty(value = "success_count", required = true)
        private Integer successCount = 0;
        @JsonProperty(value = "fail_count", required = true)
        private Integer failCount = 0;
        @JsonProperty(value = "fail_stage", required = true)
        private Integer failStage;
        @JsonProperty(value = "error_stack", required = true)
        private String errorStack;
    }
}
