// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

const (
	//NOTE: FIX_RATE_LIMIT取值为24天对应的秒数，转换为毫秒后不超过int最大值2147483647
	FIX_RATE_LIMIT = 2073600
)

var (
	CronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
)

/*
	公共的校验函数, 适用于所有模块. 包括:
		(1) 对象名称的校验;
	  	(2) 标签数组的校验;
	  	(3) 备注的校验;
	  	(4) 分页查询列表参数的校验;
		(5) 对象名称精确值和模糊值的校验.
*/

// 对象名称错误码字典, key为对象类型, value为其错误码数组
var objectNameErrorCode = map[string][]string{
	interfaces.DATA_DICT: {
		derrors.DataModel_DataDict_NullParameter_DictName,
		derrors.DataModel_DataDict_LengthExceeded_DictName,
	},
	interfaces.MODULE_TYPE_DATA_VIEW: {
		derrors.DataModel_DataView_NullParameter_ViewName,
		derrors.DataModel_DataView_LengthExceeded_ViewName,
	},
	interfaces.DATA_VIEW_GROUP: {
		derrors.DataModel_DataViewGroup_NullParameter_GroupName,
		derrors.DataModel_DataViewGroup_LengthExceeded_GroupName,
	},
	interfaces.MODULE_TYPE_DATA_VIEW_ROW_COLUMN_RULE: {
		derrors.DataModel_DataViewRowColumnRule_NullParameter_RuleName,
		derrors.DataModel_DataViewRowColumnRule_LengthExceeded_RuleName,
	},
	interfaces.METRIC_MODEL_GROUP_MODULE: {
		derrors.DataModel_MetricModelGroup_NullParameter_GrouplName,
		derrors.DataModel_MetricModelGroup_LengthExceeded_GroupName,
	},
	interfaces.METRIC_MODEL_MODULE: {
		derrors.DataModel_MetricModel_NullParameter_ModelName,
		derrors.DataModel_MetricModel_LengthExceeded_ModelName,
	},
	interfaces.METRIC_MODEL_TASK: {
		derrors.DataModel_MetricModel_NullParameter_TaskName,
		derrors.DataModel_MetricModel_LengthExceeded_TaskName,
	},
	interfaces.DATA_CONNECTION: {
		derrors.DataModel_DataConnection_NullParameter_ConnectionName,
		derrors.DataModel_DataConnection_LengthExceeded_ConnectionName,
	},
	interfaces.TRACE_MODEL: {
		derrors.DataModel_TraceModel_NullParameter_ModelName,
		derrors.DataModel_TraceModel_LengthExceeded_ModelName,
	},
	interfaces.OBJECTTYPE_OBJECTIVE_MODEL: {
		derrors.DataModel_ObjectiveModel_NullParameter_ModelName,
		derrors.DataModel_ObjectiveModel_LengthExceeded_ModelName,
	},
}

// 公共校验函数(1): 对象名称合法性校验
func validateObjectName(ctx context.Context, objectName string, objectType string) error {
	if objectName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, objectNameErrorCode[objectType][0])
	}

	if utf8.RuneCountInString(objectName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, objectNameErrorCode[objectType][1]).
			WithErrorDetails(fmt.Sprintf("The length of the %v named %v exceeds %v", objectType, objectName, interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	return nil
}

// 公共校验函数(2): 标签数组合法性校验
func validateObjectTags(ctx context.Context, tags []string) error {
	if len(tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_CountExceeded_Tags).
			WithErrorDetails(fmt.Sprintf("The length of the tag array exceeds %v", interfaces.TAGS_MAX_NUMBER))
	}

	for _, tag := range tags {
		// 去除tag的左右空格
		tag = strings.Trim(tag, " ")

		if tag == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_NullParameter_Tag)
		}

		if utf8.RuneCountInString(tag) > interfaces.OBJECT_NAME_MAX_LENGTH {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_LengthExceeded_Tag).
				WithErrorDetails(fmt.Sprintf("The length of some tags in the tag array exceeds %d", interfaces.OBJECT_NAME_MAX_LENGTH))
		}

		if isInvalid := strings.ContainsAny(interfaces.NAME_INVALID_CHARACTER, tag); isInvalid {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Tag).
				WithErrorDetails(fmt.Sprintf("The tag contains special characters, such as %s", interfaces.NAME_INVALID_CHARACTER))
		}
	}
	return nil
}

// 公共校验函数(3): 备注合法性校验
func validateObjectComment(ctx context.Context, comment string) error {
	if utf8.RuneCountInString(comment) > interfaces.COMMENT_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_LengthExceeded_Comment).
			WithErrorDetails(fmt.Sprintf("The length of the comment exceeds %v", interfaces.COMMENT_MAX_LENGTH))
	}
	return nil
}

// 公共校验函数(4): 分页参数合法性校验
func validatePaginationQueryParameters(ctx context.Context, offset, limit, sort, direction string,
	supportedSortTypes map[string]string) (interfaces.PaginationQueryParameters, error) {
	pageParams := interfaces.PaginationQueryParameters{}

	off, err := strconv.Atoi(offset)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Offset).
			WithErrorDetails(err.Error())
	}

	if off < interfaces.MIN_OFFSET {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Offset).
			WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET))
	}

	lim, err := strconv.Atoi(limit)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Limit).
			WithErrorDetails(err.Error())
	}

	if !(limit == interfaces.NO_LIMIT || (lim >= interfaces.MIN_LIMIT && lim <= interfaces.MAX_LIMIT)) {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("The number per page does not equal %s is not in the range of [%d,%d]", interfaces.NO_LIMIT, interfaces.MIN_LIMIT, interfaces.MAX_LIMIT))
	}

	_, ok := supportedSortTypes[sort]
	if !ok {
		types := make([]string, 0)
		for t := range supportedSortTypes {
			types = append(types, t)
		}
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Sort).
			WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
	}

	if direction != interfaces.DESC_DIRECTION && direction != interfaces.ASC_DIRECTION {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Direction).
			WithErrorDetails("The sort direction is not desc or asc")
	}

	return interfaces.PaginationQueryParameters{
		Offset:    off,
		Limit:     lim,
		Sort:      supportedSortTypes[sort],
		Direction: direction,
	}, nil
}

// 公共校验函数(5): 对象名称精确值和模糊值的校验
func validateNameandNamePattern(ctx context.Context, name, namePattern string) error {
	// name_pattern和name不能同时存在
	if namePattern != "" && name != "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ConflictParameter_NameAndNamePatternCoexist).
			WithErrorDetails("Parameters name_pattern and name are passed in at the same time")
	}

	return nil
}

// 公共校验函数(6): 数据标签名称合法性校验
func validateDataTagName(ctx context.Context, dataTagName string) error {
	// 去除dataTagName的左右空格
	dataTagName = strings.Trim(dataTagName, " ")

	if dataTagName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_DataTagName)
		// .WithErrorDetails("Data tag name is null")
	}

	if utf8.RuneCountInString(dataTagName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_DataTagName).
			WithErrorDetails(fmt.Sprintf("The length of the data tag name exceeds %d", interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	if isInvalid := strings.ContainsAny(interfaces.NAME_INVALID_CHARACTER, dataTagName); isInvalid {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_DataTagName).
			WithErrorDetails(fmt.Sprintf("Data tag name contains special characters, such as %s", interfaces.NAME_INVALID_CHARACTER))
	}

	return nil
}

// tags 的合法性校验
func ValidateTags(ctx context.Context, Tags []string) error {
	if len(Tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_CountExceeded_TagTotal)
	}

	for _, tag := range Tags {
		err := validateDataTagName(ctx, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsValidEventModelType(modelType string) error {
	return nil
}

func IsValidDataSourceType(ctx context.Context, dastaSourceType string) error {
	return nil
}

func IsValidDataSource(dataSource string, dastaSourceType string) error {
	return nil
}

// 事件模型持久化任务信息校验
func validateEventTask(ctx context.Context, task interfaces.EventTask) error {

	//配置持久化则进行校验
	if !common.IsTaskEmpty(task) {
		// 1. 执行频率
		if task.Schedule == (interfaces.Schedule{}) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_NullParameter_Schedule).
				WithErrorDetails("Schedule is null")
		}
		// 2.1 执行类型
		if task.Schedule.Type == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_NullParameter_ScheduleType).
				WithErrorDetails("Schedule type is empty")
		}
		if task.Schedule.Type != interfaces.SCHEDULE_TYPE_FIXED && task.Schedule.Type != interfaces.SCHEDULE_TYPE_CRON {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_ScheduleType).
				WithErrorDetails("Schedule type is invalid")
		}
		// 2.2 执行表达式
		if task.Schedule.Expression == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_NullParameter_ScheduleExpression).
				WithErrorDetails("Schedule expression is empty")
		} else {
			if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
				// 不为空时校验有效性,time duration的格式，单位支持 m - 分钟； h - 小时； d - 天
				err := validateDuration(ctx, task.Schedule.Expression, common.DurationDayHourMinuteRE,
					derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression, "schedule expression ", true)
				if err != nil {
					return err
				}
				durationV, _ := common.ParseDuration(task.Schedule.Expression, common.DurationDayHourMinuteRE, true)
				// 毫秒转秒
				stepV := int64(durationV/(time.Millisecond/time.Nanosecond)) / 1000
				if stepV > FIX_RATE_LIMIT {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression).
						WithErrorDetails("Schedule should be less than 24 days")
				}
			} else {
				// cron 表达式的校验。只支持6位，不支持年的指定。
				_, err := CronParser.Parse(task.Schedule.Expression)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression).
						WithErrorDetails(err.Error())
				}
			}

		}
		// 3. 索引库类型
		if task.StorageConfig.IndexBase == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_NullParameter_IndexBase).
				WithErrorDetails("IndexBase is null")
		}
		//4. 阻塞策略
		if task.DispatchConfig.BlockStrategy != "" {
			if _, ok := interfaces.BlockStrategy[task.DispatchConfig.BlockStrategy]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_BlockStrategy).
					WithErrorDetails("BlockStrategy type is invalid")

			}
		}

		//5. 调度策略
		if task.DispatchConfig.RouteStrategy != "" {
			if _, ok := interfaces.RouteStrategy[task.DispatchConfig.RouteStrategy]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_RouteStrategy).
					WithErrorDetails("RouteStrategy type is invalid")

			}
		}

		//6. 超时时间
		if !(task.DispatchConfig.TimeOut >= 0 && task.DispatchConfig.TimeOut <= 99999) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_TimeOut).
				WithErrorDetails("TimeOut is should be between 0 and 99999")
		}

		//7. 重试次数
		if !(task.DispatchConfig.FailRetryCount >= 0 && task.DispatchConfig.FailRetryCount <= 3) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InvalidParameter_FailRetryCount).
				WithErrorDetails("FailRetryCount should be between 0 and 9999")
		}
	}
	return nil
}

// 校验的导入模式
func validateImportMode(ctx context.Context, mode string) *rest.HTTPError {
	switch mode {
	case interfaces.ImportMode_Normal,
		interfaces.ImportMode_Ignore,
		interfaces.ImportMode_Overwrite:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_InvalidParameter_ImportMode).
			WithErrorDetails("The import_mode value can be 'overwrite', 'normal', 'ignore'")
	}

	return nil
}

func validateModelID(ctx context.Context, modelID string, builtin uint8) error {
	if modelID != "" {
		if builtin == interfaces.Builtin {
			// 内置视图 id 校验，只包含小写英文字母和数字和下划线(_)和连字符(-)，允许以下划线开头，不能超过40个字符
			re := regexp2.MustCompile(interfaces.RegexPattern_Builtin_ID, regexp2.RE2)
			match, err := re.MatchString(modelID)
			if err != nil || !match {
				errDetails := `The view id can contain only lowercase letters, digits and underscores(_),
			it cannot exceed 40 characters`
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_ID).
					WithErrorDetails(errDetails)
			}
		} else {
			// 非内置视图校验数据视图 id，只包含小写英文字母和数字和下划线(_)和连字符(-)，且不能以下划线开头，不能超过40个字符
			re := regexp2.MustCompile(interfaces.RegexPattern_NonBuiltin_ID, regexp2.RE2)
			match, err := re.MatchString(modelID)
			if err != nil || !match {
				errDetails := `The view id can contain only lowercase letters, digits and underscores(_),
			it cannot start with underscores and cannot exceed 40 characters`
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_ID).
					WithErrorDetails(errDetails)
			}
		}
	}

	return nil
}
