// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mysql provides MySQL database connector implementation.
package mysql

import "vega-backend/interfaces/data_type"

// TypeMapping maps MySQL native types to VEGA types.
var TypeMapping = map[string]string{
	// Integer types
	"tinyint":   data_type.DataType_Integer,
	"smallint":  data_type.DataType_Integer,
	"mediumint": data_type.DataType_Integer,
	"int":       data_type.DataType_Integer,
	"integer":   data_type.DataType_Integer,
	"bigint":    data_type.DataType_Integer,
	"year":      data_type.DataType_Integer,

	// Unsigned integer types
	"tinyint unsigned":   data_type.DataType_UnsignedInteger,
	"smallint unsigned":  data_type.DataType_UnsignedInteger,
	"mediumint unsigned": data_type.DataType_UnsignedInteger,
	"int unsigned":       data_type.DataType_UnsignedInteger,
	"integer unsigned":   data_type.DataType_UnsignedInteger,
	"bigint unsigned":    data_type.DataType_UnsignedInteger,

	// Float types
	"float":            data_type.DataType_Float,
	"double":           data_type.DataType_Float,
	"real":             data_type.DataType_Float,
	"double precision": data_type.DataType_Float,

	// Decimal types
	"decimal": data_type.DataType_Decimal,
	"numeric": data_type.DataType_Decimal,
	"fixed":   data_type.DataType_Decimal,
	"dec":     data_type.DataType_Decimal,

	// String types
	"char":    data_type.DataType_String,
	"varchar": data_type.DataType_String,

	// Text types
	"tinytext":   data_type.DataType_Text,
	"text":       data_type.DataType_Text,
	"mediumtext": data_type.DataType_Text,
	"longtext":   data_type.DataType_Text,

	// Date/Time types
	"date":      data_type.DataType_Date,
	"datetime":  data_type.DataType_Datetime,
	"timestamp": data_type.DataType_Datetime,
	"time":      data_type.DataType_Time,

	// Boolean
	"boolean": data_type.DataType_Boolean,
	"bool":    data_type.DataType_Boolean,
	"bit":     data_type.DataType_Boolean,

	// Binary types
	"binary":     data_type.DataType_Binary,
	"varbinary":  data_type.DataType_Binary,
	"tinyblob":   data_type.DataType_Binary,
	"blob":       data_type.DataType_Binary,
	"mediumblob": data_type.DataType_Binary,
	"longblob":   data_type.DataType_Binary,

	// JSON
	"json": data_type.DataType_Json,
}

// MapType returns VEGA type for MySQL native type.
func MapType(nativeType string) string {
	if vegaType, ok := TypeMapping[nativeType]; ok {
		return vegaType
	}
	return data_type.DataType_Other // default
}
