// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

var (
	seriesMatcher = &interfaces.Matchers{
		Start:      1655346155000,
		End:        1655350445000,
		LogGroupId: "a",
	}

	emptySeriesResult = []labels.Labels{}

	moreThanOneMatchSeries = []labels.Labels{

		{
			{Name: "cpu", Value: "0"},
			{Name: "instance", Value: "instance_abc"},
			{Name: "job", Value: "prometheus"},
			{Name: "mode", Value: "nice"},
		},
		{
			{Name: "cpu", Value: "1"},
			{Name: "instance", Value: "instance_abc"},
			{Name: "job", Value: "prometheus"},
			{Name: "mode", Value: "nice"},
		},
		{
			{Name: "device", Value: "/dev/mapper/centos-root"},
			{Name: "fstype", Value: "xfs"},
			{Name: "instance", Value: "instance_abc"},
			{Name: "job", Value: "prometheus"},
			{Name: "mountpoint", Value: "/"},
		},
		{
			{Name: "device", Value: "/dev/sda1"},
			{Name: "fstype", Value: "xfs"},
			{Name: "instance", Value: "instance_abc"},
			{Name: "job", Value: "prometheus"},
			{Name: "mountpoint", Value: "/boot"},
		},
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

	mustFilter = []interface{}{
		map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
				"analyze_wildcard": true,
			},
		},
	}
)

func TestSeries(t *testing.T) {
	Convey("test series Series", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		Convey("matchers.MatcherSet is empty ", func() {
			seriesMatcherE := &interfaces.Matchers{
				MatcherSet: [][]*labels.Matcher{},
				Start:      1655346155000,
				End:        1655350445000,
			}

			res, status, err := lnMock.Series(seriesMatcherE)

			So(res, ShouldResemble, []labels.Labels{})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ExecutePool.Submit failed", func() {
			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)

			util.ExecutePool.Release()
			defer util.MegerPool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldResemble, `ExecutePool.Submit error: this pool has been closed`)
		})

		Convey("success with empty series ", func() {

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldResemble, emptySeriesResult)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("success with more than one match[]  ", func() {

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeCpuGuestSeriesDslResult), http.StatusOK, nil)

			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(convert.MapToByte(NodeFilesystemSeriesDslResult), http.StatusOK, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`,
				`prometheus.metrics.node_filesystem_avail_bytes`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldResemble, moreThanOneMatchSeries)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("error with opensearch failed ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).Times(1).Return(nil, http.StatusInternalServerError,
				uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason("Error getting response from opensearch"))

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":500,"error":{"type":"UniQuery.InternalServerError","reason":"Error getting response from opensearch"}}`)
		})

		Convey("error with GetLogGroupQueryFilters failed ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `get request method failed`)
		})

		Convey("success with empty index pattern ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, true, nil)

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := lnMock.Series(seriesMatcher)
			So(res, ShouldResemble, emptySeriesResult)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestSeriesDSL(t *testing.T) {
	Convey("test seriesDSL ", t, func() {

		Convey("seriesDSL correctly ", func() {
			// expectDSL := fmt.Sprintf(`
			// {
			// 	"size":0,
			// 	"query": {
			// 		"bool": {
			// 			"filter": [
			// 				{
			// 					"term": {
			// 						"labels.mode.keyword": "nice"
			// 					}
			// 				 },
			// 				{
			// 					"exists": {
			// 						"field": "metrics.node_cpu_guest_seconds_total"
			// 					}
			// 				},
			// 				{
			// 					"range": {
			// 						"@timestamp": {
			// 							"gte":1655346155000,
			// 							"lte":1655350445000
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
			// 	},
			// 	"aggs": {
			// 		"%s": {
			// 			"terms": {
			// 				"field": "%s",
			// 				"size": 10000
			// 			}
			// 		}
			// 	}
			// }`, interfaces.LABELS_STR, wrapKeyWordFieldName(interfaces.LABELS_STR))

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			_, status, err := seriesDSL(seriesMatcher.MatcherSet[0], seriesMatcher, maxSearchSeriesSize, mustFilter)
			//So(replace(res.String()), ShouldEqual, replace(expectDSL))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Marshal must_filter error", func() {
			patch := ApplyFunc(sonic.Marshal,
				func(v any) ([]byte, error) {
					return nil, fmt.Errorf("a")
				},
			)
			defer patch.Reset()

			matcherSets, err := static.ParseMatchersParam([]string{`node_cpu_guest_seconds_total{mode="nice"}`})
			So(err, ShouldBeNil)
			seriesMatcher.MatcherSet = matcherSets

			res, status, err := seriesDSL(seriesMatcher.MatcherSet[0], seriesMatcher, maxSearchSeriesSize, mustFilter)

			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)
		})
	})
}
