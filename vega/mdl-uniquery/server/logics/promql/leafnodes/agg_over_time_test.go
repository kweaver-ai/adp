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
func TestAggOverTime(t *testing.T) {
	Convey("test AggOverTime overall process", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		expr := &parser.MatrixSelector{VectorSelector: &parser.VectorSelector{Name: "a", LabelMatchers: []*labels.Matcher{
			{Type: labels.MatchEqual, Name: "__name__", Value: "a"},
		}}}
		query := &interfaces.Query{}
		query.LogGroupId = "a"
		commonInterval := (15 * time.Second).Milliseconds()
		commonRange := 1 * time.Minute

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

		Convey("1. range is not an integer multiple of interval, substep=gcd(range, step)", func() {
			query.Interval = (10 * time.Second).Milliseconds()
			expr.Range = 15 * time.Second

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).Return(shards, 200, nil).AnyTimes()
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(bytesResult, 0, nil)

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)

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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
			So(val, ShouldBeNil)
			So(status, ShouldEqual, http.StatusUnprocessableEntity)
			So(err, ShouldNotBeNil)

		})

		Convey("3. failed to get query filters", func() {
			query.Interval = commonInterval
			expr.Range = commonRange

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
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
			patches := ApplyFunc(aggOverTimeMerge, func(_ MapResult, _ *parser.MatrixSelector,
				_ *interfaces.Query, _ string) (static.Matrix, error) {
				return nil, errors.New("merge error")
			})
			defer patches.Reset()

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
			So(query.Interval, ShouldEqual, int64(1500000))
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(query.Start, ShouldEqual, query.Start)
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

			val, status, err := lnMock.AggOverTime(testCtx, expr, query, interfaces.AVG_OVER_TIME)
			So(query.SubIntervalWith30min, ShouldEqual, query.Interval)
			So(val, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}

func TestAggOverTimeMerge(t *testing.T) {
	Convey("test aggOverTime merge", t, func() {

		expr := &parser.MatrixSelector{
			VectorSelector: &parser.VectorSelector{},
			Range:          15 * time.Second,
		}
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, FixedStart: 1646360670000, FixedEnd: 1646360685000}

		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 2,
							}
							}`,
				},
			},

			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 3,
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

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.AVG_OVER_TIME)
			So(mat, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("2. merge task successfully with avg_over_time", func() {
			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			expectAvg := []static.Series{
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
							V: 1.25,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectAvg...)

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.AVG_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)

		})

		Convey("3. merge task successfully with sum_over_time", func() {
			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			expectAvg := []static.Series{
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
							V: 5,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectAvg...)

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.SUM_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("4. merge task successfully with max_over_time", func() {
			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			expectAvg := []static.Series{
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
							V: 3,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectAvg...)

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.MAX_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("5. merge task successfully with min_over_time", func() {
			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			expectAvg := []static.Series{
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
							V: 2,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectAvg...)

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.MIN_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("6. merge task successfully with count_over_time", func() {
			util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			expectAvg := []static.Series{
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
							V: 5,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectAvg...)

			mat, err := aggOverTimeMerge(mapResult, expr, query, interfaces.COUNT_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestAggOverTimeMerge4InstantQuery(t *testing.T) {
	Convey("test aggOverTimeMerge4InstantQuery ", t, func() {
		query := &interfaces.Query{Interval: 15000, SubIntervalWith30min: 15000, IsInstantQuery: true}

		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 2,
							}
							}`,
				},
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360685000,"doc_count":2,
							"value":{
								"value": 3,
							}
							}`,
				},
			},

			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 7,
							}
							}`,
				},
			},
		}

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

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResultTmp, query, interfaces.AVG_OVER_TIME)
			So(mat, ShouldResemble, static.Matrix{})
			So(err, ShouldBeNil)
		})

		Convey("2. tsArr contain more then one segment wiht avg_over_time", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 2,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResult, query, interfaces.AVG_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("3. tsArr contain more then one segment wiht sum_over_time", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResult, query, interfaces.SUM_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("4. tsArr contain more then one segment wiht max_over_time", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 7,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResult, query, interfaces.MAX_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("5. tsArr contain more then one segment wiht min_over_time", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 2,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResult, query, interfaces.MIN_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("6. tsArr contain more then one segment wiht count_over_time", func() {
			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 0,
							V: 12,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{{
				Name:  interfaces.LABELS_STR,
				Value: key,
			}}

			mapResult := MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}

			mat, err := aggOverTimeMerge4InstantQuery([]string{key}, mapResult, query, interfaces.COUNT_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}

func TestAggOverTimeMerge4RangeQuery(t *testing.T) {
	Convey("test aggOverTimeMerge4RangeQuery ", t, func() {
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

		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 3,
							}
							}`,
				},
			},

			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
							"value":{
								"value": 2,
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

			mat, err := aggOverTimeMerge4RangeQuery([]string{key}, mapResult, query, expr, interfaces.AVG_OVER_TIME)
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
							V: 1.25,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := aggOverTimeMerge4RangeQuery([]string{key}, mapResult, query, expr, interfaces.AVG_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("3. tsArr contain more then one segment && search with subintervel", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
								"value":{
									"value": 3,
								}
								}`,
					},
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360685000,"doc_count":3,
								"value":{
									"value": 4,
								}
								}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
								"value":{
									"value": 2,
								}
								}`,
					},
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360685000,"doc_count":3,
								"value":{
									"value": 3,
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
			expr.Range = 30 * time.Second

			expectSeries := []static.Series{
				{
					Metric: []*labels.Label{
						{Name: "cluster", Value: "txy"},
						{Name: "name", Value: "node-1"},
					},
					Points: []static.Point{
						{
							T: 1646360670000,
							V: 1.2,
						},
						{
							T: 1646360685000,
							V: 1.1666666666666667,
						},
					},
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := aggOverTimeMerge4RangeQuery([]string{key}, mapResult, query, expr, interfaces.AVG_OVER_TIME)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

	})
}

// ok
func TestAggOverTimeMergePointWithSameKey(t *testing.T) {
	Convey("test agg_over_time merge bucket with same keyTS", t, func() {
		tsValueMapK := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
					"value":{"value" : 2.0}}`,
				},
			},
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
					"value":{"value" : 3.0}}`,
				},
			},
			{
				{
					Type: gjson.JSON,
					Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
					"value":{"value" : 4.0}}`,
				},
			},
		}
		Convey("1. merge with avg_over_time", func() {

			expectPointMap := make(map[int64]*static.AGGPoint, 0)
			expectPointMap[1646360670000] = &static.AGGPoint{
				Count: 6,
				Value: 9,
			}

			pointMap := make(map[int64]*static.AGGPoint, 0)
			aggOverTimeMergePointWithSameKey(interfaces.AVG_OVER_TIME, pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("2. merge with sum_over_time", func() {

			expectPointMap := make(map[int64]*static.AGGPoint, 0)
			expectPointMap[1646360670000] = &static.AGGPoint{
				Count: 6,
				Value: 9,
			}

			pointMap := make(map[int64]*static.AGGPoint, 0)
			aggOverTimeMergePointWithSameKey(interfaces.SUM_OVER_TIME, pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("3. merge with max_over_time", func() {

			expectPointMap := make(map[int64]*static.AGGPoint, 0)
			expectPointMap[1646360670000] = &static.AGGPoint{
				Count: 6,
				Value: 4,
			}

			pointMap := make(map[int64]*static.AGGPoint, 0)
			aggOverTimeMergePointWithSameKey(interfaces.MAX_OVER_TIME, pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("4. merge with min_over_time", func() {

			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 2.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 3.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 1.0}}`,
					},
				},
			}
			expectPointMap := make(map[int64]*static.AGGPoint, 0)
			expectPointMap[1646360670000] = &static.AGGPoint{
				Count: 6,
				Value: 1,
			}

			pointMap := make(map[int64]*static.AGGPoint, 0)
			aggOverTimeMergePointWithSameKey(interfaces.MIN_OVER_TIME, pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

		Convey("5. merge with count_over_time", func() {

			tsValueMapK := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 2.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 2.0}}`,
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw: `{"key_as_string":"2022-03-04T02:24:30.000Z","key":1646360670000,"doc_count":2,
						"value":{"value" : 2.0}}`,
					},
				},
			}

			expectPointMap := make(map[int64]*static.AGGPoint, 0)
			expectPointMap[1646360670000] = &static.AGGPoint{
				Count: 6,
				Value: 6,
			}

			pointMap := make(map[int64]*static.AGGPoint, 0)
			aggOverTimeMergePointWithSameKey(interfaces.COUNT_OVER_TIME, pointMap, tsValueMapK)

			So(pointMap, ShouldResemble, expectPointMap)
		})

	})
}
