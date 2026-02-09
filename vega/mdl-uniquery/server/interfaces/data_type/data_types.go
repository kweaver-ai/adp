// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_type

const (
	KEYWORD_SUFFIX = "keyword"
)

const (
	// 整数型
	DataType_Integer         = "integer"
	DataType_UnsignedInteger = "unsigned integer"

	// 浮点型
	DataType_Float = "float"

	// 任意精度数
	DataType_Decimal = "decimal"

	// 字符串型
	DataType_String = "string"
	DataType_Text   = "text"

	// 时间型
	DataType_Date      = "date"
	DataType_Time      = "time"
	DataType_Datetime  = "datetime"
	DataType_Timestamp = "timestamp"

	// ip类型
	DataType_Ip = "ip"

	// 布尔型
	DataType_Boolean = "boolean"

	// 二进制数据类型
	DataType_Binary = "binary"

	// json类型
	DataType_Json = "json"

	// 空间类型
	DataType_Point = "point"
	DataType_Shape = "shape"

	// 向量类型
	DataType_Vector = "vector"
)

var (
	STRING_TYPES = map[string]struct{}{
		DataType_String: {},
		DataType_Text:   {},
	}

	NUMBER_TYPES = map[string]struct{}{
		DataType_Integer:         {},
		DataType_UnsignedInteger: {},
		DataType_Float:           {},
		DataType_Decimal:         {},
	}

	// BOOLEAN_TYPES = map[string]struct{}{
	// 	DATATYPE_BOOLEAN: {},
	// }

	DATE_TYPES = map[string]struct{}{
		DataType_Date:      {},
		DataType_Time:      {},
		DataType_Datetime:  {},
		DataType_Timestamp: {},
	}

	// IP_TYPES = map[string]struct{}{
	// 	DATATYPE_IP: {},
	// }

	// GEO_POINT_TYPES = map[string]struct{}{
	// 	DATATYPE_GEO_POINT: {},
	// }

	// GEO_SHAPE_TYPES = map[string]struct{}{
	// 	DATATYPE_GEO_SHAPE: {},
	// }
)

func DataType_IsString(t string) bool {
	_, ok := STRING_TYPES[t]
	return ok
}

func DataType_IsNumber(t string) bool {
	_, ok := NUMBER_TYPES[t]
	return ok
}

func DataType_IsDate(t string) bool {
	_, ok := DATE_TYPES[t]
	return ok
}

// region 业务大类型 与 虚拟化引擎类型 映射
var SimpleTypeMapping = map[string]string{
	//region 字符型
	"string":           DataType_String,
	"char":             DataType_String,
	"varchar":          DataType_String,
	"json":             DataType_Json,
	"text":             DataType_Text,
	"tinytext":         DataType_Text,
	"mediumtext":       DataType_Text,
	"longtext":         DataType_Text,
	"uuid":             DataType_String,
	"name":             DataType_String,
	"jsonb":            DataType_Json,
	"bpchar":           DataType_String,
	"uniqueidentifier": DataType_String,
	"xml":              DataType_String,
	"sysname":          DataType_String,
	"nvarchar":         DataType_String,
	"enum":             DataType_String,
	"set":              DataType_String,
	"ntext":            DataType_String,
	"nchar":            DataType_String,
	"rowid":            DataType_String,
	"urowid":           DataType_String,
	"varchar2":         DataType_String,
	"nvarchar2":        DataType_String,
	"fixedstring":      DataType_String,
	"nclob":            DataType_String,
	"ipaddress":        DataType_Ip,
	// 	//endregion

	// 	//region 整数型
	"number":             DataType_Integer,
	"tinyint":            DataType_Integer,
	"smallint":           DataType_Integer,
	"integer":            DataType_Integer,
	"bigint":             DataType_Integer,
	"int":                DataType_Integer,
	"mediumint":          DataType_Integer,
	"int unsigned":       DataType_UnsignedInteger,
	"tinyint unsigned":   DataType_UnsignedInteger,
	"smallint unsigned":  DataType_UnsignedInteger,
	"mediumint unsigned": DataType_UnsignedInteger,
	"bigint unsigned":    DataType_UnsignedInteger,
	"int8":               DataType_Integer,
	"int4":               DataType_Integer,
	"int2":               DataType_Integer,
	"int16":              DataType_Integer,
	"int32":              DataType_Integer,
	"int64":              DataType_Integer,
	"int128":             DataType_Integer,
	"int256":             DataType_Integer,
	"long":               DataType_Integer,

	"real":             DataType_Float,
	"double":           DataType_Float,
	"float":            DataType_Float,
	"double precision": DataType_Float,
	"float4":           DataType_Float,
	"float8":           DataType_Float,
	"float16":          DataType_Float,
	"float32":          DataType_Float,
	"float64":          DataType_Float,
	"binary_double":    DataType_Float,
	"binary_float":     DataType_Float,

	"decimal": DataType_Decimal,
	"numeric": DataType_Decimal,
	"dec":     DataType_Decimal,

	"boolean": DataType_Boolean,
	"bit":     DataType_Boolean,
	"bool":    DataType_Boolean,

	"date": DataType_Date,
	"year": DataType_Date,

	"datetime":                 DataType_Datetime,
	"datetime2":                DataType_Datetime,
	"smalldatetime":            DataType_Datetime,
	"timestamp":                DataType_Datetime,
	"timestamptz":              DataType_Datetime,
	"timestamp with time zone": DataType_Datetime,
	"interval":                 DataType_Datetime, // 跨度
	"interval year to month":   DataType_Datetime, // 年和月的跨度
	"interval day to second":   DataType_Datetime, // 天数、小时、分钟、秒和毫秒的跨度

	"time":                DataType_Time,
	"timetz":              DataType_Time,
	"time with time zone": DataType_Time,

	"binary":      DataType_Binary,
	"blob":        DataType_Binary,
	"tinyblob":    DataType_Binary,
	"mediumblob":  DataType_Binary,
	"longblob":    DataType_Binary,
	"bytea":       DataType_Binary,
	"image":       DataType_Binary,
	"hierarchyid": DataType_Binary,
	"geography":   DataType_Binary,
	"geometry":    DataType_Binary,
	"varbinary":   DataType_Binary,
	"raw":         DataType_Binary,
	"map":         DataType_Binary,
	"array":       DataType_Binary,
	"struct":      DataType_Binary,

	"money":       SimpleOther,
	"smallmoney":  SimpleOther,
	"oid":         SimpleOther,
	"smallserial": SimpleOther,
	"serial4":     SimpleOther,
	"bigserial":   SimpleOther,
	"serial":      SimpleOther,
	"row":         SimpleOther,
	"hyperloglog": SimpleOther,
}

//endregion

const (
	// DATATYPE_KEYWORD = "keyword"
	// DATATYPE_TEXT    = "text"
	// DATATYPE_BINARY  = "binary"

	// DATATYPE_BYTE       = "byte"
	// DATATYPE_SHORT      = "short"
	// DATATYPE_INTEGER    = "integer"
	// DATATYPE_LONG       = "long"
	// DATATYPE_HALF_FLOAT = "half_float"
	// DATATYPE_FLOAT      = "float"
	// DATATYPE_DOUBLE     = "double"

	// DATATYPE_BOOLEAN = "boolean"

	// DATATYPE_DATE     = "date"
	// DATATYPE_DATETIME = "datetime"

	// DATATYPE_IP        = "ip"
	// DATATYPE_GEO_POINT = "geo_point"
	// DATATYPE_GEO_SHAPE = "geo_shape"

	//字符型
	// CHAR    = "char"
	// VARCHAR = "varchar"
	// STRING  = "string"
	//整数型
	// NUMBER   = "number"
	// TINYINT  = "tinyint"
	// SMALLINT = "smallint"
	// INTEGER  = "integer"
	// INT      = "int" //INTEGER 别名
	// BIGINT   = "bigint"
	//小数型
	// REAL            = "real"
	// FLOAT           = "float" //REAL 别名
	// DOUBLE          = "double"
	// DOUBLEPRECISION = "double precision" //DOUBLE 别名
	//高精度型
	// DECIMAL = "decimal"
	// NUMERIC = "numeric" //DECIMAL 别名
	// DEC     = "dec"     //DECIMAL 别名
	//布尔型
	// BOOLEAN = "boolean"
	// //日期型
	// DATE = "date"
	// //日期时间型
	// TIME                     = "time"
	// TIME_WITH_TIME_ZONE      = "time with time zone"
	// DATETIME                 = "datetime"
	// TIMESTAMP                = "timestamp"
	// TIMESTAMP_WITH_TIME_ZONE = "timestamp with time zone"

	// region 业务大类型
	// SimpleChar     = "char"
	// SimpleInt      = "int"
	// SimpleFloat    = "float"
	// SimpleDecimal  = "decimal"
	// SimpleBool     = "bool"
	// SimpleDate     = "date"
	// SimpleDatetime = "datetime"
	// SimpleTime     = "time"
	// SimpleBinary   = "binary"
	SimpleOther = "other"

// endregion
)
