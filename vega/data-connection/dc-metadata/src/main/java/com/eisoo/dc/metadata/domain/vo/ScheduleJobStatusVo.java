package com.eisoo.dc.metadata.domain.vo;


import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.deserializer.StringDeserializer;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Pattern;

/**
 * @author Tian.lan
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class ScheduleJobStatusVo {
    @NotBlank(message = "定时扫描任务Id" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @JsonProperty("schedule_id")
    private String scheduleId;
    @Pattern(regexp = "open|close", message = "open|close")
    @JsonDeserialize(using = StringDeserializer.class)
    private String status;
}
