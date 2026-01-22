package com.eisoo.dc.common.util.jdbc.db;


import com.eisoo.dc.common.constant.CatalogConstant;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.util.jdbc.db.impl.*;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

/**
 * 数据库连接策略工厂：根据数据源类型匹配对应的连接策略
 *
 * @author Tian.lan
 */
public class DbConnectionStrategyFactory {
    // 缓存策略对象，避免重复创建
    public static Set<String> SUPPORT_NEW_SCAN = new HashSet<>();
    public static final Map<String, DbClientInterface> DB_CLIENT_MAP = new HashMap<>();
    public static final Map<String, String> DRIVER_CLASS_MAP = new HashMap<>();
    public static final Map<String, String> TABLE_METADATA_SQL_TEMPLATE_MAP = new HashMap<>();

    public static final String INCEPTOR_JDBC_DB = "SELECT * FROM (SELECT database_name,table_name, commentstring as remarks, \"table\" as table_type\n" +
            "               FROM `system`.`tables_v` tv\n" +
            "               union all\n" +
            "               SELECT database_name,view_name as table_name, \"\" as remarks, \"view\" as table_type\n" +
            "               FROM `system`.`views_v` tv\n" +
            "               )t\n" +
            "         WHERE t.database_name = '%s'\n" +
            "ORDER BY table_type,table_name LIMIT %d,1000";
    public static final String ORACLE_DB = "SELECT * \n" +
            "   FROM (\n" +
            "       SELECT t.*, ROWNUM rn \n" +
            "           FROM\n" +
            "(\n" +
            "   SELECT NULL AS table_cat,\n" +
            "       o.owner AS table_schem,\n" +
            "       o.object_name AS table_name,\n" +
            "       o.object_type AS table_type,\n" +
            "       NULL AS remarks,\n" +
            "       %d/1000 AS page_size" +
            "  FROM all_objects o\n" +
            "  WHERE o.owner='%s'\n" +
            "    AND o.object_type IN ('TABLE', 'VIEW')\n" +
            "  ORDER BY table_type, table_schem, table_name\n" +
            "  )t\n" +
            "    WHERE ROWNUM <= page_size * 1000\n" +
            ") \n" +
            "WHERE rn >= (page_size-1)*1000+1";
    public static final String MARIA_DB = "SELECT\n" +
            "\tTABLE_SCHEMA TABLE_CAT,\n" +
            "\tNULL TABLE_SCHEM,\n" +
            "\tTABLE_NAME,\n" +
            "\tIF(TABLE_TYPE = 'BASE TABLE'or TABLE_TYPE = 'SYSTEM VERSIONED','TABLE',IF(TABLE_TYPE = 'TEMPORARY','LOCAL TEMPORARY',TABLE_TYPE)) as TABLE_TYPE,\n" +
            "\tTABLE_COMMENT REMARKS,\n" +
            "\tNULL TYPE_CAT,\n" +
            "\tNULL TYPE_SCHEM,\n" +
            "\tNULL TYPE_NAME,\n" +
            "\tENGINE,TABLE_ROWS,CREATE_TIME,UPDATE_TIME,DATA_LENGTH/1024/1024 AS DATA_LENGTH,INDEX_LENGTH/1024/1024 AS INDEX_LENGTH,\n" +
            "\tNULL SELF_REFERENCING_COL_NAME,\n" +
            "\tNULL REF_GENERATION\n" +
            "FROM\n" +
            "\tINFORMATION_SCHEMA.TABLES\n" +
            "WHERE\n" +
            "\tTABLE_SCHEMA = '%s'\n" +
            "\tAND TABLE_TYPE IN ('BASE TABLE', 'SYSTEM VERSIONED','VIEW')\n" +
            "ORDER BY\n" +
            "\tTABLE_TYPE,\n" +
            "\tTABLE_SCHEMA,\n" +
            "\tTABLE_NAME LIMIT %d,1000";
    public static final String MYSQL_DB = "SELECT\n" +
            "\tTABLE_SCHEMA AS TABLE_CAT,\n" +
            "\tNULL AS TABLE_SCHEM,\n" +
            "\tTABLE_NAME,\n" +
            "\tCASE\n" +
            "\t\tWHEN TABLE_TYPE = 'BASE TABLE' THEN \n" +
            "\t\t\tCASE\n" +
            "\t\t\t\tWHEN TABLE_SCHEMA = 'mysql' OR TABLE_SCHEMA = 'performance_schema' THEN 'SYSTEM TABLE' ELSE 'TABLE'\n" +
            "\t\t\tEND\n" +
            "\t\tWHEN TABLE_TYPE = 'TEMPORARY' THEN 'LOCAL_TEMPORARY'\n" +
            "\t\tELSE TABLE_TYPE\n" +
            "\tEND AS TABLE_TYPE,\n" +
            "\tTABLE_COMMENT AS REMARKS,\n" +
            "\tNULL AS TYPE_CAT,\n" +
            "\tNULL AS TYPE_SCHEM,\n" +
            "\tNULL AS TYPE_NAME,\n" +
            "\tENGINE,TABLE_ROWS,CREATE_TIME,UPDATE_TIME,DATA_LENGTH/1024/1024 AS DATA_LENGTH,INDEX_LENGTH/1024/1024 AS INDEX_LENGTH,\n" +
            "\tNULL AS SELF_REFERENCING_COL_NAME,\n" +
            "\tNULL AS REF_GENERATION\n" +
            "FROM\n" +
            "\tINFORMATION_SCHEMA.TABLES\n" +
            "WHERE TABLE_SCHEMA='%s'\n" +
            "HAVING\n" +
            "\tTABLE_TYPE IN ('TABLE','VIEW','SYSTEM TABLE')\n" +
            "ORDER BY\n" +
            "\tTABLE_TYPE,\n" +
            "\tTABLE_SCHEMA,\n" +
            "\tTABLE_NAME LIMIT %d,1000";


    // 静态初始化：注册不同数据源的连接策略
    static {
        SUPPORT_NEW_SCAN.add(CatalogConstant.MARIA_CATALOG);
        SUPPORT_NEW_SCAN.add(CatalogConstant.MYSQL_CATALOG);
        SUPPORT_NEW_SCAN.add(CatalogConstant.ORACLE_CATALOG);
        SUPPORT_NEW_SCAN.add(CatalogConstant.OPENSEARCH_CATALOG);
        SUPPORT_NEW_SCAN.add(CatalogConstant.INCEPTOR_JDBC_CATALOG);

        //  注册client
        DB_CLIENT_MAP.put(CatalogConstant.MYSQL_CATALOG, new MySqlJdbcClient());
        DB_CLIENT_MAP.put(CatalogConstant.MARIA_CATALOG, new MariaJdbcClient());
        DB_CLIENT_MAP.put(CatalogConstant.ORACLE_CATALOG, new OracleJdbcClient());
        DB_CLIENT_MAP.put(CatalogConstant.OPENSEARCH_CATALOG, new OpenSearchClient());
        DB_CLIENT_MAP.put(CatalogConstant.INCEPTOR_JDBC_CATALOG, new Inceptor2JdbcClient());

        // 注册getTable的sql模板
        TABLE_METADATA_SQL_TEMPLATE_MAP.put(CatalogConstant.MARIA_CATALOG, MARIA_DB);
        TABLE_METADATA_SQL_TEMPLATE_MAP.put(CatalogConstant.MYSQL_CATALOG, MYSQL_DB);
        TABLE_METADATA_SQL_TEMPLATE_MAP.put(CatalogConstant.ORACLE_CATALOG, ORACLE_DB);
        TABLE_METADATA_SQL_TEMPLATE_MAP.put(CatalogConstant.INCEPTOR_JDBC_CATALOG, INCEPTOR_JDBC_DB);

        // MySQL：5.x版本驱动类为com.mysql.jdbc.Driver，8.0+为com.mysql.cj.jdbc.Driver（推荐）
        DRIVER_CLASS_MAP.put(CatalogConstant.MYSQL_CATALOG, "com.mysql.cj.jdbc.Driver");
        DRIVER_CLASS_MAP.put(CatalogConstant.MARIA_CATALOG, "org.mariadb.jdbc.Driver");
        // Oracle：ojdbc8及以上版本推荐用oracle.jdbc.OracleDriver（替代旧的oracle.jdbc.driver.OracleDriver）
        DRIVER_CLASS_MAP.put(CatalogConstant.ORACLE_CATALOG, "oracle.jdbc.OracleDriver");
        DRIVER_CLASS_MAP.put(CatalogConstant.INCEPTOR_JDBC_CATALOG, "org.apache.hive.jdbc.HiveDriver");


        // PostgreSQL：主流版本统一为org.postgresql.Driver
        DRIVER_CLASS_MAP.put(CatalogConstant.POSTGRESQL_CATALOG, "org.postgresql.Driver");
        // SQL Server：2017+版本用com.microsoft.sqlserver.jdbc.SQLServerDriver
        DRIVER_CLASS_MAP.put(CatalogConstant.SQLSERVER_CATALOG, "com.microsoft.sqlserver.jdbc.SQLServerDriver");
    }

    /**
     * 根据数据源类型获取连接策略
     *
     * @param dbType 数据源类型（mysql/oracle/postgresql）
     * @return 对应的连接策略
     * @throws IllegalArgumentException 不支持的数据源类型
     */
    public static DbClientInterface getStrategy(String dbType) {
        DbClientInterface strategy = DB_CLIENT_MAP.get(dbType.toLowerCase());
        if (strategy == null) {
            throw new IllegalArgumentException("不支持的数据源类型：" + dbType);
        }
        return strategy;
    }

    /**
     * 根据数据源配置获取数据库连接URL
     *
     * @param dataSourceEntity 数据源配置
     * @return 数据库连接URL
     */
    public static String getDriverURL(DataSourceEntity dataSourceEntity) {
        String fType = dataSourceEntity.getFType();
        switch (fType) {
            case "mysql":
                return "jdbc:mysql://" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + "/" + dataSourceEntity.getFDatabase() + "?useSSL=false&serverTimezone=UTC";
            case "maria":
                return "jdbc:mariadb://" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + "/" + dataSourceEntity.getFDatabase() + "?useSSL=false&serverTimezone=UTC";
            case "oracle":
                return "jdbc:oracle:thin:@" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + ":" + dataSourceEntity.getFDatabase();
            case "postgresql":
                return "jdbc:postgresql://" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + "/" + dataSourceEntity.getFDatabase();
            case CatalogConstant.INCEPTOR_JDBC_CATALOG:
                return "jdbc:inceptor2://" + dataSourceEntity.getFHost() + ":" + dataSourceEntity.getFPort() + "/" + dataSourceEntity.getFDatabase();
            default:
                throw new IllegalArgumentException("不支持的数据源类型：" + fType);
        }
    }

    public static boolean supportNewScan(String dbType) {
        return SUPPORT_NEW_SCAN.contains(dbType);
    }
}
