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

// ok
func TestDeltaAgg(t *testing.T) {
	Convey("test DeltaAggs overall process", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		expr := &parser.MatrixSelector{VectorSelector: &parser.VectorSelector{}}
		query := &interfaces.Query{
			Start:      1652320533000,
			End:        1652320557000,
			FixedStart: 1652320533000,
			FixedEnd:   1652320557000,
			Interval:   3000,
		}
		query.LogGroupId = "a"
		commonInterval := (15 * time.Second).Milliseconds()
		commonRange := 1 * time.Minute
		call := static.FunctionCalls["delta"]

		indexShardsArr := []interfaces.IndexShards{
			{
				IndexName: "metricbeat-*",
				Pri:       "2",
			},
		}
		shards, _ := sonic.Marshal(indexShardsArr)
		bytesResult := convert.MapToByte(NotEmptyDslResult)

		Convey("1. range is not an integer multiple of interval, substep=gcd(range, step)", func() {
			query.Interval = (10 * time.Second).Milliseconds()
			expr.Range = 15 * time.Second

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)

			So(query.SubIntervalWith30min, ShouldEqual, (5 * time.Second).Milliseconds())
			So(expr.Range, ShouldEqual, 15*time.Second)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("2. type assertion failed", func() {
			expr.VectorSelector = &parser.MatrixSelector{}
			query.Interval = commonInterval
			expr.Range = commonRange

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("3. failed to get query filters", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("4. return static.Matrix{}: getIndicesNumberOfShards indexShardsArr == 0", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			query.Interval = commonInterval
			expr.Range = commonRange
			emptyIndexShardsArr, _ := sonic.Marshal([]interfaces.IndexShards{})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(emptyIndexShardsArr, 200, nil)

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(val, ShouldResemble, static.PageMatrix{Matrix: static.Matrix{}})
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("5. failed to merge", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			// patch submit 会在execute dsl 时就返回错误
			patches := ApplyFunc(deltaMerge, func(_ MapResult, _ *parser.MatrixSelector,
				_ *interfaces.Query, _ static.FunctionCall) (static.Matrix, error) {
				return nil, errors.New("merge error")
			})
			defer patches.Reset()

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("6. is instant query ", func() {
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

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(query.Interval, ShouldEqual, stepRangeMs)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(query.Start, ShouldEqual, 10e12)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("7. is range query", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			// 缓存里已有的话, 就不会去os里查了
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("8. is instant query && auto is true", func() {
			query.IsInstantQuery = true
			query.Interval = commonInterval
			expr.Range = 0
			expr.Auto = true
			query.Start = 1652319900000
			query.End = 1652321400000

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(query.Interval, ShouldEqual, int64(1500000))
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(query.Start, ShouldEqual, 1652319900000)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("9. is range query && auto is true", func() {
			query.Interval = commonInterval
			expr.Range = 0
			expr.Auto = true

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			// 缓存里已有的话, 就不会去os里查了
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.DeltaAggs(testCtx, expr, query, call)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestDeltaMerge(t *testing.T) {
	Convey("test delta merge", t, func() {
		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, FixedStart: 1646360670000, FixedEnd: 1646360685000, End: 1646360670000}
		call := static.FunctionCalls["delta"]

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
						V: 8.00053336889126,
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

			mat, err := deltaMerge(mapResult, expr, query, call)
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

			mat, err := deltaMerge(mapResult, expr, query, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestDeltaMerge4InstantQuery(t *testing.T) {
	Convey("test delta instant query merge", t, func() {
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, IsInstantQuery: true}
		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		call := static.FunctionCalls["delta"]

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

			mat, err := deltaMerge4InstantQuery([]string{key}, mapResultTmp, query, expr, call)
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
							V: 2.4842105263157896,
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

			mat, err := deltaMerge4InstantQuery([]string{key}, mapResult, query, expr, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestDeltaMerge4RangeQuery(t *testing.T) {
	Convey("test delta range query merge ", t, func() {
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
		call := static.FunctionCalls["delta"]

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
			mat, err := deltaMerge4RangeQuery([]string{key}, mapResult, query, expr, call)
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
							V: 8.00053336889126,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			chs := make(chan map[string]static.Series, len(mapResult.TsValueMap))
			defer close(chs)
			mat, err := deltaMerge4RangeQuery([]string{key}, mapResult, query, expr, call)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

// ok
func TestDeltaMergePointWithSameKey(t *testing.T) {
	Convey("test delta merge bucket with same keyTS", t, func() {
		Convey("1. merge b into a, firstT取最小; lastT相等取lastV最大", func() {
			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count": 2,
						"value":{"firstValue": 8.0,"firstTimestamp":1646360671000,"lastValue":10.0,"lastTimestamp":1646360708000}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"firstValue":2.0,"firstTimestamp":1646360670000,"lastValue": 8.0,"lastTimestamp":1646360708000}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360670000,
				FirstValue:     2.0,
				LastTimestamp:  1646360708000,
				LastValue:      10.0,
				PointsCount:    4,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360669000,
				FirstValue:     2.0,
				LastTimestamp:  1646360708000,
				LastValue:      10.0,
				PointsCount:    3,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360669000,
				FirstValue:     2.0,
				LastTimestamp:  1646360708000,
				LastValue:      10.0,
				PointsCount:    8,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360671000,
				FirstValue:     8.0,
				LastTimestamp:  1646360708000,
				LastValue:      12.0,
				PointsCount:    4,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360671000,
				FirstValue:     8.0,
				LastTimestamp:  1646360709000,
				LastValue:      12.0,
				PointsCount:    92,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360671000,
				FirstValue:     9.0,
				LastTimestamp:  1646360708000,
				LastValue:      8.0,
				PointsCount:    2,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360670000,
				FirstValue:     2.0,
				LastTimestamp:  1646360708000,
				LastValue:      8.0,
				PointsCount:    4,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

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

			expectPointMap := make(map[int64]*static.DeltaPoint, 0)
			expectPointMap[1646360670000] = &static.DeltaPoint{
				FirstTimestamp: 1646360671000,
				FirstValue:     8.0,
				LastTimestamp:  1646360708000,
				LastValue:      8.0,
				PointsCount:    3,
			}

			pointMap := make(map[int64]*static.DeltaPoint, 0)
			deltaMergePointWithSameKey(pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})
	})

}
