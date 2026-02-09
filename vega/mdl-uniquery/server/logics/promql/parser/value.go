// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package parser

import "math"

// Value is a generic interface for values resulting from a query evaluation.
type Value interface {
	Type() ValueType
	String() string
}

// ValueType describes a type of a value.
type ValueType string

// The valid value types.
const (
	ValueTypeNone   ValueType = "none"
	ValueTypeVector ValueType = "vector"
	ValueTypeScalar ValueType = "scalar"
	ValueTypeMatrix ValueType = "matrix"
	ValueTypeString ValueType = "string"
	//ValueTypeSeries ValueType = "series"
)

// DocumentedType returns the internal type to the equivalent
// user facing terminology as defined in the documentation.
func DocumentedType(t ValueType) string {
	switch t {
	case ValueTypeVector:
		return "instant vector"
	case ValueTypeMatrix:
		return "range vector"
	default:
		return string(t)
	}
}

const (
	// NormalNaN is a quiet NaN. This is also math.NaN().
	NormalNaN uint64 = 0x7ff8000000000001

	// StaleNaN is a signaling NaN, due to the MSB of the mantissa being 0.
	// This value is chosen with many leading 0s, so we have scope to store more
	// complicated values in the future. It is 2 rather than 1 to make
	// it easier to distinguish from the NormalNaN by a human when debugging.
	StaleNaN uint64 = 0x7ff0000000000002
)

// IsStaleNaN returns true when the provided NaN value is a stale marker.
func IsStaleNaN(v float64) bool {
	return math.Float64bits(v) == StaleNaN
}
