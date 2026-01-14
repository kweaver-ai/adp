package com.eisoo.dc.common.constant;

public class Detail {
    public final static String DB_ERROR = "访问数据库发生异常。";
    public final static String BUILT_IN_CATALOG_DEL_UNSUPPORTED = "内置数据源不允许删除。";
    public final static String FIELD_ERROR = "%s的值[%s]不符合接口要求。";
    public final static String PASSWORD_ERROR = "密码解密错误。";
    public final static String URL_ERROR = "URL解析异常:%s。";
    public final static String CATALOG_TYPE_ERROR = "数据源[%s]存在但数据源类型不一致无法更新。";
    public final static String USERNAME_OR_PASSWORD_ERROR = "用户名或密码错误。";
    public final static String USERNAME_OR_PASSWORD_NOT_EMPLOY = "用户名或密码不能为空。";
    public final static String SCHEMA_NOT_NULL = "schema不能为空";
    public final static String ID_NOT_EXISTS = "数据源ID不存在。";
    public final static String CREATE_DATASOURCE_FAILED = "创建数据源失败。";
    public final static String DATASOURCE_NAME_EXIST = "数据源名称已存在:%s。";
    public final static String DATABASE_NAME_NOT_EMPLOY = "数据库名称不能为空。";
    public final static String DATASOURCE_AUTH_NOT_EMPLOY = "数据源身份认证方式必须为用户名密码认证和Token认证其中一种。";
    public final static String EXCEL_BASE_AND_PROTOCOL_NOT_EMPLOY = "excel存储介质及存储路径不能为空。";
    public final static String EXCEL_PROTOCOL_ILLEGAL = "excel存储介质不合法。";
    public final static String BASE_NOT_EMPLOY = "存储路径不能为空。";
    public final static String CONNECTOR_PROTOCOL_UNSUPPORTED = "连接方式不支持。";
    public final static String STORAGE_PROTOCOL_EDIT_ERROR = "数据源存储介质不能修改。";
    public final static String CONNECT_PROTOCOL_EDIT_ERROR = "数据源连接方式不能修改。";
    public final static String RESOURCE_PERMISSION_ERROR = "缺少资源权限[%s]。";
    public final static String AUTHORIZATION_FORMAT_ERROR = "请求头中Authorization鉴权串格式错误，如：Authorization: Bearer xxxx。";
    public final static String AUTHORIZATION_NOT_EXIST = "请求头中Authorization鉴权串不存在。";
    public final static String DATASOURCE_SCAN_IS_RUNNING = "整个数据源正在扫描:%s。";
    public final static String META_DATA_SCAN_NOT_FOUND = "元数据扫描任务不存在:%s。";
    public final static String META_DATA_SCAN_UNSUPPORTED = "元数据扫描任务不支持:%s。";
    public final static String META_DATA_SCAN__TASK_UNSUPPORTED = "元数据扫描任务不支持:type=%s。";
    public final static String DS_NOT_EXIST = "数据源不存在:%s。";

    public final static String DS_NOT_EXIST_TABLE = "table:%s 所属数据源不存在:%s。";
    public final static String TABLE_NOT_EXIST_FIELD = "tableId:%s; tableName:%s的列不存在";

    public final static String TABLE_NOT_EXIST = "数据表不存在:%s。";
    public final static String TASK_NOT_EXIST = "任务不存在:%s。";

    public final static String DS_NOT_UNSUPPORTED = "数据源类型不支持:%s。";

    public final static String TABLES_NOT_EMPLOY = "table不能为空。";

    public final static String META_DATA_QUERY_UNSUPPORTED = "数据查询类型不支持:%s。";
    public final static String SOURCE_TYPE_OR_INDEX_NULL = "原始数据类型sourceTypeName为空或类型索引index为空。";
    public final static String RUNNING_SCAN_TASK_EXIST = "数据源[%s]存在正在执行的扫描任务。";
    public final static String HOST_NOT_EMPLOY = "连接地址不能为空。";
    public final static String PORT_NOT_EMPLOY = "连接端口不能为空。";
    public final static String BUILT_IN_DATASOURCE_CANNOT_MODIFY = "内置数据源[%s]不能修改。";
    public final static String BUILT_IN_DATASOURCE_CANNOT_DELETE = "内置数据源[%s]不能删除。";
    public static final String NOT_SUPPORT_CONNECTOR = "不支持的数据源类型:%s。若要支持请安装Etrino可选包";
}
