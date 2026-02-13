// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/golang/mock/gomock"
	rest "github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func TestGetSpanList(t *testing.T) {
	Convey("Test GetSpanList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'convertQueryParams'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", expectedErr
				})
			defer patch.Reset()

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method 'GetDataConnectionByID'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", nil
				})
			defer patch.Reset()

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, false, expectedErr)

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the data connection was not found", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", nil
				})
			defer patch.Reset()

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, false, nil)

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
				WithErrorDetails(fmt.Sprintf("Data connection whose id equal to %d was not found", 1)))
		})

		Convey("Get failed, caused by the error from method 'processDataConnection'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return expectedErr
				})
			defer patch2.Reset()

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the error from the method 'GetTraceList'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceList",
				func(ctx context.Context, cfg TingYunClientConfig) (traceList []map[string]any, total int64, err error) {
					return []map[string]any{}, int64(0), expectedErr
				})
			defer patch3.Reset()

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryParams",
				func(ctx context.Context, params interfaces.SpanListQueryParams) (string, error) {
					return "", nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceList",
				func(ctx context.Context, cfg TingYunClientConfig) (traceList []map[string]any, total int64, err error) {
					return []map[string]any{}, int64(0), nil
				})
			defer patch3.Reset()

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			_, _, err := tyAdapter.GetSpanList(testCtx, model, interfaces.SpanListQueryParams{})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetSpan(t *testing.T) {
	Convey("Test GetSpan", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{}
			params := interfaces.SpanQueryParams{}
			_, err := tyAdapter.GetSpan(testCtx, model, params)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetSpanMap(t *testing.T) {
	Convey("Test GetSpanMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetDataConnectionByID'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{}
			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, false, expectedErr)

			params := interfaces.TraceQueryParams{}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the data connection was not found", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}
			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, false, nil)

			params := interfaces.TraceQueryParams{}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
				WithErrorDetails(fmt.Sprintf("Data connection whose id equal to %d was not found", 1)))
		})

		Convey("Get failed, caused by the error from method 'processDataConnection'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}
			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return expectedErr
				})
			defer patch.Reset()

			params := interfaces.TraceQueryParams{}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the error from the method 'GetTraceDetail'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceDetail",
				func(ctx context.Context, cfg TingYunClientConfig) (traceDetail map[string]any, isExist bool, err error) {
					return map[string]any{}, false, expectedErr
				})
			defer patch2.Reset()

			params := interfaces.TraceQueryParams{}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Get failed, caused by the trace was not found", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceDetail",
				func(ctx context.Context, cfg TingYunClientConfig) (traceDetail map[string]any, isExist bool, err error) {
					return map[string]any{}, false, nil
				})
			defer patch2.Reset()

			params := interfaces.TraceQueryParams{
				TraceID: "1",
			}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceNotFound).
				WithErrorDetails(fmt.Sprintf("The trace whose id equal to %v was not found in tingyun system", params.TraceID)))
		})

		Convey("Get failed, caused by the error from method 'parseRawTraceDetail'", func() {
			expectedErr := errors.New("some errors")
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "parseRawTraceDetail",
				func(ctx context.Context, parentSeqStr string, traceID string, rawTraceDetail map[string]any, briefSpanMap map[string]*interfaces.BriefSpan_, detailSpanMap map[string]interfaces.SpanDetail) error {
					return expectedErr
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceDetail",
				func(ctx context.Context, cfg TingYunClientConfig) (traceDetail map[string]any, isExist bool, err error) {
					return map[string]any{}, true, nil
				})
			defer patch3.Reset()

			params := interfaces.TraceQueryParams{
				TraceID: "1",
			}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						ID: "1",
					},
				},
			}

			mockDCAccess.EXPECT().GetDataConnectionByID(gomock.Any(), gomock.Any()).
				Return(&interfaces.DataConnection{}, true, nil)

			patch1 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "parseRawTraceDetail",
				func(ctx context.Context, parentSeqStr string, traceID string, rawTraceDetail map[string]any, briefSpanMap map[string]*interfaces.BriefSpan_, detailSpanMap map[string]interfaces.SpanDetail) error {
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "processDataConnection",
				func(ctx context.Context, conn *interfaces.DataConnection) error {
					return nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "getTraceDetail",
				func(ctx context.Context, cfg TingYunClientConfig) (traceDetail map[string]any, isExist bool, err error) {
					return map[string]any{}, true, nil
				})
			defer patch3.Reset()

			params := interfaces.TraceQueryParams{
				TraceID: "1",
			}
			_, _, err := tyAdapter.GetSpanMap(testCtx, model, params)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetRelatedLogCountMap(t *testing.T) {
	Convey("Test GetRelatedLogCountMap", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{}

			params := interfaces.TraceQueryParams{}
			_, err := tyAdapter.GetRelatedLogCountMap(testCtx, model, params)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetSpanRelatedLogList(t *testing.T) {
	Convey("Test GetSpanRelatedLogList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{}
			params := interfaces.RelatedLogListQueryParams{}

			_, _, err := tyAdapter.GetSpanRelatedLogList(testCtx, model, params)
			So(err, ShouldBeNil)
		})

	})
}

/*
	私有方法
*/

func TestConvertQueryParams(t *testing.T) {
	Convey("Test convertQueryParams", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Convert failed, caused by the invalid sort", func() {
			params := interfaces.SpanListQueryParams{
				PaginationQueryParams: interfaces.PaginationQueryParams{
					Sort: "invalid",
				},
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Sort).
				WithErrorDetails(fmt.Sprintf("TingYun does not support this sort field %v", params.Sort))

			_, err := tyAdapter.convertQueryParams(testCtx, params)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Convert failed, caused by the error from method 'convertQueryCondition'", func() {
			params := interfaces.SpanListQueryParams{
				Condition: &cond.CondCfg{},
				PaginationQueryParams: interfaces.PaginationQueryParams{
					Sort: interfaces.DEFAULT_SORT,
				},
			}
			expectedErr := errors.New("some errors")

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryCondition",
				func(ctx context.Context, cond *cond.CondCfg) (queryParams string, err error) {
					return "", expectedErr
				})
			defer patch.Reset()

			_, err := tyAdapter.convertQueryParams(testCtx, params)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("Convert succeed", func() {
			params := interfaces.SpanListQueryParams{
				Condition: &cond.CondCfg{},
				PaginationQueryParams: interfaces.PaginationQueryParams{
					Sort: interfaces.DEFAULT_SORT,
				},
			}

			patch := ApplyPrivateMethod(reflect.TypeOf(&tingYunwAdapter{}), "convertQueryCondition",
				func(ctx context.Context, cond *cond.CondCfg) (queryParams string, err error) {
					return "", nil
				})
			defer patch.Reset()

			queryParams, err := tyAdapter.convertQueryParams(testCtx, params)
			So(err, ShouldBeNil)
			So(queryParams, ShouldNotBeEmpty)
		})
	})
}

func TestConvertQueryCondition(t *testing.T) {
	Convey("Test convertQueryCondition", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Convert failed, caused by the invalid operation", func() {
			cond := &cond.CondCfg{
				Operation: "invalid",
			}
			expectedErr := fmt.Errorf("the tingyun does not support operation %v", cond.Operation)

			_, err := tyAdapter.convertQueryCondition(testCtx, cond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Convert failed, caused by the invalid valueFrom", func() {
			cond := &cond.CondCfg{
				Operation: cond.OperationEq,
				ValueOptCfg: vopt.ValueOptCfg{
					ValueFrom: "invalid",
				},
			}
			expectedErr := fmt.Errorf("the tingyun condition does not support value from type(%s)", cond.ValueFrom)

			_, err := tyAdapter.convertQueryCondition(testCtx, cond)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Convert succeed", func() {
			cond := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{
						Name:      "f1",
						Operation: cond.OperationEq,
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     "1",
						},
					},
					{
						Name:      "@timestamp",
						Operation: cond.OperationRange,
						ValueOptCfg: vopt.ValueOptCfg{
							ValueFrom: vopt.ValueFrom_Const,
							Value:     []any{float64(1718696519262), float64(1718696519362)},
						},
					},
				},
			}

			_, err := tyAdapter.convertQueryCondition(testCtx, cond)
			So(err, ShouldBeNil)
		})
	})
}

func TestExtractRawTraceEntry(t *testing.T) {
	Convey("Test extractRawTraceEntry", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("rawTraceDetail is nil", func() {
			rawTraceEntry := map[string]any{}

			abstractSpan := tyAdapter.extractRawTraceEntry(testCtx, rawTraceEntry)
			So(abstractSpan, ShouldResemble, interfaces.AbstractSpan{
				SpanID: "_-1",
				Kind:   interfaces.SPAN_KIND_UNSPECIFIED,
				Status: interfaces.SPAN_STATUS_OK,
			})
		})
	})
}

func TestParseRawTraceDetail(t *testing.T) {
	Convey("Test parseRawTraceDetail", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("rawTraceDetail is nil", func() {
			parentSeqStr := "-1"
			traceID := ""
			rawTraceDetail := map[string]any{}
			briefSpanMap := map[string]*interfaces.BriefSpan_{}
			detailSpanMap := map[string]interfaces.SpanDetail{}

			err := tyAdapter.parseRawTraceDetail(testCtx, parentSeqStr, traceID, rawTraceDetail, briefSpanMap, detailSpanMap)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetTraceStatisticData(t *testing.T) {
	Convey("Test getTraceStatisticData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("rawSpan is nil", func() {
			res := tyAdapter.genSpanDetail(nil, interfaces.AbstractSpan{})
			So(len(res), ShouldEqual, 10)
		})
	})
}

func TestGetTraceList(t *testing.T) {
	Convey("Test GetTraceList", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := tyAdapter.getTraceList(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the respCode from method 'GetNoUnmarshal'", func() {
			expectedRespData, _ := json.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			expectedErr := fmt.Errorf("failed to get tingyun trace list: %s", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := tyAdapter.getTraceList(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("error")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, _, err := tyAdapter.getTraceList(testCtx, TingYunClientConfig{})
			So(err, ShouldNotBeNil)
		})

		Convey("Get failed, caused by the respCode from respBody", func() {
			dataInfo := struct {
				Code int `json:"code"`
				Data struct {
					Total   int              `json:"totalElements"`
					Content []map[string]any `json:"content"`
				} `json:"data"`
			}{
				Code: http.StatusBadRequest,
			}
			expectedRespData, _ := sonic.Marshal(dataInfo)
			expectedErr := fmt.Errorf("failed to get tingyun trace list: %s", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespData, nil)

			_, _, err := tyAdapter.getTraceList(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			dataInfo := struct {
				Code int `json:"code"`
				Data struct {
					Total   int              `json:"totalElements"`
					Content []map[string]any `json:"content"`
				} `json:"data"`
			}{
				Code: http.StatusOK,
			}
			expectedRespData, _ := sonic.Marshal(dataInfo)

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespData, nil)

			_, _, err := tyAdapter.getTraceList(testCtx, TingYunClientConfig{})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetTraceDetail(t *testing.T) {
	Convey("Test GetTraceDetail", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Get failed, caused by the error from method 'GetNoUnmarshal'", func() {
			expectedErr := fmt.Errorf("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, expectedErr)

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the respCode from method 'GetNoUnmarshal'", func() {
			expectedRespData := []byte(nil)
			expectedErr := fmt.Errorf("failed to get tingyun trace detail: %s", string(expectedRespData))

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, expectedRespData, nil)

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func 'sonic.Get'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch := ApplyFuncReturn(sonic.Get, ast.Node{}, expectedErr)
			defer patch.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from ast.Node method 'Int64'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(0), expectedErr)
			defer patch2.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the respCode is 404 in respBody", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(404), nil)
			defer patch2.Reset()

			_, isExist, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldBeNil)
			So(isExist, ShouldBeFalse)
		})

		Convey("Get failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(500), nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch3.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the respCode is 500 in respBody", func() {
			expectedRespBody, _ := sonic.Marshal(
				struct {
					Code int    `json:"code"`
					Msg  string `json:"msg"`
				}{
					Code: 500,
					Msg:  "some errors",
				},
			)
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespBody, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(500), nil)
			defer patch2.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, respCode is 200 in respBody, but parse respBody failed", func() {
			expectedRespBody, _ := sonic.Marshal(
				struct {
					Code int            `json:"code"`
					Data map[string]any `json:"data"`
				}{
					Code: 200,
					Data: make(map[string]any, 0),
				},
			)
			expectedErr := errors.New("some errors")

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespBody, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(200), nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch3.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			expectedRespBody, _ := sonic.Marshal(
				struct {
					Code int            `json:"code"`
					Data map[string]any `json:"data"`
				}{
					Code: 200,
					Data: make(map[string]any, 0),
				},
			)

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, expectedRespBody, nil)

			patch1 := ApplyFuncReturn(sonic.Get, ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&ast.Node{}, "Int64", int64(200), nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch3.Reset()

			_, _, err := tyAdapter.getTraceDetail(testCtx, TingYunClientConfig{})
			So(err, ShouldBeNil)
		})
	})
}

func TestProcessDataConnection(t *testing.T) {
	Convey("Test processDataConnection", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDCAccess := umock.NewMockDataConnectionAccess(mockCtrl)
		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		tyAdapter := tingYunwAdapter{
			appSetting: &common.AppSetting{
				ThirdParty: common.ThirdParty{
					TingYunMaxTimePeriod: 172800,
				},
			},
			dcAccess:   mockDCAccess,
			httpClient: mockHttpClient,
		}

		Convey("Process failed, caused by the invalid data_source_type", func() {
			invalidType := "invalid"
			expectedErr := fmt.Errorf("Invalid data_source_type: %v", invalidType)

			err := tyAdapter.processDataConnection(testCtx, &interfaces.DataConnection{
				DataSourceType: invalidType,
			})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Process failed, caused by the error from func 'sonic.Marshal'", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(sonic.Marshal, []byte(nil), expectedErr)
			defer patch.Reset()

			err := tyAdapter.processDataConnection(testCtx, &interfaces.DataConnection{
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Process failed, caused by the error from func 'sonic.Unmarshal'", func() {
			expectedErr := errors.New("some errors")
			conn := interfaces.DataConnection{
				DataSourceType:   interfaces.SOURCE_TYPE_TINGYUN,
				DataSourceConfig: TingYunDetailedConfig{},
			}

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			err := tyAdapter.processDataConnection(testCtx, &conn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			conn := interfaces.DataConnection{
				DataSourceType:   interfaces.SOURCE_TYPE_TINGYUN,
				DataSourceConfig: TingYunDetailedConfig{},
			}

			err := tyAdapter.processDataConnection(testCtx, &conn)
			So(err, ShouldBeNil)
		})
	})
}
