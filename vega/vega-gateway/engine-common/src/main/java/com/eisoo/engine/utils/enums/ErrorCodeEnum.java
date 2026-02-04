package com.eisoo.engine.utils.enums;

import lombok.Getter;
import org.springframework.http.HttpStatus;

@Getter
public enum ErrorCodeEnum {
    SUCCESS(HttpStatus.OK, "VirtualizationEngine.Success.", "成功。"),
    InvalidParameter(HttpStatus.BAD_REQUEST, "VirtualizationEngine.InvalidParameter.", "参数错误。"),
    MethodArgumentTypeMismatchException(HttpStatus.BAD_REQUEST, "VirtualizationEngine.MethodArgumentTypeMismatchException.", "参数值校验不通过。"),
    HttpRequestMethodNotSupportedException(HttpStatus.BAD_REQUEST, "VirtualizationEngine.HttpRequestMethodNotSupportedException.", "接口请求方式不支持。"),
    Duplicated(HttpStatus.BAD_REQUEST, "VirtualizationEngine.Duplicated.", "数据重复。"),
    DeadlockException(HttpStatus.BAD_REQUEST, "VirtualizationEngine.DeadlockError.", "数据表死锁异常。"),
    CatalogNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.CatalogNotExist.", "数据源不存在。"),
    SchemaNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.SchemaNotExist.", "Schema不存在。"),
    TableNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TableNotExist.", "Table不存在。"),
    ViewNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.ViewNotExist.", "视图不存在。"),
    CatalogExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.CatalogExist.", "数据源已存在。"),
    SchemaExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.SchemaExist.", "Schema已存在。"),
    TableExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TableExist.", "Table已存在。"),
    ViewExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.ViewExist.", "视图已存在。"),
    SqlSyntaxError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.SqlSyntaxError.", "Sql语句异常。"),
    NotFound(HttpStatus.BAD_REQUEST, "VirtualizationEngine.NotFound.", "访问的资源不存在。"),
    AuthServiceError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.AuthServiceError.", "数据虚拟化引擎获取行列权限失败。"),
    MetadataCollectError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.MetadataError.", "数据虚拟化引擎读取元数据失败。"),
    RuleError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.RuleError.", "探查规则异常。"),
    TableKeyError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TableKeyError.", "主键异常。"),
    TableFieldError(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TableFieldError.", "视图与源表的字段不一致。"),
    TaskNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TaskNotExist.", "任务不存在。"),
    TaskCompleted(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TaskCompleted.", "任务已经执行完成。"),
    TaskDeleted(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TaskDeleted.", "任务已删除。"),
    TaskCanceled(HttpStatus.BAD_REQUEST, "VirtualizationEngine.TaskCanceled.", "任务已取消。"),
    ReadExcelFail(HttpStatus.BAD_REQUEST,"VirtualizationEngine.ReadExcelFail", "读取Excel文件错误。"),
    SheetNotExist(HttpStatus.BAD_REQUEST,"VirtualizationEngine.SheetNotExist", "Sheet不存在。"),
    DBError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.DBError.", "数据库异常。"),
    VirEngineWorkError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.VirEngineWorkError.", "数据虚拟化引擎计算微服务工作节点异常。"),
    OpenLooKengError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.OpenLooKengError.", "数据虚拟化引擎微服务异常。"),
    OpenLooKengInitError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.OpenLooKengInitError.", "数据虚拟化引擎微服务初始化异常。"),
    DolphinSchedulerError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.DolphinSchedulerError.", "数据虚拟化引擎调度微服务异常。"),
    InternalError(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.InternalError.", "数据虚拟化引擎系统内部错误。"),
    SystemBusy(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.SystemBusy.", "数据虚拟化引擎系统繁忙。"),
    ConnectFail(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.ConnectFail", "连接失败。"),
    VirEngineNotSupportError(HttpStatus.INTERNAL_SERVER_ERROR,"VirtualizationEngine.VirEngineNotSupportError", "引擎不支持此操作。"),
    AuthTokenForbiddenError(HttpStatus.FORBIDDEN,"VirtualizationEngine.FORBIDDEN", "权限校验失败，禁止访问"),
    AuthTokenUnAuthRIZEDError(HttpStatus.UNAUTHORIZED,"VirtualizationEngine.NotFound", "请求头中缺少Authorization。"),
    AuthTokenRequestFormatError(HttpStatus.BAD_REQUEST,"VirtualizationEngine.FormatError", "请求头中鉴权串格式错误;如：Authorization: Bearer xxxx。"),
    ConnectAuthServerFail(HttpStatus.INTERNAL_SERVER_ERROR, "VirtualizationEngine.ConnectAuthServerFail", "连接hydra服务失败。"),
    RunningError(HttpStatus.BAD_REQUEST,"VirtualizationEngine.RunningError", "任务不能删除。"),
    DataSourceNotExist(HttpStatus.BAD_REQUEST, "VirtualizationEngine.DataSourceNotExist.", "数据源不存在。"),
    DataSourceNameExist(HttpStatus.BAD_REQUEST,"VirtualizationEngine.DataSourceNameExist","数据源名称存在。"),
    AbandonedQuery(HttpStatus.BAD_REQUEST,"Public.BadRequest","查询已失效。"),
    MissingAttributeError(HttpStatus.BAD_REQUEST, "Public.BadRequest", "查询中表或字段不存在。"),
    ETrinoError(HttpStatus.INTERNAL_SERVER_ERROR, "Public.InternalServerError", "ETrino服务访问异常。");



    /**
     * http状态码
     */
    private HttpStatus httpStatus;

    /**
     * 错误码
     */
    private String errorCode;

    /**
     * 错误描述
     */
    private String description;

    /**
     * 错误细节
     */
    private Object errorDetails;

    /**
     * 错误处理建议
     */
    private String solution;

    private ErrorCodeEnum(HttpStatus httpStatus, String errorCode, String description) {
        this.httpStatus = httpStatus;
        this.errorCode = errorCode;
        this.description = description;
    }

    private ErrorCodeEnum(HttpStatus httpStatus, String errorCode, String description, Object errorDetails) {
        this.httpStatus = httpStatus;
        this.errorCode = errorCode;
        this.description = description;
        this.errorDetails = errorDetails;
    }

    private ErrorCodeEnum(HttpStatus httpStatus, String errorCode, String description, Object errorDetails, String solution) {
        this.httpStatus = httpStatus;
        this.errorCode = errorCode;
        this.description = description;
        this.errorDetails = errorDetails;
        this.solution = solution;
    }

}
