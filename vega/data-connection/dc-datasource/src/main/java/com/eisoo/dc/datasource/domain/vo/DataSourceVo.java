package com.eisoo.dc.datasource.domain.vo;

import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.deserializer.StringDeserializer;
import com.eisoo.dc.common.vo.BinDataVo;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;

import javax.validation.Valid;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Pattern;
import javax.validation.constraints.Size;
import java.io.Serializable;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@ApiModel
public class DataSourceVo implements Serializable {
    @ApiModelProperty(value = "数据源名称", example = "mysql_133_144_1", dataType = "java.lang.String")
    @NotBlank(message = "数据源名称" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @Size(min = 1, max = 128, message = "数据源名称长度必须在1-128个字符之间")
    @Pattern(regexp = "^[a-zA-Z0-9\u4e00-\u9fa5-_]+$", message = "数据源名称仅支持中英文、数字、下划线和中划线")
    @JsonDeserialize(using = StringDeserializer.class)
    private String name;

    @ApiModelProperty(value = "类型", example = "maria", dataType = "java.lang.String")
    @NotBlank(message = "类型" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @Size(max = 30, message = "长度最大为30个字符")
    @Pattern(regexp = "oracle|postgresql|doris|sqlserver|hive|clickhouse|mysql|maria|mongodb|dameng|hologres|gaussdb|excel|opengauss|inceptor-jdbc|tingyun|anyshare7|maxcompute|opensearch", message = "支持传参：oracle, postgresql, doris, sqlserver, hive, clickhouse, mysql, maria, mongodb, dameng, hologres, gaussdb, excel, opengauss, inceptor-jdbc, tingyun, anyshare7, maxcompute, opensearch")
    @JsonDeserialize(using = StringDeserializer.class)
    private String type;

    @JsonProperty("bin_data")
    @NotNull(message = "数据源配置" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @Valid
    private BinDataVo binData;

    @ApiModelProperty(value = "描述", example = "信息项", dataType = "java.lang.String")
    @Size(max = 255, message = "长度最大为255个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String comment;

}
