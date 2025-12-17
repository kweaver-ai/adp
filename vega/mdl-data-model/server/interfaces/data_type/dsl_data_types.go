package data_type

const (
	IndexBase_DataType_Keyword = "keyword"
	IndexBase_DataType_Text    = "text"
	IndexBase_DataType_Binary  = "binary"

	IndexBase_DataType_Short     = "short"
	IndexBase_DataType_Integer   = "integer"
	IndexBase_DataType_Long      = "long"
	IndexBase_DataType_Float     = "float"
	IndexBase_DataType_Double    = "double"
	IndexBase_DataType_HalfFloat = "half_float"
	IndexBase_DataType_Byte      = "byte"

	IndexBase_DataType_Boolean = "boolean"

	IndexBase_DataType_Date     = "date"
	IndexBase_DataType_DateTime = "datetime"

	IndexBase_DataType_Ip        = "ip"
	IndexBase_DataType_GeoPoint  = "geo_point"
	IndexBase_DataType_GeoShape  = "geo_shape"
	IndexBase_DataType_KNNVector = "knn_vector"
)

// 索引库类型映射为统一类型
var IndexBase_DataType_Map = map[string]string{
	IndexBase_DataType_Keyword: DataType_String,
	IndexBase_DataType_Text:    DataType_Text,
	IndexBase_DataType_Binary:  DataType_Binary,

	IndexBase_DataType_Short:     DataType_Integer,
	IndexBase_DataType_Integer:   DataType_Integer,
	IndexBase_DataType_Long:      DataType_Integer,
	IndexBase_DataType_Float:     DataType_Float,
	IndexBase_DataType_Double:    DataType_Float,
	IndexBase_DataType_HalfFloat: DataType_Float,
	IndexBase_DataType_Byte:      DataType_Integer,

	IndexBase_DataType_Boolean: DataType_Boolean,

	IndexBase_DataType_Date:     DataType_Datetime,
	IndexBase_DataType_DateTime: DataType_Datetime,

	IndexBase_DataType_Ip:       DataType_Ip,
	IndexBase_DataType_GeoPoint: DataType_Point,
	IndexBase_DataType_GeoShape: DataType_Shape,
}

// var (
// 	STRING_TYPES = map[string]struct{}{
// 		DataType_String: {},
// 		DataType_Text:   {},
// 	}

// 	NUMBER_TYPES = map[string]struct{}{
// 		DataType_Integer:         {},
// 		DataType_UnsignedInteger: {},
// 		DataType_Float:           {},
// 		DataType_Decimal:         {},
// 	}

// BOOLEAN_TYPES = map[string]struct{}{
// 	DataType_Boolean: {},
// }

// DATE_TYPES = map[string]struct{}{
// 	DataType_Date:     {},
// 	DataType_DateTime: {},
// }

// IP_TYPES = map[string]struct{}{
// 	DataType_Ip: {},
// }

// GEO_POINT_TYPES = map[string]struct{}{
// 	DataType_Point: {},
// }

// GEO_SHAPE_TYPES = map[string]struct{}{
// 	DataType_Shape: {},
// }
// )
