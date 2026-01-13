package com.eisoo.dc.common.vo;

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
public class TestDataSourceVo implements Serializable {

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

}
