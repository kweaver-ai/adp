// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package promql

import (
	"context"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/data_dict"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/leafnodes"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func compareJsonString(actualString string, expectedString string) {
	actualJson := map[string]any{}
	expectedJson := map[string]any{}
	_ = sonic.UnmarshalString(actualString, &actualJson)
	_ = sonic.UnmarshalString(expectedString, &expectedJson)
	So(actualJson, ShouldResemble, expectedJson)
}

var (
	shard0            = "_shards:0"
	shard1            = "_shards:1"
	shard2            = "_shards:2"
	index             = "metricbeat-1"
	indexPattern      = "metricbeat-*"
	indicesError      = "Error _cat/indices response from opensearch"
	emptyResult       = "{\"status\":\"success\",\"data\":{\"resultType\":\"matrix\",\"result\":[]}}"
	emptyVectorResult = "{\"status\":\"success\",\"data\":{\"resultType\":\"vector\",\"result\":[]}}"
	emptySeriesResult = `{"status":"success","data":[]}`
	matResult         = make(static.Matrix, 0)
	emptyDslResult    = map[string]interface{}{
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
	irateDslResult = map[string]interface{}{
		"_shards": map[string]interface{}{
			"failed":     0,
			"skipped":    0,
			"successful": 3,
			"total":      3,
		},
		"hits": []string{},
		"aggregations": map[string]interface{}{
			interfaces.LABELS_STR: map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       "a",
						"doc_count": 4,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-24T05:10:00.000Z",
									"key":           1653369000000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"previousValue":     0,
										"previousTimestamp": 0,
										"lastValue":         171456.0,
										"lastTimestamp":     1653369141709,
									},
								},
								{
									"key_as_string": "2022-05-24T05:15:00.000Z",
									"key":           1653369300000,
									"doc_count":     3,
									"value": map[string]interface{}{
										"previousValue":     171856.0,
										"previousTimestamp": 1653369321709,
										"lastValue":         171722.0,
										"lastTimestamp":     1653369328111,
									},
								},
							},
						},
					},
					{
						"key":       "b",
						"doc_count": 3,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-24T05:10:00.000Z",
									"key":           1653369000000,
									"doc_count":     2,
									"value": map[string]interface{}{
										"previousValue":     171536.0,
										"previousTimestamp": 1653369020709,
										"lastValue":         171011.0,
										"lastTimestamp":     1653369141709,
									},
								},
								{
									"key_as_string": "2022-05-24T05:15:00.000Z",
									"key":           1653369300000,
									"doc_count":     1,
									"value": map[string]interface{}{
										"previousValue":     0,
										"previousTimestamp": 0,
										"lastValue":         171722.0,
										"lastTimestamp":     1653369328111,
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
		Start:      1646360670,
		End:        1646360700,
		Interval:   30000,
		LogGroupId: "a",
	}
	irateQuery = &interfaces.Query{
		Start:      1653368870,
		End:        1653549380,
		Interval:   300000,
		LogGroupId: "a",
		Limit:      -1,
	}
	queryInstant = &interfaces.Query{
		Start:          1652320554000,
		End:            1652320554000,
		FixedStart:     1652320554000,
		FixedEnd:       1652320554000,
		Interval:       1,
		IsInstantQuery: true,
		LogGroupId:     "a",
		Limit:          -1,
	}

	emptyJson, _ = sonic.Marshal([]*interfaces.IndexShards{})

	queryWithFixedTime = &interfaces.Query{
		Start:       1652320539000,
		End:         1652320554000,
		FixedStart:  1652320539000,
		FixedEnd:    1652320554000,
		Interval:    3000,
		IntervalStr: "3000ms",
		LogGroupId:  "a",
		Limit:       -1,
	}

	labelsStrNas   = "cluster=\"opensearch\",instance=\"instance_abc\",index=\"nas_statis-2022.05-0\",job=\"prometheus\""
	labelsStrNode  = "cluster=\"opensearch\",instance=\"instance_abc\",index=\"node_statis-2022.05-0\",job=\"prometheus\""
	labelsStrNode2 = "cluster=\"opensearch\",instance=\"instance_abc\",index=\"noee_statis-2022.05-0\",job=\"prometheus\""
	dslResult0     = map[string]interface{}{
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

	dslResultByUnusable = map[string]interface{}{
		"aggregations": map[string]interface{}{
			interfaces.TSID: map[string]interface{}{
				"buckets": []map[string]interface{}{
					{
						"key": labelsStrNas,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-12T09:50:00.000Z",
									"key":           1652320200000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320201541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:51:00.000Z",
									"key":           1652320260000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320261541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:52:00.000Z",
									"key":           1652320320000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320321541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:53:00.000Z",
									"key":           1652320380000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320381541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:54:00.000Z",
									"key":           1652320440000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320441541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:55:00.000Z",
									"key":           1652320500000,
									"value": map[string]interface{}{
										"value":     20.0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:56:00.000Z",
									"key":           1652320560000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320564541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:58:00.000Z",
									"key":           1652320680000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320684641,
									},
								},
								{
									"key_as_string": "2022-05-12T09:59:00.000Z",
									"key":           1652320740000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320744941,
									},
								},
								{
									"key_as_string": "2022-05-12T10:00:00.000Z",
									"key":           1652320800000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320805241,
									},
								},
								{
									"key_as_string": "2022-05-12T10:01:00.000Z",
									"key":           1652320860000,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320865000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:02:00.000Z",
									"key":           1652320920000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320925000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:10:00.000Z",
									"key":           1652321400000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652321405000,
									},
								},
							},
						},
					},
					{
						"key": labelsStrNode,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-12T09:50:00.000Z",
									"key":           1652320200000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320201541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:51:00.000Z",
									"key":           1652320260000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320261541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:52:00.000Z",
									"key":           1652320320000,
									"value": map[string]interface{}{
										"value":     1,
										"timestamp": 1652320321541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:53:00.000Z",
									"key":           1652320380000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320381541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:54:00.000Z",
									"key":           1652320440000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320441541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:55:00.000Z",
									"key":           1652320500000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320541541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:56:00.000Z",
									"key":           1652320560000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320564541,
									},
								},
								{
									"key_as_string": "2022-05-12T09:58:00.000Z",
									"key":           1652320680000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320684641,
									},
								},
								{
									"key_as_string": "2022-05-12T09:59:00.000Z",
									"key":           1652320740000,
									"value": map[string]interface{}{
										"value":     1,
										"timestamp": 1652320744941,
									},
								},
								{
									"key_as_string": "2022-05-12T10:00:00.000Z",
									"key":           1652320800000,
									"value": map[string]interface{}{
										"value":     1,
										"timestamp": 1652320805241,
									},
								},
								{
									"key_as_string": "2022-05-12T10:01:00.000Z",
									"key":           1652320860000,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320865000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:02:00.000Z",
									"key":           1652320920000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320925000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:10:00.000Z",
									"key":           1652321400000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652321405000,
									},
								},
							},
						},
					},
					{
						"key": labelsStrNode2,
						"time": map[string]interface{}{
							"buckets": []map[string]interface{}{
								{
									"key_as_string": "2022-05-12T10:01:00.000Z",
									"key":           1652320860000,
									"value": map[string]interface{}{
										"value":     1.0,
										"timestamp": 1652320865000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:02:00.000Z",
									"key":           1652320920000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652320925000,
									},
								},
								{
									"key_as_string": "2022-05-12T10:10:00.000Z",
									"key":           1652321400000,
									"value": map[string]interface{}{
										"value":     0,
										"timestamp": 1652321405000,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	nasLabels = []*labels.Label{
		{
			Name: "cluster", Value: "opensearch",
		},
		{
			Name: "index", Value: "nas_statis-2022.05-0",
		},
		{
			Name: "instance", Value: "instance_abc",
		},
		{
			Name: "job", Value: "prometheus",
		},
	}

	nodeLabels = []*labels.Label{
		{
			Name: "cluster", Value: "opensearch",
		},
		{
			Name: "index", Value: "node_statis-2022.05-0",
		},
		{
			Name: "instance", Value: "instance_abc",
		},
		{
			Name: "job", Value: "prometheus",
		},
	}

	noeeLabels = []*labels.Label{
		{
			Name: "cluster", Value: "opensearch",
		},
		{
			Name: "index", Value: "noee_statis-2022.05-0",
		},
		{
			Name: "instance", Value: "instance_abc",
		},
		{
			Name: "job", Value: "prometheus",
		},
	}

	timeString39 = "2022-05-12T01:55:39.000Z"
	timeString42 = "2022-05-12T01:55:42.000Z"

	indexMetricbeat111    = "metricbeat11-1"
	indexbaseMetricbeat11 = "metricbeat11-*"
	indexMetricbeat122    = "metricbeat12-2"
	indexbaseMetricbeat12 = "metricbeat12-*"

	seriesMatcher = interfaces.Matchers{
		Start:      1655346155000,
		End:        1655350445000,
		LogGroupId: "a",
	}

	NodeCpuGuestSeriesDslResult = map[string]interface{}{
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
						"key":       "cpu=0,instance=instance_abc,job=prometheus,mode=nice",
						"doc_count": 211,
					},
					{
						"key":       "cpu=1,instance=instance_abc,job=prometheus,mode=nice",
						"doc_count": 211,
					},
				},
			},
		},
		"timed_out": false,
		"took":      0,
	}

	NodeFilesystemSeriesDslResult = map[string]interface{}{
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
						"key":       "device=/dev/sda1,fstype=xfs,instance=instance_abc,job=prometheus,mountpoint=/boot",
						"doc_count": 211,
					},
					{
						"key":       "device=/dev/mapper/centos-root,fstype=xfs,instance=instance_abc,job=prometheus,mountpoint=/",
						"doc_count": 211,
					},
				},
			},
		},
		"timed_out": false,
		"took":      0,
	}

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

func mockNewPromqlService(osAccess interfaces.OpenSearchAccess, lgAccess interfaces.LogGroupAccess,
	dvService interfaces.DataViewService, mmService interfaces.MetricModelService) *promQLService {
	psMock := &promQLService{
		leafNodes: leafnodes.NewLeafNodes(&common.AppSetting{}, osAccess, lgAccess, dvService),
		mmService: mmService,
		appSetting: &common.AppSetting{
			ServerSetting: common.ServerSetting{
				FullCacheRefreshInterval: 24 * time.Hour,
			},
		},
	}

	util.InitAntsPool(common.PoolSetting{
		MegerPoolSize:       10,
		ExecutePoolSize:     10,
		BatchSubmitPoolSize: 10,
	})
	return psMock
}

func TestExec(t *testing.T) {
	Convey("test promql_service Exec with query_range", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+a", leafnodes.TsidData{
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
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Convey("expression parse error ", func() {
			query.QueryStr = `elasticsearch_os_cpu_percent{{cluster="txy",name!~"^node.*"}[5m]`

			_, res, status, err := psMock.Exec(testCtx, *query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("expression type error ", func() {
			query.QueryStr = `"a"`

			_, res, status, err := psMock.Exec(testCtx, *query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err.Error(), ShouldEqual, `invalid expression type "string" for range query, must be Scalar,range or instant Vector`)
		})

		Convey("invalid expression type for range query ", func() {
			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}[5m]`

			_, res, status, err := psMock.Exec(testCtx, *query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("success Exec with IndexShards is empty", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyJson, http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *query)

			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), emptyResult)
			So(err, ShouldBeNil)
		})

		Convey("success Exec with IndexShards is not empty ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *query)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), emptyResult)
			So(err, ShouldBeNil)
		})

		Convey("Exec failed with GetIndicesNumberOfShards error for range query ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":500,"error":{"type":"UniQuery.InternalServerError","reason":"Error _cat/indices response from opensearch"}}`)
		})

		Convey("Exec failed with GetIndicesNumberOfShards error for instant query ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			queryInstant.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *queryInstant)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":500,"error":{"type":"UniQuery.InternalServerError","reason":"Error _cat/indices response from opensearch"}}`)
		})

		Convey("success Exec with metric model", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indexShardsArr, nil, 200, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query1 := interfaces.Query{
				Start:         1646360670,
				End:           1646360700,
				Interval:      30000,
				LogGroupId:    "a",
				QueryStr:      `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`,
				IsMetricModel: true,
			}
			res, resBytes, status, err := psMock.Exec(testCtx, query1)

			So(res, ShouldResemble, interfaces.PromQLResponse{
				Status: "success",
				Data:   QueryData{ResultType: parser.ValueType("matrix"), Result: static.PageMatrix{Matrix: static.Matrix{}}}})
			So(resBytes, ShouldResemble, []byte(nil))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})

	Convey("test promql_service Exec with query ", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)
		Convey("expression parse error with query ", func() {
			queryInstant.QueryStr = `elasticsearch_os_cpu_percent{{cluster="txy",name!~"^node.*"}[5m]`

			_, res, status, err := psMock.Exec(testCtx, *queryInstant)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("success Exec with IndexShards is empty", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyJson, http.StatusOK, nil)

			queryInstant.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *queryInstant)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), emptyVectorResult)
			So(err, ShouldBeNil)
		})

		Convey("success Exec with IndexShards is not empty ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			queryInstant.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *queryInstant)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), emptyVectorResult)
			So(err, ShouldBeNil)
		})

		Convey("success Exec with IndexShards is not empty && dsl result is not empty ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			queryInstant.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			_, res, status, err := psMock.Exec(testCtx, *queryInstant)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), "{\"status\":\"success\",\"data\":{\"resultType\":\"vector\",\"result\":[{\"metric\":{\"cluster\":\"opensearch\",\"index\":\"nas_statis-2022.05-0\",\"instance\":\"instance_abc\",\"job\":\"prometheus\"},\"value\":[1652320554,\"1\"]},{\"metric\":{\"cluster\":\"opensearch\",\"index\":\"node_statis-2022.05-0\",\"instance\":\"instance_abc\",\"job\":\"prometheus\"},\"value\":[1652320554,\"1\"]}]}}")
			So(err, ShouldBeNil)
		})
	})
}

func TestEval(t *testing.T) {
	Convey("test promql_service eval", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+a", leafnodes.TsidData{
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
		var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		Convey("VectorSelector expression ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `elasticsearch_os_cpu_percent{cluster="txy",name!~"^node.*"}`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, query)
			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matResult})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("scalar expression with 1 ", func() {
			queryWithFixedTime.QueryStr = `1`

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label(nil),
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
			}

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.eval(testCtx, expr, queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("unary vector expression -vector", func() {
			dslResultUnary := map[string]interface{}{
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
								"key":       "cluster=\"opensearch\",cluster=\"opensearch\",instance=\"instance_abc\",index=\"nas_statis-2022.05-0\",job=\"prometheus\"",
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
								"key":       "cluster=\"opensearch\",cluster=\"opensearch\",instance=\"instance_abc\",index=\"nas_statis-2022.05-0\",job=\"prometheus\"",
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
						},
					},
				},
			}

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "cluster", Value: "opensearch",
						},
						{
							Name: "cluster", Value: "opensearch",
						},
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
						{
							Name: "instance", Value: "instance_abc",
						},
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: -20,
						},
						{
							T: 1652320542000,
							V: -10,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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
				shard0).Times(1).Return(convert.MapToByte(dslResultUnary), http.StatusOK, nil)

			queryWithFixedTime.QueryStr = `-opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, queryWithFixedTime)
			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: expectMat, TotalSeries: 1})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar expression (vector+1) * 3 with mock emptyDslResult ", func() {

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			queryWithFixedTime.QueryStr = `(prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.index=~"a.*|node.*"}+1)*3`
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, matResult)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op on vector expression vector / vector with mock emptyDslResult ", func() {

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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
				shard0).Times(2).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			queryWithFixedTime.QueryStr = `prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.index=~"a.*|node.*"} / on(prometheus.labels.index) ` +
				`prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, matResult)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("AggregateExpr expression sum by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 60,
						},
						{
							T: 1652320542000,
							V: 30,
						},
						{
							T: 1652320545000,
							V: 62,
						},
						{
							T: 1652320548000,
							V: 32,
						},
						{
							T: 1652320551000,
							V: 64,
						},
						{
							T: 1652320554000,
							V: 34,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			queryWithFixedTime.QueryStr = `sum(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)

			res, status, err := psMock.eval(testCtx, expr, queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("IrateMetrixSelector expression", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, query)
			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matResult})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("IrateMetrixSelector expression not empty result", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(irateDslResult), http.StatusOK, nil)

			irateQuery.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, irateQuery.QueryStr)
			irateQuery.FixedStart = int64(math.Floor(float64(irateQuery.Start*1000)/float64(irateQuery.Interval))) * irateQuery.Interval
			irateQuery.FixedEnd = int64(math.Floor(float64(irateQuery.End*1000)/float64(irateQuery.Interval))) * irateQuery.Interval
			res, status, err := psMock.eval(testCtx, expr, irateQuery)
			mat, ok := res.(static.PageMatrix)
			matRes := static.Matrix{
				static.Series{
					Metric: labels.Labels{&labels.Label{
						Name:  interfaces.LABELS_STR,
						Value: "a",
					}},
					Points: []static.Point{{T: 1653369000000, V: 26823.180256169948},
						{T: 1653369300000, V: 26823.180256169948}},
				},
				static.Series{
					Metric: labels.Labels{&labels.Label{
						Name:  interfaces.LABELS_STR,
						Value: "b",
					}},
					Points: []static.Point{{T: 1653368700000, V: 1413.314049586777}, {T: 1653369000000, V: 3.8143367560433905}},
				},
			}

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matRes, TotalSeries: 2})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("IrateMetrixSelector step can't be divisible by range", func() {
			irateSubstepDslResult := map[string]interface{}{
				"_shards": map[string]interface{}{
					"failed":     0,
					"skipped":    0,
					"successful": 3,
					"total":      3,
				},
				"hits": []string{},
				"aggregations": map[string]interface{}{
					interfaces.LABELS_STR: map[string]interface{}{
						"doc_count_error_upper_bound": 0,
						"sum_other_doc_count":         0,
						"buckets": []map[string]interface{}{
							{
								"key":       "a",
								"doc_count": 4,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": "2022-05-24T05:10:00.000Z",
											"key":           1653369000000,
											"doc_count":     1,
											"value": map[string]interface{}{
												"previousValue":     0,
												"previousTimestamp": 0,
												"lastValue":         171456.0,
												"lastTimestamp":     1653369000709,
											},
										},
										{
											"key_as_string": "2022-05-24T05:15:00.000Z",
											"key":           1653369001000,
											"doc_count":     3,
											"value": map[string]interface{}{
												"previousValue":     171856.0,
												"previousTimestamp": 1653369001709,
												"lastValue":         171722.0,
												"lastTimestamp":     1653369001811,
											},
										},
									},
								},
							},
							{
								"key":       "b",
								"doc_count": 3,
								"time": map[string]interface{}{
									"buckets": []map[string]interface{}{
										{
											"key_as_string": "2022-05-24T05:10:00.000Z",
											"key":           1653369000000,
											"doc_count":     2,
											"value": map[string]interface{}{
												"previousValue":     171536.0,
												"previousTimestamp": 1653369000709,
												"lastValue":         171011.0,
												"lastTimestamp":     1653369000809,
											},
										},
										{
											"key_as_string": "2022-05-24T05:15:00.000Z",
											"key":           1653369001000,
											"doc_count":     1,
											"value": map[string]interface{}{
												"previousValue":     0,
												"previousTimestamp": 0,
												"lastValue":         171722.0,
												"lastTimestamp":     1653369001811,
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
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(irateSubstepDslResult), http.StatusOK, nil)

			irateQuery.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[6m])`
			expr, _ := parser.ParseExpr(testCtx, irateQuery.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, irateQuery)
			mat, ok := res.(static.PageMatrix)
			matRes := static.Matrix{
				static.Series{
					Metric: labels.Labels{&labels.Label{
						Name:  interfaces.LABELS_STR,
						Value: "b",
					}},
					Points: []static.Point{{T: 1653368700000, V: 1710110}, {T: 1653369000000, V: 1710110}},
				},
			}
			So(mat, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: matRes, TotalSeries: 2})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		// eval rate UT, here I mock leafnodes.RateAggs, UT of RateAggs is in rate_test.go
		Convey("funcCall rate", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "rate",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name: "prometheus.metrics.node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{
							{
								Type:  labels.MatchEqual,
								Name:  "ar_dataview",
								Value: "metricbeat_node_exporter",
							},
						},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "rate(prometheus.metrics.node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}
			patches := ApplyMethodReturn(ln, "RateAggs", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

		})

		Convey("funcCall is valid", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `abc(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, " FunctionCalls is not defined, please input valid function. ")
		})

		// eval changes UT, here I mock leafnodes.ChangesAggs, UT of ChangesAggs is in changes_test.go
		Convey("funcCall changes", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "changes",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name: "prometheus.metrics.node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{
							{
								Type:  labels.MatchEqual,
								Name:  "ar_dataview",
								Value: "metricbeat_node_exporter",
							},
						},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "changes(prometheus.metrics.node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "ChangesAggs", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

		})

		// eval increase UT, here I mock leafnodes.RateAggs, UT of RateAggs is in rate_test.go
		Convey("funcCall increase", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "increase",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name: "prometheus.metrics.node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{
							{
								Type:  labels.MatchEqual,
								Name:  "ar_indexbase",
								Value: "metricbeat_node_exporter",
							},
						},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "increase(prometheus.metrics.node_cpu_seconds_total{ar_indexbase=\"metricbeat_node_exporter\"}[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "RateAggs", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

		})

		// eval delta UT, here I mock leafnodes.DeltaAggs, UT of DeltaAggs is in delta_test.go
		Convey("funcCall delta", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "delta",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name: "prometheus.metrics.node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{
							{
								Type:  labels.MatchEqual,
								Name:  "ar_dataview",
								Value: "metricbeat_node_exporter",
							},
						},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "delta(prometheus.metrics.node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "DeltaAggs", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)

		})

	})
}

type testResponse struct {
	res    string
	status int
	err    error
}

const queryString = "time()"

func TestExcTime(t *testing.T) {
	loc, _ := time.LoadLocation(os.Getenv("TZ"))
	common.APP_LOCATION = loc

	psMock := &promQLService{}
	tests := []struct {
		query  interfaces.Query
		expect testResponse
	}{
		// Start == End
		{
			query: interfaces.Query{
				QueryStr:       queryString,
				Start:          1646360670000,
				End:            1646360670000,
				Interval:       1,
				IntervalStr:    "1ms",
				IsInstantQuery: true,
			},
			expect: testResponse{
				res:    `{"status":"success","data":{"resultType":"scalar","result":[1646360670,"1646360670"]}}`,
				status: http.StatusOK,
				err:    nil,
			},
		},
		// Start != End
		{
			query: interfaces.Query{
				QueryStr:    queryString,
				Start:       1646360670000,
				End:         1646360700000,
				Interval:    30000,
				IntervalStr: "30000ms",
			},
			expect: testResponse{
				res:    `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{},"values":[[1646360670,"1646360670"],[1646360700,"1646360700"]]}]}}`,
				status: http.StatusOK,
				err:    nil,
			},
		},
	}

	for _, tt := range tests {
		_, res, status, err := psMock.Exec(testCtx, tt.query)
		if string(res) != tt.expect.res || status != tt.expect.status || err != tt.expect.err {
			t.Errorf("want: res:%s,status:%d,err:%v, \n but actual got:res:%s,status:%d,err:%v\n", tt.expect.res, tt.expect.status, tt.expect.err, res, status, err)
		}
	}

}

// func BenchmarkTime(b *testing.B) {
// 	psMock := &promQLService{}
// 	query := interfaces.Query{
// 		QueryStr: queryString,
// 		Start:    1646360670000,
// 		End:      1646360670000,
// 		Interval: 1,
// 	}

// 	want := testResponse{
// 		res:    `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{},"values":[[1646331870,"1646331870"]]}]}}`,
// 		status: 200,
// 		err:    nil,
// 	}

// 	for i := 0; i < b.N; i++ {
// 		_, res, status, err := psMock.Exec(testCtx, query)
// 		if string(res) != want.res || status != want.status || err != want.err {
// 			b.Errorf("want: res:%s,status:%d,err:%v, \n but actual get:res:%s,status:%d,err:%v\n", want.res, want.status, want.err, res, status, err)
// 		}
// 	}
// }

func TestEvalAggregateExpr(t *testing.T) {
	Convey("test promql_service evalAggregateExpr ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-1"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-1"] = time.Now()

		aggQueryWithFixedTime := &interfaces.Query{
			Start:           1652320539000,
			End:             1652320554000,
			FixedStart:      1652320539000,
			FixedEnd:        1652320554000,
			Interval:        3000,
			LogGroupId:      "a",
			IfNeedAllSeries: true,
		}

		aggQueryInstant := &interfaces.Query{
			Start:           1652320554000,
			End:             1652320554000,
			FixedStart:      1652320554000,
			FixedEnd:        1652320554000,
			Interval:        1,
			IsInstantQuery:  true,
			LogGroupId:      "a",
			IfNeedAllSeries: true,
		}

		Convey("evalAggregateExpr sum by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 60,
						},
						{
							T: 1652320542000,
							V: 30,
						},
						{
							T: 1652320545000,
							V: 62,
						},
						{
							T: 1652320548000,
							V: 32,
						},
						{
							T: 1652320551000,
							V: 64,
						},
						{
							T: 1652320554000,
							V: 34,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `sum(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			// res, status, err := psMock.evalAggregateExpr(testCtx,expr.(*parser.AggregateExpr), queryWithFixedTime)
			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr sum without", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "node_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `sum(opensearch_indices_shards_docs)without(job,cluster,instance)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr sum no by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 60,
						},
						{
							T: 1652320542000,
							V: 30,
						},
						{
							T: 1652320545000,
							V: 62,
						},
						{
							T: 1652320548000,
							V: 32,
						},
						{
							T: 1652320551000,
							V: 64,
						},
						{
							T: 1652320554000,
							V: 34,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `sum(opensearch_indices_shards_docs)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr count by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 2,
						},
						{
							T: 1652320542000,
							V: 2,
						},
						{
							T: 1652320545000,
							V: 2,
						},
						{
							T: 1652320548000,
							V: 2,
						},
						{
							T: 1652320551000,
							V: 2,
						},
						{
							T: 1652320554000,
							V: 2,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `count(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr count without", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "node_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `count(opensearch_indices_shards_docs)without(job,cluster,instance)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr count no by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 2,
						},
						{
							T: 1652320542000,
							V: 2,
						},
						{
							T: 1652320545000,
							V: 2,
						},
						{
							T: 1652320548000,
							V: 2,
						},
						{
							T: 1652320551000,
							V: 2,
						},
						{
							T: 1652320554000,
							V: 2,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `count(opensearch_indices_shards_docs)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr avg by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 30,
						},
						{
							T: 1652320542000,
							V: 15,
						},
						{
							T: 1652320545000,
							V: 31,
						},
						{
							T: 1652320548000,
							V: 16,
						},
						{
							T: 1652320551000,
							V: 32,
						},
						{
							T: 1652320554000,
							V: 17,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `avg(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr avg without", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "node_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `avg(opensearch_indices_shards_docs)without(job,cluster,instance)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr avg no by", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 30,
						},
						{
							T: 1652320542000,
							V: 15,
						},
						{
							T: 1652320545000,
							V: 31,
						},
						{
							T: 1652320548000,
							V: 16,
						},
						{
							T: 1652320551000,
							V: 32,
						},
						{
							T: 1652320554000,
							V: 17,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `avg(opensearch_indices_shards_docs)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryWithFixedTime.Start, aggQueryWithFixedTime.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("max aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `max(opensearch_indices_shards_docs{index=~"a.*|node.*"})`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("max by aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{{Name: "cluster", Value: "opensearch"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `max(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(cluster)`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("max without aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: []*labels.Label{{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `max(opensearch_indices_shards_docs{index=~"a.*|node.*"})without(cluster,instance,job)`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr topk by", func() {
			expectSeries := []static.Series{
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `topk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr topk without", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `topk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})without(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr topk no by", func() {
			expectSeries := []static.Series{
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `topk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr bottomk by", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `bottomk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr bottomk without", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `bottomk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})without(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr bottomk no by", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `bottomk(1,opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)
			mat, ok := res.(static.Matrix)
			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("evalAggregateExpr topk 9223372036854774785000", func() {
			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryInstant.QueryStr = `topk(9223372036854774785000,opensearch_indices_shards_docs{index=~"a.*|node.*"})by(job)`

			expr, _ := parser.ParseExpr(testCtx, aggQueryInstant.QueryStr)
			// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
			expr, _ = static.PreprocessExpr(expr, aggQueryInstant.Start, aggQueryInstant.End)

			res, status, err := psMock.evalAggregateExpr(testCtx, expr.(*parser.AggregateExpr), aggQueryInstant)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err.Error(), ShouldEqual, "Scalar value 9.223372036854775e+21 overflows int64")
		})

		Convey("min aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `min(opensearch_indices_shards_docs{index=~"a.*|node.*"})`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("min by aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{{Name: "cluster", Value: "opensearch"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `min(opensearch_indices_shards_docs{index=~"a.*|node.*"})by(cluster)`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("min without aggregation expression", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{{Name: "index", Value: "nas_statis-2022.05-0"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 10,
						},
						{
							T: 1652320545000,
							V: 21,
						},
						{
							T: 1652320548000,
							V: 11,
						},
						{
							T: 1652320551000,
							V: 22,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
				{
					Metric: []*labels.Label{{Name: "index", Value: "node_statis-2022.05-0"}},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 40,
						},
						{
							T: 1652320542000,
							V: 20,
						},
						{
							T: 1652320545000,
							V: 41,
						},
						{
							T: 1652320548000,
							V: 21,
						},
						{
							T: 1652320551000,
							V: 42,
						},
						{
							T: 1652320554000,
							V: 22,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			aggQueryWithFixedTime.QueryStr = `min(opensearch_indices_shards_docs{index=~"a.*|node.*"})without(cluster,instance,job)`
			expr, _ := parser.ParseExpr(testCtx, aggQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, aggQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestEvalBinaryExpr(t *testing.T) {
	Convey("test promql_service evalBinaryExpr ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)
		leafnodes.Tsids_Of_Model_Metric_Map.Store("+edc713d981893fadc91d4fc8271b94c6", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.UnixMilli(1652320539000),
			EndTime:         time.UnixMilli(1652320554000),
			Tsids:           []string{"id1"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
				},
			},
		})
		leafnodes.Tsids_Of_Model_Metric_Map.Store("+5365585e711829eb616096928c183b60", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.UnixMilli(1652320539000),
			EndTime:         time.UnixMilli(1652320554000),
			Tsids:           []string{"id1"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
				},
			},
		})
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-1"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat11-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-1"] = time.Now()

		binQueryWithFixedTime := &interfaces.Query{
			Start:           1652320539000,
			End:             1652320554000,
			FixedStart:      1652320539000,
			FixedEnd:        1652320554000,
			Interval:        3000,
			LogGroupId:      "a",
			IfNeedAllSeries: true,
		}

		Convey("scalar op vector expression 100 % vector failed when dataviewid is empty", func() {
			queryTmp := &interfaces.Query{
				Start:      1652320539000,
				End:        1652320554000,
				FixedStart: 1652320539000,
				FixedEnd:   1652320554000,
				Interval:   3000,
			}
			queryTmp.QueryStr = `100%opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, queryTmp.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), queryTmp)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `missing ar_dataview parameter or ar_dataview parameter value cannot be empty`)
		})

		Convey("scalar op vector expression 100 % vector", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 0,
						},
						{
							T: 1652320542000,
							V: 0,
						},
						{
							T: 1652320545000,
							V: 16,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 12,
						},
						{
							T: 1652320554000,
							V: 4,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 20,
						},
						{
							T: 1652320542000,
							V: 0,
						},
						{
							T: 1652320545000,
							V: 18,
						},
						{
							T: 1652320548000,
							V: 16,
						},
						{
							T: 1652320551000,
							V: 16,
						},
						{
							T: 1652320554000,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `100%opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar expression vector * 3 ", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 60,
						},
						{
							T: 1652320542000,
							V: 30,
						},
						{
							T: 1652320545000,
							V: 63,
						},
						{
							T: 1652320548000,
							V: 33,
						},
						{
							T: 1652320551000,
							V: 66,
						},
						{
							T: 1652320554000,
							V: 36,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 120,
						},
						{
							T: 1652320542000,
							V: 60,
						},
						{
							T: 1652320545000,
							V: 123,
						},
						{
							T: 1652320548000,
							V: 63,
						},
						{
							T: 1652320551000,
							V: 126,
						},
						{
							T: 1652320554000,
							V: 66,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs{index=~"a.*|node.*"}*3`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op vector expression vector / vector ", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs{index=~"a.*|node.*"} / ` +
				`opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op scalar expression (vector+1) * 3  ", func() {
			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 63,
						},
						{
							T: 1652320542000,
							V: 33,
						},
						{
							T: 1652320545000,
							V: 66,
						},
						{
							T: 1652320548000,
							V: 36,
						},
						{
							T: 1652320551000,
							V: 69,
						},
						{
							T: 1652320554000,
							V: 39,
						},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 123,
						},
						{
							T: 1652320542000,
							V: 63,
						},
						{
							T: 1652320545000,
							V: 126,
						},
						{
							T: 1652320548000,
							V: 66,
						},
						{
							T: 1652320551000,
							V: 129,
						},
						{
							T: 1652320554000,
							V: 69,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `(opensearch_indices_shards_docs{index=~"a.*|node.*"}+1)*3`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op on vector expression vector / vector ", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "node_statis-2022.05-0",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 1,
						},
						{
							T: 1652320542000,
							V: 1,
						},
						{
							T: 1652320545000,
							V: 1,
						},
						{
							T: 1652320548000,
							V: 1,
						},
						{
							T: 1652320551000,
							V: 1,
						},
						{
							T: 1652320554000,
							V: 1,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs{index=~"a.*|node.*"} / on(index) ` +
				`opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("vector op vector when found duplicate series for the match group, error ", func() {
			var indexShardsArrMetric2 = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs{index=~"a.*|node.*"} / ignoring(index) ` +
				`opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err.Error(), ShouldEqual, "found duplicate series for the match group, many-to-many only allowed for set operators")

		})

		Convey("vector op ignoring vector expression vector - vector ", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "nas_statis-2022.05-0",
						},
						{
							Name: "instance", Value: "instance_abc",
						},
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 0,
						},
						{
							T: 1652320542000,
							V: 0,
						},
						{
							T: 1652320545000,
							V: 0,
						},
						{
							T: 1652320548000,
							V: 0,
						},
						{
							T: 1652320551000,
							V: 0,
						},
						{
							T: 1652320554000,
							V: 0,
						},
					},
				},
				{
					Metric: []*labels.Label{
						{
							Name: "index", Value: "node_statis-2022.05-0",
						},
						{
							Name: "instance", Value: "instance_abc",
						},
						{
							Name: "job", Value: "prometheus",
						},
					},
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 0,
						},
						{
							T: 1652320542000,
							V: 0,
						},
						{
							T: 1652320545000,
							V: 0,
						},
						{
							T: 1652320548000,
							V: 0,
						},
						{
							T: 1652320551000,
							V: 0,
						},
						{
							T: 1652320554000,
							V: 0,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArrMetric2 = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, &interfaces.IndexShards{
				IndexName: indexMetricbeat111,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexbaseMetricbeat11)
			notEmptyJson2, _ := sonic.Marshal(indexShardsArrMetric2)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard1).Times(2).Return(convert.MapToByte(dslResult1), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat122}, gomock.Any(),
				shard2).Times(2).Return(convert.MapToByte(dslResult2), http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), []string{indexMetricbeat111}, gomock.Any(),
				shard0).Times(2).Return(convert.MapToByte(dslResultMetric2), http.StatusOK, nil)

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs{index=~"a.*|node.*"} - ignoring(cluster) ` +
				`opensearch_indices_shards_docs{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
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
			var indexShardsArrMetric2 = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, &interfaces.IndexShards{
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
				gomock.Any()).Times(1).Return(convert.MapToByte(dslResult0), http.StatusOK, nil)

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docsA - on(index) opensearch_indices_shards_docsB`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.evalBinaryExpr(testCtx, expr.(*parser.BinaryExpr), binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err.Error(), ShouldEqual, "multiple matches for labels: many-to-one matching must be explicit (group_left/group_right)")
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

			var indexShardsArrMetric2 = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArrMetric2 = append(indexShardsArrMetric2, &interfaces.IndexShards{
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

			binQueryWithFixedTime.QueryStr = `opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"} - on(index) ` +
				`opensearch_indices_shards_docs` +
				`{index=~"a.*|node.*"}`
			expr, _ := parser.ParseExpr(testCtx, binQueryWithFixedTime.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, binQueryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err.Error(), ShouldEqual, "found duplicate series for the match group, many-to-many only allowed for set operators")
		})

	})
}

func TestEvalStepInvariantExpr(t *testing.T) {
	Convey("test promql_service evalStepInvariantExpr ", t, func() {

		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		Convey("scalar op scalar expression with unsupport operator, 1 >bool 2 ", func() {
			queryWithFixedTime.QueryStr = `1 >bool 2`

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.evalStepInvariantExpr(testCtx, expr.(*parser.StepInvariantExpr), queryWithFixedTime)

			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
		})

		Convey("scalar op scalar expression with 1+2 ", func() {
			queryWithFixedTime.QueryStr = `1+2`

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label(nil),
					Points: []static.Point{
						{
							T: 1652320539000,
							V: 3,
						},
						{
							T: 1652320542000,
							V: 3,
						},
						{
							T: 1652320545000,
							V: 3,
						},
						{
							T: 1652320548000,
							V: 3,
						},
						{
							T: 1652320551000,
							V: 3,
						},
						{
							T: 1652320554000,
							V: 3,
						},
					},
				},
			}

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.evalStepInvariantExpr(testCtx, expr.(*parser.StepInvariantExpr), queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("scalar op scalar expression -3^1000", func() {
			queryWithFixedTime.QueryStr = `-3^1000`

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label(nil),
					Points: []static.Point{
						{
							T: 1652320539000,
							V: math.Inf(-1),
						},
						{
							T: 1652320542000,
							V: math.Inf(-1),
						},
						{
							T: 1652320545000,
							V: math.Inf(-1),
						},
						{
							T: 1652320548000,
							V: math.Inf(-1),
						},
						{
							T: 1652320551000,
							V: math.Inf(-1),
						},
						{
							T: 1652320554000,
							V: math.Inf(-1),
						},
					},
				},
			}

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.evalStepInvariantExpr(testCtx, expr.(*parser.StepInvariantExpr), queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("foo @ 20 * bar @ 10 when dataview is empty ", func() {
			queryTmp := &interfaces.Query{
				Start:      1652320539000,
				End:        1652320554000,
				FixedStart: 1652320539000,
				FixedEnd:   1652320554000,
				Interval:   3000,
			}

			queryTmp.QueryStr = `foo @ 20 * bar @ 10`
			expr, _ := parser.ParseExpr(testCtx, queryTmp.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryTmp.Start, queryTmp.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.evalStepInvariantExpr(testCtx, expr.(*parser.StepInvariantExpr), queryTmp)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `missing ar_dataview parameter or ar_dataview parameter value cannot be empty`)
		})
	})
}

func TestEvalUnaryExpr(t *testing.T) {
	Convey("test promql_service evalUnaryExpr ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		Convey("UnaryExpr -foo @20 when dataview is empty ", func() {
			queryTmp := &interfaces.Query{
				Start:      1652320539000,
				End:        1652320554000,
				FixedStart: 1652320539000,
				FixedEnd:   1652320554000,
				Interval:   3000,
			}

			queryTmp.QueryStr = `-foo @20`
			expr, _ := parser.ParseExpr(testCtx, queryTmp.QueryStr)

			res, status, err := psMock.evalUnaryExpr(testCtx, expr.(*parser.UnaryExpr), queryTmp)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `missing ar_dataview parameter or ar_dataview parameter value cannot be empty`)
		})

		Convey("scalar op scalar expression -3^1000", func() {
			queryWithFixedTime.QueryStr = `-3^1000`

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label(nil),
					Points: []static.Point{
						{
							T: 1652320539000,
							V: math.Inf(-1),
						},
						{
							T: 1652320542000,
							V: math.Inf(-1),
						},
						{
							T: 1652320545000,
							V: math.Inf(-1),
						},
						{
							T: 1652320548000,
							V: math.Inf(-1),
						},
						{
							T: 1652320551000,
							V: math.Inf(-1),
						},
						{
							T: 1652320554000,
							V: math.Inf(-1),
						},
					},
				},
			}

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			expr, err := static.PreprocessExpr(expr, queryWithFixedTime.Start, queryWithFixedTime.End)
			So(err, ShouldBeNil)

			res, status, err := psMock.evalStepInvariantExpr(testCtx, expr.(*parser.StepInvariantExpr), queryWithFixedTime)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, expectMat)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestSeries(t *testing.T) {
	Convey("test promql_service Series ", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+node_cpu_guest_seconds_total", leafnodes.TsidData{
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
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-1"] = time.Now()

		Convey("success series one match[] ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := psMock.Series(seriesMatcher)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), emptySeriesResult)

			So(err, ShouldBeNil)
		})

		Convey("success series with two same match[] ", func() {

			expectedRslt := "{\"status\":\"success\",\"data\":[{\"cpu\":\"0\",\"instance\":\"instance_abc\",\"job\":\"prometheus\",\"mode\":\"nice\"},{\"cpu\":\"1\",\"instance\":\"instance_abc\",\"job\":\"prometheus\",\"mode\":\"nice\"}]}"
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeCpuGuestSeriesDslResult), http.StatusOK, nil)

			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeCpuGuestSeriesDslResult), http.StatusOK, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`,
				`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := psMock.Series(seriesMatcher)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), expectedRslt)

			So(err, ShouldBeNil)
		})

		Convey("success series with two different match[] ", func() {

			expectedRslt := `{"status":"success","data":[{"cpu":"0","instance":"instance_abc","job":"prometheus","mode":"nice"},{"cpu":"1","instance":"instance_abc","job":"prometheus","mode":"nice"},{"device":"/dev/mapper/centos-root","fstype":"xfs","instance":"instance_abc","job":"prometheus","mountpoint":"/"},{"device":"/dev/sda1","fstype":"xfs","instance":"instance_abc","job":"prometheus","mountpoint":"/boot"}]}`

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeCpuGuestSeriesDslResult), http.StatusOK, nil)

			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeFilesystemSeriesDslResult), http.StatusOK, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`,
				`node_filesystem_avail_bytes`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := psMock.Series(seriesMatcher)
			So(status, ShouldEqual, http.StatusOK)
			compareJsonString(string(res), expectedRslt)
			So(err, ShouldBeNil)
		})

		Convey("error series with ar_dataview is missing ", func() {
			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			ss := interfaces.Matchers{
				Start:      1655346155000,
				End:        1655350445000,
				LogGroupId: "",
			}
			ss.MatcherSet = matcherSets

			res, status, err := psMock.Series(ss)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, "missing ar_dataview parameter or ar_dataview parameter value cannot be empty")
		})

	})
}

func TestSortAndSortDescFunctionInQueryRange(t *testing.T) {
	Convey("Test sort and sort_desc function in query_range ", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		Convey("Test sort function in query_range ", func() {
			queryWithFixedTime.QueryStr = `sort(prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.index=~"a.*|node.*"}*3)`
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			res, status, err := psMock.eval(testCtx, expr, queryWithFixedTime)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)

		})
		Convey("Test sort_desc function in query_range ", func() {
			queryWithFixedTime.QueryStr = `sort_desc(prometheus.metrics.opensearch_indices_shards_docs` +
				`{prometheus.labels.index=~"a.*|node.*"}*3)`
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTime.QueryStr)
			res, status, err := psMock.eval(testCtx, expr, queryWithFixedTime)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestTestAggOverTime(t *testing.T) {
	Convey("test promql_service eval about agg_over_time ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

		Convey("funcCall is valid", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `abc_over_time(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			res, status, err := psMock.eval(testCtx, expr, query)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, " FunctionCalls is not defined, please input valid function. ")
		})

		// eval agg_over_time UT, here I mock leafnodes.AggOverTime, UT of AggOverTime is in agg_over_time_test.go
		Convey("funcCall avg_over_time ", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "avg_over_time",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name:          "node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "avg_over_time(node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "AggOverTime", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("funcCall sum_over_time ", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "sum_over_time",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name:          "node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "sum_over_time(node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "AggOverTime", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("funcCall max_over_time ", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "max_over_time",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name:          "node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "max_over_time(node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "AggOverTime", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("funcCall min_over_time ", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "min_over_time",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name:          "node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "min_over_time(node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "AggOverTime", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("funcCall count_over_time ", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name:       "count_over_time",
					ArgTypes:   []parser.ValueType{"matrix"},
					Variadic:   0,
					ReturnType: "vector",
				},
				Args: []parser.Expr{&parser.MatrixSelector{
					VectorSelector: &parser.VectorSelector{
						Name:          "node_cpu_seconds_total",
						LabelMatchers: []*labels.Matcher{},
					},
					Range:  1 * time.Minute,
					EndPos: 148,
				},
				},
				PosRange: parser.PositionRange{Start: 0, End: 149},
			}

			query := &interfaces.Query{
				QueryStr:       "count_over_time(node_cpu_seconds_total[1m])",
				Start:          1655346045000,
				End:            1655346300000,
				Interval:       15000,
				FixedStart:     1655346045000,
				FixedEnd:       1655346300000,
				IsInstantQuery: false,
			}
			ln := &leafnodes.LeafNodes{}

			patches := ApplyMethodReturn(ln, "AggOverTime", static.Matrix{}, 200, nil)
			defer patches.Reset()

			res, status, err := psMock.eval(testCtx, expr, query)
			_, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestEvalCumulativeSum(t *testing.T) {
	Convey("test promql_service evalCumulativeSum ", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("0+a", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.UnixMilli(1652320533000),
			EndTime:         time.UnixMilli(1652320557000),
			Tsids:           []string{"id1"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
				},
			},
		})
		interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = time.Now()
		interfaces.INDEX_PATTERN_SPLIT_TIME["metricbeat-*"] = time.Now()

		Convey("CumulativeSum a", func() {

			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{T: 1652320533000, V: 0},
						{T: 1652320536000, V: 0},
						{T: 1652320539000, V: 20},
						{T: 1652320542000, V: 30},
						{T: 1652320545000, V: 51},
						{T: 1652320548000, V: 62},
						{T: 1652320551000, V: 84},
						{T: 1652320554000, V: 96},
						{T: 1652320557000, V: 96},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{T: 1652320533000, V: 0},
						{T: 1652320536000, V: 0},
						{T: 1652320539000, V: 40},
						{T: 1652320542000, V: 60},
						{T: 1652320545000, V: 101},
						{T: 1652320548000, V: 122},
						{T: 1652320551000, V: 164},
						{T: 1652320554000, V: 186},
						{T: 1652320557000, V: 186},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
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

			queryWithFixedTimeTmp := &interfaces.Query{
				Start:      1652320533000,
				End:        1652320557000,
				FixedStart: 1652320533000,
				FixedEnd:   1652320557000,
				Interval:   3000,
				LogGroupId: "a",
				QueryStr:   `cumulative_sum(a)`,
				Limit:      -1,
			}
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTimeTmp.QueryStr)

			res, status, err := psMock.evalCumulativeSum(testCtx, expr.(*parser.Call), queryWithFixedTimeTmp)
			mat, ok := res.(static.PageMatrix)

			So(ok, ShouldBeTrue)
			So(mat, ShouldResemble, static.PageMatrix{Matrix: expectMat, TotalSeries: 2})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestEvalKMinuteDowntime(t *testing.T) {
	Convey("test promql_service evalKMinuteDowntime ", t, func() {
		loc, _ := time.LoadLocation(os.Getenv("TZ"))
		common.APP_LOCATION = loc

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("123456+16ac079aedd3e5cfe54a6cee689222a1", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.UnixMilli(1652319900000),
			EndTime:         time.UnixMilli(1652321400000),
			Tsids:           []string{labelsStrNas, labelsStrNode, labelsStrNode2},
			TsidsMap: map[string]labels.Labels{
				labelsStrNas:   nasLabels,
				labelsStrNode:  nodeLabels,
				labelsStrNode2: noeeLabels,
			},
		})
		defer leafnodes.Tsids_Of_Model_Metric_Map.Delete("123456+16ac079aedd3e5cfe54a6cee689222a1")

		Convey("KMinuteDowntime a", func() {

			expectSeries := []static.Series{
				{
					Metric: nasLabels,
					Points: []static.Point{
						{T: 1652321400000, V: 14},
					},
				},
				{
					Metric: nodeLabels,
					Points: []static.Point{
						{T: 1652321400000, V: 17},
					},
				},
				{
					Metric: noeeLabels,
					Points: []static.Point{
						{T: 1652321400000, V: 20},
					},
				},
			}
			var expectMat static.Matrix
			expectMat = append(expectMat, expectSeries...)

			var indexShardsArr = make([]*interfaces.IndexShards, 0, 1)
			indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
				IndexName: index,
				Pri:       "1",
			})
			defer leafnodes.Number_Of_Shards_Map.Delete(indexPattern)
			notEmptyJson, _ := sonic.Marshal(indexShardsArr)

			interfaces.INDEX_BASE_SPLIT_TIME["metricbeat-*"] = time.Now()

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), indexPattern).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indexShardsArr, nil, 200, nil)

			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(dslResultByUnusable), http.StatusOK, nil)

			queryWithFixedTimeTmp := &interfaces.Query{
				IsInstantQuery: true,
				Start:          1652321400000,
				End:            1652321400000,
				FixedStart:     1652321400000,
				FixedEnd:       1652321400000,
				// LookBackDelta:  20 * 60 * 1000,
				LogGroupId:    "a",
				QueryStr:      `continuous_k_minute_downtime(5, -1, 0, a)`,
				Limit:         -1,
				ModelId:       "123456",
				IsMetricModel: true,
			}
			expr, _ := parser.ParseExpr(testCtx, queryWithFixedTimeTmp.QueryStr)

			res, status, err := psMock.evalKMinuteDowntime(testCtx, expr.(*parser.Call), queryWithFixedTimeTmp)
			mat, ok := res.(static.Matrix)

			So(ok, ShouldBeTrue)
			So(len(mat), ShouldResemble, len(expectMat))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetFields(t *testing.T) {
	Convey("test GetFields", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("+4cafff8d1837e84727ff440e349c4049", leafnodes.TsidData{
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

		Convey("when expression type is not vector/scalar/matrix", func() {
			query := interfaces.Query{
				QueryStr: `"test"`,
			}

			fields, status, err := psMock.GetFields(testCtx, query)

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when parse expression fails", func() {
			query := interfaces.Query{
				QueryStr: `invalid expression`,
			}

			fields, status, err := psMock.GetFields(testCtx, query)

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is valid", func() {
			query := interfaces.Query{
				QueryStr: `metric{label="value"}`,
			}

			fields, status, err := psMock.GetFields(testCtx, query)

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestEvalFieldsInfo(t *testing.T) {
	Convey("Given a promQL service", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("+6c1f145ac8feab4d4cee4daa37376bb3", leafnodes.TsidData{
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
		leafnodes.Tsids_Of_Model_Metric_Map.Store("+91612042db783c667fcaa6a87863a9c2", leafnodes.TsidData{
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

		query := &interfaces.Query{}

		Convey("when expression is AggregateExpr", func() {
			expr := &parser.AggregateExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is Call with undefined function", func() {
			query.QueryStr = `abc(metric)`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is Call with matrix function irate", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "irate",
				},
				Args: []parser.Expr{
					&parser.MatrixSelector{
						VectorSelector: &parser.VectorSelector{
							Name: "metric",
						},
						Range: 5 * time.Minute,
					},
				},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is Call with matrix function irate and invalid VectorSelector type", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "irate",
				},
				Args: []parser.Expr{
					&parser.MatrixSelector{
						VectorSelector: &parser.NumberLiteral{Val: 1}, // Invalid type
						Range:          5 * time.Minute,
					},
				},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is Call with function abs", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "abs",
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
				},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is Call with unsupported function 1", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "unsupported_function",
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
				},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			promqlErr, ok := err.(uerrors.PromQLError)
			So(ok, ShouldBeTrue)
			So(promqlErr.Typ, ShouldEqual, uerrors.ErrorBadData)
			So(promqlErr.Err.Error(), ShouldEqual, " 'unsupported_function' is not currently supported. ")
		})

		Convey("when expression is ParenExpr", func() {
			expr := &parser.ParenExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is UnaryExpr", func() {
			expr := &parser.UnaryExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is BinaryExpr", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric"},
				RHS: &parser.VectorSelector{Name: "metric1"},
			}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is NumberLiteral", func() {
			expr := &parser.NumberLiteral{Val: 1}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is StepInvariantExpr", func() {
			expr := &parser.StepInvariantExpr{}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression type is unhandled", func() {
			expr := &parser.StringLiteral{}

			fields, status, err := psMock.evalFieldsInfo(testCtx, expr, query, "")

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetFieldValues(t *testing.T) {
	Convey("Given a promQL service", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("+4cafff8d1837e84727ff440e349c4049", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1", "id2"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
					{Name: "label2", Value: "value2"},
				},
				"id2": {
					{Name: "label1", Value: "value3"},
					{Name: "label2", Value: "value4"},
				},
			},
		})

		Convey("when expression type is not vector/scalar/matrix", func() {
			query := interfaces.Query{
				QueryStr: `"test"`,
			}

			values, status, err := psMock.GetFieldValues(testCtx, query, "label1")

			So(values, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when parse expression fails", func() {
			query := interfaces.Query{
				QueryStr: `invalid expression`,
			}

			values, status, err := psMock.GetFieldValues(testCtx, query, "label1")

			So(values, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is valid", func() {
			query := interfaces.Query{
				QueryStr: `metric{label="value"}`,
			}

			values, status, err := psMock.GetFieldValues(testCtx, query, "label1")

			So(values, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetLabels(t *testing.T) {
	Convey("Given a promQL service", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("+4cafff8d1837e84727ff440e349c4049", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1", "id2"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
					{Name: "label2", Value: "value2"},
				},
				"id2": {
					{Name: "label1", Value: "value3"},
					{Name: "label2", Value: "value4"},
				},
			},
		})

		Convey("when expression type is not vector/scalar/matrix", func() {
			query := interfaces.Query{
				QueryStr: `"test"`,
			}

			labels, status, err := psMock.GetLabels(testCtx, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when parse expression fails", func() {
			query := interfaces.Query{
				QueryStr: `invalid expression`,
			}

			labels, status, err := psMock.GetLabels(testCtx, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is valid", func() {
			query := interfaces.Query{
				QueryStr: `metric{label="value"}`,
			}

			labels, status, err := psMock.GetLabels(testCtx, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestEvalLabelsInfo(t *testing.T) {
	Convey("Given a promQL service", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		mmsMock := umock.NewMockMetricModelService(mockCtrl)
		psMock := mockNewPromqlService(osaMock, lgaMock, dvsMock, mmsMock)

		leafnodes.Tsids_Of_Model_Metric_Map.Store("+6c1f145ac8feab4d4cee4daa37376bb3", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1", "id2"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
					{Name: "label2", Value: "value2"},
				},
				"id2": {
					{Name: "label1", Value: "value3"},
					{Name: "label2", Value: "value4"},
				},
			},
		})
		leafnodes.Tsids_Of_Model_Metric_Map.Store("+91612042db783c667fcaa6a87863a9c2", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1", "id2"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
					{Name: "label2", Value: "value2"},
				},
				"id2": {
					{Name: "label1", Value: "value3"},
					{Name: "label2", Value: "value4"},
				},
			},
		})
		leafnodes.Tsids_Of_Model_Metric_Map.Store("+bfbe5ba037a931f7f354d30db60542f4", leafnodes.TsidData{
			RefreshTime:     time.Now(),
			FullRefreshTime: time.Now(),
			StartTime:       time.Unix(0, 0),
			EndTime:         time.Unix(0, 0),
			Tsids:           []string{"id1", "id2"},
			TsidsMap: map[string]labels.Labels{
				"id1": {
					{Name: "label1", Value: "value1"},
					{Name: "label3", Value: "value2"},
					{Name: "label4", Value: "value41"},
				},
				"id2": {
					{Name: "label1", Value: "value3"},
					{Name: "label3", Value: "value4"},
					{Name: "label4", Value: "value41"},
				},
			},
		})

		query := &interfaces.Query{}

		Convey("when expression is AggregateExpr", func() {
			expr := &parser.AggregateExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is Call with undefined function", func() {
			query.QueryStr = `abc(metric)`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is Call with matrix function irate", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "irate",
				},
				Args: []parser.Expr{
					&parser.MatrixSelector{
						VectorSelector: &parser.VectorSelector{
							Name: "metric",
						},
						Range: 5 * time.Minute,
					},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is Call with function abs", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "abs",
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is ParenExpr", func() {
			expr := &parser.ParenExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is UnaryExpr", func() {
			expr := &parser.UnaryExpr{
				Expr: &parser.VectorSelector{Name: "metric"},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is BinaryExpr", func() {
			expr := &parser.BinaryExpr{
				LHS:            &parser.VectorSelector{Name: "metric"},
				RHS:            &parser.VectorSelector{Name: "metric1"},
				Op:             parser.ADD,
				VectorMatching: &parser.VectorMatching{},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is NumberLiteral", func() {
			expr := &parser.NumberLiteral{Val: 1}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is StepInvariantExpr", func() {
			expr := &parser.StepInvariantExpr{}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is AggregateExpr and evalLabelsInfo returns error", func() {
			expr := &parser.AggregateExpr{
				Op:   parser.SUM,
				Expr: &parser.StringLiteral{},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is AggregateExpr with grouping", func() {
			expr := &parser.AggregateExpr{
				Op:       parser.SUM,
				Expr:     &parser.VectorSelector{Name: "metric"},
				Grouping: []string{"label1", "label2"},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is AggregateExpr with grouping and without flag", func() {
			expr := &parser.AggregateExpr{
				Op:       parser.SUM,
				Expr:     &parser.VectorSelector{Name: "metric"},
				Grouping: []string{"label1"},
				Without:  true,
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeFalse)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is Call with unsupported function 1", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: "unsupported_function",
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
				},
			}

			fields, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(fields, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			promqlErr, ok := err.(uerrors.PromQLError)
			So(ok, ShouldBeTrue)
			So(promqlErr.Typ, ShouldEqual, uerrors.ErrorBadData)
			So(promqlErr.Err.Error(), ShouldEqual, " 'unsupported_function' is not currently supported. ")
		})

		Convey("when expression is Call with dict_labels function", func() {
			patch := ApplyFunc(data_dict.GetDictByName,
				func(dictName string) (interfaces.DataDict, bool) {
					return interfaces.DataDict{
						UniqueKey: true,
						DictRecords: map[string][]map[string]string{
							"host1": {
								{
									"key1":   "host1",
									"value1": "oracal",
									"value2": "15G",
									"value3": "36",
								},
							},
							"host2": {
								{
									"key1":   "host2",
									"value1": "mysql",
									"value2": "3G",
									"value3": "89",
								},
							},
							"host3": {
								{
									"key1":   "host3",
									"value1": "redis",
									"value2": "12G",
									"value3": "99",
								},
							},
						},
						Dimension: interfaces.Dimension{
							Keys: []interfaces.DimensionItem{
								{Name: "key1"},
							},
							Values: []interfaces.DimensionItem{
								{Name: "value1"},
								{Name: "value2"},
								{Name: "value3"},
							},
						},
					}, true
				},
			)
			defer patch.Reset()

			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.DICT_LABELS,
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
					&parser.StringLiteral{Val: "dict_name"},
					&parser.StringLiteral{Val: "key1"},
					&parser.StringLiteral{Val: "label1"},
					&parser.StringLiteral{Val: "value1"},
					&parser.StringLiteral{Val: "db"},
					&parser.StringLiteral{Val: "value2"},
					&parser.StringLiteral{Val: "capacity"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is Call with function dict_labels and error evaluating args", func() {
			patch := ApplyFunc(data_dict.GetDictByName,
				func(dictName string) (interfaces.DataDict, bool) {
					return interfaces.DataDict{
						UniqueKey: true,
						DictRecords: map[string][]map[string]string{
							"host1": {
								{
									"key1":   "host1",
									"value1": "oracal",
									"value2": "15G",
									"value3": "36",
								},
							},
							"host2": {
								{
									"key1":   "host2",
									"value1": "mysql",
									"value2": "3G",
									"value3": "89",
								},
							},
							"host3": {
								{
									"key1":   "host3",
									"value1": "redis",
									"value2": "12G",
									"value3": "99",
								},
							},
						},
						Dimension: interfaces.Dimension{
							Keys: []interfaces.DimensionItem{
								{Name: "key1"},
							},
							Values: []interfaces.DimensionItem{
								{Name: "value1"},
								{Name: "value2"},
								{Name: "value3"},
							},
						},
					}, true
				},
			)
			defer patch.Reset()

			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.DICT_LABELS,
				},
				Args: []parser.Expr{
					&parser.StringLiteral{}, // Invalid arg type
					&parser.StringLiteral{Val: "dict_name"},
					&parser.StringLiteral{Val: "key1"},
					&parser.StringLiteral{Val: "label1"},
					&parser.StringLiteral{Val: "value1"},
					&parser.StringLiteral{Val: "db"},
					&parser.StringLiteral{Val: "value2"},
					&parser.StringLiteral{Val: "capacity"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is Call with function dict_values", func() {
			patch := ApplyFunc(data_dict.GetDictByName,
				func(dictName string) (interfaces.DataDict, bool) {
					return interfaces.DataDict{
						UniqueKey: true,
						DictRecords: map[string][]map[string]string{
							"host1": {
								{
									"key1":   "host1",
									"value1": "oracal",
									"value2": "15G",
									"value3": "36",
								},
							},
							"host2": {
								{
									"key1":   "host2",
									"value1": "mysql",
									"value2": "3G",
									"value3": "89",
								},
							},
							"host3": {
								{
									"key1":   "host3",
									"value1": "redis",
									"value2": "12G",
									"value3": "99",
								},
							},
						},
						Dimension: interfaces.Dimension{
							Keys: []interfaces.DimensionItem{
								{Name: "key1"},
							},
							Values: []interfaces.DimensionItem{
								{Name: "value1"},
								{Name: "value2"},
								{Name: "value3"},
							},
						},
					}, true
				},
			)
			defer patch.Reset()

			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.DICT_VALUES,
				},
				Args: []parser.Expr{
					&parser.StringLiteral{Val: "dict_name"},
					&parser.StringLiteral{Val: "value3"},
					&parser.StringLiteral{Val: "key1"},
					&parser.StringLiteral{Val: "label1"},
					&parser.StringLiteral{Val: "value1"},
					&parser.StringLiteral{Val: "db"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["db"], ShouldBeTrue)
		})

		Convey("when expression is Call with label_join function", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.LABEL_JOIN,
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
					&parser.StringLiteral{Val: "dst"},
					&parser.StringLiteral{Val: ","},
					&parser.StringLiteral{Val: "src1"},
					&parser.StringLiteral{Val: "src2"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["dst"], ShouldBeTrue)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is Call with label_replace function", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.LABEL_REPLACE,
				},
				Args: []parser.Expr{
					&parser.VectorSelector{Name: "metric"},
					&parser.StringLiteral{Val: "dst"},
					&parser.StringLiteral{Val: "replacement"},
					&parser.StringLiteral{Val: "src"},
					&parser.StringLiteral{Val: "regex"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["dst"], ShouldBeTrue)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is Call with label_join function and evalLabelsInfo error", func() {
			expr := &parser.Call{
				Func: &parser.Function{
					Name: interfaces.LABEL_JOIN,
				},
				Args: []parser.Expr{
					&parser.StringLiteral{}, // This will cause evalLabelsInfo to return error
					&parser.StringLiteral{Val: "dst"},
					&parser.StringLiteral{Val: ","},
					&parser.StringLiteral{Val: "src1"},
					&parser.StringLiteral{Val: "src2"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=true", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             true,
					MatchingLabels: []string{"label1", "label2"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=true and missing label", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             true,
					MatchingLabels: []string{"label1", "missing_label"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeEmpty)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=false and ignoring labels", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             false,
					MatchingLabels: []string{"label3", "label4"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=false and mismatched non-ignored labels", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             false,
					MatchingLabels: []string{"label2"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels["label1"], ShouldBeTrue)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is BinaryExpr with VectorMatching.LeftJoin=true", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					LeftJoin:       true,
					MatchingLabels: []string{"label1"},
					Include:        []string{"label2", "label4"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
		})

		Convey("when expression is BinaryExpr with VectorMatching.LeftJoin=true and no matching labels", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					LeftJoin:       true,
					MatchingLabels: []string{"missing_label"},
					Include:        []string{"label3", "label4"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
			// Should not include group_left labels when no match
			So(labels["label3"], ShouldBeFalse)
			So(labels["label4"], ShouldBeFalse)
		})

		Convey("when expression is BinaryExpr with VectorMatching.OutJoin=true", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric2"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					OutJoin:        true,
					MatchingLabels: []string{"label1", "label2"},
					IncludeLeft:    []string{"label3"},
					IncludeRight:   []string{"label4"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			// Should include matching labels that exist in either side
			So(labels["label1"], ShouldBeTrue)
			So(labels["label2"], ShouldBeTrue)
			// Should include select_left labels that exist in left side
			So(labels["label3"], ShouldBeFalse)
			// Should include select_right labels that exist in right side
			So(labels["label4"], ShouldBeTrue)
		})

		Convey("when expression is BinaryExpr with VectorMatching.OutJoin=true and IncludeLeft labels exist in labelsMap", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric2"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					OutJoin:        true,
					MatchingLabels: []string{"label1"},
					IncludeLeft:    []string{"label3", "label4"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			// Should include matching labels
			So(labels["label1"], ShouldBeTrue)
			// Should include IncludeLeft labels that exist in labelsMap
			So(labels["label3"], ShouldBeTrue)
			So(labels["label4"], ShouldBeTrue)
		})

		Convey("when expression is unhandled type", func() {
			expr := &parser.StringLiteral{
				Val: "test",
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unhandled expression of type")
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=false and MatchingLabels not empty and key not in rFields", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric2"},
				RHS: &parser.VectorSelector{Name: "metric"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             false,
					MatchingLabels: []string{"non_existent_label"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("when expression is BinaryExpr with VectorMatching.On=false and MatchingLabels not empty and key not in lFields", func() {
			expr := &parser.BinaryExpr{
				LHS: &parser.VectorSelector{Name: "metric1"},
				RHS: &parser.VectorSelector{Name: "metric2"},
				Op:  parser.ADD,
				VectorMatching: &parser.VectorMatching{
					On:             false,
					MatchingLabels: []string{"label3"},
				},
			}

			labels, status, err := psMock.evalLabelsInfo(testCtx, expr, query)

			So(labels, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}
