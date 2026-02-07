// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package parser

import (
	"context"
	"fmt"
	"strings"
	"uniquery/interfaces"
	"uniquery/logics"
)

// Function represents a function of the expression language and is
// used by function nodes.
type Function struct {
	Name          string
	ArgTypes      []ValueType
	Variadic      int
	ReturnType    ValueType
	SemanticCheck SemanticCheck // 语义检查
}

type SemanticCheck = func(context.Context, Expressions) error

// Functions is a list of all functions supported by PromQL, including their types.
var Functions = map[string]*Function{
	"abs": {
		Name:       "abs",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"absent": {
		Name:       "absent",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"absent_over_time": {
		Name:       "absent_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"present_over_time": {
		Name:       "present_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"avg_over_time": {
		Name:       "avg_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"ceil": {
		Name:       "ceil",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"changes": {
		Name:       "changes",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"clamp": {
		Name:       "clamp",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"clamp_max": {
		Name:       "clamp_max",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"clamp_min": {
		Name:       "clamp_min",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"count_over_time": {
		Name:       "count_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"days_in_month": {
		Name:       "days_in_month",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"day_of_month": {
		Name:       "day_of_month",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"day_of_week": {
		Name:       "day_of_week",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"delta": {
		Name:       "delta",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"deriv": {
		Name:       "deriv",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"dict_labels": {
		Name:       "dict_labels",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeString},
		Variadic:   -1,
		ReturnType: ValueTypeVector,
		SemanticCheck: func(ctx context.Context, args Expressions) error {
			// 判断参数是否为偶数个并且不少于6个
			if len(args)%2 != 0 || len(args) < 6 {
				return fmt.Errorf("invalid parameter in dict_labels()")
			}
			if args[1].Type() == ValueTypeString {
				dictName := strings.Trim(args[1].String(), "\"")
				// 加载字典数据
				err := logics.DDService.LoadDict(ctx, dictName)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("invalid dict_name in dict_labels()")
			}
			return nil
		},
	},
	"dict_values": {
		Name:       "dict_values",
		ArgTypes:   []ValueType{ValueTypeString},
		Variadic:   -1,
		ReturnType: ValueTypeVector,
		SemanticCheck: func(ctx context.Context, args Expressions) error {
			// 判断参数是否为偶数个并且不少于6个
			if len(args)%2 != 0 || len(args) < 4 {
				return fmt.Errorf("invalid parameter in dict_values(),actual number of arguments: %d, expected must greater than 6", len(args))
			}
			if args[0].Type() == ValueTypeString {
				dictName := strings.Trim(args[0].String(), "\"")
				// 加载字典数据
				err := logics.DDService.LoadDict(ctx, dictName)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("invalid dict_name in dict_values()")
			}
			return nil
		},
	},
	"exp": {
		Name:       "exp",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"floor": {
		Name:       "floor",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"histogram_quantile": {
		Name:       "histogram_quantile",
		ArgTypes:   []ValueType{ValueTypeScalar, ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"holt_winters": {
		Name:       "holt_winters",
		ArgTypes:   []ValueType{ValueTypeMatrix, ValueTypeScalar, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"hour": {
		Name:       "hour",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"idelta": {
		Name:       "idelta",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"increase": {
		Name:       "increase",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"irate": {
		Name:       "irate",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"label_replace": {
		Name:       "label_replace",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeString, ValueTypeString, ValueTypeString, ValueTypeString},
		ReturnType: ValueTypeVector,
	},
	"label_join": {
		Name:       "label_join",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeString, ValueTypeString, ValueTypeString},
		Variadic:   -1,
		ReturnType: ValueTypeVector,
	},
	"last_over_time": {
		Name:       "last_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"ln": {
		Name:       "ln",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"log10": {
		Name:       "log10",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"log2": {
		Name:       "log2",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"max_over_time": {
		Name:       "max_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"min_over_time": {
		Name:       "min_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"minute": {
		Name:       "minute",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"month": {
		Name:       "month",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"predict_linear": {
		Name:       "predict_linear",
		ArgTypes:   []ValueType{ValueTypeMatrix, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"percent_rank": {
		Name:       "percent_rank",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"quantile_over_time": {
		Name:       "quantile_over_time",
		ArgTypes:   []ValueType{ValueTypeScalar, ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"rate": {
		Name:       "rate",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"rank": {
		Name:       "rank",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"resets": {
		Name:       "resets",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"round": {
		Name:       "round",
		ArgTypes:   []ValueType{ValueTypeVector, ValueTypeScalar},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"scalar": {
		Name:       "scalar",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeScalar,
	},
	"sgn": {
		Name:       "sgn",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"sort": {
		Name:       "sort",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"sort_desc": {
		Name:       "sort_desc",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"sqrt": {
		Name:       "sqrt",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"stddev_over_time": {
		Name:       "stddev_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"stdvar_over_time": {
		Name:       "stdvar_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"sum_over_time": {
		Name:       "sum_over_time",
		ArgTypes:   []ValueType{ValueTypeMatrix},
		ReturnType: ValueTypeVector,
	},
	"time": {
		Name:       "time",
		ArgTypes:   []ValueType{},
		ReturnType: ValueTypeScalar,
	},
	"timestamp": {
		Name:       "timestamp",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"vector": {
		Name:       "vector",
		ArgTypes:   []ValueType{ValueTypeScalar},
		ReturnType: ValueTypeVector,
	},
	"year": {
		Name:       "year",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   1,
		ReturnType: ValueTypeVector,
	},
	"cumulative_sum": {
		Name:       "cumulative_sum",
		ArgTypes:   []ValueType{ValueTypeVector},
		ReturnType: ValueTypeVector,
	},
	"greatest": {
		Name:       "greatest",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   -1,              // todo: Variadic的含义？？？
		ReturnType: ValueTypeVector, // todo: 考虑是否支持标量间的最大值，标量间的最大值的返回值是否是vector
	},
	"least": {
		Name:       "least",
		ArgTypes:   []ValueType{ValueTypeVector},
		Variadic:   -1,              // todo: Variadic的含义？？？
		ReturnType: ValueTypeVector, // todo: 考虑是否支持标量间的最大值，标量间的最大值的返回值是否是vector
	},
	"continuous_k_minute_downtime": {
		Name:       "continuous_k_minute_downtime",
		ArgTypes:   []ValueType{ValueTypeScalar, ValueTypeScalar, ValueTypeScalar, ValueTypeVector},
		Variadic:   -1,
		ReturnType: ValueTypeVector, // k := int(args[1].(*parser.NumberLiteral).Val)
		SemanticCheck: func(ctx context.Context, args Expressions) error {
			if len(args) < 4 {
				return fmt.Errorf("invalid parameter in continuous_k_minute_downtime(), the number of arguments must gte 4, actual number is %d", len(args))
			}

			// 最多五个v进行比较
			if len(args) > 8 {
				return fmt.Errorf("invalid parameter in continuous_k_minute_downtime(), the max number of eval vector must lte 5, actual is %d", len(args)-3)
			}

			// 判断参数，第一个参数是k分钟，k >= 1
			k := int(args[0].(*NumberLiteral).Val)
			if k < 1 || k > interfaces.MAX_K_MINUTE {
				return fmt.Errorf("invalid k parameter in continuous_k_minute_downtime(),expected k gte 1 and lte 60, actual is %d", k)
			}
			// 当前序时间段的数据缺失时使用的策略。 -1: 将数据缺失的时间段视为异常（服务不可用）；1：将数据缺失的时间段视为正常（服务可用）
			precedingMissingPolicy := int(args[1].(*NumberLiteral).Val)
			if precedingMissingPolicy != -1 && precedingMissingPolicy != 1 {
				return fmt.Errorf("invalid preceding_missing_policy parameter in continuous_k_minute_downtime(),expected preceding_missing_policy  is one of [-1, 1] gte 1, actual is %d", precedingMissingPolicy)
			}
			// 当中间段的数据缺失时，给定的填充值，必须。 -1: 将数据缺失的时间段视为异常（服务不可用）；1：将数据缺失的时间段视为正常（服务可用）；0：数据缺失的时间段的状态沿用前一分钟的状态。
			middleMissingPolicy := int(args[2].(*NumberLiteral).Val)
			if middleMissingPolicy != -1 && middleMissingPolicy != 0 && middleMissingPolicy != 1 {
				return fmt.Errorf("invalid preceding_missing_policy parameter in continuous_k_minute_downtime(),expected preceding_missing_policy  is one of [-1, 0, 1] gte 1, actual is %d", middleMissingPolicy)
			}
			return nil
		},
	},
	"metric_model": {
		Name:       "metric_model",
		ArgTypes:   []ValueType{ValueTypeString},
		ReturnType: ValueTypeVector,
	},
}

// getFunction returns a predefined Function object for the given name.
func getFunction(name string) (*Function, bool) {
	function, ok := Functions[name]
	return function, ok
}
