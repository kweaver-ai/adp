package com.eisoo.engine.utils.common;

import com.eisoo.engine.utils.enums.ErrorCodeEnum;

public class Message {

    public final static String MESSAGE_INPUT_NOT_EMPTY = "输入不能为空";

    public final static String MESSAGE_AUTH_SERVICE_ERROR_SOLUTION = "请检查权限微服务(auth-service)状态是否正常。具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_PARAM_ERROR_SOLUTION = "请使用请求参数构造规范化的请求字符串，详细信息参见产品 API 文档。";

    public final static String MESSAGE_DATANOTEXIST_ERROR_SOLUTION = "请检查参数是否正确或对应数据是否存在，详细信息参见产品 API 文档。";

    public final static String MESSAGE_PATHVARIABLE_ERROR_SOLUTION = "请使用正确的占位符参数构造URL，详细信息参见产品 API 文档。";

    public final static String MESSAGE_MaxUploadSizeExceededException = "上传的文件大小超出了有效值，详细信息参见产品 API 文档。";

    public final static String MESSAGE_MultipartException = "未上传文件或上传文件异常，详细信息参见产品 API 文档。";

    public final static String MESSAGE_REQUEST_METHOD_UNSUPPORTED_SOLUTION = "当前请求方式为：[%s]，支持的接口请求方式有：[%s]，请使用接口支持的请求方式重试，详细信息参见产品 API 文档。";

    public final static String MESSAGE_Duplicated_SOLUTION = "数据重复，请检查数据或重试，重试再次失败，请联系技术人员进行检查处理。";

    public final static String MESSAGE_CATALOG_MONGO_SEEDS_RULE_SOLUTION = "请检查mongodb.seeds参数格式，正确格式为ip:port，多个地址逗号隔开，详细信息参见产品 API 文档。";

    public final static String MESSAGE_CATALOG_TYPES_SOLUTION = "支持的数据源connectorNames:%s，详细信息参见产品 API 文档。";

    public final static String MESSAGE_CATALOG_NAME_SOLUTION = "catalogName规则：由小写字母，数字或者下划线组成的字符串，只能包含小写字母，数字和下划线，且不能以数字开头，且catalogName长度不能超过100，详细信息参见产品 API 文档。";

    public final static String MESSAGE_CATALOG_MONGO_REPLICA_SET_SOLUTION = "副本集模式下填写该参数，其他部署模式不需要该参数，详细信息参见产品 API 文档。";

    public final static String MESSAGE_CATALOG_MONGO_CREDENTIALS_SOLUTION = "请检查mongodb.credentials参数格式，正确格式为username:password@database，详细信息参见产品 API 文档。";

    public final static String MESSAGE_ENUMS_SOLUTION = "支持的枚举值为：%s，详细信息参见产品 API 文档。";

    public final static String MESSAGE_QUERY_TYPE_SOLUTION = "目前仅支持type=0,随机取样，详细信息参见产品 API 文档。";

    public final static String MESSAGE_QUERY_MISSING_ATTRIBUTE_SOLUTION = "检查查询语句中的字段和表信息，是否存在拼写错误或不存在的字段或表，详细信息参见产品 API 文档。";

    public final static String MESSAGE_QUERY_VIEW_STALE_SOLUTION = "检查字段数量及字段名称，可重新生成视图，详细信息参见产品 API 文档。";

    public final static String MESSAGE_QUERY_SQL_SOLUTION = "请检查sql语句，详细信息参见产品 API 文档。";

    public final static String MESSAGE_TABLE_NAME_SOLUTION = "tableName规则：长度不能超过128,只能包含字母，数字和下划线，且必须以字母开头，详细信息参见产品 API 文档。";

    public final static String MESSAGE_TYPE_LENGTH_SOLUTION = "请检查类型长度是否存在问题，详细信息参见产品 API 文档。";

    public final static String MESSAGE_METADATA_COLLECTION_LOG_SOLUTION = "请查看数据虚拟化引擎计算微服务(vega-calculate-coordinator, vega-calculate-worker)元数据采集日志，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_VIEW_NAME_SOLUTION = "规则：由小写字母，数字，下划线，中文字符以及空格组成的字符串，只能包含小写字母，数字，下划线，中文字符或空格，且长度不能超过100，详细信息参见产品 API 文档。";

    public final static String MESSAGE_COLUMN_NAME_SOLUTION = "规则：不能使用 \\ / : * ? \" < > |，且不能使用大写字母，且长度不能超过100，详细信息参见产品 API 文档。";

    public final static String MESSAGE_INTERNAL_ERROR = "未具体分类的内部错误，请联系技术支持。";

    public final static String MESSAGE_SYSTEM_BUSY = "请稍后重试。如果重试后仍无法解决您的问题，请联系技术支持。";

    public final static String MESSAGE_DATABASE_ERROR_SOLUTION = "请检查数据库服务状态和相关日志，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_OPENLOOKENG_ERROR = "请检查数据虚拟化引擎计算微服务(vega-calculate-coordinator, vega-calculate-worker)配置及状态是否正常，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_WORKER_ERROR = "请检查数据虚拟化引擎计算微服务工作节点(vega-calculate-worker)状态是否正常，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_SCHEDULER_SERVICE_SOLUTION = "请检查数据虚拟化引擎调度微服务(vega-scheduler)配置及状态是否正常，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_CELL_RANGE_SOLUTION = "行数最大限制为1048576，列数最大限制为16384，详细信息参见产品 API 文档。";

    public final static String EXCEL_FILENAME_DECODE_SOLUTION = "文件名参数编码格式为UTF-8，详细信息参见产品 API 文档。";

    public final static String MESSAGE_COLUMN_LENGTH_SOLUTION = "列名长度不超过64，详细信息参见产品 API 文档。";

    public final static String MESSAGE_COLUMN_TYPE_SOLUTION = "字段类型仅支持bigint,double,varchar,boolean,timestamp，详细信息参见产品 API 文档。";

    public final static String MESSAGE_JSON_SOLUTION = "请检查报文参数，详细信息参见产品 API 文档。";
    public final static String MESSAGE_OPENLOOKENG_INIT_ERROR = "请检查数据虚拟化引擎计算微服务(vega-calculate-coordinator, vega-calculate-worker)是否启动完成，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public final static String MESSAGE_QUERY_SOLUTION = "请重新发起SQL请求，详细信息参见产品 API 文档。";

    public final static String MESSAGE_ETRINO_SOLUTION = "请检查ETrino模块是否安装且服务运行正常，具体操作可参考：《数据虚拟化引擎故障排除指南》。";

    public static String getMessageDetail(ErrorCodeEnum errorCodeEnum) {
        switch (errorCodeEnum) {
            case AuthTokenForbiddenError:
                return "鉴权不通过，禁止访问";
            case AuthTokenUnAuthRIZEDError:
                return "请求头中缺少Authorized";
            case AuthTokenRequestFormatError:
                return "请求头Authorization格式错误";
            default:
                return errorCodeEnum.getDescription();
        }
    }
    public static String getMessageSolutions(ErrorCodeEnum errorCodeEnum) {
        switch (errorCodeEnum) {
            case AuthTokenForbiddenError:
                return "token不合法，请更换token";
            case AuthTokenUnAuthRIZEDError:
                return "请添加鉴权头，如：Authorization：xxx";
            case AuthTokenRequestFormatError:
                return "请修改鉴权格式，如：Authorization：Bearer xxx";
            default:
                return errorCodeEnum.getDescription();
        }
    }
}
