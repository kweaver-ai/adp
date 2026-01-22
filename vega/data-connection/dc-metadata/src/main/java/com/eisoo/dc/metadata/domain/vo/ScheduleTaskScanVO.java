package com.eisoo.dc.metadata.domain.vo;

import com.eisoo.dc.common.deserializer.StringDeserializer;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.swagger.annotations.ApiModel;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.Pattern;
import java.util.List;

/**
 * @author Tian.lan
 */
@Data
@AllArgsConstructor
@NoArgsConstructor
@JsonInclude(JsonInclude.Include.NON_NULL)
@ApiModel
public class ScheduleTaskScanVO {
    @JsonProperty("schedule_id")
    private String scheduleId;
    @JsonProperty("cron_expression")
    private TaskScanVO.CronExpressionObj cronExpression;
    @JsonProperty(value = "scan_strategy")
    private List<String> scanStrategy;
    @Pattern(regexp = "open|close", message = "open|close")
    @JsonDeserialize(using = StringDeserializer.class)
    private String status;
}
