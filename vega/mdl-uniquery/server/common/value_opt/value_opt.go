// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package value_opt

const (
	ValueFrom_Const = "const"
	ValueFrom_Field = "field"
	ValueFrom_User  = "user"
)

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
	RealValue any    `json:"real_value,omitempty" mapstructure:"real_value"`
}
