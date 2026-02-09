// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

// 链路模型
func Test_ValidateTraceModel_ValidateTraceModelsWhenCreate(t *testing.T) {
	Convey("Test validateTraceModelsWhenCreate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("Invalid models, caused by the error from func `validateTraceModelsWithName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateTraceModelsWithName, expectedErr)
			defer patch.Reset()

			err := validateTraceModelsWhenCreate(testCtx, handler, []interfaces.TraceModel(nil))
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid model, caused by the error from func `validateTraceModelsWithoutName`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateTraceModelsWithName, expectedErr)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWithoutName, expectedErr)
			defer patch2.Reset()

			err := validateTraceModelsWhenCreate(testCtx, handler, []interfaces.TraceModel(nil))
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid model", func() {
			patch1 := ApplyFuncReturn(validateTraceModelsWithName, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWithoutName, nil)
			defer patch2.Reset()

			err := validateTraceModelsWhenCreate(testCtx, handler, []interfaces.TraceModel(nil))
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateTraceModelWhenUpdate(t *testing.T) {
	Convey("Test validateTraceModelWhenUpdate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("Invalid model, caused by the error from func `validateTraceModelsWithName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateTraceModelsWithName, expectedErr)
			defer patch.Reset()

			err := validateTraceModelWhenUpdate(testCtx, handler, "1", interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid model, caused by the error from func `validateTraceModelsWithoutName`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateTraceModelsWithName, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWithoutName, expectedErr)
			defer patch2.Reset()

			err := validateTraceModelWhenUpdate(testCtx, handler, "1", interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid model", func() {
			patch1 := ApplyFuncReturn(validateTraceModelsWithName, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateTraceModelsWithoutName, nil)
			defer patch2.Reset()

			err := validateTraceModelWhenUpdate(testCtx, handler, "1", interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateTraceModelsWithName(t *testing.T) {
	Convey("Test validateTraceModelsWithName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tms := dmock.NewMockTraceModelService(mockCtrl)
		hydra := rmock.NewMockHydra(mockCtrl)

		handler := MockNewTraceModelRestHandler(appSetting, hydra, tms)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		Convey("Invalid models, caused by the error from func `validateObjectName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			models := make([]interfaces.TraceModel, 1)
			err := validateTraceModelsWithName(testCtx, handler, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, because the model name is not unique", func() {
			models := []interfaces.TraceModel{
				{Name: "m1"},
				{Name: "m1"},
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_NotUniqueInBatch_ModelName).
				WithErrorDetails(fmt.Sprintf("The trace model named %v is not unique among all the trace models to be created or imported", models[0].Name))

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			err := validateTraceModelsWithName(testCtx, handler, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, caused by the error from tmService method `GetSimpleTraceModelMapByNames`", func() {
			models := []interfaces.TraceModel{
				{Name: "m1"},
			}
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			tms.EXPECT().GetSimpleTraceModelMapByNames(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			err := validateTraceModelsWithName(testCtx, handler, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, because some model name have already exist", func() {
			models := []interfaces.TraceModel{
				{Name: "m1"},
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
				WithErrorDetails(fmt.Sprintf("The trace model whose name is in %v already exist in the database!", []string{models[0].Name}))
			expectedModelMap := map[string]interfaces.TraceModel{"m1": {}}

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			tms.EXPECT().GetSimpleTraceModelMapByNames(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)

			err := validateTraceModelsWithName(testCtx, handler, models)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Valid models", func() {
			models := []interfaces.TraceModel{
				{Name: "m1"},
			}

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			tms.EXPECT().GetSimpleTraceModelMapByNames(gomock.Any(), gomock.Any()).Return(nil, nil)

			err := validateTraceModelsWithName(testCtx, handler, models)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateTraceModelsWithoutName(t *testing.T) {
	Convey("Test validateTraceModelsWithoutName", t, func() {

		Convey("Invalid models, caused by the error from func `validateObjectTags`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectTags, expectedErr)
			defer patch.Reset()

			models := make([]interfaces.TraceModel, 1)
			err := validateTraceModelsWithoutName(testCtx, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, caused by the error from func `validateObjectComment`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, expectedErr)
			defer patch2.Reset()

			models := make([]interfaces.TraceModel, 1)
			err := validateTraceModelsWithoutName(testCtx, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, caused by the error from func `validateSpanConfig`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateSpanConfig, expectedErr)
			defer patch3.Reset()

			models := make([]interfaces.TraceModel, 1)
			err := validateTraceModelsWithoutName(testCtx, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid models, caused by the error from func `validateRelatedLogConfig`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateSpanConfig, nil)
			defer patch3.Reset()

			patch4 := ApplyFuncReturn(validateRelatedLogConfig, expectedErr)
			defer patch4.Reset()

			models := []interfaces.TraceModel{
				{
					EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
				},
			}
			err := validateTraceModelsWithoutName(testCtx, models)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid models", func() {
			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateSpanConfig, nil)
			defer patch3.Reset()

			models := make([]interfaces.TraceModel, 1)
			err := validateTraceModelsWithoutName(testCtx, models)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateSpanConfig(t *testing.T) {
	Convey("Test validateSpanConfig", t, func() {

		Convey("Invalid span_config with span_source_type is data_view, caused by the error from func `validateObjectName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			span := interfaces.SpanConfigWithDataView{}
			err := validateSpanConfig(testCtx, span, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid span_config with span_source_type is data_connection, caused by the error from func `validateObjectName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			span := interfaces.SpanConfigWithDataConnection{}
			err := validateSpanConfig(testCtx, span, interfaces.SOURCE_TYPE_DATA_CONNECTION)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid span_config, caused by the error from func `validateSpanBasicAttrs`", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectName, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateSpanBasicAttrs, expectedErr)
			defer patch2.Reset()

			span := interfaces.SpanConfigWithDataView{}
			err := validateSpanConfig(testCtx, span, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid span_config", func() {
			patch1 := ApplyFuncReturn(validateObjectName, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateSpanBasicAttrs, nil)
			defer patch2.Reset()

			span := interfaces.SpanConfigWithDataView{}
			err := validateSpanConfig(testCtx, span, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateSpanBasicAttrs(t *testing.T) {
	Convey("Test validateSpanBasicAttrs", t, func() {

		Convey("Invalid spanBasicAttrs, because the traceID corresponding field name is empty", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
				WithErrorDetails("The field name is empty")

			attrs := interfaces.SpanConfigWithDataView{}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because the spanID corresponding field name array is empty", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails("The field name array is empty")

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because the spanID corresponding field name array has some empty string", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails("Some field names in the field name array are empty")

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2", ""},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID configs is empty array", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails("The configuration array is empty")

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID has invalid precond", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(fmt.Sprintf("Condition err: %v", expectedErr))

			patch := ApplyFuncReturn(validPrecond, expectedErr)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID corresponding field name array is empty", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails("The field name array is empty")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID corresponding field name array has some empty string", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails("Some field names in the field name array are empty")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: make([]string, 1),
					},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID corresponding field name is empty string", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
				WithErrorDetails("The field name is empty")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the startTime corresponding field name is empty string", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
				WithErrorDetails("The field name is empty")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the startTime corresponding time format is invalid", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
				WithErrorDetails(fmt.Sprintf("The time format is invalid, valid time format is in %v", interfaces.VALID_TIME_FORMATS))

			patch1 := ApplyFuncReturn(validPrecond, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(isValidTimeFormat, false)
			defer patch2.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the serviceName corresponding field name is empty string", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
				WithErrorDetails("The field name is empty")

			patch1 := ApplyFuncReturn(validPrecond, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(isValidTimeFormat, true)
			defer patch2.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the endTime corresponding time format is invalid", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
				WithErrorDetails(fmt.Sprintf("The time format is invalid, valid time format is in %v", interfaces.VALID_TIME_FORMATS))

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName:   "f7",
					FieldFormat: "invalid",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the duration corresponding unit is invalid", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
				WithErrorDetails(fmt.Sprintf("The duration unit is invalid, valid duration unit is in %v", interfaces.VALID_DURATION_UNITS))

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName:   "f7",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
					FieldUnit: "invalid",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because both endTime and duration are configured", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig).
				WithErrorDetails("Both end_time and duration are configured")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName:   "f7",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
					FieldUnit: interfaces.MS,
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because both endTime and duration are not configured", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig).
				WithErrorDetails("Both end_time and duration are not configured")

			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Valid spanBasicAttrs", func() {
			patch := ApplyFuncReturn(validPrecond, nil)
			defer patch.Reset()

			attrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "f5",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName:   "f7",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
			}
			err := validateSpanBasicAttrs(testCtx, attrs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateTraceModel_ValidateRelatedLogConfig(t *testing.T) {
	Convey("Test validateRelatedLogConfig", t, func() {

		Convey("Invalid relatedLogConfig, caused by the error from func `validateObjectName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			relatedLog := interfaces.RelatedLogConfigWithDataView{}
			err := validateRelatedLogConfig(testCtx, relatedLog, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid relatedLogConfig, because the traceID corresponding field name is empty", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
				WithErrorDetails("The field name is empty")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			relatedLog := interfaces.RelatedLogConfigWithDataView{}
			err := validateRelatedLogConfig(testCtx, relatedLog, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid relatedLogConfig, because the spanID corresponding field name array is empty", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails("The field name array is empty")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			err := validateRelatedLogConfig(testCtx, relatedLog, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid relatedLogConfig, because the spanID corresponding field name array has some empty string", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails("Some field names in the field name array are empty")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f1", ""},
				},
			}
			err := validateRelatedLogConfig(testCtx, relatedLog, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid relatedLogConfig", func() {
			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f1"},
				},
			}
			err := validateRelatedLogConfig(testCtx, relatedLog, interfaces.SOURCE_TYPE_DATA_VIEW)
			So(err, ShouldBeNil)
		})
	})
}
