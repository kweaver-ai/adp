// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	cond "uniquery/common/condition"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func TestGetSpanList2(t *testing.T) {
	Convey("Test GetSpanList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Get failed, caused by the error from method 'RetrieveSingleViewData'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, expectedErr)

			_, _, err := dvAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed, but result is nil", func() {
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, nil)

			_, total, err := dvAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{
					Total: int64(1),
					Datas: make([]*ast.Node, 1),
				}, nil)

			_, total, err := dvAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
		})
	})
}

func TestGetSpan2(t *testing.T) {
	Convey("Test GetSpanMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Get failed, caused by the spanID fails to be split by the default separator", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SplitSpanIDFailed).
				WithErrorDetails(fmt.Sprintf("The spanID fails to be split by the default separator %v, please pass spanID in the correct format or check whether spanID is configured correctly in the trace model", interfaces.DEFAULT_SEPARATOR))
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			params := interfaces.SpanQueryParams{
				TraceID: "1",
				SpanID:  "1" + interfaces.DEFAULT_SEPARATOR + "1",
			}
			_, err := dvAdapter.GetSpan(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'RetrieveSingleViewData'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, expectedErr)

			params := interfaces.SpanQueryParams{
				TraceID: "1",
				SpanID:  "1",
			}
			_, err := dvAdapter.GetSpan(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the span was not found", func() {
			params := interfaces.SpanQueryParams{
				TraceID: "1",
				SpanID:  "1",
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_SpanNotFound).
				WithErrorDetails("The span whose spanID equals " + params.SpanID + " was not found!")
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, nil)

			_, err := dvAdapter.GetSpan(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig: interfaces.SpanConfigWithDataView{
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{
					Total: int64(1),
					Datas: make([]*ast.Node, 1),
				}, nil)

			params := interfaces.SpanQueryParams{
				TraceID: "1",
				SpanID:  "1",
			}
			_, err := dvAdapter.GetSpan(testCtx, model, params)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetSpanRelatedLogList2(t *testing.T) {
	Convey("Test GetSpanRelatedLogList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Get failed, caused by the spanID fails to be split by the default separator", func() {
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_SplitSpanIDFailed).
				WithErrorDetails(fmt.Sprintf("The spanID fails to be split by the default separator %v, please pass spanID in the correct format or check whether spanID is configured correctly in the trace model", interfaces.DEFAULT_SEPARATOR))
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
			}

			_, _, err := dvAdapter.GetSpanRelatedLogList(testCtx, model, interfaces.RelatedLogListQueryParams{
				TraceID: "1",
				SpanID:  "1$_$1",
			})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'RetrieveSingleViewData'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
					TraceID: interfaces.TraceIDConfig{
						FieldName: "traceID",
					},
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, expectedErr)

			_, _, err := dvAdapter.GetSpanRelatedLogList(testCtx, model, interfaces.RelatedLogListQueryParams{
				TraceID: "1",
				SpanID:  "1",
			})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed, but return empty result", func() {
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
					TraceID: interfaces.TraceIDConfig{
						FieldName: "traceID",
					},
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, nil)

			_, total, err := dvAdapter.GetSpanRelatedLogList(testCtx, model, interfaces.RelatedLogListQueryParams{
				TraceID: "1",
				SpanID:  "1",
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
					TraceID: interfaces.TraceIDConfig{
						FieldName: "traceID",
					},
					SpanID: interfaces.SpanIDConfig{
						FieldNames: []string{"spanID"},
					},
				},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{
					Total: int64(1),
					Datas: make([]*ast.Node, 1),
				}, nil)

			_, total, err := dvAdapter.GetSpanRelatedLogList(testCtx, model, interfaces.RelatedLogListQueryParams{
				TraceID: "1",
				SpanID:  "1",
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
		})

	})
}

func TestGetSpanMap2(t *testing.T) {
	Convey("Test GetSpanMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Get failed, caused by the error from method 'RetrieveSingleViewData'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				SpanConfig:     interfaces.SpanConfigWithDataView{},
			}

			mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ViewInternalResponse{}, expectedErr)

			params := interfaces.TraceQueryParams{}
			_, _, err := dvAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		// Convey("Get succeed", func() {
		// 	model := interfaces.TraceModel{
		// 		SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
		// 		SpanConfig:     interfaces.SpanConfigWithDataView{},
		// 	}

		// 	mockDVService.EXPECT().RetrieveSingleViewData(gomock.Any(), gomock.Any()).
		// 		Return(&interfaces.ViewInternalResponse{}, nil)
		// 	patch := ApplyPrivateMethod(reflect.TypeOf(dataViewAdapter{}), "clearScrollIDs",
		// 		func(scrollIDs []string) {})
		// 	defer patch.Reset()

		// 	_, _, err := dvAdapter.GetSpanMap(testCtx,model, "")
		// 	So(err, ShouldBeNil)
		// })

	})
}

func TestGetRelatedLogCountMap2(t *testing.T) {
	Convey("Test GetRelatedLogCountMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Get failed, caused by the error from method 'CountMultiFields'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
			}

			mockDVService.EXPECT().CountMultiFields(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, expectedErr)

			params := interfaces.TraceQueryParams{}
			_, err := dvAdapter.GetRelatedLogCountMap(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
				RelatedLogConfig:     interfaces.RelatedLogConfigWithDataView{},
			}

			mockDVService.EXPECT().CountMultiFields(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil)

			params := interfaces.TraceQueryParams{}
			_, err := dvAdapter.GetRelatedLogCountMap(testCtx, model, params)
			So(err, ShouldBeNil)
		})
	})
}

/*
	私有方法
*/

func TestExtractRawSpan(t *testing.T) {
	Convey("Test extractRawSpan", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Extract succeed", func() {
			precond := &cond.CondCfg{
				Operation: cond.OperationOr,
				SubConds: []*cond.CondCfg{
					{
						Operation: cond.OperationAnd,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationNotEq,
								Name:      "category",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "aaa",
								},
							},
							{
								Operation: cond.OperationEq,
								Name:      "Status.CodeDesc",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "ok",
								},
							},
						},
					},
					{
						Operation: cond.OperationRange,
						Name:      "__write_time",
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{"2024-06-18T06:50:49.903Z", "2024-06-18T07:50:49.904Z"},
						},
					},
				},
			}
			parentSpanIDconf := []interfaces.ParentSpanIDConfig{
				{
					Precond:    precond,
					FieldNames: []string{"category", "Status.CodeDesc"},
				},
			}

			spanConfig := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "traceID",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"spanID"},
				},
				ParentSpanID: parentSpanIDconf,
				Name: interfaces.NameConfig{
					FieldName: "spanName",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "startTime",
					FieldFormat: interfaces.UNIX_NANOS,
				},
				Duration: interfaces.DurationConfig{
					FieldName: "duration",
					FieldUnit: interfaces.NS,
				},
				Kind: interfaces.KindConfig{
					FieldName: "spanKind",
				},
				Status: interfaces.StatusConfig{
					FieldName: "status",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "service",
				},
			}

			astNode, _ := sonic.GetFromString(`{"category": "bbb", "Status": {"CodeDesc": "ok"}, "__write_time": "2024-06-18T08:50:49.904Z"}`)
			res := dvAdapter.extractRawSpan(spanConfig, &astNode, true)
			So(res, ShouldEqual, interfaces.AbstractSpan{
				ParentSpanID: "bbb" + interfaces.DEFAULT_SEPARATOR + "ok",
				Kind:         interfaces.SPAN_KIND_UNSPECIFIED,
				Status:       interfaces.SPAN_STATUS_UNSET,
			})
		})
	})
}

func TestParseParentSpanID(t *testing.T) {
	Convey("Test parseParentSpanID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Parse succeed", func() {
			precond := &cond.CondCfg{
				Operation: cond.OperationOr,
				SubConds: []*cond.CondCfg{
					{
						Operation: cond.OperationAnd,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationNotEq,
								Name:      "category",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "aaa",
								},
							},
							{
								Operation: cond.OperationEq,
								Name:      "Status.CodeDesc",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "ok",
								},
							},
						},
					},
					{
						Operation: cond.OperationRange,
						Name:      "__write_time",
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{"2024-06-18T06:50:49.903Z", "2024-06-18T07:50:49.904Z"},
						},
					},
				},
			}
			confs := []interfaces.ParentSpanIDConfig{
				{
					Precond:    precond,
					FieldNames: []string{"category", "Status.CodeDesc"},
				},
			}

			astNode, _ := sonic.GetFromString(`{"category": "bbb", "Status": {"CodeDesc": "ok"}, "__write_time": "2024-06-18T08:50:49.904Z"}`)
			res := dvAdapter.parseParentSpanID(confs, &astNode)
			So(res, ShouldEqual, "bbb"+interfaces.DEFAULT_SEPARATOR+"ok")
		})

		Convey("Parse failed", func() {
			precond := &cond.CondCfg{
				Operation: cond.OperationOr,
				SubConds: []*cond.CondCfg{
					{
						Operation: cond.OperationAnd,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationNotEq,
								Name:      "category",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "aaa",
								},
							},
							{
								Operation: cond.OperationEq,
								Name:      "Status.CodeDesc",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "ok",
								},
							},
						},
					},
					{
						Operation: cond.OperationRange,
						Name:      "__write_time",
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{"2024-06-18T06:50:49.903Z", "2024-06-18T07:50:49.904Z"},
						},
					},
				},
			}
			confs := []interfaces.ParentSpanIDConfig{
				{
					Precond:    precond,
					FieldNames: []string{"category", "Status.CodeDesc"},
				},
			}

			astNode, _ := sonic.GetFromString(`{"category": "aaa", "Status": {"CodeDesc": "ok"}, "__write_time": "2024-06-18T08:50:49.904Z"}`)
			res := dvAdapter.parseParentSpanID(confs, &astNode)
			So(res, ShouldBeEmpty)
		})
	})
}

func TestIsSatisfyCondition(t *testing.T) {
	Convey("Test isSatisfyCondition", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Satisfy", func() {
			precond := &cond.CondCfg{
				Operation: cond.OperationOr,
				SubConds: []*cond.CondCfg{
					{
						Operation: cond.OperationAnd,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationNotEq,
								Name:      "category",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "aaa",
								},
							},
							{
								Operation: cond.OperationEq,
								Name:      "Status.CodeDesc",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "ok",
								},
							},
						},
					},
					{
						Operation: cond.OperationRange,
						Name:      "__write_time",
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{"2024-06-18T06:50:49.903Z", "2024-06-18T07:50:49.904Z"},
						},
					},
				},
			}
			astNode, _ := sonic.GetFromString(`{"category": "bbb", "Status": {"CodeDesc": "ok"}, "__write_time": "2024-06-18T08:50:49.904Z"}`)
			_, ok := dvAdapter.isSatisfyCondition(precond, &astNode)
			So(ok, ShouldBeTrue)
		})

		Convey("Not satisfy", func() {
			precond := &cond.CondCfg{
				Operation: cond.OperationOr,
				SubConds: []*cond.CondCfg{
					{
						Operation: cond.OperationAnd,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationNotEq,
								Name:      "category",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "aaa",
								},
							},
							{
								Operation: cond.OperationEq,
								Name:      "Status.CodeDesc",
								ValueOptCfg: vopt.ValueOptCfg{
									ValueFrom: vopt.ValueFrom_Const,
									Value:     "ok",
								},
							},
						},
					},
					{
						Operation: cond.OperationRange,
						Name:      "__write_time",
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{"2024-06-18T06:50:49.903Z", "2024-06-18T07:50:49.904Z"},
						},
					},
				},
			}
			astNode, _ := sonic.GetFromString(`{"category": "aaa", "Status": {"CodeDesc": "ok"}, "__write_time": "2024-06-18T08:50:49.904Z"}`)
			_, ok := dvAdapter.isSatisfyCondition(precond, &astNode)
			So(ok, ShouldBeFalse)
		})
	})
}

func TestParseTimeAndDuration(t *testing.T) {
	Convey("Test parseTimeAndDuration", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("endTimeConf is nil", func() {
			spanConf := interfaces.SpanConfigWithDataView{
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "startTime",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				Duration: interfaces.DurationConfig{
					FieldName: "duration",
					FieldUnit: interfaces.MS,
				},
			}
			astNode, _ := sonic.GetFromString(`{"startTime": 1718721919000, "duration": 20}`)
			startTime, endTime, duration := dvAdapter.parseTimeAndDuration(spanConf, &astNode)
			So(startTime, ShouldEqual, int64(1718721919000*1000))
			So(endTime, ShouldEqual, int64(1718721919020*1000))
			So(duration, ShouldEqual, int64(20)*1000)
		})

		Convey("endTimeConf is not nil", func() {
			spanConf := interfaces.SpanConfigWithDataView{
				StartTime: interfaces.StartTimeConfig{
					FieldName:   "startTime",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName:   "endTime",
					FieldFormat: interfaces.UNIX_MILLIS,
				},
			}
			astNode, _ := sonic.GetFromString(`{"startTime": 1718721919000, "endTime": 1718721920000}`)
			startTime, endTime, duration := dvAdapter.parseTimeAndDuration(spanConf, &astNode)
			So(startTime, ShouldEqual, int64(1718721919000*1000))
			So(endTime, ShouldEqual, int64(1718721920000*1000))
			So(duration, ShouldEqual, int64(1000)*1000)
		})
	})
}

func TestModifySpanMap(t *testing.T) {
	Convey("Test modifySpanMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("rawSpan is nil", func() {
			l := &sync.Mutex{}
			astNodes := make([]*ast.Node, 1)
			spanConf := interfaces.SpanConfigWithDataView{}
			spanMap := make(map[string]*interfaces.BriefSpan_)

			patch := ApplyPrivateMethod(reflect.TypeOf(&dataViewAdapter{}), "extractRawSpan",
				func(spanConfig interfaces.SpanConfigWithDataView, astNode *ast.Node, simpleInfo bool) interfaces.AbstractSpan {
					return interfaces.AbstractSpan{}
				})
			defer patch.Reset()

			dvAdapter.modifySpanMap(testCtx, l, astNodes, spanConf, spanMap)
			So(len(spanMap), ShouldEqual, 1)
		})
	})
}

func TestGenSpanDetail(t *testing.T) {
	Convey("Test genSpanDetail", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		dvAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("rawSpan is nil", func() {
			res := dvAdapter.genSpanDetail(nil, interfaces.AbstractSpan{})
			So(len(res), ShouldEqual, 10)
		})
	})
}

func TestGenRelatedLogDetail(t *testing.T) {
	Convey("Test genRelatedLogDetail", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		tyAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("rawRelatedLog is nil", func() {
			res := tyAdapter.genRelatedLogDetail(nil, interfaces.AbstractRelatedLog{})
			So(len(res), ShouldEqual, 2)
		})
	})
}

func TestExtractRawRelatedLog(t *testing.T) {
	Convey("Test extractRawRelatedLog", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDVService := umock.NewMockDataViewService(mockCtrl)
		mockDSLAccess := umock.NewMockDslService(mockCtrl)

		tyAdapter := dataViewAdapter{
			dvService:  mockDVService,
			dslService: mockDSLAccess,
		}

		Convey("Two fields in the span configuration", func() {
			relatedLogConfig := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "traceID",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"spanID", "f1"},
				},
			}

			astNode, _ := sonic.GetFromString(`{"traceID": "bbb", "spanID": "ccc", "f1": "ddd"}`)

			res := tyAdapter.extractRawRelatedLog(relatedLogConfig, &astNode)
			So(res.TraceID, ShouldEqual, "bbb")
			So(res.SpanID, ShouldEqual, "ccc"+interfaces.DEFAULT_SEPARATOR+"ddd")
		})

		Convey("One field in the span configuration", func() {
			relatedLogConfig := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "traceID",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"spanID"},
				},
			}

			astNode, _ := sonic.GetFromString(`{"traceID": "bbb", "spanID": "ccc", "f1": "ddd"}`)

			res := tyAdapter.extractRawRelatedLog(relatedLogConfig, &astNode)
			So(res.TraceID, ShouldEqual, "bbb")
			So(res.SpanID, ShouldEqual, "ccc")
		})
	})
}
