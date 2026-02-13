// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

var (
	// 持久化任务的固定频率调度时间和时间窗口的正则匹配，支持单位有分钟，小时，天
	DurationDayHourMinuteRE = regexp.MustCompile("^(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?$")

	// 持久化任务的追溯时长的正则匹配，支持单位有小时，天
	DurationDayHourRE = regexp.MustCompile("^(([0-9]+)d)?(([0-9]+)h)?$")
)

// 时间区间解析
func ParseDuration(s string, reg *regexp.Regexp, containMinute bool) (time.Duration, error) {
	if d, err := strconv.ParseFloat(s, 64); err == nil {
		ts := d * float64(time.Second)
		if ts > float64(math.MaxInt64) || ts < float64(math.MinInt64) {
			return 0, fmt.Errorf("cannot parse %q to a valid duration. It overflows int64", s)
		}
		return time.Duration(ts), nil
	}

	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	matches := reg.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("not a valid duration string: %q", s)
	}
	var dur time.Duration

	// Parse the match at pos `pos` in the regex and use `mult` to turn that
	// into ms, then add that value to the total parsed duration.
	var overflowErr error
	m := func(pos int, mult time.Duration) {
		if matches[pos] == "" {
			return
		}
		n, _ := strconv.Atoi(matches[pos])

		// Check if the provided duration overflows time.Duration (> ~ 290years).
		if n > int((1<<63-1)/mult/time.Millisecond) {
			overflowErr = errors.New("duration out of range")
		}
		d := time.Duration(n) * time.Millisecond
		dur += d * mult

		if dur < 0 {
			overflowErr = errors.New("duration out of range")
		}
	}

	m(2, 1000*60*60*24) // d
	m(4, 1000*60*60)    // h
	if containMinute {
		m(6, 1000*60) // m
	}

	if overflowErr == nil {
		return time.Duration(dur), nil
	}
	return 0, fmt.Errorf("cannot parse %q to a valid duration", s)
}

func DurationMilliseconds(d time.Duration) int64 {
	return int64(d / (time.Millisecond / time.Nanosecond))
}
func TimeMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond/time.Nanosecond)
}

func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// []string 转 []uint64
func StringSliceToInt64Slice(strArr []string) ([]uint64, error) {
	res := make([]uint64, len(strArr))
	var err error
	for index, val := range strArr {
		res[index], err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return []uint64{}, fmt.Errorf("[]string %v to []uint64 parse failed", strArr)
		}
	}
	return res, nil
}
