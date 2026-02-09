// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"net/http"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

var (
	dslStart = int64(1686218262951)
	dslEnd   = int64(1686218562951)

	dslStart2 = int64(1686117584856)
	dslEnd2   = int64(1686121184856)

	dslStart3 = int64(1692512527037)
	dslEnd3   = int64(1695190927037)

	dslStart4 = int64(1722691260000)
	dslEnd4   = int64(1722730600000)

	step_1min    = "1m"
	step_30m     = "30m"
	step_50m     = "50m"
	step_1hour   = "1h"
	step_1day    = "1d"
	step_1w      = "1w"
	step_1M      = "1M"
	setp_week    = "week"
	step_month   = "month"
	step_1q      = "1q"
	step_quarter = "quarter"
	step_1y      = "1y"
	step_year    = "year"
	step_60000   = int64(60000)
	step_3000000 = int64(3000000)
	step_1800000 = int64(1800000)
	// time5Min             = 5 * time.Minute
	StepsMap = map[string]string{
		"1m":  "1m",
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

	DslResultWithTopHits = `{
		"took" : 7,
		"timed_out" : false,
		"_shards" : {
		  "total" : 6,
		  "successful" : 6,
		  "skipped" : 0,
		  "failed" : 0
		},
		"hits" : {
		  "total" : {
			"value" : 480,
			"relation" : "eq"
		  },
		  "max_score" : null,
		  "hits" : [ ]
		},
		"aggregations" : {
		  "cpu2" : {
			"doc_count_error_upper_bound" : 0,
			"sum_other_doc_count" : 0,
			"buckets" : [
			  {
				"key" : "1",
				"doc_count" : 240,
				"mode" : {
				  "doc_count_error_upper_bound" : 0,
				  "sum_other_doc_count" : 0,
				  "buckets" : [
					{
					  "key" : "iowait",
					  "doc_count" : 240,
					  "time2" : {
						"buckets" : [
						  {
							"key_as_string" : "2023-06-07 05:50:00.000",
							"key" : 1686117000000,
							"doc_count" : 161,
							"value" : {
							  "hits" : {
								"total" : {
								  "value" : 161,
								  "relation" : "eq"
								},
								"max_score" : null,
								"hits" : [
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "nIeVlIgBltw65OWEQgPG",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208242.43
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119988665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "RYeVlIgBltw65OWECAND",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208242.43
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119973665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "Z4eUlIgBltw65OWEzQKj",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208242.42
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119958666
									]
								  }
								]
							  }
							}
						  },
						  {
							"key_as_string" : "2023-06-07 06:40:00.000",
							"key" : 1686120000000,
							"doc_count" : 79,
							"value" : {
							  "hits" : {
								"total" : {
								  "value" : 79,
								  "relation" : "eq"
								},
								"max_score" : null,
								"hits" : [
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "6YenlIgBltw65OWEVy_F",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208243.02
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121173666
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "G4enlIgBltw65OWEHS8c",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208243.01
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121158665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "xIemlIgBltw65OWE4i6h",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 208243.01
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121143665
									]
								  }
								]
							  }
							}
						  }
						]
					  }
					}
				  ]
				}
			  },
			  {
				"key" : "2",
				"doc_count" : 240,
				"mode" : {
				  "doc_count_error_upper_bound" : 0,
				  "sum_other_doc_count" : 0,
				  "buckets" : [
					{
					  "key" : "iowait",
					  "doc_count" : 240,
					  "time2" : {
						"buckets" : [
						  {
							"key_as_string" : "2023-06-07 05:50:00.000",
							"key" : 1686117000000,
							"doc_count" : 161,
							"value" : {
							  "hits" : {
								"total" : {
								  "value" : 161,
								  "relation" : "eq"
								},
								"max_score" : null,
								"hits" : [
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "7oeVlIgBltw65OWEQgPK",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212854.73
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119988665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "A4eVlIgBltw65OWECAM_",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212854.73
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119973665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "3YeUlIgBltw65OWEzQKn",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212854.72
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686119958666
									]
								  }
								]
							  }
							}
						  },
						  {
							"key_as_string" : "2023-06-07 06:40:00.000",
							"key" : 1686120000000,
							"doc_count" : 79,
							"value" : {
							  "hits" : {
								"total" : {
								  "value" : 79,
								  "relation" : "eq"
								},
								"max_score" : null,
								"hits" : [
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "n4enlIgBltw65OWEVy_D",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212855.24
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121173666
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "ioenlIgBltw65OWEHS8g",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212855.24
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121158665
									]
								  },
								  {
									"_index" : "cnao_metric_node_exporter2-2023.06-0",
									"_type" : "_doc",
									"_id" : "y4emlIgBltw65OWE4i6h",
									"_score" : null,
									"_routing" : "936733",
									"_source" : {
									  "metrics" : {
										"node_cpu_seconds_total" : 212855.24
									  },
									  "labels" : {
										"job" : "prometheus"
									  }
									},
									"sort" : [
									  1686121143665
									]
								  }
								]
							  }
							}
						  }
						]
					  }
					}
				  ]
				}
			  }
			]
		  }
		}
	  }
	  `

	DslResultWithTopHits4Instant = `{
		"took" : 7,
		"timed_out" : false,
		"_shards" : {
		  "total" : 6,
		  "successful" : 6,
		  "skipped" : 3,
		  "failed" : 0
		},
		"hits" : {
		  "total" : {
			"value" : 76,
			"relation" : "eq"
		  },
		  "max_score" : null,
		  "hits" : [ ]
		},
		"aggregations" : {
		  "cpu2" : {
			"doc_count_error_upper_bound" : 0,
			"sum_other_doc_count" : 0,
			"buckets" : [
			  {
				"key" : "1",
				"doc_count" : 38,
				"mode" : {
				  "doc_count_error_upper_bound" : 0,
				  "sum_other_doc_count" : 0,
				  "buckets" : [
					{
					  "key" : "user",
					  "doc_count" : 38,
					  "value" : {
						"hits" : {
						  "total" : {
							"value" : 38,
							"relation" : "eq"
						  },
						  "max_score" : null,
						  "hits" : [
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "f4ZwlIgBltw65OWEvqk6",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1690429.97
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117573666
							  ]
							},
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "C4ZwlIgBltw65OWEvqk5",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1690426.44
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117558665
							  ]
							},
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "0YZwlIgBltw65OWEvqg5",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1690418.3
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117543665
							  ]
							}
						  ]
						}
					  }
					}
				  ]
				}
			  },
			  {
				"key" : "2",
				"doc_count" : 38,
				"mode" : {
				  "doc_count_error_upper_bound" : 0,
				  "sum_other_doc_count" : 0,
				  "buckets" : [
					{
					  "key" : "user",
					  "doc_count" : 38,
					  "value" : {
						"hits" : {
						  "total" : {
							"value" : 38,
							"relation" : "eq"
						  },
						  "max_score" : null,
						  "hits" : [
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "eoZwlIgBltw65OWEvqk6",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1696406.23
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117573666
							  ]
							},
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "FoZwlIgBltw65OWEvqk5",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1696401.81
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117558665
							  ]
							},
							{
							  "_index" : "cnao_metric_node_exporter2-2023.06-0",
							  "_type" : "_doc",
							  "_id" : "qYZwlIgBltw65OWEvqg5",
							  "_score" : null,
							  "_routing" : "936731",
							  "_source" : {
								"metrics" : {
								  "node_cpu_seconds_total" : 1696393.52
								},
								"labels" : {
								  "job" : "prometheus"
								}
							  },
							  "sort" : [
								1686117543665
							  ]
							}
						  ]
						}
					  }
					}
				  ]
				}
			  }
			]
		  }
		}
	  }
	  `

	DslResultWithFiltersSum = `{
		"took" : 6,
		"timed_out" : false,
		"_shards" : {
		  "total" : 6,
		  "successful" : 6,
		  "skipped" : 0,
		  "failed" : 0
		},
		"hits" : {
		  "total" : {
			"value" : 5760,
			"relation" : "eq"
		  },
		  "max_score" : null,
		  "hits" : [ ]
		},
		"aggregations" : {
		  "NAME1" : {
			"buckets" : {
			  "system" : {
				"doc_count" : 720,
				"NAME2" : {
				  "buckets" : [
					{
					  "key_as_string" : "2023-06-07T05:50:00.000Z",
					  "key" : 1686117000000,
					  "doc_count" : 483,
					  "NAME3" : {
						"value" : 8.14395374925486E14,
						"value_as_string" : "+27777-02-28T09:28:45.486Z"
					  }
					},
					{
					  "key_as_string" : "2023-06-07T06:40:00.000Z",
					  "key" : 1686120000000,
					  "doc_count" : 237,
					  "NAME3" : {
						"value" : 3.99610579513734E14,
						"value_as_string" : "+14633-02-26T10:45:13.734Z"
					  }
					}
				  ]
				}
			  },
			  "user" : {
				"doc_count" : 720,
				"NAME2" : {
				  "buckets" : [
					{
					  "key_as_string" : "2023-06-07T05:50:00.000Z",
					  "key" : 1686117000000,
					  "doc_count" : 483,
					  "NAME3" : {
						"value" : 8.14395374925486E14,
						"value_as_string" : "+27777-02-28T09:28:45.486Z"
					  }
					},
					{
					  "key_as_string" : "2023-06-07T06:40:00.000Z",
					  "key" : 1686120000000,
					  "doc_count" : 237,
					  "NAME3" : {
						"value" : 3.99610579513734E14,
						"value_as_string" : "+14633-02-26T10:45:13.734Z"
					  }
					}
				  ]
				}
			  }
			}
		  }
		}
	  }
	  `

	DslResultWithFiltersSum4Instant = `{
		"took" : 5,
		"timed_out" : false,
		"_shards" : {
		  "total" : 6,
		  "successful" : 6,
		  "skipped" : 0,
		  "failed" : 0
		},
		"hits" : {
		  "total" : {
			"value" : 912,
			"relation" : "eq"
		  },
		  "max_score" : null,
		  "hits" : [ ]
		},
		"aggregations" : {
		  "NAME1" : {
			"buckets" : {
			  "system" : {
				"doc_count" : 114,
				"NAME3" : {
				  "value" : 1.92217369062858E14,
				  "value_as_string" : "8061-02-15T01:37:42.858Z"
				}
			  },
			  "user" : {
				"doc_count" : 114,
				"NAME3" : {
				  "value" : 1.92217369062858E14,
				  "value_as_string" : "8061-02-15T01:37:42.858Z"
				}
			  }
			}
		  }
		}
	  }
	  `
)

func TestParseDslQuery(t *testing.T) {
	Convey("Test parseDslQuery", t, func() {

		param := interfaces.MetricModelQueryParameters{IgnoringMemoryCache: true}
		Convey("parse failed, because dsl Unmarshal error ", func() {
			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param, Formula: `"size":0`}, &interfaces.DataView{})
			So(res, ShouldNotBeNil)
		})

		Convey("parse failed, because parse agg error ", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Formula,
					Description:  "Invalid Formula",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "The aggregation of dsl is not a map",
				},
			}

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param,
				Formula: `{"aggs": 1,"size":0}`}, &interfaces.DataView{})
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse failed, because parse query error ", func() {

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param, Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"NAME2": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "1d"
					  },
					  "aggs": {
						"NAME3": {
						  "value_count": {
							"field": "1"
						  }
						}
					  }
					}
				  }
				}
			  }}`,
				Filters: []interfaces.Filter{{}}}, &dataview) // filters 被重写为空
			So(res, ShouldBeNil)
		})

		Convey("parse success when instant query ", func() {

			dslInfo, err := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param, Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"NAME2": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "1d"
					  },
					  "aggs": {
						"NAME3": {
						  "value_count": {
							"field": "1"
						  }
						}
					  }
					}
				  }
				}
			  }}`,
				QueryTimeParams: interfaces.QueryTimeParams{Time: 1000},
			}, &dataview)

			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(dslInfo.RangeQueryDSL, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("parse failed, because aggs not exists ", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Formula,
					Description:  "Invalid Formula",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "dsl missing aggregation.",
				},
			}

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param, Formula: `{"size":0}`}, &dataview)
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse-parseAggs failed, because multiple aggs ", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Formula,
					Description:  "Invalid Formula",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "Multiple aggregation is not supported",
				},
			}

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param,
				Formula: `{"size":0,"aggs":{"terms1":{},"terms2":{}}}`}, &dataview)
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse-parseAggs failed, because aggs is not a map", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Formula,
					Description:  "Invalid Formula",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "The aggregation of dsl is not a map",
				},
			}

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param,
				Formula: `{"size":0,"aggs":{"terms1":1}}`}, &dataview)
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse-parseAggs failed, because sub-aggs is not a map", func() {

			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusBadRequest,
				Language: "en-US",
				BaseError: rest.BaseError{
					ErrorCode:    uerrors.Uniquery_MetricModel_InvalidParameter_Formula,
					Description:  "Invalid Formula",
					Solution:     "Please check whether the parameter is correct.",
					ErrorLink:    "None",
					ErrorDetails: "The sub-aggregation of dsl is not a map",
				},
			}

			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{MetricModelQueryParameters: param, Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": 1
				}
			  }}`}, &dataview)
			So(res, ShouldResemble, expectedErr)
		})

		Convey("parse-parseAggs failed, because top_hits unmarshal error", func() {
			_, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "1h"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": "1",
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})
			So(res, ShouldNotBeNil)
		})

		Convey("success with top hits instant query", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{IsInstantQuery: true},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "{{__interval}}"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with top hits instant query when fixed_interval is cons", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{IsInstantQuery: true},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "6m"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with top hits range query", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{StepStr: &fiveMin},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "{{__interval}}",
						"time_zone":"Asia/Shanghai"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with top hits range query when fixed_interval is cons", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{StepStr: &fiveMin},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "6m",
						"time_zone":"Asia/Shanghai"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with top hits range query calendar_interval is variable", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &minute,
				},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"calendar_interval": "{{__interval}}",
						"time_zone":"Asia/Shanghai"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with top hits range query calendar_interval is const", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{StepStr: &hour},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"calendar_interval": "minute",
						"time_zone":"Asia/Shanghai"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with filters", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Formula: `{
					"size": 0,
					  "query": {
					  "match_bool_prefix": {
						"__labels_str": "cpu mode iowait"
					  }
					},
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query no must", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
						  "filter": [
							{
							  "match_bool_prefix": {
								"__labels_str": "cpu mode iowait"
							  }
							}
						  ]
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query must object", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
							"must": {
								"match_bool_prefix": {
								  "__labels_str": "cpu mode iowait"
								}
							}
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &dataview)

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query must array", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
							"must": [
								{
								"match_bool_prefix": {
									"__labels_str": "cpu mode iowait"
								}
								},
								{
								"range": {
									"@timestamp": {
									"gte": 1686218386212
									}
								}
								}
							]
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &dataview)

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query no filters", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Filters: []interfaces.Filter{{Name: "a", Operation: interfaces.OPERATION_EQ, Value: 1}},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
						  "must": [
							{
							  "match_bool_prefix": {
								"__labels_str": "cpu mode iowait"
							  }
							}
						  ]
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query filter object", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Filters: []interfaces.Filter{{Name: "a", Operation: cond.OperationNotEq, Value: 1}},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
							"filter": {
								"match_bool_prefix": {
								   "__labels_str": "cpu mode iowait"
								}
							}
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &dataview)

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with query filter array", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart,
					End:     &dslEnd,
					StepStr: &fiveMin,
				},
				Filters: []interfaces.Filter{{Name: "a", Operation: cond.OperationNotEq, Value: 1}},
				Formula: `{
					"size": 0,
					"query": {
						"bool": {
							"filter": [
								{
								"match_bool_prefix": {
									"__labels_str": "cpu mode iowait"
								}
								},
								{
								"range": {
									"@timestamp": {
									"gte": 1686218386212
									}
								}
								}
							]
						}
					  },
					"aggs": {
					  "NAME1": {
						"terms": {
						  "field": "labels.cpu.keyword",
						  "size": 10
						},
						"aggs": {
						  "time": {
							"date_histogram": {
							  "field": "@timestamp",
							  "fixed_interval": "{{__interval}}",
							  "time_zone":"Asia/Shanghai"
							},
							"aggs": {
							  "NAME": {
								"top_hits": {
								  "size": 3,
									"sort": [
									  {
										"@timestamp": {
										  "order": "desc"
										}
									  }
									],
									"_source": {
									  "includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
									}
								  }
							  }
							}
						  }
						}
					  }
					}
				  }`}, &dataview)

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 3)
			So(res, ShouldBeNil)
		})

		Convey("success with terms,filters,range,date_range", func() {

			dslInfo, res := parseDsl(testENCtx, &interfaces.MetricModelQuery{
				MetricModelQueryParameters: param,
				QueryTimeParams:            interfaces.QueryTimeParams{StepStr: &fiveMin},
				Formula: `{"size":0,"aggs": {
				"NAME1": {
				  "terms": {
					"field": "labels.cpu.keyword",
					"size": 10
				  },
				  "aggs": {
					"mode": {
					  "filters": {
						"filters": {
						  "idle": {
							"match": {
							  "labels.mode": "idle"
							}
						  },
						  "system": {
							"match": {
							  "labels.mode": "system"
							}
						  }
						}
					  },
					  "aggs": {
						"date_rg": {
						  "date_range": {
							"field": "@timestamp",
							"format": "yyyy-MM-dd",
							"ranges": [
							  {
								"from": "now-51d/d",
								"to": "now-50d/d"
							  },
							  {
								"from": "now-52d/d",
								"to": "now-51d/d"
							  }
							]
						  },
				  "aggs": {
					"time": {
					  "date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "{{__interval}}",
						"time_zone":"Asia/Shanghai"
					  },
					  "aggs": {
						"NAME": {
						  "top_hits": {
							"size": 3,
							  "sort": [
								{
								  "@timestamp": {
									"order": "desc"
								  }
								}
							  ],
							  "_source": {
								"includes": [ "labels.cpu","metrics.node_cpu_seconds_total" ]
							  }
							}
						}
					  }
					}
				  }
				}
			  }}}}}}`}, &interfaces.DataView{})

			_, err := sonic.Marshal(dslInfo.RangeQueryDSL)
			So(err, ShouldBeNil)
			So(len(dslInfo.AggInfos), ShouldEqual, 5)
			So(dslInfo.BucketSeriesNum, ShouldEqual, 40)
			So(res, ShouldBeNil)
		})

		// Convey("parse success with cache ", func() {
		// 	Dsl_Info_Of_Model.Store(uint64(1), DSLInfoCache{RefreshTime: time.Now(), DslInfo: interfaces.DslInfo{}})
		// 	dslInfo, err := parseDsl(testENCtx
		// 		&interfaces.MetricModelQuery{
		// 			MetricModelID:              1,
		// 			MetricModelQueryParameters: interfaces.MetricModelQueryParameters{IgnoringMemoryCache: false},
		// 			Formula: `{"size":0,"aggs": {
		// 		"NAME1": {
		// 		  "terms": {
		// 			"field": "labels.cpu.keyword",
		// 			"size": 10
		// 		  },
		// 		  "aggs": {
		// 			"NAME2": {
		// 			  "date_histogram": {
		// 				"field": "@timestamp",
		// 				"fixed_interval": "1d"
		// 			  },
		// 			  "aggs": {
		// 				"NAME3": {
		// 				  "value_count": {
		// 					"field": "1"
		// 				  }
		// 				}
		// 			  }
		// 			}
		// 		  }
		// 		}
		// 	  }}`,
		// 			DataView:        dataview,
		// 			Time:            1000,
		// 			ModelUpdateTime: time.Now().Add(-1 * time.Hour),
		// 		})

		// 	So(len(dslInfo.AggInfos), ShouldEqual, 0)
		// 	So(err, ShouldBeNil)
		// })

	})
}

func TestParseDSLResult2Uniresponse(t *testing.T) {
	Convey("Test parseDSLResult2Uniresponse", t, func() {
		dslMetric := interfaces.MetricModel{
			QueryType: interfaces.DSL,
		}
		defaultqueryZone, _ := time.LoadLocation("")
		Convey("Test parseDSLResult2Uniresponse top_hits contian labels field ", func() {
			aggInfos := map[int]interfaces.AggInfo{
				1: {
					AggName: "cpu2",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				2: {
					AggName: "mode",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				3: {
					AggName:     "time2",
					AggType:     interfaces.BUCKET_TYPE_DATE_HISTOGRAM,
					IsDateField: true,
				},
				4: {
					AggName:       "value",
					AggType:       interfaces.AGGR_TYPE_TOP_HITS,
					IncludeFields: []string{"labels.job"},
				},
			}
			dslInfo := interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: defaultqueryZone}}
			res, err := parseDSLResult2Uniresponse(testENCtx, []byte(DslResultWithTopHits), dslInfo,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						IsInstantQuery: false,
						Start:          &dslStart2,
						End:            &dslEnd2,
						StepStr:        &step_50m,
						Step:           &step_3000000,
					},
					MeasureField: "metrics.node_cpu_seconds_total",
				}, dslMetric)

			So(len(res.Datas), ShouldEqual, 2)
			So(len(res.Datas[0].Labels), ShouldEqual, 3)
			So(len(res.Datas[0].Times), ShouldEqual, 2)
			So(err, ShouldBeNil)

		})

		Convey("Test parseDSLResult2Uniresponse top_hits contian labels field instant query ", func() {
			aggInfos := map[int]interfaces.AggInfo{
				1: {
					AggName: "cpu2",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				2: {
					AggName: "mode",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				3: {
					AggName:     "time2",
					AggType:     interfaces.BUCKET_TYPE_DATE_HISTOGRAM,
					IsDateField: true,
				},
				4: {
					AggName:       "value",
					AggType:       interfaces.AGGR_TYPE_TOP_HITS,
					IncludeFields: []string{"labels.job"},
				},
			}
			starti := int64(1686117284856)
			dslInfo := interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: defaultqueryZone}}
			res, err := parseDSLResult2Uniresponse(testENCtx, []byte(DslResultWithTopHits4Instant), dslInfo,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						IsInstantQuery: true,
						// Start: ,
						Start: &starti,
						End:   &dslStart2,
					},
					MeasureField: "metrics.node_cpu_seconds_total",
				}, dslMetric)

			So(len(res.Datas), ShouldEqual, 2)
			So(len(res.Datas[0].Labels), ShouldEqual, 3)
			So(len(res.Datas[0].Times), ShouldEqual, 1)
			So(err, ShouldBeNil)

		})

		Convey("Test parseDSLResult2Uniresponse top_hits only contian measure field ", func() {
			aggInfos := map[int]interfaces.AggInfo{
				1: {
					AggName: "cpu2",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				2: {
					AggName: "mode",
					AggType: interfaces.BUCKET_TYPE_TERMS,
				},
				3: {
					AggName:     "time2",
					AggType:     interfaces.BUCKET_TYPE_DATE_HISTOGRAM,
					IsDateField: true,
				},
				4: {
					AggName: "value",
					AggType: interfaces.AGGR_TYPE_TOP_HITS,
				},
			}
			dslInfo := interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: defaultqueryZone}}
			res, err := parseDSLResult2Uniresponse(testENCtx, []byte(DslResultWithTopHits), dslInfo,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						IsInstantQuery: false,
						Start:          &dslStart2,
						End:            &dslEnd2,
						StepStr:        &step_50m,
						Step:           &step_3000000,
					},
					MeasureField: "metrics.node_cpu_seconds_total",
				}, dslMetric)

			So(len(res.Datas), ShouldEqual, 2)
			So(len(res.Datas[0].Labels), ShouldEqual, 2)
			So(len(res.Datas[0].Times), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("Test parseDSLResult2Uniresponse filters sum ", func() {
			aggInfos := map[int]interfaces.AggInfo{
				1: {
					AggName: "NAME1",
					AggType: interfaces.BUCKET_TYPE_FILTERS,
				},
				2: {
					AggName:     "NAME2",
					AggType:     interfaces.BUCKET_TYPE_DATE_HISTOGRAM,
					IsDateField: true,
				},
				3: {
					AggName: "NAME3",
					AggType: interfaces.AGGR_TYPE_SUM,
				},
			}
			dslInfo := interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: defaultqueryZone}}
			res, err := parseDSLResult2Uniresponse(testENCtx, []byte(DslResultWithFiltersSum), dslInfo,
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						IsInstantQuery: false,
						Start:          &dslStart2,
						End:            &dslEnd2,
						StepStr:        &step_50m,
						Step:           &step_3000000,
					},
					MeasureField: "NAME3",
				}, dslMetric)

			So(len(res.Datas), ShouldEqual, 2)
			So(len(res.Datas[0].Labels), ShouldEqual, 1)
			So(len(res.Datas[0].Times), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("Test parseDSLResult2Uniresponse filters sum 4 instant query", func() {
			aggInfos := map[int]interfaces.AggInfo{
				1: {
					AggName: "NAME1",
					AggType: interfaces.BUCKET_TYPE_FILTERS,
				},
				2: {
					AggName:     "NAME2",
					AggType:     interfaces.BUCKET_TYPE_DATE_HISTOGRAM,
					IsDateField: true,
				},
				3: {
					AggName: "NAME3",
					AggType: interfaces.AGGR_TYPE_SUM,
				},
			}
			starti := int64(1686117284856)
			res, err := parseDSLResult2Uniresponse(testENCtx, []byte(DslResultWithFiltersSum4Instant), interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: defaultqueryZone}},
				interfaces.MetricModelQuery{
					QueryTimeParams: interfaces.QueryTimeParams{
						IsInstantQuery: true,
						Start:          &starti,
						End:            &dslStart2,
					},
					MeasureField: "NAME3",
				}, dslMetric)

			So(len(res.Datas), ShouldEqual, 2)
			So(len(res.Datas[0].Labels), ShouldEqual, 1)
			So(len(res.Datas[0].Times), ShouldEqual, 1)
			So(err, ShouldBeNil)

		})

		// 	Convey("Test parseDSLResult2Uniresponse failed, because timezone is invalid", func() {
		// 		aggInfos := map[int]interfaces.AggInfo{
		// 			1: {
		// 				AggName: "NAME1",
		// 				AggType: interfaces.FILTERS,
		// 			},
		// 			2: {
		// 				AggName:     "NAME2",
		// 				AggType:     interfaces.DATE_HISTOGRAM,
		// 				IsDateField: true,
		// 			},
		// 			3: {
		// 				AggName: "NAME3",
		// 				AggType: interfaces.SUM,
		// 			},
		// 		}
		// 		expectedErr := &rest.HTTPError{
		// 			HTTPCode: http.StatusInternalServerError,
		// 			Language: "en-US",
		// 			BaseError: rest.BaseError{
		// 				ErrorCode:    uerrors.Uniquery_MetricModel_InvaliEnvironmentVariable_TZ,
		// 				Description:  "Invali Environment Variable Timezone",
		// 				Solution:     "Please check whether the environment variable is correct.",
		// 				ErrorLink:    "None",
		// 				ErrorDetails: "unknown time zone 123",
		// 			},
		// 		}

		// 		zone123, _ := time.LoadLocation("123")
		// 		res, err := parseDSLResult2Uniresponse(testENCtx,[]byte(DslResultWithFiltersSum4Instant), interfaces.DslInfo{AggInfos: aggInfos, DateHistogram: interfaces.AggInfo{ZoneLocation: zone123}}, interfaces.MetricModel{},
		// 			interfaces.MetricModelQuery{
		// 				IsInstantQuery: true,
		// 				Time:           1686117584856,
		// 				MeasureField:   "NAME3",
		// 				// TimeZone:       "123",
		// 			})

		// 		So(len(res.Datas), ShouldEqual, 0)
		// 		So(err, ShouldResemble, expectedErr)

		// 	})
	})
}

func TestCorrectingTime(t *testing.T) {
	Convey("Test correctingTime", t, func() {

		Convey("Test correctingTime ,isCalendar is true and step is 1m  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1min,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692512520000)
			So(fixEnd, ShouldEqual, 1695190920000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is minute  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1min,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692512520000)
			So(fixEnd, ShouldEqual, 1695190920000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1h  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1hour,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692511200000)
			So(fixEnd, ShouldEqual, 1695189600000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is hour ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &hour,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692511200000)
			So(fixEnd, ShouldEqual, 1695189600000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is day ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &day,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692460800000)
			So(fixEnd, ShouldEqual, 1695139200000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1d ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1day,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692460800000)
			So(fixEnd, ShouldEqual, 1695139200000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1w ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1w,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1691942400000)
			So(fixEnd, ShouldEqual, 1694966400000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is week ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &setp_week,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1691942400000)
			So(fixEnd, ShouldEqual, 1694966400000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1M ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1M,
				},
				// //TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1690819200000)
			So(fixEnd, ShouldEqual, 1693497600000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is month ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_month,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1690819200000)
			So(fixEnd, ShouldEqual, 1693497600000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1q ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1q,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1688140800000)
			So(fixEnd, ShouldEqual, 1688140800000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is quarter ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_quarter,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1688140800000)
			So(fixEnd, ShouldEqual, 1688140800000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is 1y ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1y,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1672502400000)
			So(fixEnd, ShouldEqual, 1672502400000)
		})

		Convey("Test correctingTime ,isCalendar is true and step is year ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: true,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_year,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1672502400000)
			So(fixEnd, ShouldEqual, 1672502400000)
		})

		Convey("Test correctingTime ,isCalendar is false ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar: false,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart3,
					End:     &dslEnd3,
					StepStr: &step_1day,
				},
				//TimeZone:   "Asia/Shanghai",
			}
			fixedStart, fixEnd := correctingTime(query, interfaces.DEFAULT_QUERY_TIME_ZONE)
			So(fixedStart, ShouldEqual, 1692460800000)
			So(fixEnd, ShouldEqual, 1695139200000)
		})
	})
}

func TestGetNextPointTime(t *testing.T) {

	Convey("Test correctingTime", t, func() {

		var currentTime int64 = 1692512520000

		Convey("Test getNextPointTime ,isCalendar is true and step is 1m  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1min},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692512580000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is minute  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &minute},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692512580000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1h  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1hour},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692516120000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is hour  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &hour},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692516120000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1d  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1day},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692598920000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is day  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &day},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692598920000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1w  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1w},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1693117320000)
		})

		Convey("Test getNextPointTime ,isCalendar is true and step is week  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &setp_week},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1693117320000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1M  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1M},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1695190920000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is month  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_month},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1695190920000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1q  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1q},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1700461320000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is quarter  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_quarter},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1700461320000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is 1y  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_1y},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1724134920000)
		})
		Convey("Test getNextPointTime ,isCalendar is true and step is year  ", func() {
			query := interfaces.MetricModelQuery{
				IsCalendar:      true,
				QueryTimeParams: interfaces.QueryTimeParams{StepStr: &step_year},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1724134920000)
		})
		Convey("Test getNextPointTime ,isCalendar is false ", func() {
			stepi := int64(86400000)
			query := interfaces.MetricModelQuery{
				IsCalendar: false,
				QueryTimeParams: interfaces.QueryTimeParams{
					StepStr: &step_1day,
					Step:    &stepi,
				},
			}
			nextPointTime := getNextPointTime(query, currentTime)
			So(nextPointTime, ShouldEqual, 1692598920000)
		})
	})
}

func TestProcessDateHistogram(t *testing.T) {
	Convey("Test processDateHistogram", t, func() {
		common.FixedStepsMap = StepsMap

		Convey("processDateHistogram success with instant query && fixed variable ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram failed with instant query && fixed constant error ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "2a",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram failed with instant query && fixed constant error -2s ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "0",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar variable ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant minute ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "minute",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant hour ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "hour",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant day ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "day",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant week ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "week",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant month ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "month",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant quarter ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "quarter",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram success with instant query && calendar constant year ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "year",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram error with instant query && calendar constant yeara ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "yeara",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram error with instant query && invalid interval type ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendara_interval": "year",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = true
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram success with range query && fixed variable ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &step_30m,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram error with range query && fixed variable unsupport step ", func() {
			stepi := "31m"
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &stepi,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram failed with range query && fixed constant exceeded 11000 ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &step_30m,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "1ms",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram failed with range query && fixed constant parse interval error ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &step_30m,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "fixed_interval": "1a",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram success with range query && calendar variable ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &minute,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

		Convey("processDateHistogram failed with range query && calendar variable && unsupport step", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start:   &dslStart4,
					End:     &dslEnd4,
					Step:    &step_1800000,
					StepStr: &step_30m,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "{{__interval}}",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldNotBeNil)
		})

		Convey("processDateHistogram success with range query && calendar constant ", func() {
			query1 := interfaces.MetricModelQuery{
				MetricType: interfaces.ATOMIC_METRIC,
				QueryType:  interfaces.DSL,
				QueryTimeParams: interfaces.QueryTimeParams{
					Start: &dslStart4,
					End:   &dslEnd4,
					Step:  &step_1800000,
				},
				QueryTimeNum: 22,
				Formula: `{
					"size": 0,
					"query": {
					  "bool": {
						"filter": [
						  {
							"exists": {
							  "field": "metrics.node_cpu_seconds_total"
							}
						  }
						],
						"must": []
					  }
					},
					"aggs": {
					  "job": {
						"terms": {
						  "field": "labels.job",
						  "size": 1,
						  "order": {
							"_key": "asc"
						  }
						},
						"aggs": {
						  "mode": {
							"terms": {
							  "field": "labels.mode",
							  "size": 3,
							  "order": {
								"_key": "desc"
							  }
							},
							"aggs": {
							  "instance": {
								"terms": {
								  "field": "labels.instance",
								  "size": 100,
								  "order": {
									"_key": "desc"
								  }
								},
								"aggs": {
								  "cpu": {
									"terms": {
									  "field": "labels.cpu",
									  "size": 25,
									  "order": {
										"_key": "desc"
									  }
									},
									"aggs": {
									  "time": {
										"date_histogram": {
										  "field": "@timestamp",
										  "calendar_interval": "year",
										  "min_doc_count": 1
										},
										"aggs": {
										  "value": {
											"avg": {
											  "field": "metrics.node_cpu_seconds_total"
											}
										  }
										}
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }`,
				MeasureField: "value"}
			query1.IsInstantQuery = false
			query1.IgnoringMemoryCache = true

			dslInfo1, _ := parseDsl(testENCtx, &query1, &dataview)

			_, errP := processDateHistogram(testENCtx, &query1, dslInfo1, 5*time.Minute)
			So(errP, ShouldBeNil)
		})

	})
}
