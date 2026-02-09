// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics"
	"uniquery/logics/promql"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/leafnodes"
	"uniquery/logics/promql/util"
)

func replace(s string) string {
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}

func compareJsonString(actualString string, expectedString string) {
	actualJson := map[string]any{}
	expectedJson := map[string]any{}
	_ = sonic.UnmarshalString(actualString, &actualJson)
	_ = sonic.UnmarshalString(expectedString, &expectedJson)
	So(actualJson, ShouldResemble, expectedJson)
}

var (
	shard0        = "_shards:0"
	shard1        = "_shards:1"
	shard2        = "_shards:2"
	labelsStrNas  = "cluster=\"opensearch\",instance=\"instance_abc\",index=\"nas_statis-2022.05-0\",job=\"prometheus\""
	labelsStrNode = "cluster=\"opensearch\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\""
	timeString39  = "2022-05-12T01:55:39.000Z"
	timeString42  = "2022-05-12T01:55:42.000Z"
	dslResult0    = map[string]interface{}{
		"took":      16,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      3,
			"successful": 3,
			"skipped":    0,
			"failed":     0,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       labelsStrNas,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320539000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320542000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     10.0,
										"timestamp": 1652320544541,
									},
								},
							},
						},
					},
					{
						"key":       labelsStrNode,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320539000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     40.0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320542000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320544541,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	dslResult1 = map[string]interface{}{
		"took":      16,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      3,
			"successful": 3,
			"skipped":    0,
			"failed":     0,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       labelsStrNas,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320545000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     21.0,
										"timestamp": 1652320547541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320548000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     11.0,
										"timestamp": 1652320550541,
									},
								},
							},
						},
					},
					{
						"key":       labelsStrNode,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320545000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     41.0,
										"timestamp": 1652320547541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320548000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     21.0,
										"timestamp": 1652320550541,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	dslResult2 = map[string]interface{}{
		"took":      16,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      3,
			"successful": 3,
			"skipped":    0,
			"failed":     0,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       labelsStrNas,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320551000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     22.0,
										"timestamp": 1652320553541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320554000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     12.0,
										"timestamp": 1652320556541,
									},
								},
							},
						},
					},
					{
						"key":       labelsStrNode,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320551000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     42.0,
										"timestamp": 1652320553541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320554000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     22.0,
										"timestamp": 1652320556541,
									},
								},
							},
						},
					},
				},
			},
		},
	}

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

	dslResultMetric2 = map[string]interface{}{
		"took":      16,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      3,
			"successful": 3,
			"skipped":    0,
			"failed":     0,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       labelsStrNas,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320539000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     2.0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320542000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320544541,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:45.000Z",
									"key":           1652320545000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320544641,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:48.000Z",
									"key":           1652320548000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320544941,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:51.000Z",
									"key":           1652320551000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320545241,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:54.000Z",
									"key":           1652320554000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320555000,
									},
								},
							},
						},
					},
					{
						"key":       labelsStrNode,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": timeString39,
									"key":           1652320539000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     4.0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": timeString42,
									"key":           1652320542000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     2.0,
										"timestamp": 1652320544541,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:45.000Z",
									"key":           1652320545000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320544641,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:48.000Z",
									"key":           1652320548000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320544941,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:51.000Z",
									"key":           1652320551000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320545241,
									},
								},
								{
									"key_as_string": "2022-05-12T01:55:54.000Z",
									"key":           1652320554000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320555000,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	indexMetricbeat111    = "metricbeat11-1"
	indexMetricbeat122    = "metricbeat12-2"
	indexbaseMetricbeat11 = "metricbeat11-*"
	indexbaseMetricbeat12 = "metricbeat12-*"

	indexPattern         = "metricbeat-*"
	index                = "metricbeat-1"
	logGroupQueryFilters = interfaces.LogGroup{
		IndexPattern: []string{
			indexPattern,
		},
		MustFilter: []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
					"analyze_wildcard": true,
				},
			},
		},
	}

	logGroupQueryFiltersWith1112 = interfaces.LogGroup{
		IndexPattern: []string{
			"metricbeat11-*",
			"metricbeat12-*",
		},
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

func mockNewPromQLRestHandler(
	appSetting *common.AppSetting, hydra rest.Hydra,
	promqlService interfaces.PromQLService) (r *restHandler) {
	r = &restHandler{
		appSetting:    appSetting,
		hydra:         hydra,
		promqlService: promqlService,
	}
	r.InitMetric()
	return r
}

func TestTimeSeriesSelector(t *testing.T) {
	Convey("Test time series selector ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		Convey("Success time series selector with index has 1 shard \n", func() {

			dslResult := map[string]interface{}{
				"took":      16,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      3,
					"successful": 3,
					"skipped":    0,
					"failed":     0,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       labelsStrNas,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544541,
											},
										},
									},
								},
							},
							{
								"key":       labelsStrNode,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
									},
								},
							},
						},
					},
				},
			}

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"2"
								],
								[
									1652320542,
									"1"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"4"
								],
								[
									1652320542,
									"2"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(dslResult), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{cluster="opensearch",index=~"a.*|node.*"}` +
				`&start=1652064729&end=1652064735&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success time series selector with index has 3 shard \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"20"
								],
								[
									1652320542,
									"10"
								],
								[
									1652320545,
									"21"
								],
								[
									1652320548,
									"11"
								],
								[
									1652320551,
									"22"
								],
								[
									1652320554,
									"12"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{cluster="opensearch",index=~"a.*|node.*"}` +
				`&start=1652064729&end=1652064735&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("Success time series selector with no data \n", func() {
			expected := `{
					"status": "success",
					"data": {
						"resultType": "matrix",
						"result": []
					}
				}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{cluster="opensearch",index=~"a.*|node.*"}` +
				`&start=1652064729&end=1652064735&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestArithmeticBinaryOperatorsInQueryRange(t *testing.T) {
	Convey("Test Arithmetic binary operators in query_range ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)

		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		Convey("scalar op scalar expression with 1+2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"3"
								],
								[
									1646360700,
									"3"
								],
								[
									1646360730,
									"3"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=1%2B2&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey(`scalar op scalar expression with 1+"a" \n`, func() {
			expected := `{"status":"error","errorType":"bad_data","error":"1:3: parse error: binary expression must contain only scalar and instant vector types"}`
			body := []byte(`query=1%2B%22a%22&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3^-1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"0"
								],
								[
									1646360700,
									"0"
								],
								[
									1646360730,
									"0"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=3^-1000&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3^1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"+Inf"
								],
								[
									1646360700,
									"+Inf"
								],
								[
									1646360730,
									"+Inf"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=3^1000&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with -3^1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"-Inf"
								],
								[
									1646360700,
									"-Inf"
								],
								[
									1646360730,
									"-Inf"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=-3^1000&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 6.7108864e+07/1024/1024 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"64"
								],
								[
									1646360700,
									"64"
								],
								[
									1646360730,
									"64"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=6.7108864e%2B07/1024/1024&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3%0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"NaN"
								],
								[
									1646360700,
									"NaN"
								],
								[
									1646360730,
									"NaN"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=3%250&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3/0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"+Inf"
								],
								[
									1646360700,
									"+Inf"
								],
								[
									1646360730,
									"+Inf"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=3/0&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3*0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"0"
								],
								[
									1646360700,
									"0"
								],
								[
									1646360730,
									"0"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=3*0&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with (5.2818978e+07+1.0766302e+07)/(11*24*60*60) \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"66.9037037037037"
								],
								[
									1646360700,
									"66.9037037037037"
								],
								[
									1646360730,
									"66.9037037037037"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=(5.2818978e%2B07%2B1.0766302e%2B07)/(11*24*60*60)&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with (5.2818978e+07-1.0766302e+07)/10*2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1646360670,
									"8410535.2"
								],
								[
									1646360700,
									"8410535.2"
								],
								[
									1646360730,
									"8410535.2"
								]
							]
						}
					]
				}
			}`

			body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30s`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector * 3 \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"60"
								],
								[
									1652320542,
									"30"
								],
								[
									1652320545,
									"63"
								],
								[
									1652320548,
									"33"
								],
								[
									1652320551,
									"66"
								],
								[
									1652320554,
									"36"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"120"
								],
								[
									1652320542,
									"60"
								],
								[
									1652320545,
									"123"
								],
								[
									1652320548,
									"63"
								],
								[
									1652320551,
									"126"
								],
								[
									1652320554,
									"66"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}*3` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector not exists  \n", func() {
			expected := `{
					"status": "success",
					"data": {
						"resultType": "matrix",
						"result": []
					}
				}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=1%2Ba&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector % 3 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"2"
								],
								[
									1652320542,
									"1"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"2"
								],
								[
									1652320551,
									"1"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"1"
								],
								[
									1652320542,
									"2"
								],
								[
									1652320545,
									"2"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"1"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}%253` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with 100%vector \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"16"
								],
								[
									1652320548,
									"1"
								],
								[
									1652320551,
									"12"
								],
								[
									1652320554,
									"4"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"20"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"18"
								],
								[
									1652320548,
									"16"
								],
								[
									1652320551,
									"16"
								],
								[
									1652320554,
									"12"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=100%25opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with 3^vector \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"3486784401"
								],
								[
									1652320542,
									"59049"
								],
								[
									1652320545,
									"10460353203"
								],
								[
									1652320548,
									"177147"
								],
								[
									1652320551,
									"31381059609"
								],
								[
									1652320554,
									"531441"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"12157665459056929000"
								],
								[
									1652320542,
									"3486784401"
								],
								[
									1652320545,
									"36472996377170790000"
								],
								[
									1652320548,
									"10460353203"
								],
								[
									1652320551,
									"109418989131512370000"
								],
								[
									1652320554,
									"31381059609"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=3^opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector^3 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"8000"
								],
								[
									1652320542,
									"1000"
								],
								[
									1652320545,
									"9261"
								],
								[
									1652320548,
									"1331"
								],
								[
									1652320551,
									"10648"
								],
								[
									1652320554,
									"1728"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"64000"
								],
								[
									1652320542,
									"8000"
								],
								[
									1652320545,
									"68921"
								],
								[
									1652320548,
									"9261"
								],
								[
									1652320551,
									"74088"
								],
								[
									1652320554,
									"10648"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}^3` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector/2/2/2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"2.5"
								],
								[
									1652320542,
									"1.25"
								],
								[
									1652320545,
									"2.625"
								],
								[
									1652320548,
									"1.375"
								],
								[
									1652320551,
									"2.75"
								],
								[
									1652320554,
									"1.5"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"5"
								],
								[
									1652320542,
									"2.5"
								],
								[
									1652320545,
									"5.125"
								],
								[
									1652320548,
									"2.625"
								],
								[
									1652320551,
									"5.25"
								],
								[
									1652320554,
									"2.75"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}/2/2/2` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// (1-(opensearch_os_mem_free_bytes{name='xx',fstype=~"ext.*|xfs"} / node_filesystem_size_bytes{name='xx',fstype=~"ext.*|xfs"}))*100 - 0
		Convey("vector op scalar expression with (1-vector/vector)*100-0) \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=(1-(opensearch_os_mem_used_bytes` +
				`/ opensearch_os_mem_total_bytes))*100 - 0` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op vector expression vector - vector ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op on vector expression vector - vector ", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(index) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op vector when found duplicate series for the match group, error ", func() {
			expected := "{\"status\":\"error\",\"errorType\":\"execution\",\"error\":\"found duplicate series for the match group, many-to-many only allowed for set operators\"}"
			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- ignoring(index) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op ignoring vector expression vector - vector ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- ignoring(cluster) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op on vector when matcher is many to one ", func() {
			dslResultMetric21 := map[string]interface{}{
				"took":      16,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      3,
					"successful": 3,
					"skipped":    0,
					"failed":     0,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       labelsStrNas,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       labelsStrNode,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       "cluster=\"opensearch2\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\"",
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
						},
					},
				},
			}
			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(dslResultMetric21), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			expected := "{\"status\":\"error\",\"errorType\":\"execution\",\"error\":\"multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)\"}"

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(index) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op on vector when matcher is one to many ", func() {
			dslResultMetric21 := map[string]interface{}{
				"took":      16,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      3,
					"successful": 3,
					"skipped":    0,
					"failed":     0,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       labelsStrNas,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       labelsStrNode,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       "cluster=\"opensearch2\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\"",
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
						},
					},
				},
			}

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(dslResultMetric21), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(index) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			expected := `{"status":"error","errorType":"execution","error":"found duplicate series for the match group, many-to-many only allowed for set operators"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)

		})

		Convey("vector op on vector expression vector - vector 2 ", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"0"
								],
								[
									1652320542,
									"0"
								],
								[
									1652320545,
									"0"
								],
								[
									1652320548,
									"0"
								],
								[
									1652320551,
									"0"
								],
								[
									1652320554,
									"0"
								]
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(cluster,index) opensearch_os_mem_total_bytes` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestArithmeticBinaryOperatorsInQuery(t *testing.T) {
	Convey("Test Arithmetic binary operators in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)

		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		Convey("scalar op scalar expression with 1+2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"3"
					]
				}
			}`

			body := []byte(`query=1%2B2&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey(`scalar op scalar expression with 1+"a" \n`, func() {

			expected := `{"status":"error","errorType":"bad_data","error":"1:3: parse error: binary expression must contain only scalar and instant vector types"}`

			body := []byte(`query=1%2B%22a%22&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3^-1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"0"
					]
				}
			}`

			body := []byte(`query=3^-1000&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3^1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"+Inf"
					]
				}
			}`

			body := []byte(`query=3^1000&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with -3^1000 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"-Inf"
					]
				}
			}`

			body := []byte(`query=-3^1000&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 6.7108864e+07/1024/1024 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"64"
					]
				}
			}`

			body := []byte(`query=6.7108864e%2B07/1024/1024&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3%0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"NaN"
					]
				}
			}`

			body := []byte(`query=3%250&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3/0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"+Inf"
					]
				}
			}`

			body := []byte(`query=3/0&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with 3*0 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"0"
					]
				}
			}`

			body := []byte(`query=3*0&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with (5.2818978e+07+1.0766302e+07)/(11*24*60*60) \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"66.9037037037037"
					]
				}
			}`

			body := []byte(`query=(5.2818978e%2B07%2B1.0766302e%2B07)/(11*24*60*60)&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("scalar op scalar expression with (5.2818978e+07-1.0766302e+07)/10*2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "scalar",
					"result":[
						1646360730,
						"8410535.2"
					]
				}
			}`

			body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&time=1646360730`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector * 3 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"36"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"66"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}*3` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector not exists  \n", func() {
			expected := `{
					"status": "success",
					"data": {
						"resultType": "vector",
						"result": []
					}
				}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=1%2Ba&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector % 3 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"1"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}%253` +
				`&time=1652320554&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with 100%vector \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
								1652320555,
								"4"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
								1652320555,
								"12"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=100%25opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}` +
				`&time=1652320555&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with 3^vector \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"531441"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"31381059609"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=3^opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}` +
				`&time=1652320554&ar_dataview="a"`)
			// body := []byte(`query=(5.2818978e%2B07-1.0766302e%2B07)/10*2&start=1646360670&end=1646360730&step=30`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector^3 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"1728"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"10648"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}^3` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op scalar expression with vector/2/2/2 \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"1.5"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"2.75"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}/2/2/2` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// (1-(opensearch_os_mem_free_bytes{name='xx',fstype=~"ext.*|xfs"} / node_filesystem_size_bytes{name='xx',fstype=~"ext.*|xfs"}))*100 - 0
		Convey("vector op scalar expression with (1-vector/vector)*100-0) \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=(1-(opensearch_os_mem_used_bytes` +
				`/ opensearch_os_mem_total_bytes))*100 - 0` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op vector expression vector - vector ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op on vector expression vector - vector ", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"0"
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(index) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op vector when found duplicate series for the match group, error ", func() {
			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- ignoring(index) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			expected := "{\"status\":\"error\",\"errorType\":\"execution\",\"error\":\"found duplicate series for the match group, many-to-many only allowed for set operators\"}"

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op ignoring vector expression vector - vector ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320554,
									"0"
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- ignoring(cluster) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("vector op on vector when matcher is many to one ", func() {
			dslResultMetric21 := map[string]interface{}{
				"took":      16,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      3,
					"successful": 3,
					"skipped":    0,
					"failed":     0,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       labelsStrNas,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       labelsStrNode,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       "cluster=\"opensearch2\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\"",
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
						},
					},
				},
			}
			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{index}, gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(dslResultMetric21), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{index}, gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes - on(index) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			expected := "{\"status\":\"error\",\"errorType\":\"execution\",\"error\":\"multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)\"}"

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)

		})

		Convey("vector op on vector when matcher is one to many ", func() {
			dslResultMetric21 := map[string]interface{}{
				"took":      16,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      3,
					"successful": 3,
					"skipped":    0,
					"failed":     0,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       labelsStrNas,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       labelsStrNode,
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
							{
								"key":       "cluster=\"opensearch2\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\"",
								"doc_count": 12,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": timeString39,
											"key":           1652320539000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     4.0,
												"timestamp": 1652320541541,
											},
										},
										{
											"key_as_string": timeString42,
											"key":           1652320542000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     2.0,
												"timestamp": 1652320544541,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:45.000Z",
											"key":           1652320545000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544641,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:48.000Z",
											"key":           1652320548000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320544941,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:51.000Z",
											"key":           1652320551000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320545241,
											},
										},
										{
											"key_as_string": "2022-05-12T01:55:54.000Z",
											"key":           1652320554000,
											"doc_count":     6,
											"value": map[string]interface{}{
												"value":     1.0,
												"timestamp": 1652320555000,
											},
										},
									},
								},
							},
						},
					},
				},
			}

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric21), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(index) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			expected := `{"status":"error","errorType":"execution","error":"found duplicate series for the match group, many-to-many only allowed for set operators"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnprocessableEntity)
			compareJsonString(w.Body.String(), expected)

		})

		Convey("vector op on vector expression vector - vector 2 ", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"0"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"0"
							]
						}
					]
				}
			}`

			var indexShardsArrMetric2 = make([]interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFiltersWith1112, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat11).AnyTimes().Return(
				notEmptyJson2, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				gomock.Any()).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: indexMetricbeat122,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat12)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexbaseMetricbeat12).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=opensearch_os_mem_used_bytes` +
				`- on(cluster,index) opensearch_os_mem_total_bytes` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestSumAggregationsInQueryRange(t *testing.T) {
	Convey("Test aggregation in query_range ", t, func() {
		test := setGinMode()
		defer test()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		Convey("aggregate op sum ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"index":"nas_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"20"
								],
								[
									1652320542,
									"10"
								],
								[
									1652320545,
									"21"
								],
								[
									1652320548,
									"11"
								],
								[
									1652320551,
									"22"
								],
								[
									1652320554,
									"12"
								]
							]
						},
						{
							"metric": {
								"index":"node_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=sum(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(index)` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("agregation op avg \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
							},
							"values": [
								[
									1652320539,
									"30"
								],
								[
									1652320542,
									"15"
								],
								[
									1652320545,
									"31"
								],
								[
									1652320548,
									"16"
								],
								[
									1652320551,
									"32"
								],
								[
									1652320554,
									"17"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=avg(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"node_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(1,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op bottomk ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"nas_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"values": [
								[
									1652320539,
									"20"
								],
								[
									1652320542,
									"10"
								],
								[
									1652320545,
									"21"
								],
								[
									1652320548,
									"11"
								],
								[
									1652320551,
									"22"
								],
								[
									1652320554,
									"12"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=bottomk(1,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with param invalid ", func() {
			body := []byte(`query=topk("1",opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(index)` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			expected := `{"status":"error","errorType":"bad_data","error":"1:6: parse error: expected type scalar in aggregation parameter, got string"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with k lt 1 ", func() {
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(0.5,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(index)` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			expected := `{"status":"success","data":{"resultType":"matrix","result":[]}}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with k == 1.6", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"node_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(1.6,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestSumAggregationsInQuery(t *testing.T) {
	Convey("Test aggregation in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		Convey("aggregate op sum ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"index":"nas_statis-2022.05-0"
							},
							"value": [
								1652320554,
								"12"
							]
						},
						{
							"metric": {
								"index":"node_statis-2022.05-0"
							},
							"value": [
								1652320554,
								"22"
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=sum(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(index)` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("agregation op avg \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
							},
							"value": [
								1652320554,
								"17"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=avg(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"node_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(1,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op bottomk ", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"nas_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"value": [
									1652320554,
									"12"
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=bottomk(1,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with param invalid ", func() {
			body := []byte(`query=topk("1",opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)

			expected := `{"status":"error","errorType":"bad_data","error":"1:6: parse error: expected type scalar in aggregation parameter, got string"}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with k lt 1 ", func() {
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(0.5,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)

			expected := `{"status":"success","data":{"resultType":"vector","result":[]}}`

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("aggregate op topk with k == 1.6", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster":"opensearch",
								"index":"node_statis-2022.05-0",
								"instance":"instance_abc",
								"job":"prometheus"
							},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=topk(1.6,opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestMaxOperatorInQueryRange(t *testing.T) {
	Convey("Test max operator in query_range ", t, func() {
		test := setGinMode()
		defer test()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		Convey("max aggregation expression \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("max by aggregation expression \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(cluster)` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("max without aggregation expression \n", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"20"
								],
								[
									1652320542,
									"10"
								],
								[
									1652320545,
									"21"
								],
								[
									1652320548,
									"11"
								],
								[
									1652320551,
									"22"
								],
								[
									1652320554,
									"12"
								]
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0"
							},
							"values": [
								[
									1652320539,
									"40"
								],
								[
									1652320542,
									"20"
								],
								[
									1652320545,
									"41"
								],
								[
									1652320548,
									"21"
								],
								[
									1652320551,
									"42"
								],
								[
									1652320554,
									"22"
								]
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})without(cluster,instance,job)` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestMaxOperatorInQuery(t *testing.T) {
	Convey("Test max operator in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		Convey("max aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("max by aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch"
							},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(cluster)` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("max without aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"12"
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=max(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})without(cluster,instance,job)` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestSortFunctionInQueryRange(t *testing.T) {
	Convey("Test sort function in query_range ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)

		body := []byte(`query=sort(opensearch_indices_shards_docs` +
			`{index=~"a.*|node.*"})` +
			`&start=1652320539&end=1652320554&step=3s`)

		expected := `{"status": "error", "errorType": "bad_data", "error": " 'sort' can not be used in the query_range requests. "}`

		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		compareJsonString(w.Body.String(), expected)
	})
}

func TestSortDescFunctionInQueryRange(t *testing.T) {
	Convey("Test sort_desc function in query_range ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query_range"

		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)

		body := []byte(`query=sort_desc(opensearch_indices_shards_docs` +
			`{index=~"a.*|node.*"})` +
			`&start=1652320539&end=1652320554&step=3s`)

		expected := `{"status":"error","errorType":"bad_data","error":" 'sort_desc' can not be used in the query_range requests. "}`

		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		compareJsonString(w.Body.String(), expected)
	})
}

func TestSortFunctionInQuery(t *testing.T) {
	Convey("Test sort function in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		expected := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {
							"cluster": "opensearch",
							"index": "nas_statis-2022.05-0",
							"instance": "instance_abc",
							"job": "prometheus"
						},
						"value": [
								1652320554,
								"36"
						]
					},
					{
						"metric": {
							"cluster": "opensearch",
							"index": "node_statis-2022.05-0",
							"instance": "instance_abc",
							"job": "prometheus"
						},
						"value": [
								1652320554,
								"66"
						]
					}
				]
			}
		}`

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: index,
			Pri:       "3",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
			logGroupQueryFilters, true, nil)
		dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
			notEmptyJson, http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

		body := []byte(`query=sort(opensearch_indices_shards_docs` +
			`{index=~"a.*|node.*"}*3)` +
			`&time=1652320554&ar_dataview="a"`)
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		compareJsonString(w.Body.String(), expected)
	})
}

func TestSortDescFunctionInQuery(t *testing.T) {
	Convey("Test sort_desc function in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		expected := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {
							"cluster": "opensearch",
							"index": "node_statis-2022.05-0",
							"instance": "instance_abc",
							"job": "prometheus"
						},
						"value": [
								1652320554,
								"66"
						]
					},
					{
						"metric": {
							"cluster": "opensearch",
							"index": "nas_statis-2022.05-0",
							"instance": "instance_abc",
							"job": "prometheus"
						},
						"value": [
								1652320554,
								"36"
						]
					}
				]
			}
		}`

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: index,
			Pri:       "3",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
			logGroupQueryFilters, true, nil)
		dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
			notEmptyJson, http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

		body := []byte(`query=sort_desc(opensearch_indices_shards_docs` +
			`{index=~"a.*|node.*"}*3)` +
			`&time=1652320554&ar_dataview="a"`)
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		compareJsonString(w.Body.String(), expected)
	})
}

func TestMinOperatorInQuery(t *testing.T) {
	Convey("Test min operator in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		Convey("max aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"12"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=min(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("min by aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch"
							},
							"value": [
									1652320554,
									"12"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=min(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})by(cluster)` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("min without aggregation expression \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"index": "nas_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"12"
							]
						},
						{
							"metric": {
								"index": "node_statis-2022.05-0"
							},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			body := []byte(`query=min(opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"})without(cluster,instance,job)` +
				`&time=1652320554&ar_dataview="a"`)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

// rate 整个算子测试，从接口开始 ,ok
func TestFuncRate(t *testing.T) {
	Convey("test rate operator\n", t, func() {
		test := setGinMode()
		defer test()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		rateDslResult0 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320539541,
											"lastValue":         20.0,
											"lastTimestamp":     1652320541541,
											"counterCorrection": 2.3,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320543541,
											"lastValue":         18.0,
											"lastTimestamp":     1652320544541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        4.0,
											"firstTimestamp":    1652320540541,
											"lastValue":         40.0,
											"lastTimestamp":     1652320541541,
											"counterCorrection": 0.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        2.0,
											"firstTimestamp":    1652320542541,
											"lastValue":         20.0,
											"lastTimestamp":     1652320544541,
											"counterCorrection": 1.29,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult1 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        19.0,
											"firstTimestamp":    1652320545001,
											"lastValue":         21.0,
											"lastTimestamp":     1652320547541,
											"counterCorrection": 7.0,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        22.0,
											"firstTimestamp":    1652320549541,
											"lastValue":         19.0,
											"lastTimestamp":     1652320550541,
											"counterCorrection": 22.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        20.0,
											"firstTimestamp":    1652320546541,
											"lastValue":         41.0,
											"lastTimestamp":     1652320547541,
											"counterCorrection": 0.0,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        42.0,
											"firstTimestamp":    1652320550141,
											"lastValue":         44.0,
											"lastTimestamp":     1652320550541,
											"counterCorrection": 2.0,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult2 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320551000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        20.0,
											"firstTimestamp":    1652320551141,
											"lastValue":         22.0,
											"lastTimestamp":     1652320553541,
											"counterCorrection": 2.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        23.0,
											"firstTimestamp":    1652320555141,
											"lastValue":         33.0,
											"lastTimestamp":     1652320556541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 7,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320551000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320551141,
											"lastValue":         42.0,
											"lastTimestamp":     1652320553541,
											"counterCorrection": 2.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     1,
										"value": map[string]interface{}{
											"firstValue":        12.0,
											"firstTimestamp":    1652320556541,
											"lastValue":         12.0,
											"lastTimestamp":     1652320556541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		emptyDslResult = map[string]interface{}{
			"aggregations": map[string]interface{}{
				"labels": map[string]interface{}{"buckets": []string{}},
			},
		}

		// rate query with range query, no data
		Convey("1. rate with no data and range query", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=rate(opensearch_indices_shards_docs[6s])` +
				`&start=1652032053&end=1652032153&step=3s&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// rate query with range query, have data
		Convey("2. rate with multiple point and range query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"5.743135454545455"
								],
								[
									1652320542,
									"3.272727272727272"
								],
								[
									1652320545,
									"5.234657039711192"
								],
								[
									1652320548,
									"4.363636363636363"
								],
								[
									1652320551,
									"2.7777777777777772"
								],
								[
									1652320554,
									"2"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"10.64901515151515"
								],
								[
									1652320542,
									"7.664770333333333"
								],
								[
									1652320545,
									"4.727272727272727"
								],
								[
									1652320548,
									"8.727272727272727"
								],
								[
									1652320551,
									"8.518518518518517"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(rateDslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(rateDslResult2), http.StatusOK, nil)

			body := []byte(`query=rate(opensearch_indices_shards_docs[6s])` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// rate query with instant query, no data
		Convey("3. rate with no data and instant query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=rate(opensearch_indices_shards_docs[6s])` +
				`&time=1652032053&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// rate query with instant query, have data
		Convey("4. rate with multiple point and instant query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320544,
									"6.06"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320544,
									"12.999187803030303"
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=rate(opensearch_indices_shards_docs[3s])` +
				`&time=1652320544&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

var (
	keyMetricsInf      = "handler=\"/metrics\",le=\"+Inf\""
	keyApiInf          = "handler=\"/api/v1/query\",le=\"+Inf\""
	keyMetrics02       = "handler=\"/metrics\",le=\"0.2\""
	keyApi02           = "handler=\"/api/v1/query\",le=\"0.2\""
	keyApi01           = "handler=\"/api/v1/query\",le=\"0.1\""
	keyMetrics01       = "handler=\"/metrics\",le=\"0.1\""
	key20220425T104000 = "2022-04-25T10:40:00.000Z"
	key20220425T104500 = "2022-04-25T10:45:00.000Z"

	nDslResult0 = map[string]interface{}{
		"took":      1,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      1,
			"successful": 1,
			"skipped":    0,
			"failed":     0,
		},
		"hits": map[string]interface{}{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       keyMetricsInf,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     8,
									"value": map[string]interface{}{
										"value":     353.0,
										"timestamp": 1650883485202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     4,
									"value": map[string]interface{}{
										"value":     361.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi01,
						"doc_count": 11,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     9,
									"value": map[string]interface{}{
										"value":     1353.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     1357.0,
										"timestamp": 1650883535202,
									},
								},
							},
						},
					},
					{
						"key":       keyApiInf,
						"doc_count": 10,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     9,
									"value": map[string]interface{}{
										"value":     1354.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     1361.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetrics01,
						"doc_count": 10,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     8,
									"value": map[string]interface{}{
										"value":     352.0,
										"timestamp": 1650883485202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     358.0,
										"timestamp": 1650883545202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi02,
						"doc_count": 8,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1345.0,
										"timestamp": 1650883415202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     1356.0,
										"timestamp": 1650883525202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetrics02,
						"doc_count": 6,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     4,
									"value": map[string]interface{}{
										"value":     345.0,
										"timestamp": 1650883415202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     357.0,
										"timestamp": 1650883535202,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	nDslResult1 = map[string]interface{}{
		"took":      1,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      1,
			"successful": 1,
			"skipped":    0,
			"failed":     0,
		},
		"hits": map[string]interface{}{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       keyApiInf,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     7,
									"value": map[string]interface{}{
										"value":     1351.0,
										"timestamp": 1650883465202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     5,
									"value": map[string]interface{}{
										"value":     1360.0,
										"timestamp": 1650883555202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetrics01,
						"doc_count": 12,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     10,
									"value": map[string]interface{}{
										"value":     353.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     357.0,
										"timestamp": 1650883535202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi02,
						"doc_count": 11,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     9,
									"value": map[string]interface{}{
										"value":     1351.0,
										"timestamp": 1650883475202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     1359.0,
										"timestamp": 1650883555202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetricsInf,
						"doc_count": 11,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     9,
									"value": map[string]interface{}{
										"value":     352.0,
										"timestamp": 1650883475202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     360.0,
										"timestamp": 1650883555202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi01,
						"doc_count": 8,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     1351.0,
										"timestamp": 1650883475202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"value":     1360.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetrics02,
						"doc_count": 8,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     4,
									"value": map[string]interface{}{
										"value":     352.0,
										"timestamp": 1650883485202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     4,
									"value": map[string]interface{}{
										"value":     360.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	nDslResult2 = map[string]interface{}{
		"took":      1,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      1,
			"successful": 1,
			"skipped":    0,
			"failed":     0,
		},
		"hits": map[string]interface{}{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       keyMetrics02,
						"doc_count": 16,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     15,
									"value": map[string]interface{}{
										"value":     353.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     358.0,
										"timestamp": 1650883545202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi01,
						"doc_count": 11,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     8,
									"value": map[string]interface{}{
										"value":     1352.0,
										"timestamp": 1650883485202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     3,
									"value": map[string]interface{}{
										"value":     1359.0,
										"timestamp": 1650883555202,
									},
								},
							},
						},
					},
					{
						"key":       keyApi02,
						"doc_count": 11,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     8,
									"value": map[string]interface{}{
										"value":     1353.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     3,
									"value": map[string]interface{}{
										"value":     1360.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
					{
						"key":       keyApiInf,
						"doc_count": 8,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     7,
									"value": map[string]interface{}{
										"value":     1349.0,
										"timestamp": 1650883445202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     1355.0,
										"timestamp": 1650883505202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetrics01,
						"doc_count": 8,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     5,
									"value": map[string]interface{}{
										"value":     350.0,
										"timestamp": 1650883465202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     3,
									"value": map[string]interface{}{
										"value":     360.0,
										"timestamp": 1650883565202,
									},
								},
							},
						},
					},
					{
						"key":       keyMetricsInf,
						"doc_count": 7,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": key20220425T104000,
									"key":           1650883200000,
									"doc_count":     6,
									"value": map[string]interface{}{
										"value":     354.0,
										"timestamp": 1650883495202,
									},
								},
								{
									"key_as_string": key20220425T104500,
									"key":           1650883500000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"value":     356.0,
										"timestamp": 1650883515202,
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

// histogram_quantile 整个算子测试，从接口开始 ,ok
func TestFuncHistogramQuantile(t *testing.T) {
	Convey("Test HistogramQuantile function in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+prometheus_http_request_duration_seconds_bucket", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		expected := `{
		  "status": "success",
		  "data": {
			"resultType": "vector",
			"result": [
			  {
				"metric": {
				  "handler": "/api/v1/query"
				},
				"value": [
				  1652320554,
				  "0.09006617647058825"
				]
			  },
			  {
				"metric": {
				  "handler": "/metrics"
				},
				"value": [
				  1652320554,
				  "0.09025000000000001"
				]
			  }
			]
		  }
		}`

		type Metric struct {
			Handler string `json:"handler"`
		}

		type Result struct {
			Metric Metric        `json:"metric"`
			Value  []interface{} `json:"value"`
		}

		type Data struct {
			ResultType string   `json:"resultType"`
			Result     []Result `json:"result"`
		}

		type ExpectedResult struct {
			Status string `json:"status"`
			Data   Data   `json:"data"`
		}

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: index,
			Pri:       "3",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
			logGroupQueryFilters, true, nil)
		dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
			notEmptyJson, http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard0).Times(1).Return(convert.MapToByte(nDslResult0), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard1).Times(1).Return(convert.MapToByte(nDslResult1), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard2).Times(1).Return(convert.MapToByte(nDslResult2), http.StatusOK, nil)

		body := []byte(`query=histogram_quantile(0.9, prometheus_http_request_duration_seconds_bucket)` +
			`&time=1652320554&ar_dataview="a"`)
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		var actual ExpectedResult
		err := sonic.Unmarshal(w.Body.Bytes(), &actual)
		So(err, ShouldBeNil)

		sort.Slice(actual.Data.Result, func(i, j int) bool {
			return actual.Data.Result[i].Metric.Handler < actual.Data.Result[j].Metric.Handler
		})
		str, _ := sonic.Marshal(actual)

		So(replace(string(str)), ShouldEqual, replace(expected))
	})
}

// increase 整个算子测试，从接口开始 ,ok
func TestFuncIncrease(t *testing.T) {
	Convey("test increase operator\n", t, func() {
		test := setGinMode()
		defer test()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		rateDslResult0 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320539541,
											"lastValue":         20.0,
											"lastTimestamp":     1652320541541,
											"counterCorrection": 2.3,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320543541,
											"lastValue":         18.0,
											"lastTimestamp":     1652320544541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        4.0,
											"firstTimestamp":    1652320540541,
											"lastValue":         40.0,
											"lastTimestamp":     1652320541541,
											"counterCorrection": 0.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        2.0,
											"firstTimestamp":    1652320542541,
											"lastValue":         20.0,
											"lastTimestamp":     1652320544541,
											"counterCorrection": 1.29,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult1 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        19.0,
											"firstTimestamp":    1652320545001,
											"lastValue":         21.0,
											"lastTimestamp":     1652320547541,
											"counterCorrection": 7.0,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        22.0,
											"firstTimestamp":    1652320549541,
											"lastValue":         19.0,
											"lastTimestamp":     1652320550541,
											"counterCorrection": 22.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        20.0,
											"firstTimestamp":    1652320546541,
											"lastValue":         41.0,
											"lastTimestamp":     1652320547541,
											"counterCorrection": 0.0,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        42.0,
											"firstTimestamp":    1652320550141,
											"lastValue":         44.0,
											"lastTimestamp":     1652320550541,
											"counterCorrection": 2.0,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult2 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320551000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":        20.0,
											"firstTimestamp":    1652320551141,
											"lastValue":         22.0,
											"lastTimestamp":     1652320553541,
											"counterCorrection": 2.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        23.0,
											"firstTimestamp":    1652320555141,
											"lastValue":         33.0,
											"lastTimestamp":     1652320556541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 7,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320551000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":        10.0,
											"firstTimestamp":    1652320551141,
											"lastValue":         42.0,
											"lastTimestamp":     1652320553541,
											"counterCorrection": 2.0,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     1,
										"value": map[string]interface{}{
											"firstValue":        12.0,
											"firstTimestamp":    1652320556541,
											"lastValue":         12.0,
											"lastTimestamp":     1652320556541,
											"counterCorrection": 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		emptyDslResult = map[string]interface{}{
			"aggregations": map[string]interface{}{
				"labels": map[string]interface{}{"buckets": []string{}},
			},
		}

		// increase query with range query, no data
		Convey("1. increase with no data and range query", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=increase(opensearch_indices_shards_docs[6s])` +
				`&start=1652032053&end=1652032153&step=3s&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// increase query with range query, have data
		Convey("2. increase with multiple point and range query", func() {
			expected := `{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "cluster": "opensearch",
          "index": "nas_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "values": [
          [
            1652320539,
            "34.45881272727273"
          ],
          [
            1652320542,
            "19.636363636363633"
          ],
          [
            1652320545,
            "31.40794223826715"
          ],
          [
            1652320548,
            "26.18181818181818"
          ],
          [
            1652320551,
            "16.666666666666664"
          ],
          [
            1652320554,
            "12"
          ]
        ]
      },
      {
        "metric": {
          "cluster": "opensearch",
          "index": "node_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "values": [
          [
            1652320539,
            "63.8940909090909"
          ],
          [
            1652320542,
            "45.988622"
          ],
          [
            1652320545,
            "28.36363636363636"
          ],
          [
            1652320548,
            "52.36363636363636"
          ],
          [
            1652320551,
            "51.1111111111111"
          ]
        ]
      }
    ]
  }
}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(rateDslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(rateDslResult2), http.StatusOK, nil)

			body := []byte(`query=increase(opensearch_indices_shards_docs[6s])` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// increase query with instant query, no data
		Convey("3. increase with no data and instant query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=increase(opensearch_indices_shards_docs[6s])` +
				`&time=1652032053&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// increase query with instant query, have data
		Convey("4. increase with multiple point and instant query", func() {
			expected := `{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "cluster": "opensearch",
          "index": "nas_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "value": [
          1652320544,
          "18.18"
        ]
      },
      {
        "metric": {
          "cluster": "opensearch",
          "index": "node_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "value": [
          1652320544,
          "38.99756340909091"
        ]
      }
    ]
  }
}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=increase(opensearch_indices_shards_docs[3s])` +
				`&time=1652320544&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

// changes 整个算子测试，从接口开始 ,ok
func TestFuncChanges(t *testing.T) {
	Convey("test changes operator\n", t, func() {
		test := setGinMode()
		defer test()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		rateDslResult0 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     10.0,
											"firstTimestamp": 1652320539541,
											"lastValue":      20.0,
											"lastTimestamp":  1652320541541,
											"changes":        2,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     10.0,
											"firstTimestamp": 1652320543541,
											"lastValue":      18.0,
											"lastTimestamp":  1652320544541,
											"changes":        3,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320539000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     4.0,
											"firstTimestamp": 1652320540541,
											"lastValue":      40.0,
											"lastTimestamp":  1652320541541,
											"changes":        4,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320542000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     2.0,
											"firstTimestamp": 1652320542541,
											"lastValue":      20.0,
											"lastTimestamp":  1652320544541,
											"changes":        5,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult1 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":     19.0,
											"firstTimestamp": 1652320545001,
											"lastValue":      21.0,
											"lastTimestamp":  1652320547541,
											"changes":        6,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":     22.0,
											"firstTimestamp": 1652320549541,
											"lastValue":      19.0,
											"lastTimestamp":  1652320550541,
											"changes":        7,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320545000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":     20.0,
											"firstTimestamp": 1652320546541,
											"lastValue":      41.0,
											"lastTimestamp":  1652320547541,
											"changes":        8,
										},
									},
									{
										"key":       1652320548000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":     42.0,
											"firstTimestamp": 1652320550141,
											"lastValue":      44.0,
											"lastTimestamp":  1652320550541,
											"changes":        9,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		rateDslResult2 := map[string]interface{}{
			"aggregations": map[string]interface{}{
				interfaces.LABELS_STR: map[string]interface{}{
					"buckets": []map[string]interface{}{
						{
							"key":       labelsStrNas,
							"doc_count": 12,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key":       1652320551000,
										"doc_count": 6,
										"value": map[string]interface{}{
											"firstValue":     20.0,
											"firstTimestamp": 1652320551141,
											"lastValue":      22.0,
											"lastTimestamp":  1652320553541,
											"changes":        10,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     23.0,
											"firstTimestamp": 1652320555141,
											"lastValue":      33.0,
											"lastTimestamp":  1652320556541,
											"changes":        11,
										},
									},
								},
							},
						},
						{
							"key":       labelsStrNode,
							"doc_count": 7,
							"time": map[string]interface{}{
								"buckets": []map[string]interface{}{
									{
										"key_as_string": timeString39,
										"key":           1652320551000,
										"doc_count":     6,
										"value": map[string]interface{}{
											"firstValue":     10.0,
											"firstTimestamp": 1652320551141,
											"lastValue":      42.0,
											"lastTimestamp":  1652320553541,
											"changes":        12,
										},
									},
									{
										"key_as_string": timeString42,
										"key":           1652320554000,
										"doc_count":     1,
										"value": map[string]interface{}{
											"firstValue":     12.0,
											"firstTimestamp": 1652320556541,
											"lastValue":      12.0,
											"lastTimestamp":  1652320556541,
											"changes":        13,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		emptyDslResult = map[string]interface{}{
			"aggregations": map[string]interface{}{
				"labels": map[string]interface{}{"buckets": []string{}},
			},
		}

		// changes query with range query, no data
		Convey("1. changes with no data and range query", func() {

			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=changes(opensearch_indices_shards_docs[6s])` +
				`&start=1652032053&end=1652032153&step=3s&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// changes query with range query, have data
		Convey("2. changes with multiple point and range query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"6"
								],
								[
									1652320542,
									"10"
								],
								[
									1652320545,
									"14"
								],
								[
									1652320548,
									"18"
								],
								[
									1652320551,
									"22"
								],
								[
									1652320554,
									"11"
								]
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"values": [
								[
									1652320539,
									"10"
								],
								[
									1652320542,
									"13"
								],
								[
									1652320545,
									"18"
								],
								[
									1652320548,
									"22"
								],
								[
									1652320551,
									"26"
								],
								[
									1652320554,
									"13"
								]
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(rateDslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(rateDslResult2), http.StatusOK, nil)

			body := []byte(`query=changes(opensearch_indices_shards_docs[6s])` +
				`&start=1652320539&end=1652320554&step=3s&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query_range"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// changes query with instant query, no data
		Convey("3. changes with no data and instant query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`

			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=changes(opensearch_indices_shards_docs[6s])` +
				`&time=1652032053&ar_dataview="a"`)
			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		// changes query with instant query, have data
		Convey("4. changes with multiple point and instant query", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"cluster": "opensearch",
								"index": "nas_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320544,
									"6"
							]
						},
						{
							"metric": {
								"cluster": "opensearch",
								"index": "node_statis-2022.05-0",
								"instance": "instance_abc",
								"job": "prometheus"
							},
							"value": [
									1652320544,
									"10"
							]
						}
					]
				}
			}`
			var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
				IndexName: index,
				Pri:       "3",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard0).Times(1).Return(convert.MapToByte(rateDslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			body := []byte(`query=changes(opensearch_indices_shards_docs[3s])` +
				`&time=1652320544&ar_dataview="a"`)

			url := "/api/mdl-uniquery/v1/promql/query"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

// Math 整个算子测试，从接口开始 ,ok
func TestFuncMath(t *testing.T) {
	Convey("Test Math function in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"22"
							]
						}
					]
				}
			}`

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: index,
			Pri:       "3",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
			logGroupQueryFilters, true, nil)
		dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
			notEmptyJson, http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard0).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard1).Times(1).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard2).Times(1).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

		Convey("floor simple \n", func() {
			body := []byte(`query=floor(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("ceil simple \n", func() {
			body := []byte(`query=ceil(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)

		})

		Convey("abs simple \n", func() {
			body := []byte(`query=abs(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("exp simple \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"3584912847"
							]
						}
					]
				}
			}`
			body := []byte(`query=ceil(exp(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"})))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("sqrt simple \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"4.69041575982343"
							]
						}
					]
				}
			}`
			body := []byte(`query=sqrt(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("ln simple \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"3.091042453358316"
							]
						}
					]
				}
			}`
			body := []byte(`query=ln(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("log2 simple \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"4.459431618637297"
							]
						}
					]
				}
			}`
			body := []byte(`query=log2(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("log10 simple \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
									1652320554,
									"1.3424226808222064"
							]
						}
					]
				}
			}`
			body := []byte(`query=log10(max(opensearch_indices_shards_docs{labels.index="nas_statis-2022.05-0"}))` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}

func TestFuncLogic(t *testing.T) {
	Convey("Test logic function in query ", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		logics.SetOpenSearchAccess(osaMock)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		logics.SetLogGroupAccess(lgaMock)

		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
			ViewPoolSize:        10,
		})

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
			PoolSetting: common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
				ViewPoolSize:        10,
			}, PromqlSetting: common.PromqlSetting{}}

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+opensearch_indices_shards_docs", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat12-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat12-*"] = time.Now()

		dvsMock := umock.NewMockDataViewService(mockCtrl)
		ln := leafnodes.NewLeafNodes(appSetting, osaMock, lgaMock, dvsMock)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		ps := promql.NewPromQLServiceRaw(appSetting, ln, mmsMock)

		hydraMock := rmock.NewMockHydra(mockCtrl)
		handler := mockNewPromQLRestHandler(appSetting, hydraMock, ps)
		handler.RegisterPublic(engine)

		hydraMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-uniquery/v1/promql/query"

		expected := `{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "cluster": "opensearch",
          "index": "nas_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "value": [
          1652320554,
          "12"
        ]
      },
      {
        "metric": {
          "cluster": "opensearch",
          "index": "node_statis-2022.05-0",
          "instance": "instance_abc",
          "job": "prometheus"
        },
        "value": [
          1652320554,
          "22"
        ]
      }
    ]
  }
}`

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: index,
			Pri:       "3",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
			logGroupQueryFilters, true, nil)
		dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
			notEmptyJson, http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

		dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

		Convey("binary and \n", func() {
			body := []byte(`query=opensearch_indices_shards_docs and opensearch_indices_shards_docs` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("binary or \n", func() {
			body := []byte(`query=opensearch_indices_shards_docs or opensearch_indices_shards_docs` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})

		Convey("binary unless \n", func() {
			expected := `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`
			body := []byte(`query=opensearch_indices_shards_docs unless opensearch_indices_shards_docs` +
				`&time=1652320554&ar_dataview="a"`)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_FORM)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			compareJsonString(w.Body.String(), expected)
		})
	})
}
