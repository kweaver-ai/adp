package com.eisoo.dc.metadata.domain.vo;

import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.deserializer.IntegerDeserializer;
import com.eisoo.dc.common.deserializer.StringDeserializer;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.AssertTrue;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Pattern;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.List;

/**
 * @author Tian.lan
 */
@Data
@AllArgsConstructor
@NoArgsConstructor
@JsonInclude(JsonInclude.Include.NON_NULL)
@ApiModel
public class TaskScanVO implements Serializable {
    @ApiModelProperty(value = "扫描任务名称", example = "mysql_scan_1", dataType = "java.lang.String")
    @NotBlank(message = "扫描任务名称" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @Size(min = 1, max = 128, message = "扫描任务名称长度必须在1-128个字符之间")
    @JsonDeserialize(using = StringDeserializer.class)
    @JsonProperty("scan_name")
    private String scanName;
    @ApiModelProperty(value = "任务类型", example = "0", dataType = "java.lang.Integer")
    @JsonDeserialize(using = IntegerDeserializer.class)
    private Integer type;

    @JsonProperty("ds_info")
    private DsInfo dsInfo;

    @JsonProperty("use_default_template")
    @AssertTrue(message = "必须使用默认配置模板")
    private Boolean useDefaultTemplate = true;
    @JsonProperty("field_list_when_change")
    private List<String> fieldListWhenChange;

    @JsonProperty("use_multi_threads")
    @AssertTrue(message = "必须使用多线程采集")
    private Boolean useMultiThreads = true;


    @JsonProperty("cron_expression")
    private CronExpressionObj cronExpression;
    private List<String> tables;
    @Pattern(regexp = "open|close", message = "open|close")
    @NotBlank(message = "扫描任务状态" + Message.MESSAGE_INPUT_NOT_EMPTY)
    private String status;


    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class DsInfo {
        @JsonProperty("ds_id")
        private String dsId;
        @JsonProperty("ds_type")
        private String dsType;
        @JsonProperty(value = "scan_strategy")
        private List<String> scanStrategy;
    }

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class CronExpressionObj {
        @JsonProperty("type")
        private String type;
        @JsonProperty("expression")
        private String expression;
    }

}
