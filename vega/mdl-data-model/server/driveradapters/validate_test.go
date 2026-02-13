// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
)

var (
	StepsMap = map[string]string{
		"5m":  "5m",
		"10m": "10m",
		"15m": "15m",
		"20m": "20m",
		"30m": "30m",
		"1h":  "1h",
		"2h":  "2h",
		"3h":  "3h",
		"6h":  "6h",
		"12h": "12h",
		"1d":  "1d",
	}
)

func Test_Validate_ValidateObjectTags(t *testing.T) {
	Convey("Test validateObjectTags", t, func() {

		Convey("Validate failed, because the number of tags exceeds the upper limit", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_CountExceeded_Tags).
				WithErrorDetails(fmt.Sprintf("The length of the tag array exceeds %v", interfaces.TAGS_MAX_NUMBER))

			httpErr := validateObjectTags(testCtx, []string{"a", "b", "c", "d", "e", "f"})
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, because some tags are null", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_NullParameter_Tag)

			httpErr := validateObjectTags(testCtx, []string{"a", "", "c"})
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, because some tags exceeds the upper length limit", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_LengthExceeded_Tag).
				WithErrorDetails(fmt.Sprintf("The length of some tags in the tag array exceeds %d", interfaces.OBJECT_NAME_MAX_LENGTH))

			httpErr := validateObjectTags(testCtx, []string{"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRST"})
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, because some tags have special character", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_Tag).
				WithErrorDetails(fmt.Sprintf("The tag contains special characters, such as %s", interfaces.NAME_INVALID_CHARACTER))

			httpErr := validateObjectTags(testCtx, []string{"%[]#"})
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate succeed", func() {
			httpErr := validateObjectTags(testCtx, []string{"test"})
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateObjectComment(t *testing.T) {
	Convey("Test validateObjectComment", t, func() {

		Convey("Validate failed, because the comment exceeds the upper length limit", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_LengthExceeded_Comment).
				WithErrorDetails(fmt.Sprintf("The length of the comment exceeds %v", interfaces.COMMENT_MAX_LENGTH))

			str := ""
			for i := 0; i < interfaces.COMMENT_MAX_LENGTH+10; i++ {
				str += "a"
			}

			httpErr := validateObjectComment(testCtx, str)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate succeed", func() {
			httpErr := validateObjectComment(testCtx, "test")
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidatePaginationQueryParameters(t *testing.T) {
	Convey("Test validatePaginationQueryParameters", t, func() {

		Convey("Validate failed, because the offset cannot be converted to int", func() {
			offset := "a"
			limit := "1000"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Offset)
		})

		Convey("Validate failed, because the offset is not greater than MIN_OFFSET", func() {
			offset := "-1"
			limit := "1000"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Offset)
		})

		Convey("Validate failed, because the limit cannot be converted to int", func() {
			offset := "0"
			limit := "a"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Limit)
		})

		Convey("Validate failed, because the limit is not in the range of [MIN_LIMIT,MAX_LIMIT]", func() {
			offset := "0"
			limit := "1100"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Limit)
		})

		Convey("Validate failed, because the sort type does not belong to any item in set METRIC_MODEL_SORT", func() {
			offset := "0"
			limit := "800"
			sort := "update_time1"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Sort)
		})

		Convey("Validate failed, because the sort direction is not DESC or ASC", func() {
			offset := "0"
			limit := "800"
			sort := "update_time"
			direction := "abc"
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_Direction)
		})

		Convey("Validate succeed", func() {
			offset := "0"
			limit := "800"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION
			supportedSortTypes := interfaces.METRIC_MODEL_SORT

			_, res := validatePaginationQueryParameters(testCtx, offset, limit, sort, direction, supportedSortTypes)
			So(res, ShouldBeEmpty)
		})
	})
}

func Test_Validate_ValidateNameandNamePattern(t *testing.T) {
	Convey("Test validateNameandNamePattern", t, func() {

		Convey("Validate failed, because the name_pattern and name are passed in at the same time", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_ConflictParameter_NameAndNamePatternCoexist).
				WithErrorDetails("Parameters name_pattern and name are passed in at the same time")

			httpErr := validateNameandNamePattern(testCtx, "1", "2")
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate succeed", func() {
			httpErr := validateNameandNamePattern(testCtx, "1", "")
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateDataTagName(t *testing.T) {
	Convey("Test validateDataTagName", t, func() {

		Convey("Validate failed, because data tag name is null", func() {
			res := validateDataTagName(testCtx, "")

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_DataTagName)
		})

		Convey("Validate failed, because the length of the data tag name exceeds the limit", func() {
			res := validateDataTagName(testCtx, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRST")

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_DataTagName)
		})

		Convey("Validate failed, because the data tag name contains special characters", func() {
			res := validateDataTagName(testCtx, "a*")

			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_DataTagName)
		})

		Convey("Validate succeed", func() {
			res := validateDataTagName(testCtx, "a")
			So(res, ShouldEqual, nil)
		})
	})
}

func Test_Validate_ValidateMeasureName(t *testing.T) {
	Convey("Test validateMeasureName", t, func() {

		Convey("Validate failed, because measure name length > 40", func() {
			res := validateMeasureName(testCtx, "111111111111111111111111111111111111111111111111111111111")
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_LengthExceeded_MeasureName)
		})

		Convey("Validate failed, because measure name is not begin with __m.", func() {
			res := validateMeasureName(testCtx, "111111111111111111111111111111")
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureName)
		})

		Convey("Validate failed, because measure name is begin with __m._", func() {
			res := validateMeasureName(testCtx, "__m._11111111111111111111111111111111")
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureName)
		})

		Convey("Validate failed, because regex Compile failed", func() {
			res := validateMeasureName(testCtx, "__m._11111111111111111111111")
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_MeasureName)
		})

		Convey("Validate succeed", func() {
			res := validateMeasureName(testCtx, "__m.a")
			So(res, ShouldEqual, nil)
		})
	})
}

func Test_Validate_ValidateTasks(t *testing.T) {
	Convey("Test validateTask", t, func() {
		c := context.Background()

		common.PersistStepsMap = StepsMap

		// Convey("Validate failed, because task name is null", func() {
		// 	task := interfaces.MetricTask{
		// 		TaskName: "",
		// 	}

		// 	res := validateTask(c, interfaces.PROMQL, task)
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_TaskName)
		// })

		Convey("Validate failed, because task name length > 40", func() {
			task := interfaces.MetricTask{
				TaskName: "1111111111111111111111111111111111111111111",
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_LengthExceeded_TaskName)
		})

		Convey("Validate failed, because schedule is null", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_Schedule)
		})

		Convey("Validate failed, because schedule type is null", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: "", Expression: "1m"},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_ScheduleType)
		})

		Convey("Validate failed, because schedule type is not support", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: "fixed", Expression: "1m"},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_ScheduleType)
		})

		Convey("Validate failed, because schedule expression is null", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: ""},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_ScheduleExpression)
		})

		Convey("Validate failed, because schedule fix rate expression is invalid", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1M"},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_ScheduleExpression)
		})

		Convey("Validate failed, because schedule cron expression is invalid", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_CRON, Expression: "1-2 * * * * ? *"},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_ScheduleExpression)
		})

		// Convey("Validate failed, because filter operation is null", func() {
		// 	tasks := []interfaces.MetricTask{
		// 		{
		// 			TaskName: "task",
		// 			Schedule: interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
		// 			Filters:  []interfaces.Filter{{Name: "a"}},
		// 		},
		// 	}

		// 	res := validateTask(c, interfaces.PROMQL, tasks)
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_NullParameter_FilterValue)
		// })

		Convey("Validate failed, because query type is promql and time windows is not empty", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"3m"},
			}

			res := validateTask(c, interfaces.PROMQL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_TimeWindows)
		})

		Convey("Validate failed, because query type is dsl and time windows is empty", func() {
			task := interfaces.MetricTask{
				TaskName: "task",
				Schedule: interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_TimeWindows)
		})

		Convey("Validate failed, because query type is dsl and time windows is null", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{""},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_TimeWindows)
		})

		Convey("Validate failed, because time windows is invalid", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"1M"},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_TimeWindows)
		})

		Convey("Validate failed, because step is null", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"1m"},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_Step)
		})

		Convey("Validate failed, because step is invalid", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"1m"},
				Steps:       []string{"1M"},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_Step)
		})

		Convey("Validate failed, because index base is null", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"1m"},
				Steps:       []string{"5m"},
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_NullParameter_IndexBase)
		})

		Convey("Validate failed, because retrace duration is invalid", func() {
			task := interfaces.MetricTask{
				TaskName:        "task",
				Schedule:        interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows:     []string{"1m"},
				Steps:           []string{"5m"},
				IndexBase:       "base1",
				RetraceDuration: "1M",
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration)
		})

		// Convey("Validate failed, because duplicate task name", func() {
		// 	task := interfaces.MetricTask{
		// 		TaskName:    "task",
		// 		Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
		// 		TimeWindows: []string{"1m"},
		// 		Steps:       []string{"5m"},
		// 		IndexBase:   "base1",
		// 	}

		// 	res := validateTask(c, interfaces.DSL, task)
		// 	So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		// 	So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_Duplicated_TaskName)
		// })

		Convey("Validate failed, because task comment > 255", func() {
			task := interfaces.MetricTask{
				TaskName:    "task",
				Schedule:    interfaces.Schedule{Type: interfaces.SCHEDULE_TYPE_FIXED, Expression: "1m"},
				TimeWindows: []string{"1m"},
				Steps:       []string{"5m"},
				IndexBase:   "base1",
				Comment: `111111111111111111111111111111111111111111111111111111111111111111111111111111111111111
					1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111
					1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111
					11111111111111111111111111111111`,
			}

			res := validateTask(c, interfaces.DSL, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_LengthExceeded_TaskComment)
		})
	})
}

func Test_Validate_ValidateEventTask(t *testing.T) {
	Convey("Test ValidateEventTask", t, func() {

		Convey("task is null", func() {
			task := interfaces.EventTask{}
			err := validateEventTask(testCtx, task)
			So(err, ShouldBeNil)
		})

		Convey("schedule is null", func() {
			task := interfaces.EventTask{
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_NullParameter_Schedule)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Schedule is null")
		})
		Convey("Schedule type is empty", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "",
					Expression: "5m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_NullParameter_ScheduleType)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Schedule type is empty")
		})
		Convey("Schedule type is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "111",
					Expression: "5m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_ScheduleType)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Schedule type is invalid")
		})

		Convey("Schedule expression is empty", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_NullParameter_ScheduleExpression)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Schedule expression is empty")
		})

		Convey("Validate failed, because schedule fix rate expression is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "11a",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression)
		})

		Convey("Validate failed, because schedule cron expression is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "CRON",
					Expression: "1-2 * * * * ? *",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}

			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression)
		})

		Convey("Schedule lager than 24 days", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "99999d",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "xxx",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_ScheduleExpression)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "Schedule should be less than 24 days")

		})
		Convey("IndexBase is null", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "10m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase: "",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_NullParameter_IndexBase)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "IndexBase is null")

		})

		Convey("BlockStrategy type is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "10m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:  "xxx",
					DataViewID: "yyy",
				},
				DispatchConfig: interfaces.DispatchConfig{
					BlockStrategy: "11",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_BlockStrategy)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "BlockStrategy type is invalid")

		})
		Convey("RouteStrategy type is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "10m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:  "xxx",
					DataViewID: "yyy",
				},
				DispatchConfig: interfaces.DispatchConfig{
					RouteStrategy: "11",
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_RouteStrategy)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "RouteStrategy type is invalid")

		})
		Convey("TimeOut is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "10m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:  "xxx",
					DataViewID: "yyy",
				},
				DispatchConfig: interfaces.DispatchConfig{
					TimeOut: 1000000,
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_TimeOut)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "TimeOut is should be between 0 and 99999")

		})
		Convey("FailRetryCount is invalid", func() {
			task := interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: "10m",
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:  "xxx",
					DataViewID: "yyy",
				},
				DispatchConfig: interfaces.DispatchConfig{
					FailRetryCount: 10000,
				},
			}
			res := validateEventTask(testCtx, task)
			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_EventModel_InvalidParameter_FailRetryCount)
			So(res.(*rest.HTTPError).BaseError.ErrorDetails, ShouldEqual, "FailRetryCount should be between 0 and 9999")

		})
	})

}

func Test_Validate_IsValidTimeFormat(t *testing.T) {
	Convey("Test isValidTimeFormat", t, func() {
		Convey("Invalid", func() {
			flag := isValidTimeFormat("invalid")
			So(flag, ShouldBeFalse)
		})

		Convey("Valid", func() {
			flag := isValidTimeFormat(interfaces.UNIX_MILLIS)
			So(flag, ShouldBeTrue)
		})
	})
}

func Test_Validate_IsValidDurationUnit(t *testing.T) {
	Convey("Test isValidDurationUnit", t, func() {
		Convey("Invalid", func() {
			flag := isValidDurationUnit("invalid")
			So(flag, ShouldBeFalse)
		})

		Convey("Valid", func() {
			flag := isValidDurationUnit(interfaces.MS)
			So(flag, ShouldBeTrue)
		})
	})
}

func Test_Validate_ValidPrecond(t *testing.T) {
	Convey("Test ValidPrecond", t, func() {
		Convey("Invalid precond, because precond.Name is empty string", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: dcond.OperationEq,
						Name:      "",
						ValueOptCfg: interfaces.ValueOptCfg{
							Value: "v1",
						},
					},
				},
			}
			expectedErr := errors.New("name is empty")

			err := validPrecond(precond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid precond, because precond.ValueFrom is invalid", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: "invalid",
							Value:     "v1",
						},
					},
				},
			}
			expectedErr := fmt.Errorf("invalid value_from, valid value_from is in %v", interfaces.VALID_PRECONDITION_VALUE_FROM)

			err := validPrecond(precond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid precond, because precond.Value is invalid", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: dcond.ValueFrom_Field,
							Value:     1,
						},
					},
				},
			}
			expectedErr := errors.New("value_from is field, but value is invalid")

			err := validPrecond(precond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid precond, because precond.Operation is invalid", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: dcond.ValueFrom_Field,
							Value:     "f2",
						},
					},
					{
						Operation: "invalid",
					},
				},
			}
			expectedErr := fmt.Errorf("invalid operation, valid operation is in %v", interfaces.VALID_PRECONDITION_OPERATIONS)

			err := validPrecond(precond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid precond", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: dcond.ValueFrom_Field,
							Value:     "f2",
						},
					},
				},
			}

			err := validPrecond(precond)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_IsValidValueFrom(t *testing.T) {
	Convey("Test isValidValueFrom", t, func() {
		Convey("Invalid", func() {
			flag := isValidValueFrom("invalid")
			So(flag, ShouldBeFalse)
		})

		Convey("Valid", func() {
			flag := isValidValueFrom(dcond.ValueFrom_Const)
			So(flag, ShouldBeTrue)
		})
	})
}

func Test_Validate_ValidateSpanSourceType(t *testing.T) {
	Convey("Test SpanSourceType", t, func() {

		Convey("Valid spanSourceType, field value is data_view", func() {
			err := validateSpanSourceType(testCtx, "data_view")
			So(err, ShouldBeNil)
		})

		Convey("Valid spanSourceType, field value is data_connection", func() {
			err := validateSpanSourceType(testCtx, "data_connection")
			So(err, ShouldBeNil)
		})

		Convey("Invalid spanSourceType, field value is xxx", func() {
			err := validateSpanSourceType(testCtx, "xxx")
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails("span_source_type is invalid, valid span_source_type is data_view or data_connection"))
		})
	})
}

func Test_Validate_ValidateObjectiveModel(t *testing.T) {
	Convey("Test ValidateObjectiveModel", t, func() {

		common.PersistStepsMap = StepsMap

		Convey("When id is invalid", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelID:       "%%%",
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_InvalidParameter_ID)
		})

		Convey("When model name is empty", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName: "",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ModelName)
		})

		Convey("When model name is too long", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName: strings.Repeat("a", 129),
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_LengthExceeded_ModelName)
		})

		Convey("When objective type is empty", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: "",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveType)
		})

		Convey("When objective type is invalid", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: "invalid_type",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_UnsupportObjectiveType)
		})

		Convey("When objective config is empty", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:       "test-model",
					ObjectiveType:   interfaces.SLO,
					ObjectiveConfig: nil,
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveConfig)
		})

		Convey("When objective config marshal error for SLO type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:       "test-model",
					ObjectiveType:   interfaces.SLO,
					ObjectiveConfig: make(chan int), // Invalid type that will fail JSON marshaling
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("When objective config unmarshal error for SLO type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: map[string]interface{}{
						"objective": "invalid", // Invalid type that will fail unmarshaling
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("When objective is empty for SLO type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_Objective)
		})

		Convey("When objective is 0 for SLO type", func() {
			var objective float64
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_Objective)
		})

		Convey("When period is empty for SLO type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_Period)
		})

		Convey("When period is 0 for SLO type", func() {
			objective := float64(99)
			period := int64(0)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_Period)
		})
		Convey("When good metric model is empty for SLO type", func() {
			objective := float64(99)
			period := int64(90)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective:       &objective,
						Period:          &period,
						GoodMetricModel: nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_GoodMetricModel)
		})

		Convey("When good metric model id is empty for SLO type", func() {
			objective := float64(99)
			period := int64(90)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "",
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_GoodMetricModelID)
		})

		Convey("When total metric model is empty for SLO type", func() {
			objective := float64(99)
			period := int64(90)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "1",
						},
						TotalMetricModel: nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_TotalMetricModel)
		})

		Convey("When total metric model id is empty for SLO type", func() {
			objective := float64(99)
			period := int64(90)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "1",
						},
						TotalMetricModel: &interfaces.BundleMetricModel{
							ID: "",
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_TotalMetricModelID)
		})

		Convey("When status config is not null but status ranges is empty for SLO type", func() {
			objective := float64(99)
			period := int64(90)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.SLO,
					ObjectiveConfig: interfaces.SLOObjective{
						Objective: &objective,
						Period:    &period,
						GoodMetricModel: &interfaces.BundleMetricModel{
							ID: "1",
						},
						TotalMetricModel: &interfaces.BundleMetricModel{
							ID: "2",
						},
						StatusConfig: &interfaces.ObjectiveStatusConfig{
							Ranges: []interfaces.Range{
								{
									Status: "",
								},
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges)
		})

		Convey("When objective config marshal error for KPI type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:       "test-model",
					ObjectiveType:   interfaces.KPI,
					ObjectiveConfig: make(chan int), // Invalid type that will fail JSON marshaling
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("When objective config unmarshal error for KPI type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: map[string]interface{}{
						"objective": "invalid", // Invalid type that will fail unmarshaling
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig)
		})

		Convey("When objective is empty for KPI type", func() {
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_Objective)
		})

		Convey("When objective is 0 for KPI type", func() {
			var objective float64
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_Objective)
		})

		Convey("When objective unit is empty for KPI type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveUnit)
		})

		Convey("When objective unit is invalid for KPI type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "invalid_unit",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveUnit)
		})
		Convey("When comprehensive metric models is nil for KPI type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective:                 &objective,
						Unit:                      "%",
						ComprehensiveMetricModels: nil,
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModels)
		})

		Convey("When comprehensive metric model id is empty for KPI type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "",
								Weight: nil,
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModelID)
		})

		Convey("When comprehensive metric model weight is nil for KPI type", func() {
			objective := float64(99)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: nil,
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveWeight)
		})

		Convey("When comprehensive metric model weight is negative for KPI type", func() {
			objective := float64(99)
			weight := int64(-10)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight)
		})

		Convey("When comprehensive metric model weights sum is not 100 for KPI type", func() {
			objective := float64(99)
			weight1 := int64(40)
			weight2 := int64(40)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id-1",
								Weight: &weight1,
							},
							{
								ID:     "test-id-2",
								Weight: &weight2,
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight)
		})

		Convey("When additional metric model id is empty for KPI type", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
						AdditionalMetricModels: []interfaces.BundleMetricModel{
							{
								ID: "",
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_AdditionalMetricModelID)
		})

		Convey("When status ranges overlap for KPI type", func() {
			objective := float64(99)
			weight := int64(100)
			from1 := float64(0)
			to1 := float64(50)
			from2 := float64(40)
			to2 := float64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
						StatusConfig: &interfaces.ObjectiveStatusConfig{
							Ranges: []interfaces.Range{
								{
									Status: "good",
									From:   &from1,
									To:     &to1,
								},
								{
									Status: "bad",
									From:   &from2, // Overlaps with previous range
									To:     &to2,
								},
							},
						},
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges)
		})

		Convey("When tags exceed maximum allowed count", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
					Tags: []string{
						"tag1", "tag2", "tag3", "tag4", "tag5",
						"tag6", "tag7", "tag8", "tag9", "tag10",
						"tag11", "tag12", "tag13", "tag14", "tag15",
						"tag16", "tag17", "tag18", "tag19", "tag20",
						"tag21",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModel_CountExceeded_TagTotal)
		})

		Convey("When objective comment exceeds maximum allowed length", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
					Comment: strings.Repeat("a", 1025),
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_LengthExceeded_Comment)
		})

		Convey("When task is empty", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: nil,
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_Task)
		})

		Convey("When schedule type is empty", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type: "",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ScheduleType)
		})

		Convey("When schedule type is invalid", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type: "invalid_type",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleType)
		})

		Convey("When schedule expression is empty", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_FIXED,
						Expression: "",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_ScheduleExpression)
		})

		Convey("When schedule expression is invalid for fixed type", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_FIXED,
						Expression: "invalid",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression)
		})

		Convey("When schedule expression is invalid for cron type", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "invalid cron",
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression)
		})

		Convey("When schedule expression is invalid for cron type with invalid year field", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * * 2024", // Invalid 7-field cron with year
					},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression)
		})

		Convey("When retrace duration is invalid", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					RetraceDuration: "invalid",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration)
		})

		Convey("When retrace duration parse fails", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					RetraceDuration: "1y", // Invalid duration unit
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration)
		})

		Convey("When task steps is empty", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps: []string{},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_Step)
		})

		Convey("When task step is invalid", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps: []string{"invalid_step"},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_Step)
		})

		Convey("When task steps have duplicates", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps: []string{"1h", "1h"},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_Duplicated_TaskStep)
		})

		Convey("When task step has invalid duration format", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps: []string{"invalid_duration"},
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_Step)
		})

		Convey("When task retrace duration exceeds maximum allowed", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "10001h",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration)
		})

		Convey("When index base is empty", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_NullParameter_IndexBase)
		})

		Convey("When task comment length exceeds limit", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "test-index",
					Comment:         strings.Repeat("a", 1025),
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_ObjectiveModel_LengthExceeded_TaskComment)
		})

		Convey("When all parameters are valid", func() {
			objective := float64(99)
			weight := int64(100)
			model := &interfaces.ObjectiveModel{
				ObjectiveModelInfo: interfaces.ObjectiveModelInfo{
					ModelName:     "test-model",
					ObjectiveType: interfaces.KPI,
					ObjectiveConfig: interfaces.KPIObjective{
						Objective: &objective,
						Unit:      "%",
						ComprehensiveMetricModels: []interfaces.ComprehensiveMetricModel{
							{
								ID:     "test-id",
								Weight: &weight,
							},
						},
					},
				},
				Task: &interfaces.MetricTask{
					Schedule: interfaces.Schedule{
						Type:       interfaces.SCHEDULE_TYPE_CRON,
						Expression: "* * * * * *",
					},
					Steps:           []string{"1h"},
					RetraceDuration: "24h",
					IndexBase:       "test-index",
					Comment:         "test comment",
				},
			}
			err := ValidateObjectiveModel(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}
