// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

type LogGroup struct {
	IndexPattern []string          `json:"indices"`
	MustFilter   interface{}       `json:"must_filters"`
	Fields       map[string]string `json:"fields"` // 日志分组的字段信息
}

// 日志分组结构体
type LogGroupInfo struct {
	Id   string `json:"groupId"`
	Name string `json:"groupName"`
}

type QueryFilters struct {
	Indices     Indices                    `json:"indices"`
	MustFilter  interface{}                `json:"must_filters"`
	Name        string                     `json:"groupName"`
	ArrayFields map[string][]LogGroupField `json:"fields"`
	Fields      map[string]string          `json:"-"`
}

type Indices struct {
	IndexPattern []string `json:"index_pattern"`
	ManualIndex  []string `json:"manual_index"`
}

type LogGroupField struct {
	Name string      `json:"field"`
	Type interface{} `json:"type"`
}

//go:generate mockgen -source ../interfaces/log_group_access.go -destination ../interfaces/mock/mock_log_group_access.go
type LogGroupAccess interface {
	// GetLogGroupQueryFilters(groupId string) (LogGroup, int, error)
	GetLogGroupQueryFilters(logGroupID string) (LogGroup, bool, error)
	GetLogGroupByName(logGroupName string) ([]LogGroupInfo, error)
	GetLogGroupQueryFiltersAndFields(groupId string) (LogGroup, bool, error)
}
