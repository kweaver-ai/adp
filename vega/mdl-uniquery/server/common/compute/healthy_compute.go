// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package compute

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"uniquery/interfaces"
)

type HealthyLevel struct {
	Level      string
	Id         int
	UpperBound float64
	LowerBound float64
}

var EventModelMap = map[int]string{
	1: "Critical",
	2: "Major",
	3: "Minor",
	4: "Warning",
	5: "Indeterminate",
	6: "Cleared",
}

var LevelMap = map[string]HealthyLevel{
	"Indeterminate": {
		Level:      "Indeterminate",
		Id:         5,
		UpperBound: 0.0,
		LowerBound: 0.0,
	},

	"Critical": {
		Level:      "Unavailable",
		Id:         1,
		UpperBound: 40.0,
		LowerBound: 20.0,
	},

	"Major": {
		Level:      "Fault",
		Id:         2,
		UpperBound: 60.0,
		LowerBound: 40.0,
	},

	"Minor": {
		Level:      "Error",
		Id:         3,
		UpperBound: 80.0,
		LowerBound: 60.0,
	},
	"Warning": {
		Level:      "Warning",
		Id:         4,
		UpperBound: 100.0,
		LowerBound: 80.0,
	},
	"Cleared": {
		Level:      "Healthy",
		Id:         6,
		UpperBound: 100.0,
		LowerBound: 100.0,
	},
}

func MaxLevelMap(sr interfaces.Records) (interfaces.Records, int, float64) {
	//NOTE: 最高等级映射算法

	var LevelMapSet = map[string]interfaces.Records{
		"Critical":      {},
		"Major":         {},
		"Minor":         {},
		"Warning":       {},
		"Indeterminate": {},
		"Cleared":       {},
	}
	for _, s := range sr {
		level_id, _ := s["level"].(int)
		level := EventModelMap[level_id]
		LevelMapSet[level] = append(LevelMapSet[level], s)
	}
	for _, level := range []string{"Critical", "Major", "Minor", "Warning", "Indeterminate", "Cleared"} {
		if len(LevelMapSet[level]) > 0 {
			hit_records := LevelMapSet[level]
			hits_level := LevelMap[level].Id
			hits_score := math.Max(LevelMap[level].UpperBound-float64(len(LevelMapSet[level])), LevelMap[level].LowerBound)
			return hit_records, hits_level, hits_score
		}

	}

	return nil, 0, 0
}

func EventDataGroupAggregation(sr interfaces.Records, group_fields []string) map[string]interfaces.Records {
	// group_field := strings.Join(group_fields, ",")
	var GroupRecords = make(map[string]interfaces.Records, len(sr)/2)
	if len(group_fields) <= 0 {
		return map[string]interfaces.Records{}
	}

	for _, r := range sr {
		level, _ := r["level"].(int)
		labelDimison := map[string]string{
			"title":            r["title"].(string),
			"event_model_name": r["event_model_name"].(string),
			"event_model_id":   strconv.FormatUint(r["event_model_id"].(uint64), 10),
			"type":             r["type"].(string),
			"level":            strconv.FormatInt(int64(level), 10),
		}
		groupFieldValueStr := GetGroupFieldValue(labelDimison, group_fields)
		GroupRecords[groupFieldValueStr] = append(GroupRecords[groupFieldValueStr], r)
	}
	return GroupRecords
}

func GetGroupFieldValue(labelDimison map[string]string, group_fields []string) string {
	group_values := []string{}
	for _, field := range group_fields {
		if labelDimison[field] == "" {
			continue
		} else {
			group_values = append(group_values, labelDimison[field])
		}
	}
	return strings.Join(group_values, ",")
}

func SourceDataGroupAggregation(sr interfaces.Records, group_fields []string) map[string]interfaces.Records {
	// group_field := strings.Join(group_fields, ",")
	var GroupRecords = make(map[string]interfaces.Records, len(sr)/2)
	if len(group_fields) <= 0 {
		return map[string]interfaces.Records{}
	}
	var labelDimison = make(map[string]string, 10)
	for _, r := range sr {
		triggerData := r["trigger_data"].(interfaces.Records)[0]

		// if _, ok := reflect.TypeOf(triggerData).FieldByName("label"); ok {
		// 	lables := reflect.ValueOf(triggerData).FieldByName("label").Interface().(map[string]string)
		// 	for key,value := range lables{
		// 		labelDimison[key] = value
		// 	}
		// }
		// labelDimison, _ = triggerData["label"].(map[string]string)

		for key, value := range triggerData {
			if v, ok := value.(string); ok {
				labelDimison[key] = v
			}
		}

		// labelDimison["value"] = strconv.FormatFloat(triggerData["value"].(float64), 'f', -1, 64)
		labelDimison["level"] = strconv.Itoa(r["level"].(int))
		groupFieldValueStr := GetGroupFieldValue(labelDimison, group_fields)

		// []map[string]any, labelDimison
		// []map[string]any             map[string]string
		labelDimisonStr, _ := json.Marshal(labelDimison)
		var labelDimisonT = make(map[string]any)
		_ = json.Unmarshal(labelDimisonStr, &labelDimisonT)

		GroupRecords[groupFieldValueStr] = append(GroupRecords[groupFieldValueStr], labelDimisonT)
	}
	return GroupRecords
}
