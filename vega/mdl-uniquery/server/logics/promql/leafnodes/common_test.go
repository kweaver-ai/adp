// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)
var (
	maxSearchSeriesSize = 10000

	onlyIndexPatternFilters = interfaces.LogGroup{
		IndexPattern: []string{
			"hahaha-*",
			"test_topic_r_0gx_e_0000-*",
		},
	}

	onlyManualIndexFilters = interfaces.LogGroup{
		IndexPattern: []string{"a", "b"},
	}

	bothFilters = interfaces.LogGroup{
		IndexPattern: []string{
			"hahaha-*",
			"test_topic_r_0gx_e_0000-*",
			"a", "b",
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

func TestGetIndicesNumberOfShards(t *testing.T) {
	Convey("test getIndicesNumberOfShards ", t, func() {

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
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Convey("success load from Number_Of_Shards_Map ", func() {
			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
				IndexName: indexPattern,
				Pri:       "1",
			})
			Number_Of_Shards_Map.Store(indexPattern, indexShardsArr)
			defer Number_Of_Shards_Map.Delete(indexPattern)

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(len(indexArr), ShouldBeGreaterThan, 0)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("load from opensearch failed ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(indexArr, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)

		})

		Convey("load from opensearch successfully ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(len(indexArr), ShouldBeGreaterThan, 0)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Unmarshal failed ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]byte{}, http.StatusOK, nil)

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{"12"})
			So(indexArr, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRefresh(t *testing.T) {
	Convey("test refresh ", t, func() {
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
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Number_Of_Shards_Map.Store(indexPattern, indexShardsArr)
		defer Number_Of_Shards_Map.Delete(indexPattern)

		value, ok := Number_Of_Shards_Map.Load(indexPattern)
		So(ok, ShouldBeTrue)
		Convey("load from opensearch failed ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			bo := lnMock.Refresh(indexPattern, value)

			So(bo, ShouldBeFalse)
		})

		Convey("Unmarshal failed ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]byte{}, http.StatusOK, nil)

			bo := lnMock.Refresh(indexPattern, value)
			So(bo, ShouldBeFalse)
		})

		Convey("load from opensearch successfully ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			bo := lnMock.Refresh(indexPattern, value)

			So(bo, ShouldBeTrue)
		})
	})
}

func TestRefreshShards(t *testing.T) {
	Convey("test RefreshShards ", t, func() {

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
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Convey("success RefreshShards when Number_Of_Shards_Map is empty ", func() {
			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
				IndexName: indexPattern,
				Pri:       "1",
			})
			Number_Of_Shards_Map.Store(indexPattern, indexShardsArr)
			defer Number_Of_Shards_Map.Delete(indexPattern)

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(len(indexArr), ShouldBeGreaterThan, 0)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("load from opensearch failed ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(indicesError))

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(indexArr, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)

		})

		Convey("load from opensearch successfully ", func() {
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			indexArr, status, err := lnMock.GetIndicesNumberOfShards(testCtx, []string{indexPattern})

			So(len(indexArr), ShouldBeGreaterThan, 0)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func replace(s string) string {
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}

func TestMakeDsl(t *testing.T) {
	Convey("test MakeDsl ", t, func() {

		Convey("makeAggregation error ", func() {
			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, "ratem", mustFilter, false)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("Marshal must_filter error", func() {
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					return nil, fmt.Errorf("a")
				},
			)
			defer patch.Reset()

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg,
				mustFilter, false)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("AppendFilters error, unsupport operation", func() {
			query1 := &interfaces.Query{
				Start:         1646360670,
				End:           1646360700,
				Interval:      30000,
				LogGroupId:    "a",
				QueryStr:      `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`,
				Filters:       []interfaces.Filter{{Operation: "a"}},
				IsMetricModel: true,
			}
			expr, _ := parser.ParseExpr(testCtx, query1.QueryStr)

			// query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			// expr, _ := parser.ParseExpr(testCtx,query.QueryStr)
			// query.Filters = []interfaces.Filter{{Operation: "a"}}
			// query.IsMetricModel = true

			res, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query1, groupby, samplingAgg, mustFilter, false)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

			query.Filters = []interfaces.Filter{}
		})

		Convey("MakeDSL correctly ", func() {
			// 	expectDSL := fmt.Sprintf(`
			// 	{
			// 		"size":0,
			// 		"query": {
			// 			"bool": {
			// 				"filter": [
			// 					{
			// 						"term": {
			// 							"labels.cluster.keyword": "txy"
			// 						}
			// 					 },
			// 					{
			// 						"bool": {
			// 							"must_not": {
			// 								"regexp": {
			// 									"labels.name.keyword": "^node.*"
			// 								}
			// 							}
			// 						}
			// 					},
			// 					{
			// 						"exists": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 					},
			// 					{
			// 						"range": {
			// 							"@timestamp": {
			// 								"gte":1646360670,
			// 								"lt":1646360700
			// 							}
			// 						}
			// 					}
			// 				],
			// 				"must": [
			// 					{
			// 						"query_string": {
			// 							"analyze_wildcard": true,
			// 							"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 						}
			// 					}
			// 				]
			// 			}
			// 		}
			// 		,
			// 			"aggs": {
			// 				"%s": {
			// 					"terms": {
			// 						"field": "%s",
			// 						"size": 10000
			// 					},

			// 			"aggs": {
			// 				"time": {
			// 					"date_histogram": {
			// 						"field": "@timestamp",
			// 						"fixed_interval": "30000ms",
			// 						"min_doc_count": 1,
			// 						"time_zone":"Asia/Shanghai",
			// 						"order": {
			// 							"_key": "asc"
			// 						}
			// 					},
			// 				"aggs": {
			// 					"value": {

			// 							"sampling": {
			// 								"value": {
			// 									"field": "metrics.elasticsearch_os_cpu_percent"
			// 								},
			// 								"timestamp": {
			// 									"field": "@timestamp"
			// 								}
			// 							}
			// 					}
			// 				}
			// 			}
			// 		}

			// 			}
			// 		}
			// 	}
			// `, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// 	expectDSL = replace(expectDSL)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with empty filter ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "30000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"sampling": {
			// 							"value": {
			// 								"field": "metrics.elasticsearch_os_cpu_percent"
			// 							},
			// 							"timestamp": {
			// 								"field": "@timestamp"
			// 							}
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, samplingAgg, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with rate agg ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "30000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"rate_sampling": {
			// 							"value": {
			// 								"field": "metrics.elasticsearch_os_cpu_percent"
			// 							},
			// 							"timestamp": {
			// 								"field": "@timestamp"
			// 							}
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 30000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.RATE_AGG, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with changes agg ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"changes_sampling": {
			// 							"value": {
			// 								"field": "metrics.elasticsearch_os_cpu_percent"
			// 							},
			// 							"timestamp": {
			// 								"field": "@timestamp"
			// 							}
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.CHANGES_AGG, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with avg_over_time ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"sum": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.AVG_OVER_TIME, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with sum_over_time ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"sum": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.SUM_OVER_TIME, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with max_over_time ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"max": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.MAX_OVER_TIME, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with min_over_time ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"min": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.MIN_OVER_TIME, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with count_over_time ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "3000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"value_count": {
			// 							"field": "metrics.elasticsearch_os_cpu_percent"
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 3000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.COUNT_OVER_TIME, mustFilter, false)
			//So(replace(res.String()), ShouldEqual, expectDSL)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("MakeDSL correctly with delta agg ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"exists": {
			// 						"field": "metrics.elasticsearch_os_cpu_percent"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1646360670,
			// 							"lt":1646360700
			// 						}
			// 					}
			// 				}
			// 			],
			// 			"must": [
			// 				{
			// 					"query_string": {
			// 						"analyze_wildcard": true,
			// 						"query": "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")"
			// 					}
			// 				}
			// 			]
			// 		}
			// 	}
			// 	,
			// 		"aggs": {
			// 			"%s": {
			// 				"terms": {
			// 					"field": "%s",
			// 					"size": 10000
			// 				},
			// 		"aggs": {
			// 			"time": {
			// 				"date_histogram": {
			// 					"field": "@timestamp",
			// 					"fixed_interval": "30000ms",
			// 					"min_doc_count": 1,
			// 					"time_zone":"Asia/Shanghai",
			// 					"order": {
			// 						"_key": "asc"
			// 					}
			// 				},
			// 			"aggs": {
			// 				"value": {
			// 						"delta_sampling": {
			// 							"value": {
			// 								"field": "metrics.elasticsearch_os_cpu_percent"
			// 							},
			// 							"timestamp": {
			// 								"field": "@timestamp"
			// 							}
			// 						}
			// 				}
			// 			}
			// 		}
			// 	}

			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			// expectDSL = replace(expectDSL)

			query.SubIntervalWith30min = 30000
			query.QueryStr = `elasticsearch_os_cpu_percent`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			_, status, err := makeDSL(*(expr.(*parser.VectorSelector)), query, groupby, interfaces.DELTA_AGG, mustFilter, false)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestMakeDSLQuery(t *testing.T) {
	Convey("test makeDSLQuery ", t, func() {

		Convey("makeDSLQuery correctly ", func() {
			expectDslQuery := `{
				"size":0,
				"query": {
					"bool": {
						"filter": [
							{
								"term": {
									"labels.cluster.keyword": "txy"
								}
							 },
							{
								"bool" : {
									"must_not" : {
										"term" : {
											"labels.host.keyword": "localhost"
										}
									}
								}
							},
							{
								"regexp": {
									"labels.instance.keyword": "localhost:.*"
								}
							},
							{
								"bool": {
									"must_not": {
										"regexp": {
											"labels.name.keyword": "^node.*"
										}
									}
								}
							},
							{
								"exists": {
									"field": "metrics.elasticsearch_os_cpu_percent"
								}
							},`
			expectDslQuery = replace(expectDslQuery)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",host!="localhost",instance=~"localhost:.*",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			resDslQuery, metricName := makeDSLQuery(*(expr.(*parser.VectorSelector)))

			So(replace(resDslQuery), ShouldEqual, expectDslQuery)
			So(metricName, ShouldNotBeEmpty)
		})

	})
}

func TestMakeAggregation(t *testing.T) {
	Convey("test makeAggregation ", t, func() {

		Convey("groupby is empty", func() {
			var resAgg bytes.Buffer
			status, err := makeAggregation([]string{}, samplingAgg,
				"elasticsearch_os_cpu_percent", 30000, maxSearchSeriesSize, &resAgg)
			So(resAgg, ShouldNotBeEmpty)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("groupby is not __labels_str", func() {
			var resAgg bytes.Buffer
			status, err := makeAggregation([]string{"cluster"}, samplingAgg,
				"elasticsearch_os_cpu_percent", 30000, maxSearchSeriesSize, &resAgg)
			So(resAgg, ShouldNotBeEmpty)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("aggregationType is empty", func() {
			var resAgg bytes.Buffer
			status, err := makeAggregation(groupby, "", "elasticsearch_os_cpu_percent",
				30000, maxSearchSeriesSize, &resAgg)
			// So(resAgg, ShouldBeEmpty)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})
		Convey("aggregationType is sampling", func() {
			expectStr := fmt.Sprintf(`,
				"aggs": {
					"%s": {
						"terms": {
							"field": "%s",
							"order": { "_key": "asc" },
							"size": 10000
						},

				"aggs": {
					"time": {
						"date_histogram": {
							"field": "@timestamp",
							"fixed_interval": "30000ms",
							"min_doc_count": 1,
							"time_zone":"Asia/Shanghai",
							"order": {
								"_key": "asc"
							}
						},
					"aggs": {
						"value": {

								"sampling": {
									"value": {
										"field": "metrics.elasticsearch_os_cpu_percent"
									},
									"timestamp": {
										"field": "@timestamp"
									}
								}
						}
					}
				}
			}

				}
			}
		}
	`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))
			expectStr = replace(expectStr)

			var resAgg bytes.Buffer
			status, err := makeAggregation(groupby, samplingAgg, "elasticsearch_os_cpu_percent", 30000, maxSearchSeriesSize, &resAgg)

			So(replace(resAgg.String()), ShouldEqual, expectStr)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestAppendFilters(t *testing.T) {
	Convey("test appendFilters ", t, func() {

		// fields := map[string]string{
		// 	"a":              "text",
		// 	"b":              "text",
		// 	"c":              "keyword",
		// 	"d":              "float",
		// 	"d2":             "keyword",
		// 	"d2_desensitize": "keyword",
		// 	"e":              "text",
		// 	"e_desensitize":  "keyword",
		// 	"f":              "long",
		// 	"f_desensitize":  "long",
		// 	"g":              "keyword",
		// 	"h":              "long",
		// }

		fieldsMap := map[string]*cond.ViewField{
			"a":              {Name: "a", Type: "text", OriginalName: "a"},
			"b":              {Name: "a", Type: "text", OriginalName: "b"},
			"c":              {Name: "a", Type: "keyword", OriginalName: "c"},
			"d":              {Name: "a", Type: "float", OriginalName: "d"},
			"d2":             {Name: "a", Type: "keyword", OriginalName: "d2"},
			"d2_desensitize": {Name: "a", Type: "keyword", OriginalName: "d2_desensitize"},
			"e":              {Name: "a", Type: "text", OriginalName: "e"},
			"e_desensitize":  {Name: "a", Type: "keyword", OriginalName: "e_desensitize"},
			"f":              {Name: "a", Type: "long", OriginalName: "f"},
			"f_desensitize":  {Name: "a", Type: "long", OriginalName: "f_desensitize"},
			"g":              {Name: "a", Type: "keyword", OriginalName: "g"},
			"h":              {Name: "a", Type: "long", OriginalName: "h"},
		}
		Convey("query.IsMetricModel is false return \"\" ", func() {
			filtersStr, status, err := AppendFilters(interfaces.Query{IsMetricModel: false, Filters: []interfaces.Filter{}})
			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("filters is empty", func() {
			filtersStr, status, err := AppendFilters(interfaces.Query{IsMetricModel: true, Filters: []interfaces.Filter{}})
			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("operation is unsupport", func() {
			filtersStr, status, err := AppendFilters(interfaces.Query{
				IsMetricModel: true,
				Filters:       []interfaces.Filter{{Operation: "<>"}},
			})
			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("filters length eq 1", func() {
			expectedStr := `{
				"bool": {
					"should": [
						{
							"term": {
								"a.keyword": {
									"value": "a"
								}
							}
						},
						{
							"term": {
								"a.keyword": {
									"value": "b"
								}
							}
						}
					]
				}
			}, `
			expectedStr = replace(expectedStr)
			filtersStr, status, err := AppendFilters(interfaces.Query{
				IsMetricModel: true,
				Filters: []interfaces.Filter{
					{
						Name:      "a",
						Value:     []interface{}{"a", "b"},
						Operation: interfaces.OPERATION_IN,
					},
				},
				DataView: interfaces.DataView{FieldsMap: fieldsMap}})

			So(replace(filtersStr), ShouldEqual, expectedStr)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("filters length gte 1", func() {
			expectedStr := `{
				"bool": {
					"should": [
						{
							"term": {
								"a.keyword": {
									"value": "a"
								}
							}
						},
						{
							"term": {
								"a.keyword": {
									"value": "b"
								}
							}
						}
					]
				}
			},
			{
				"bool": {
					"should": [
						{
							"term": {
								"b.keyword": {
									"value": "d"
								}
							}
						}
					]
				}
			},
			{
				"term": {
					"c": {
						"value": "d"
					}
				}
			},
			{
				"bool": {
					"must_not": [
						{
							"term": {
								"d": {
									"value": 2
								}
							}
						}
					]
				}
			},
			{
				"bool": {
					"must_not": [
						{
							"term": {
								"d2_desensitize": {
									"value": "2"
								}
							}
						}
					]
				}
			},
			{
				"range": {
					"e_desensitize.keyword": {
						"gte": "1",
						"lt": "2"
					}
				}
			},
			{
				"range": {
					"f_desensitize": {
						"gte": 1,
						"lt": 2
					}
				}
			},
			{
				"bool": {
					"should": [
						{
							"range": {
								"g": {
									"lt": "1"
								}
							}
						},
						{
							"range": {
								"g": {
									"gte": "2"
								}
							}
						}
					]
				}
			},
			{
				"bool": {
					"should": [
						{
							"range": {
								"h": {
									"lt": 1
								}
							}
						},
						{
							"range": {
								"h": {
									"gte": 2
								}
							}
						}
					]
				}
			},
			{
				"regexp": {
					"d": ".*2.*"
				}
			},
			{
				"bool": {
					"must_not": [
						{
							"regexp": {
								"d2_desensitize": ".*2.*"
							}
						}
					]
				}
			},
			{
				"range": {
					"d": {
						"gt": 2
					}
				}
			},
			{
				"range": {
					"d2_desensitize": {
						"gt": "2"
					}
				}
			},
			{
				"range": {
					"g": {
						"gte": "3"
					}
				}
			},
			{
				"range": {
					"h": {
						"gte": 3
					}
				}
			},
			{
				"range": {
					"g": {
						"lt": "3"
					}
				}
			},
			{
				"range": {
					"h": {
						"lt": 3
					}
				}
			},
			{
				"range": {
					"g": {
						"lte": "3"
					}
				}
			},
			{
				"range": {
					"h": {
						"lte": 3
					}
				}
			},  `
			expectedStr = replace(expectedStr)
			filtersStr, status, err := AppendFilters(interfaces.Query{
				IsMetricModel: true,
				Filters: []interfaces.Filter{
					{Name: "a", Value: []interface{}{"a", "b"}, Operation: interfaces.OPERATION_IN},
					{Name: "b", Value: []interface{}{"d"}, Operation: interfaces.OPERATION_IN},
					{Name: "c", Value: "d", Operation: interfaces.OPERATION_EQ},
					{Name: "d", Value: 2, Operation: cond.OperationNotEq},
					{Name: "d2", Value: "2", Operation: cond.OperationNotEq},
					{Name: "e", Value: []interface{}{"1", "2"}, Operation: cond.OperationRange},
					{Name: "f", Value: []interface{}{1, 2}, Operation: cond.OperationRange},
					{Name: "g", Value: []interface{}{"1", "2"}, Operation: cond.OperationOutRange},
					{Name: "h", Value: []interface{}{1, 2}, Operation: cond.OperationOutRange},
					{Name: "d", Value: 2, Operation: cond.OperationLike},
					{Name: "d2", Value: "2", Operation: cond.OperationNotLike},
					{Name: "d", Value: 2, Operation: cond.OperationGt},
					{Name: "d2", Value: "2", Operation: cond.OperationGt},
					{Name: "g", Value: "3", Operation: cond.OperationGte},
					{Name: "h", Value: 3, Operation: cond.OperationGte},
					{Name: "g", Value: "3", Operation: cond.OperationLt},
					{Name: "h", Value: 3, Operation: cond.OperationLt},
					{Name: "g", Value: "3", Operation: cond.OperationLte},
					{Name: "h", Value: 3, Operation: cond.OperationLte},
				},
				DataView: interfaces.DataView{FieldsMap: fieldsMap}})

			So(replace(filtersStr), ShouldEqual, expectedStr)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("filters error, range operation and value size is ne 2", func() {

			filtersStr, status, err := AppendFilters(interfaces.Query{
				IsMetricModel: true,
				DataView: interfaces.DataView{
					FieldsMap: map[string]*cond.ViewField{
						"a": {Name: "a", Type: dtype.DataType_Text},
					},
				},
				Filters: []interfaces.Filter{{Name: "a", Value: []interface{}{"a"},
					Operation: cond.OperationRange}}})

			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("filters error, out_range operation and value size is ne 2", func() {

			filtersStr, status, err := AppendFilters(interfaces.Query{IsMetricModel: true,
				Filters: []interfaces.Filter{{Name: "a", Value: []interface{}{"a"},
					Operation: cond.OperationOutRange}}})

			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("filters error, unsupport operation", func() {

			filtersStr, status, err := AppendFilters(interfaces.Query{
				IsMetricModel: true,
				DataView: interfaces.DataView{
					FieldsMap: map[string]*cond.ViewField{
						"a": {Name: "a", Type: dtype.DataType_Text},
					},
				},
				Filters: []interfaces.Filter{{Name: "a", Value: []interface{}{"a", "b"},
					Operation: "a"}}})

			So(filtersStr, ShouldEqual, "")
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

	})
}

func TestParseLabelsStr(t *testing.T) {
	Convey("test ParseLabelsStr ", t, func() {
		Convey("the Connector of labelsStr is not '=' ", func() {
			key := "cluster:\"txy\",name=\"node-1\""
			labelsMap := make(map[string][]*labels.Label)
			lab := labels.Labels{
				{
					Name:  interfaces.LABELS_STR,
					Value: "cluster:\"txy\",name=\"node-1\"",
				},
			}
			labelArr := lab
			labelsMap[key] = labelArr

			actualLabelArr := parseLabelsStr(key, labelsMap)
			So(actualLabelArr, ShouldResemble, labelArr)

		})

		Convey("the Connector of labelsStr is '=' ", func() {
			expectLabelArr := labels.Labels{
				{
					Name:  "cluster",
					Value: "txy",
				},
				{
					Name:  "name",
					Value: "node-1",
				},
			}

			labelsMap := make(map[string][]*labels.Label)
			lab := &labels.Label{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}
			labelArr := []*labels.Label{lab}
			labelsMap[key] = labelArr

			actualLabelArr := parseLabelsStr(key, labelsMap)
			So(actualLabelArr, ShouldResemble, expectLabelArr)
		})

		Convey("the Connector of labelsStr is '=' 2 ", func() {
			expectLabelArr := labels.Labels{
				{
					Name:  "cluster",
					Value: "txy",
				},
				{
					Name:  "name",
					Value: "node-1",
				},
			}

			labelsMap := make(map[string][]*labels.Label)
			lab := &labels.Label{
				Name:  interfaces.LABELS_STR,
				Value: "labels.cluster=\"txy\",labels.name=\"node-1\"",
			}
			labelArr := []*labels.Label{lab}
			labelsMap[key] = labelArr

			actualLabelArr := parseLabelsStr(key, labelsMap)
			So(actualLabelArr, ShouldResemble, expectLabelArr)
		})

		Convey("__labels_str is empty ", func() {
			labelsMap := make(map[string][]*labels.Label)
			lab := &labels.Label{
				Name:  interfaces.LABELS_STR,
				Value: "",
			}
			labelArr := []*labels.Label{lab}
			labelsMap[key] = labelArr

			actualLabelArr := parseLabelsStr("", labelsMap)
			So(actualLabelArr, ShouldBeNil)
		})
	})
}

func TestIteratorTermsAgg(t *testing.T) {
	Convey("test IteratorTermsAgg", t, func() {
		Convey("bucket's number more than 0 ", func() {
			expectLabels := map[string][]*labels.Label{
				key: {
					{
						Name:  interfaces.LABELS_STR,
						Value: key,
					},
				},
			}

			agg := gjson.Result{
				Type: gjson.JSON,
				Raw: fmt.Sprintf("{\"%s\":{\"doc_count_error_upper_bound\":0,\"sum_other_doc_count\":0,\"buckets\":[{\"key\":\"cluster=\\\"txy\\\",name=\\\"node-1\\\"\",\"doc_count\":8,\"time\":{\"buckets\":[{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}},{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}},{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}},{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}]}}]}}",
					interfaces.LABELS_STR),
			}
			mapResult := &MapResult{
				LabelsMap:  make(map[string][]*labels.Label),
				TsValueMap: make(map[string][][]gjson.Result),
			}
			iteratorTermsAgg(agg, "", make([]*labels.Label, 0), 0, []string{interfaces.LABELS_STR}, mapResult)

			So(mapResult.LabelsMap, ShouldResemble, expectLabels)
			So(len(mapResult.TsValueMap[key]), ShouldEqual, 1)
			So(len(mapResult.TsValueMap[key][0]), ShouldEqual, 4)
		})

		Convey("groupby's len more than 1 ", func() {
			expectLabels := map[string][]*labels.Label{
				"txy|node-1": {
					{
						Name:  "cluster",
						Value: "txy",
					},
					{
						Name:  "name",
						Value: "node-1",
					},
				},
			}

			key := "txy|node-1"
			agg := gjson.Result{
				Type: gjson.JSON,
				Raw:  "{\"cluster\":{\"doc_count_error_upper_bound\":0,\"sum_other_doc_count\":0,\"buckets\":[{\"key\":\"txy\",\"doc_count\":8,\"name\":{\"doc_count_error_upper_bound\":0,\"sum_other_doc_count\":0,\"buckets\":[{\"key\":\"node-1\",\"doc_count\":8,\"time\":{\"buckets\":[{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}},{\"key_as_string\":\"2022-03-04T02:25:00.000Z\",\"key\":1646360700000,\"doc_count\":2,\"value\":{\"value\":9.0,\"timestamp\":1646360728347}},{\"key_as_string\":\"2022-03-04T02:25:30.000Z\",\"key\":1646360730000,\"doc_count\":2,\"value\":{\"value\":14.0,\"timestamp\":1646360758347}},{\"key_as_string\":\"2022-03-04T02:26:00.000Z\",\"key\":1646360760000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360788347}}]}}]}}]}}",
			}
			mapResult := &MapResult{
				LabelsMap:  make(map[string][]*labels.Label),
				TsValueMap: make(map[string][][]gjson.Result),
			}
			iteratorTermsAgg(agg, "", make([]*labels.Label, 0), 0,
				[]string{"cluster", "name"}, mapResult)

			So(mapResult.LabelsMap, ShouldResemble, expectLabels)
			So(len(mapResult.TsValueMap[key]), ShouldEqual, 1)
			So(len(mapResult.TsValueMap[key][0]), ShouldEqual, 4)
		})
	})
}

func TestCheckInstantQueryInterval(t *testing.T) {
	Convey("test check instant query interval", t, func() {
		Convey("1. interval is <= 120min, and is not divisible by 5m", func() {
			interval := 19 * time.Minute
			err := checkInstantQueryInterval(interval)
			So(err, ShouldBeNil)
		})
		Convey("2. interval is <= 120min, and is divisible by 5m", func() {
			interval := 120 * time.Minute
			err := checkInstantQueryInterval(interval)
			So(err, ShouldBeNil)
		})

		Convey("3. interval is <= 120min, and is divisible by 5m", func() {
			interval := 5 * time.Minute
			err := checkInstantQueryInterval(interval)
			So(err, ShouldBeNil)
		})

		Convey("4. interval is > 120min, and is not divisible by 5m", func() {
			interval := 121 * time.Minute
			err := checkInstantQueryInterval(interval)
			So(err, ShouldNotBeNil)
		})

		Convey("5. interval is > 120min, and is divisible by 5m", func() {
			interval := 125 * time.Minute
			err := checkInstantQueryInterval(interval)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetDataViewQueryFilters(t *testing.T) {
	Convey("test GetLogGroupQueryFilters ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		Convey("dataViewIds is \"\"", func() {
			dataView, status, err := lnMock.GetLogGroupQueryFilters(testCtx, "")

			So(len(dataView.IndexPattern), ShouldEqual, 0)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, "missing ar_dataview parameter or ar_dataview parameter value cannot be empty")
		})

		Convey("response with only index_pattern ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				onlyIndexPatternFilters, true, nil)

			dataview, status, err := lnMock.GetLogGroupQueryFilters(testCtx, "a")

			So(dataview.IndexPattern, ShouldResemble, []string{"hahaha-*", "test_topic_r_0gx_e_0000-*"})
			So(dataview.MustFilter, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("response with only manual_index ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				onlyManualIndexFilters, true, nil)

			dataview, status, err := lnMock.GetLogGroupQueryFilters(testCtx, "a")

			So(dataview.IndexPattern, ShouldResemble, []string{"a", "b"})
			So(dataview.MustFilter, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("response with index_pattern, manual_index and must_filter ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				bothFilters, true, nil)

			dataview, status, err := lnMock.GetLogGroupQueryFilters(testCtx, "a")

			So(dataview.IndexPattern, ShouldResemble, []string{"hahaha-*", "test_topic_r_0gx_e_0000-*", "a", "b"})
			So(dataview.MustFilter, ShouldResemble, mustFilter)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetPointsStatsInfo(t *testing.T) {
	Convey("test getPointsStatsInfo ", t, func() {
		Convey("tsArr is empty ", func() {

			start, end, totalNum := getPointsStatsInfo([][]gjson.Result{})
			So(start, ShouldEqual, math.MaxInt64)
			So(end, ShouldEqual, math.MinInt64)
			So(totalNum, ShouldEqual, 0)
		})

		Convey("tsArr contain one element ", func() {
			tsArr := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360670000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}
			start, end, totalNum := getPointsStatsInfo(tsArr)
			So(start, ShouldEqual, 1646360670000)
			So(end, ShouldEqual, 1646360670000)
			So(totalNum, ShouldEqual, 1)
		})

		Convey("tsArr contain one element && key is string ", func() {
			tsArr := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":\"1646360670000\",\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}
			start, end, totalNum := getPointsStatsInfo(tsArr)
			So(start, ShouldEqual, 1646360670000)
			So(end, ShouldEqual, 1646360670000)
			So(totalNum, ShouldEqual, 1)
		})

		Convey("tsArr contain one element && key is string 2 ", func() {
			tsArr := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":\"1646360670000a\",\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}
			start, end, totalNum := getPointsStatsInfo(tsArr)
			So(start, ShouldEqual, 0)
			So(end, ShouldEqual, 0)
			So(totalNum, ShouldEqual, 1)
		})

		Convey("tsArr contain one element && donot contain key of 'key' ", func() {
			tsArr := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key1\":\"1646360670000a\",\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}
			start, end, totalNum := getPointsStatsInfo(tsArr)
			So(start, ShouldEqual, 0)
			So(end, ShouldEqual, 0)
			So(totalNum, ShouldEqual, 1)
		})

		Convey("tsArr contain more than one element ", func() {
			tsArr := [][]gjson.Result{
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
						Raw:  "{\"key_as_string\":\"2022-03-04T02:24:30.000Z\",\"key\":1646360680000,\"doc_count\":2,\"value\":{\"value\":8.0,\"timestamp\":1646360698347}}",
					},
				},
			}
			start, end, totalNum := getPointsStatsInfo(tsArr)
			So(start, ShouldEqual, 1646360670000)
			So(end, ShouldEqual, 1646360680000)
			So(totalNum, ShouldEqual, 3)
		})

	})
}

func TestGetProcessParam(t *testing.T) {
	Convey("test processParam ", t, func() {

		Convey("query is instant query && range = 121m ", func() {
			queryInstant1 := &interfaces.Query{
				Start:          1652320554000,
				End:            1652320554000,
				Interval:       1,
				IsInstantQuery: true,
				LogGroupId:     "a",
			}

			queryInstant1.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[121m])`
			expr, _ := parser.ParseExpr(testCtx, queryInstant1.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), queryInstant1)

			So(exprV, ShouldBeNil)
			So(newQuery, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, "if instant query and range > 2h, range should a multiple of 5 minutes")
		})

		Convey("query is instant query && range = 125m ", func() {
			queryInstant1 := &interfaces.Query{
				Start:          1652320554000,
				End:            1652320554000,
				Interval:       1,
				IsInstantQuery: true,
				LogGroupId:     "a",
			}

			queryInstant1.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[125m])`
			expr, _ := parser.ParseExpr(testCtx, queryInstant1.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), queryInstant1)

			So(exprV, ShouldResemble, expr.(*parser.Call).Args[0].(*parser.MatrixSelector).VectorSelector.(*parser.VectorSelector))
			So(newQuery.Start, ShouldEqual, 1652313054000)
			So(queryInstant1.SubIntervalWith30min, ShouldEqual, 1500000)
			So(queryInstant1.Start, ShouldEqual, 1652320554000)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("query is instant query && range = 5m ", func() {
			queryInstant1 := &interfaces.Query{
				Start:          1652320554000,
				End:            1652320554000,
				Interval:       1,
				IsInstantQuery: true,
				LogGroupId:     "a",
			}

			queryInstant1.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, queryInstant1.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), queryInstant1)

			So(exprV, ShouldResemble, expr.(*parser.Call).Args[0].(*parser.MatrixSelector).VectorSelector.(*parser.VectorSelector))
			So(newQuery.Start, ShouldEqual, 1652320254000)
			So(queryInstant1.SubIntervalWith30min, ShouldEqual, 300000)
			So(queryInstant1.Start, ShouldEqual, 1652320554000)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("query is range query && range = 5m ", func() {
			rangeQuery := &interfaces.Query{
				Start:      1646360670,
				End:        1646360700,
				Interval:   30000,
				LogGroupId: "a",
			}
			rangeQuery.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, rangeQuery.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), rangeQuery)

			So(exprV, ShouldResemble, expr.(*parser.Call).Args[0].(*parser.MatrixSelector).VectorSelector.(*parser.VectorSelector))
			So(newQuery.Start, ShouldEqual, 1646360670)
			So(rangeQuery.SubIntervalWith30min, ShouldEqual, 30000)
			So(rangeQuery.Start, ShouldEqual, 1646360670)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("4. type assertion failed", func() {
			expr := &parser.MatrixSelector{VectorSelector: &parser.MatrixSelector{}}

			val, newQuery, status, err := processParam(expr, query)
			So(val, ShouldBeNil)
			So(newQuery, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("query is instant query && range = auto ", func() {
			queryInstant1 := &interfaces.Query{
				Start:          1652320254000,
				End:            1652320554000,
				Interval:       1,
				IsInstantQuery: true,
				LogGroupId:     "a",
			}

			queryInstant1.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[auto])`
			expr, _ := parser.ParseExpr(testCtx, queryInstant1.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), queryInstant1)

			So(exprV, ShouldResemble, expr.(*parser.Call).Args[0].(*parser.MatrixSelector).VectorSelector.(*parser.VectorSelector))
			So(newQuery.Start, ShouldEqual, 1652320254000)
			So(queryInstant1.SubIntervalWith30min, ShouldEqual, 300000)
			So(queryInstant1.Start, ShouldEqual, 1652320254000)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("query is range query && range = auto ", func() {
			rangeQuery := &interfaces.Query{
				Start:      1646360670,
				End:        1646360700,
				Interval:   30000,
				LogGroupId: "a",
			}
			rangeQuery.QueryStr = `rate(elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[auto])`
			expr, _ := parser.ParseExpr(testCtx, rangeQuery.QueryStr)

			exprV, newQuery, status, err := processParam(expr.(*parser.Call).Args[0].(*parser.MatrixSelector), rangeQuery)

			So(exprV, ShouldResemble, expr.(*parser.Call).Args[0].(*parser.MatrixSelector).VectorSelector.(*parser.VectorSelector))
			So(newQuery.Start, ShouldEqual, 1646360670)
			So(rangeQuery.SubIntervalWith30min, ShouldEqual, 30000)
			So(rangeQuery.Start, ShouldEqual, 1646360670)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestCommonProcess(t *testing.T) {
	Convey("test commonProcess", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		Tsids_Of_Model_Metric_Map.Store("0+directTsid_query", TsidData{
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

		var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Convey("MakeDSL error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, "ratem")

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("GetLogGroupQueryFilters error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, samplingAgg)
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

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, samplingAgg)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("getIndicesNumberOfShards indexShardsArr == 0 ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			emptyIndexShardsArr, _ := sonic.Marshal([]*interfaces.IndexShards{})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyIndexShardsArr, http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, samplingAgg)

			So(res, ShouldResemble, MapResult{})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("executeDslAndProcess error ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, samplingAgg)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("executeDslAndProcess success ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), query, samplingAgg)

			mat, ok := res.(MapResult)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, MapResult{LabelsMap: map[string][]*labels.Label{}, TsValueMap: map[string][][]gjson.Result{}})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("executeDslAndProcess success with tsid ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			updateTime, _ := time.Parse(libCommon.RFC3339Milli, "2022-03-04T10:28:30.000+08:00")
			interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = updateTime
			defer delete(interfaces.INDEX_BASE_SPLIT_TIME, "metricbeat-*")

			directTsidQuery := &interfaces.Query{
				Start:      1646360670000,
				End:        1646370700000,
				Interval:   30000,
				LogGroupId: "a",
				Limit:      -1,
			}
			directTsidQuery.QueryStr = `directTsid_query{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, directTsidQuery.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), directTsidQuery, samplingAgg)

			mat, ok := res.(MapResult)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, MapResult{LabelsMap: map[string][]*labels.Label{}, TsValueMap: map[string][][]gjson.Result{}})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("executeDslAndProcess success with tsid batch", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			Tsids_Of_Model_Metric_Map.Store("+f2b0d64d970b80107317e333a6f1e3f6", TsidData{
				RefreshTime:     time.Now(),
				FullRefreshTime: time.Now(),
				StartTime:       time.Unix(0, 0),
				EndTime:         time.Unix(0, 0),
				Tsids:           []string{"id1", "id2", "id3", "id4", "id5", "id5", "id6"},
				TsidsMap: map[string]labels.Labels{
					"id1": {
						{Name: "label1", Value: "value1"},
					},
				},
			})

			updateTime, _ := time.Parse(libCommon.RFC3339Milli, "2022-03-04T10:28:30.000+08:00")
			interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = updateTime
			interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = updateTime
			defer delete(interfaces.INDEX_BASE_SPLIT_TIME, "metricbeat-*")

			directTsidQuery := &interfaces.Query{
				Start:               1646360670000,
				End:                 1646370700000,
				Interval:            3000,
				LogGroupId:          "a",
				Limit:               -1,
				IgnoringMemoryCache: false,
			}
			directTsidQuery.QueryStr = `directTsid_query2{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, directTsidQuery.QueryStr)

			res, status, err := lnMock.commonProcess(testCtx, expr.(*parser.VectorSelector), directTsidQuery, samplingAgg)

			mat, ok := res.(MapResult)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, MapResult{LabelsMap: map[string][]*labels.Label{}, TsValueMap: map[string][][]gjson.Result{}, TotalSeries: 0})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestWrapKeyWordFieldName(t *testing.T) {
	Convey("test wrapKeyWordFieldName ", t, func() {
		Convey("fileds has one element ", func() {

			str := wrapKeyWordFieldName("__labels_str")
			So(str, ShouldEqual, "__labels_str.keyword")
		})

		Convey("fileds has more than one element ", func() {

			str := wrapKeyWordFieldName("labels", "cpu")
			So(str, ShouldEqual, "labels.cpu.keyword")
		})

		Convey("fileds has more than one element 2", func() {

			str := wrapKeyWordFieldName("labels", "")
			So(str, ShouldEqual, "")
		})

	})
}

func TestWrapMetricsFieldName(t *testing.T) {
	Convey("test wrapMetricsFieldName ", t, func() {
		Convey("success ", func() {
			str := wrapMetricsFieldName("cpu_seconds_total")
			So(str, ShouldEqual, "metrics.cpu_seconds_total")

		})

		Convey("empty string ", func() {
			str := wrapMetricsFieldName("")
			So(str, ShouldEqual, "")

		})
	})
}
