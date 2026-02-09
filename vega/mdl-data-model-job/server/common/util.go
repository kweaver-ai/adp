// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"strings"
	"time"
)

// 获取当前时间的时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// string 转 []string
func StringToStringSlice(str string) []string {
	if str == "" {
		return []string{}
	} else {
		//去括号
		str = strings.Trim(str, "{} <>")

		//分割
		return strings.Split(str, ",")
	}
}

// 生成数据库更新时间-字符串
func GenerateUpdateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
