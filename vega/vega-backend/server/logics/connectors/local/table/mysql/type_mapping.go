// Package mysql provides MySQL database connector implementation.
package mysql

// TypeMapping maps MySQL native types to VEGA types.
var TypeMapping = map[string]string{
	// Integer types
	"tinyint":   "integer",
	"smallint":  "integer",
	"mediumint": "integer",
	"int":       "integer",
	"integer":   "integer",
	"bigint":    "integer",
	"year":      "integer",

	// Unsigned integer types
	"tinyint unsigned":   "unsigned_integer",
	"smallint unsigned":  "unsigned_integer",
	"mediumint unsigned": "unsigned_integer",
	"int unsigned":       "unsigned_integer",
	"integer unsigned":   "unsigned_integer",
	"bigint unsigned":    "unsigned_integer",

	// Float types
	"float":            "float",
	"double":           "float",
	"real":             "float",
	"double precision": "float",

	// Decimal types
	"decimal": "decimal",
	"numeric": "decimal",
	"fixed":   "decimal",
	"dec":     "decimal",

	// String types
	"char":    "string",
	"varchar": "string",

	// Text types
	"tinytext":   "text",
	"text":       "text",
	"mediumtext": "text",
	"longtext":   "text",

	// Date/Time types
	"date":      "date",
	"datetime":  "datetime",
	"timestamp": "datetime",
	"time":      "time",

	// Boolean
	"boolean": "boolean",
	"bool":    "boolean",
	"bit":     "boolean",

	// Binary types
	"binary":     "binary",
	"varbinary":  "binary",
	"tinyblob":   "binary",
	"blob":       "binary",
	"mediumblob": "binary",
	"longblob":   "binary",

	// JSON
	"json": "json",
}

// MapType returns VEGA type for MySQL native type.
func MapType(nativeType string) string {
	if vegaType, ok := TypeMapping[nativeType]; ok {
		return vegaType
	}
	return "unsupported" // default
}
