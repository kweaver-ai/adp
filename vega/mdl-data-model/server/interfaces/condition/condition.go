// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

// new condition operation
const (
	OperationAnd = "and"
	OperationOr  = "or"

	OperationEq          = "=="
	OperationNotEq       = "!="
	OperationGt          = ">"
	OperationGte         = ">="
	OperationLt          = "<"
	OperationLte         = "<="
	OperationIn          = "in"
	OperationNotIn       = "not_in"
	OperationLike        = "like"
	OperationNotLike     = "not_like"
	OperationContain     = "contain"
	OperationNotContain  = "not_contain"
	OperationRange       = "range"
	OperationOutRange    = "out_range"
	OperationExist       = "exist"
	OperationNotExist    = "not_exist"
	OperationEmpty       = "empty"
	OperationNotEmpty    = "not_empty"
	OperationRegex       = "regex"
	OperationMatch       = "match"
	OperationMatchPhrase = "match_phrase"
	OperationPrefix      = "prefix"
	OperationNotPrefix   = "not_prefix"
	OperationNull        = "null"
	OperationNotNull     = "not_null"
	OperationTrue        = "true"
	OperationFalse       = "false"
	OperationBefore      = "before"
	OperationCurrent     = "current"
	OperationBetween     = "between"
)

const (
	MaxSubCondition = 10
)

const (
	ValueFrom_Const = "const"
	ValueFrom_Field = "field"
)

var (
	OperationMap = map[string]struct{}{
		OperationAnd:         {},
		OperationOr:          {},
		OperationEq:          {},
		OperationNotEq:       {},
		OperationGt:          {},
		OperationGte:         {},
		OperationLt:          {},
		OperationLte:         {},
		OperationIn:          {},
		OperationNotIn:       {},
		OperationLike:        {},
		OperationNotLike:     {},
		OperationContain:     {},
		OperationNotContain:  {},
		OperationRange:       {},
		OperationOutRange:    {},
		OperationExist:       {},
		OperationNotExist:    {},
		OperationEmpty:       {},
		OperationNotEmpty:    {},
		OperationRegex:       {},
		OperationMatch:       {},
		OperationMatchPhrase: {},
		OperationPrefix:      {},
		OperationNotPrefix:   {},
		OperationNull:        {},
		OperationNotNull:     {},
		OperationTrue:        {},
		OperationFalse:       {},
		OperationBefore:      {},
		OperationCurrent:     {},
		OperationBetween:     {},
	}

	OperationMatchMap = map[string]struct{}{
		OperationMatch:       {},
		OperationMatchPhrase: {},
	}

	NotRequiredValueOperationMap = map[string]struct{}{
		OperationExist:    {},
		OperationNotExist: {},
		OperationEmpty:    {},
		OperationNotEmpty: {},
		OperationNull:     {},
		OperationNotNull:  {},
		OperationTrue:     {},
		OperationFalse:    {},
	}

	HavingOperationMap = map[string]struct{}{
		OperationEq:       {},
		OperationNotEq:    {},
		OperationGt:       {},
		OperationGte:      {},
		OperationLt:       {},
		OperationLte:      {},
		OperationIn:       {},
		OperationNotIn:    {},
		OperationRange:    {},
		OperationOutRange: {},
	}
)

// old filters operation
const (
	Operation_IN        string = "in"
	Operation_NOT_IN    string = "not in"
	Operation_EQ        string = "="
	Operation_NE        string = "!="
	Operation_RANGE     string = "range"
	Operation_OUT_RANGE string = "out_range"
	Operation_LIKE      string = "like"
	Operation_NOT_LIKE  string = "not_like"
	Operation_GT        string = ">"
	Operation_GTE       string = ">="
	Operation_LT        string = "<"
	Operation_LTE       string = "<="
	Operation_EXIST     string = "exist"
	Operation_NOT_EXIST string = "not_exist"
	Operation_REGEX     string = "regex"
)

const (
	FILTERS_MAX_NUMBER = 5
)
