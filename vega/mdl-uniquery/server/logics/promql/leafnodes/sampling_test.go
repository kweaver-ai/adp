// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

var (
	indexPattern   = "metricbeat-*"
	matResult      = make(static.Matrix, 0)
	emptyDslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 0,
			"total":      0,
		},
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    0,
				"relation": "eq",
			},
			"max_score": 0,
			"hits":      []string{},
		},
		"aggregations": map[string]interface{}{
			"labels": map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets":                     []string{},
			},
		},
		"timed_out": false,
		"took":      0,
	}

	NotEmptyDslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 0,
			"total":      0,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       "a",
						"doc_count": 2,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-09T02:52:09.000Z",
									"key":           1652064729000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     11132.0,
										"timestamp": 1652064729796,
									},
								},
								{
									"key_as_string": "2022-05-09T02:52:09.000Z",
									"key":           1652064732000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     11133.0,
										"timestamp": 1652064734695,
									},
								},
							},
						},
					},
					{
						"key":       "b",
						"doc_count": 2,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-09T02:52:09.000Z",
									"key":           1652064729000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     11142.0,
										"timestamp": 1652064729796,
									},
								},
								{
									"key_as_string": "2022-05-09T02:52:09.000Z",
									"key":           1652064732000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     11143.0,
										"timestamp": 1652064734695,
									},
								},
							},
						},
					},
				},
			},
		},
		"timed_out": false,
		"took":      0,
	}

	query = &interfaces.Query{
		Start:               1646360670,
		End:                 1646360700,
		Interval:            30000,
		LogGroupId:          "a",
		Limit:               -1,
		MaxSearchSeriesSize: maxSearchSeriesSize,
	}

	queryInstant = &interfaces.Query{
		Start:               1652320554000,
		End:                 1652320554000,
		Interval:            1,
		IsInstantQuery:      true,
		LogGroupId:          "a",
		MaxSearchSeriesSize: maxSearchSeriesSize,
	}

	groupby     = []string{interfaces.LABELS_STR}
	samplingAgg = interfaces.SAMPLING_AGG

	indicesError = "Error _cat/indices response from opensearch"

	key = "cluster=\"txy\",name=\"node-1\""

	logGroupQueryFilters = interfaces.LogGroup{
		IndexPattern: []string{indexPattern},
		MustFilter: []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
					"analyze_wildcard": true,
				},
			},
		},
	}
)

func mockLeafNodes(osAccess interfaces.OpenSearchAccess, lgAccess interfaces.LogGroupAccess,
	dvService interfaces.DataViewService) (leafNodes *LeafNodes) {

	leafNodes = NewLeafNodes(&common.AppSetting{
		ServerSetting: common.ServerSetting{
			FullCacheRefreshInterval: 24 * time.Hour,
		},
	}, osAccess, lgAccess, dvService)

	util.InitAntsPool(common.PoolSetting{
		MegerPoolSize:       10,
		ExecutePoolSize:     10,
		BatchSubmitPoolSize: 10,
	})
	return leafNodes
}

func TestGetQueryData(t *testing.T) {
	Convey("test sampling getQueryData", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		Tsids_Of_Model_Metric_Map.Store("0+a", TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
				},
			},
		})
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		Convey("GetLogGroupQueryFilters error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, query)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("getIndicesNumberOfShards error ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, query)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("executeDslAndProcess error ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("MakeDSL error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, "ratem", query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("Merge error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(NotEmptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			util.MegerPool.Release()
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("getQueryData success with query range ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, query)

			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matResult})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("getQueryData success with instant query", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, queryInstant)

			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matResult})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("return static.Matrix{}: getIndicesNumberOfShards indexShardsArr == 0", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			emptyIndexShardsArr, _ := sonic.Marshal([]interfaces.IndexShards{})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyIndexShardsArr, http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelector(testCtx, expr.(*parser.VectorSelector), groupby, samplingAgg, queryInstant)

			So(res, ShouldResemble, static.PageMatrix{Matrix: matResult})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestExecuteDslAndProcess(t *testing.T) {
	Convey("test executeDslAndProcess ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})

		query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`

		Convey("strconv.Atoi failed ", func() {

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			patches := ApplyFunc(strconv.Atoi,
				func(s string) (int, error) {
					return 0, errors.New("error")
				},
			)
			defer patches.Reset()

			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			dsl, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

			mapResult, err := lnMock.ExecuteDslAndProcess(testCtx, *dsl, indexShardsArr, groupby)

			So(mapResult.LabelsMap, ShouldBeNil)
			So(mapResult.TsValueMap, ShouldBeNil)
			So(err.Error(), ShouldResemble, `error`)

		})

		Convey("ExecutePool.Submit failed", func() {
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			dsl, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

			util.ExecutePool.Release()
			defer util.MegerPool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			mapResult, err := lnMock.ExecuteDslAndProcess(testCtx, *dsl, indexShardsArr, groupby)

			So(mapResult.LabelsMap, ShouldBeNil)
			So(mapResult.TsValueMap, ShouldBeNil)
			So(err.Error(), ShouldResemble, `ExecutePool.Submit error: this pool has been closed`)
		})

		Convey("GetDataFromOpenSearchWithBuffer failed", func() {
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			dsl, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

			mapResult, err := lnMock.ExecuteDslAndProcess(testCtx, *dsl, indexShardsArr, groupby)

			So(mapResult.LabelsMap, ShouldBeNil)
			So(mapResult.TsValueMap, ShouldBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{"type":"UniQuery.InternalServerError",`+
				`"reason":"Error getting response from opensearch"}}`)
		})

		Convey("executeDslAndProcess ok ", func() {
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			dsl, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

			mapResult, err := lnMock.ExecuteDslAndProcess(testCtx, *dsl, indexShardsArr, groupby)

			So(mapResult.TsValueMap, ShouldResemble, map[string][][]gjson.Result{})
			So(mapResult.LabelsMap, ShouldResemble, map[string][]*labels.Label{})
			So(err, ShouldBeNil)
		})

		Convey("executeDslAndProcess ok with result is not empty ", func() {
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(NotEmptyDslResult), http.StatusOK, nil)

			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			dsl, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

			mapResult, err := lnMock.ExecuteDslAndProcess(testCtx, *dsl, indexShardsArr, groupby)

			expectedLabelsMap := make(map[string][]*labels.Label)
			expectedLabelsMap["a"] = []*labels.Label{
				{
					Name:  interfaces.LABELS_STR,
					Value: "a",
				},
			}
			expectedLabelsMap["b"] = []*labels.Label{
				{
					Name:  interfaces.LABELS_STR,
					Value: "b",
				},
			}

			So(mapResult.LabelsMap, ShouldResemble, expectedLabelsMap)
			So(len(mapResult.TsValueMap), ShouldEqual, 2)
			So(len(mapResult.TsValueMap["a"][0]), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergeSamples(t *testing.T) {
	Convey("test MergeSamples ", t, func() {
		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
		})

		Convey("mergePool submit err ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			util.MegerPool.Release()
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 0})
			So(mat, ShouldBeNil)
			So(err.Error(), ShouldResemble, `this pool has been closed`)
		})

		Convey("success with 1 sample ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 8,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 10})
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with more than 1 sample ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360698350}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":10.0,\"timestamp\":1646360728352}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},

					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":15.0,\"timestamp\":1646360758353}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360788356}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 9,
						},
						{
							T: 1646360700000,
							V: 10,
						},
						{
							T: 1646360730000,
							V: 15,
						},
						{
							T: 1646360760000,
							V: 9,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 30000})
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with more than 1 sample 2 ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360698350}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":15.0,\"timestamp\":1646360758353}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360788356}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 9,
						},
						{
							T: 1646360700000,
							V: 9,
						},
						{
							T: 1646360730000,
							V: 15,
						},
						{
							T: 1646360760000,
							V: 9,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 30000})
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success when there is empty samples in a time window ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360698350}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360788356}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 9,
						},
						{
							T: 1646360700000,
							V: 9,
						},
						{
							T: 1646360730000,
							V: 9,
						},
						{
							T: 1646360760000,
							V: 9,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 30000})
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with 1 sample with instant query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 8,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, queryInstant)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with more than 1 sample 2 with instant query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360698350}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":15.0,\"timestamp\":1646360758353}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":10.0,\"timestamp\":1646360788356}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 10,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, queryInstant)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success when there is empty samples in a time window with instant query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360698350}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360788356}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name:  "cluster",
							Value: "txy",
						},
						{
							Name:  "name",
							Value: "node-1",
						},
					},
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 9,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := samplingMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, queryInstant)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestEvalVectorSelectorFields(t *testing.T) {
	Convey("test sampling EvalVectorSelectorFields", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		Tsids_Of_Model_Metric_Map.Store("+5f1634064bfc952772fb67c6363f4ee6", TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
				},
			},
		})
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()

		Convey("GetDataFromOpenSearchWithBuffer error", func() {
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError, errors.New("opensearch 查询失败"))

			query.QueryStr = `a1`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.EvalVectorSelectorFields(testCtx, expr.(*parser.VectorSelector), query, "")
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("Success case", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `a`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			res, status, err := lnMock.EvalVectorSelectorFields(testCtx, expr.(*parser.VectorSelector), query, "")
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]bool{"label1": true})
		})
	})
}
