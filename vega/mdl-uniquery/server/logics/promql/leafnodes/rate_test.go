// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/panjf2000/ants/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	"uniquery/common"
	"uniquery/common/convert"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

func TestRateAggs(t *testing.T) {
	Convey("test rate overall process", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		expr := &parser.MatrixSelector{VectorSelector: &parser.VectorSelector{}}
		query := &interfaces.Query{}
		query.LogGroupId = "a"
		commonInterval := (15 * time.Second).Milliseconds()
		commonRange := 1 * time.Minute
		call := static.FunctionCalls["rate"]

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

		indexShardsArr := []interfaces.IndexShards{
			{
				IndexName: "metricbeat-*",
				Pri:       "2",
			},
		}
		shards, _ := sonic.Marshal(indexShardsArr)
		bytesResult := convert.MapToByte(NotEmptyDslResult)

		Convey("1. instant query && stepRange > 120m and is not divisible by 5m", func() {
			query.IsInstantQuery = true
			expr.Range = 121 * time.Minute

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)

			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("2. range is not an integer multiple of interval, substep=gcd(range, step)", func() {
			query.Interval = (10 * time.Second).Milliseconds()
			expr.Range = 15 * time.Second

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)

			So(query.SubIntervalWith30min, ShouldEqual, (5 * time.Second).Milliseconds())
			So(expr.Range, ShouldEqual, 15*time.Second)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("3. interval is greater than routing, need SubInterval", func() {
			query.Interval = (140 * time.Minute).Milliseconds()
			expr.Range = 10 * time.Minute

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)

			So(query.SubIntervalWith30min, ShouldEqual, (10 * time.Minute).Milliseconds())
			So(expr.Range, ShouldEqual, 10*time.Minute)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("4. type assertion failed", func() {
			expr.VectorSelector = &parser.MatrixSelector{}
			query.Interval = commonInterval
			expr.Range = commonRange

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("5. failed to get query filters", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("6. failed to get shards count", func() {
			// 缓存里已有的话, 就不会去os里查了,这里删掉缓存
			Number_Of_Shards_Map.Delete(indexPattern)
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return([]byte{}, http.StatusInternalServerError, errors.New("os error"))

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("7. failed to execute dsl", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return([]byte{}, 0, errors.New("search submit error"))

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("8. return static.Matrix{}: getIndicesNumberOfShards indexShardsArr == 0", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			query.Interval = commonInterval
			expr.Range = commonRange
			emptyIndexShardsArr, _ := sonic.Marshal([]interfaces.IndexShards{})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(emptyIndexShardsArr, 200, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldResemble, static.PageMatrix{Matrix: static.Matrix{}})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("9. failed to merge", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			// patch submit 会在execute dsl 时就返回错误
			patches := ApplyFunc(rateMerge, func(_ MapResult, _ *parser.MatrixSelector,
				_ *interfaces.Query, _ static.FunctionCall) (static.Matrix, error) {
				return nil, errors.New("merge error")
			})
			defer patches.Reset()

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("10. is instant query && stepRange is <= 120m ", func() {
			query.IsInstantQuery = true
			query.Interval = commonInterval
			expr.Range = commonRange
			query.Start = 10e12
			stepRangeMs := commonRange.Milliseconds()

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(query.Interval, ShouldEqual, stepRangeMs)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(query.Start, ShouldEqual, 10e12)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("11. is instant query && stepRange is > 120m and is divisible by 5m", func() {
			query.IsInstantQuery = true
			expr.Range = 125 * time.Minute
			stepRangeMs := expr.Range.Milliseconds()
			query.Start = 1234567891000

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(query.Interval, ShouldEqual, stepRangeMs)
			So(query.Start, ShouldEqual, 1234567891000)
			So(query.SubIntervalWith30min, ShouldEqual, (25 * time.Minute).Milliseconds())
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("12. is range query", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			// 缓存里已有的话, 就不会去os里查了
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.RateAggs(testCtx, expr, query, call)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestCalcSubInterval(t *testing.T) {
	Convey("test calculate subInterval", t, func() {
		Convey("1. loop 30 times", func() {
			interval, _ := time.ParseDuration("11113948m300s900000ms60000000000ns")
			stepRange := 5 * time.Minute
			subInterval := calcSubInterval(interval, stepRange, interfaces.SHARD_ROUTING_2H)
			So(subInterval, ShouldEqual, 1*time.Minute)
		})

		Convey("2. loop 30 times", func() {
			interval := 59 * time.Minute
			stepRange := 5 * time.Minute
			subInterval := calcSubInterval(interval, stepRange, interfaces.SHARD_ROUTING_2H)
			So(subInterval, ShouldEqual, 1*time.Minute)
		})

		Convey("3. loop 26 times", func() {
			interval := 160 * time.Minute
			stepRange := 5 * time.Minute
			subInterval := calcSubInterval(interval, stepRange, interfaces.SHARD_ROUTING_2H)
			So(subInterval, ShouldEqual, 5*time.Minute)
		})

		Convey("4. loop 1 times", func() {
			interval := 60 * time.Minute
			stepRange := 30 * time.Minute
			subInterval := calcSubInterval(interval, stepRange, interfaces.SHARD_ROUTING_2H)
			So(subInterval, ShouldEqual, 30*time.Minute)
		})

		Convey("5. loop 2 times", func() {
			interval := 158 * time.Minute
			stepRange := 158 * time.Minute
			subInterval := calcSubInterval(interval, stepRange, interfaces.SHARD_ROUTING_2H)
			So(subInterval, ShouldEqual, 79*time.Minute)
		})

	})
}

func TestRateMerge(t *testing.T) {
	Convey("test rate merge", t, func() {
		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, FixedStart: 1646360670000, FixedEnd: 1646360685000, End: 1646360670000}
		call := static.FunctionCalls["rate"]

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
						V: 0.5333688912594173,
					},
				},
			},
		}

		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"firstValue":8.0,
								"firstTimestamp":1646360671000,
								"lastValue":10.0,
								"lastTimestamp":1646360684999,
								"counterCorrection": 0.0,
							}
							}`,
				},
			},

			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"firstValue":2.0,
								"firstTimestamp":1646360670000,
								"lastValue":8.0,
								"lastTimestamp":1646360684998,
								"counterCorrection": 0.0,
							}
							}`,
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
		mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

		Convey("1. instant query", func() {
			query.IsInstantQuery = true

			mat, err := rateMerge(mapResult, expr, query, call)
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("2. range query", func() {
			query.IsInstantQuery = false

			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := rateMerge(mapResult, expr, query, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestInstantQueryMerge(t *testing.T) {
	Convey("test rate instant query merge", t, func() {
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, IsInstantQuery: true}
		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		call := static.FunctionCalls["rate"]

		Convey("1. tsArr is empty", func() {

			resultArra := [][]gjson.Result{}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResultTmp := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := instantQueryMerge([]string{key}, mapResultTmp, query, expr, call)
			So(mat, ShouldResemble, static.Matrix{})
			So(err, ShouldBeNil)
		})

		Convey("2. tsArr contain more then one segment", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 0.47614035087719303,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
								"value":{
									"firstValue":8.0,
									"firstTimestamp":1646360671000,
									"lastValue":10.0,
									"lastTimestamp":1646360684999,
									"counterCorrection": 3.0,
								}
								}`,
					},
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360685000,"doc_count":2,
								"value":{
									"firstValue":8.0,
									"firstTimestamp":1646360685000,
									"lastValue":10.0,
									"lastTimestamp":1646360689000,
									"counterCorrection": 2.0,
								}
								}`,
					},
				},

				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
								"value":{
									"firstValue":2.0,
									"firstTimestamp":1646360670000,
									"lastValue":8.0,
									"lastTimestamp":1646360684998,
									"counterCorrection": 2.0,
								}
								}`,
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

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := instantQueryMerge([]string{key}, mapResult, query, expr, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestRangeQueryMerge(t *testing.T) {
	Convey("test rate range query merge ", t, func() {
		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
		})

		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, FixedStart: 1646360670000, FixedEnd: 1646360685000}
		call := static.FunctionCalls["rate"]

		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"firstValue":8.0,
								"firstTimestamp":1646360671000,
								"lastValue":10.0,
								"lastTimestamp":1646360684999,
								"counterCorrection": 3.0,
							}
							}`,
				},
			},

			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"firstValue":2.0,
								"firstTimestamp":1646360670000,
								"lastValue":8.0,
								"lastTimestamp":1646360684998,
								"counterCorrection": 2.0,
							}
							}`,
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

		mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

		Convey("1. failed to submit task", func() {
			patches := ApplyMethodReturn(&ants.Pool{}, "Submit", errors.New("submit error"))
			defer patches.Reset()

			chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
			defer close(chs)
			mat, err := rangeQueryMerge(chs, []string{key}, mapResult, query, expr, call)
			So(mat, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("2. tsArr contain more then one segment", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 0.7333822254816987,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
			defer close(chs)
			mat, err := rangeQueryMerge(chs, []string{key}, mapResult, query, expr, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergePointWithSameKeyTS(t *testing.T) {
	Convey("test merge bucket with same keyTS", t, func() {
		Convey("1. merge b into a, firstT取最小; lastT相等取lastV最大;两区间重叠,counterCorrection取最大", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 2,
						"value":{"firstValue": 8.0,"firstTimestamp":1646360671000,"lastValue":10.0,"lastTimestamp":1646360708000,"counterCorrection" : 0.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":2.0,"firstTimestamp":1646360670000,"lastValue": 8.0,"lastTimestamp":1646360708000,"counterCorrection" : 4.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360670000,
				FirstValue:        2.0,
				LastTimestamp:     1646360708000,
				LastValue:         10.0,
				CounterCorrection: 4.0,
				PointsCount:       4,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("2. merge b into a, firstT取最小; lastT取最大;两区间不重叠(b在前),counterCorrection取两者相加以及两段之间的修正量", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 1,
						"value":{"firstValue": 8.0,"firstTimestamp":1646360671000,"lastValue":10.0,"lastTimestamp":1646360708000,"counterCorrection" : 6.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":2.0,"firstTimestamp":1646360669000,"lastValue": 9.0,"lastTimestamp":1646360670000,"counterCorrection" : 0.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360669000,
				FirstValue:        2.0,
				LastTimestamp:     1646360708000,
				LastValue:         10.0,
				CounterCorrection: 15.0,
				PointsCount:       3,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("3. merge b into a, firstT取最小; lastT取最大;两区间不重叠(a在前),counterCorrection取两者相加以及两段之间的修正量", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 6,
						"value":{"firstValue": 2.0,"firstTimestamp":1646360669000,"lastValue":6.0,"lastTimestamp":1646360670000,"counterCorrection" : 0.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":8.0,"firstTimestamp":1646360671000,"lastValue": 10.0,"lastTimestamp":1646360708000,"counterCorrection" : 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360669000,
				FirstValue:        2.0,
				LastTimestamp:     1646360708000,
				LastValue:         10.0,
				CounterCorrection: 6.0,
				PointsCount:       8,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("4. merge b into a, firstT相等取值最大; lastT相等取lastV最大;两区间重叠,counterCorrection取最大", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 2,
						"value":{"firstValue": 2.0,"firstTimestamp":1646360671000,"lastValue":12.0,"lastTimestamp":1646360708000,"counterCorrection" : 4.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":8.0,"firstTimestamp":1646360671000,"lastValue": 10.0,"lastTimestamp":1646360708000,"counterCorrection" : 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360671000,
				FirstValue:        8.0,
				LastTimestamp:     1646360708000,
				LastValue:         12.0,
				CounterCorrection: 6.0,
				PointsCount:       4,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("5. merge b into a, firstT相等取值最大; lastT取最大;两区间重叠,counterCorrection取最大", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 2,
						"value":{"firstValue": 2.0,"firstTimestamp":1646360671000,"lastValue":12.0,"lastTimestamp":1646360709000,"counterCorrection" : 8.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":90,
						"value":{"firstValue":8.0,"firstTimestamp":1646360671000,"lastValue": 24.0,"lastTimestamp":1646360708000,"counterCorrection" : 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360671000,
				FirstValue:        8.0,
				LastTimestamp:     1646360709000,
				LastValue:         12.0,
				CounterCorrection: 8.0,
				PointsCount:       92,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("6. merge b into a, firstT取最小; lastT取最大;两区间各自只有一个点(a在前),counterCorrection取两者相加以及两段之间的修正量", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 1,
						"value":{"firstValue": 9.0,"firstTimestamp":1646360671000,"lastValue":9.0,"lastTimestamp":1646360671000,"counterCorrection" : 8.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":1,
						"value":{"firstValue":8.0,"firstTimestamp":1646360708000,"lastValue": 8.0,"lastTimestamp":1646360708000,"counterCorrection" : 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360671000,
				FirstValue:        9.0,
				LastTimestamp:     1646360708000,
				LastValue:         8.0,
				CounterCorrection: 23.0,
				PointsCount:       2,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("7. merge b into a, firstT取最小; lastT取最大;其中有一个区间只有一个点,点和区间不重叠(a在前),counterCorrection取两者相加以及两段之间的修正量", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 1,
						"value":{"firstValue": 2.0,"firstTimestamp":1646360670000,"lastValue":2.0,"lastTimestamp":1646360670000,"counterCorrection": 8.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":3,
						"value":{"firstValue":8.0,"firstTimestamp":1646360671000,"lastValue": 8.0,"lastTimestamp":1646360708000,"counterCorrection": 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360670000,
				FirstValue:        2.0,
				LastTimestamp:     1646360708000,
				LastValue:         8.0,
				CounterCorrection: 14.0,
				PointsCount:       4,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})
		Convey("8. merge b into a, firstT取最小; lastT取最大;其中有一个区间只有一个点,点和区间重叠,counterCorrection取最大的", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 1,
						"value":{"firstValue": 2.0,"firstTimestamp":1646360672000,"lastValue":2.0,"lastTimestamp":1646360672000,"counterCorrection": 8.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":8.0,"firstTimestamp":1646360671000,"lastValue": 8.0,"lastTimestamp":1646360708000,"counterCorrection": 6.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.RatePoint, 0)
			expectPointMap[1646360670000] = &static.RatePoint{
				FirstTimestamp:    1646360671000,
				FirstValue:        8.0,
				LastTimestamp:     1646360708000,
				LastValue:         8.0,
				CounterCorrection: 8.0,
				PointsCount:       3,
			}

			pointMap := make(map[int64]*static.RatePoint, 0)
			mergePointWithSameKeyTS(&pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})
	})

}
