package com.eisoo.dc.common.vo;

import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.deserializer.IntegerDeserializer;
import com.eisoo.dc.common.deserializer.StringDeserializer;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonNaming;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;

import javax.validation.constraints.Max;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Pattern;
import javax.validation.constraints.Size;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
@ApiModel
public class BinDataVo {

    @ApiModelProperty(value = "数据库名称", example = "af_openlookeng", dataType = "java.lang.String")
    @JsonProperty("database_name")
    @Size(max = 100, message = "长度最大为100个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String databaseName;

    @ApiModelProperty(value = "数据库名称", example = "public", dataType = "java.lang.String")
    @Size(max = 100, message = "长度最大为100个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String schema;

    @ApiModelProperty(value = "连接方式", dataType = "java.lang.String")
    @NotBlank(message = "连接方式" + Message.MESSAGE_INPUT_NOT_EMPTY)
    @Pattern(regexp = "https|http|thrift|jdbc", message = "可选参数值：https、http、thrift、jdbc")
    @JsonProperty("connect_protocol")
    @JsonDeserialize(using = StringDeserializer.class)
    private String connectProtocol;

    @ApiModelProperty(value = "连接地址", dataType = "java.lang.String")
    @JsonDeserialize(using = StringDeserializer.class)
    private String host;

    @ApiModelProperty(value = "端口号", example = "3306", dataType = "int")
    @Max(value = 65535)
    @JsonDeserialize(using = IntegerDeserializer.class)
    private int port;

    @ApiModelProperty(value = "用户名", example = "root", dataType = "java.lang.String")
    @Size(max = 128, message = "长度最大为128个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String account;

    @ApiModelProperty(value = "密码", example = "hGCcAVLScCkS/d2wNBq2G51VyjY0ho4GMPVafTTa8eftoDNmTmNV1D1loPU9DaLaNWzte2lMSiunycuilMaT7TOtinA6cAmZIbb4C4YZc0W5L+D+vZb910Hl6FTSGf2+PH55ccpLGvAegri5PuG02Hei0CyM7DL5o3k34jJPPQs=", dataType = "java.lang.String")
    @Size(max = 1024, message = "长度最大为1024个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String password;

    @ApiModelProperty(value = "token", example = "", dataType = "java.lang.String")
    @Size(max = 100, message = "长度最大为100个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String token;

    @ApiModelProperty(value = "存储介质", example = "anyshare", dataType = "java.lang.String")
    @JsonProperty("storage_protocol")
    @Size(max = 100, message = "长度最大为100个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String storageProtocol;

    @ApiModelProperty(value = "存储路径", example = "/opt/excel", dataType = "java.lang.String")
    @JsonProperty("storage_base")
    @Size(max = 1024, message = "长度最大为1024个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String storageBase;

    @ApiModelProperty(value = "副本集", example = "rs0", dataType = "java.lang.String")
    @JsonProperty("replica_set")
    @Size(max = 100, message = "长度最大为100个字符")
    @JsonDeserialize(using = StringDeserializer.class)
    private String replicaSet;
}
