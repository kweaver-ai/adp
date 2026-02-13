// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"testing"
	"time"

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
	irateMatResult = make(static.Matrix, 0)

	irateQuery = &interfaces.Query{
		Start:      1653368870,
		End:        1653549380,
		Interval:   300000,
		LogGroupId: "a",
		Limit:      -1,
	}

	irateQueryInstant = &interfaces.Query{
		Start:          1653369600,
		End:            1653369600,
		Interval:       1,
		IsInstantQuery: true,
		LogGroupId:     "a",
	}

	irateQueryFix = &interfaces.Query{
		Start:      1653368870,
		End:        1653369700,
		Interval:   60000,
		LogGroupId: "a",
	}
	selRange = time.Minute * 5

	irateAgg = interfaces.IRATE_AGG
)

func unwrapParenExpr(expr *parser.Expr) {
	for {
		if p, ok := (*expr).(*parser.ParenExpr); ok {
			*expr = p.Expr
		} else {
			break
		}
	}
}

func unwrapStepInvariantExpr(expr parser.Expr) parser.Expr {
	if p, ok := expr.(*parser.StepInvariantExpr); ok {
		return p.Expr
	}
	return expr
}

func TestLeafNodes_IrateEval(t *testing.T) {
	Convey("test LeafNodes IrateEval", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osaMock := umock.NewMockOpenSearchAccess(mockCtrl)
		lgaMock := umock.NewMockLogGroupAccess(mockCtrl)
		dvsMock := umock.NewMockDataViewService(mockCtrl)
		lnMock := mockLeafNodes(osaMock, lgaMock, dvsMock)

		var indexShardsArr = make([]interfaces.IndexShards, 0, 1)
		indexShardsArr = append(indexShardsArr, interfaces.IndexShards{
			IndexName: indexPattern,
			Pri:       "1",
		})
		var (
			matrixArgIndex int
		)
		notEmptyJson, _ := sonic.Marshal(indexShardsArr)

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

		Convey("GetLogGroupQueryFilters error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get request method failed"))
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)
				So(res, ShouldBeNil)
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("getIndicesNumberOfShards error ", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.InternalServerError).
					WithReason(indicesError))

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)
				So(res, ShouldBeNil)
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(err, ShouldNotBeNil)
			}

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

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)

				So(res, ShouldBeNil)
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(err, ShouldNotBeNil)
			}

		})

		Convey("Merge error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(NotEmptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)

			util.MegerPool.Release()
			defer util.ExecutePool.Release()
			defer util.InitAntsPool(common.PoolSetting{
				MegerPoolSize:       10,
				ExecutePoolSize:     10,
				BatchSubmitPoolSize: 10,
			})
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)

				So(res, ShouldBeNil)
				So(status, ShouldEqual, http.StatusUnprocessableEntity)
				So(err, ShouldNotBeNil)
			}

		})

		Convey("MakeDSL error ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, "irateAgg", irateQuery)

				So(res, ShouldBeNil)
				So(status, ShouldEqual, http.StatusUnprocessableEntity)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("irateEval success with query range ", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)

				mat, ok := res.(static.PageMatrix)

				So(ok, ShouldBeTrue)
				So(mat, ShouldResemble, static.PageMatrix{Matrix: irateMatResult})
				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
			}

		})

		Convey("irateEval success with instant query", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQueryInstant)

				mat, ok := res.(static.PageMatrix)

				So(ok, ShouldBeTrue)
				So(mat, ShouldResemble, static.PageMatrix{Matrix: irateMatResult})
				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
			}

		})

		Convey("return static.Matrix{}: getIndicesNumberOfShards indexShardsArr == 0", func() {
			Number_Of_Shards_Map.Delete(indexPattern)
			emptyIndexShardsArr, _ := sonic.Marshal([]interfaces.IndexShards{})

			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				emptyIndexShardsArr, http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[5m])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQueryInstant)

				So(res, ShouldResemble, static.PageMatrix{Matrix: irateMatResult})
				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
			}
		})

		Convey("8. is instant query && auto is true", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[auto])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQueryInstant)

				mat, ok := res.(static.PageMatrix)

				So(ok, ShouldBeTrue)
				So(mat, ShouldResemble, static.PageMatrix{Matrix: irateMatResult})
				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
			}

		})

		Convey("9. is range query && auto is true", func() {
			lgaMock.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				logGroupQueryFilters, true, nil)
			dvsMock.EXPECT().LoadIndexShards(gomock.Any(), gomock.Any()).AnyTimes().Return(
				notEmptyJson, http.StatusOK, nil)
			dvsMock.EXPECT().GetDataFromOpenSearchWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(convert.MapToByte(emptyDslResult), http.StatusOK, nil)

			query.QueryStr = `irate(opensearch_indices_indexing_index_total{cluster="txy",name!~"^node.*"}[auto])`
			expr, _ := parser.ParseExpr(testCtx, query.QueryStr)
			switch expression := expr.(type) {
			case *parser.Call:
				unwrapParenExpr(&expression.Args[matrixArgIndex])
				arg := unwrapStepInvariantExpr(expression.Args[matrixArgIndex])
				unwrapParenExpr(&arg)
				sel := arg.(*parser.MatrixSelector)
				res, status, err := lnMock.IrateEval(testCtx, sel, groupby, irateAgg, irateQuery)

				mat, ok := res.(static.PageMatrix)

				So(ok, ShouldBeTrue)
				So(mat, ShouldResemble, static.PageMatrix{Matrix: irateMatResult})
				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
			}
		})

	})
}

func TestLeafNodes_IrateMerge(t *testing.T) {
	Convey("test IrateMerge ", t, func() {
		util.InitAntsPool(common.PoolSetting{
			MegerPoolSize:       10,
			ExecutePoolSize:     10,
			BatchSubmitPoolSize: 10,
		})
		resultArra := [][]gjson.Result{
			{
				{
					Type: gjson.JSON,
					Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
				},
			},
			{
				{
					Type: gjson.JSON,
					Raw:  "{\"key_as_string\":\"2022-05-24T05:15:00.000Z\",\"key\":1653369300000,\"doc_count\":3,\"value\":{\"previousValue\":171856.0,\"previousTimestamp\":1653369321709,\"lastValue\":171722.0,\"lastTimestamp\":1653369328111}}",
				},
			},
			{
				{
					Type: gjson.JSON,
					Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171456.0,\"lastTimestamp\":1653369141709}}",
				},
			},
		}
		Convey("mergePool submit err ", func() {

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

			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 0}, selRange)
			So(mat, ShouldBeNil)
			So(err.Error(), ShouldResemble, `this pool has been closed`)
		})

		Convey("success with 1 sample ", func() {

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{
				{
					Name:  interfaces.LABELS_STR,
					Value: key,
				}}
			//根据向右取值的时间点的值计算irate逻辑，因此两个时间点的值相同
			irateQuery2 := &interfaces.Query{
				Start:      1653368870,
				End:        1653369700,
				Interval:   60000,
				LogGroupId: "a",
			}
			irateRange := selRange.Milliseconds()
			irateQuery2.SubIntervalWith30min = getCommonDivisor(irateRange, irateQuery2.Interval)
			irateQuery2.FixedStart = int64(math.Floor(float64(irateQuery2.Start*1000)/float64(irateQuery2.Interval))) * irateQuery2.Interval
			irateQuery2.FixedEnd = int64(math.Floor(float64(irateQuery2.End*1000)/float64(irateQuery2.Interval))) * irateQuery2.Interval
			start := irateQuery2.FixedStart
			end := irateQuery2.FixedEnd
			calseriesPoint := make(map[int64]static.IratePoint, 1)
			j := 1
			for ts := start; ts <= end; ts += irateQuery2.Interval {
				if ts > 1653369300000 {
					break
				}
				if j <= 3 {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369020709,
						PreviousV: 171536,
						LastT:     1653369141709,
						LastV:     171456,
					}
				} else {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369321709,
						PreviousV: 171856,
						LastT:     1653369328111,
						LastV:     171722,
					}
				}
				j++
			}
			calSeries := []static.IrateSeries{
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
					Points: calseriesPoint,
				},
			}
			finalPoint := make([]static.Point, 0)

			points := []static.Point{}
			i := 1
			for ts := start; ts <= end; ts += irateQuery2.Interval {
				if ts > 1653369300000 {
					break
				}
				if i <= 3 {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				} else {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				}
				i++
			}
			expectSeries := []static.Series{
				{
					Metric: calSeries[0].Metric,
					Points: points,
				},
			}
			finalPoint = make([]static.Point, 0)
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, irateQuery2, selRange)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with more than 1 sample ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171011.0,\"lastTimestamp\":1653369141709}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:15:00.000Z\",\"key\":1653369300000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171722.0,\"lastTimestamp\":1653369328111}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
					},
				},
			}

			tsValueMap := make(map[string][][]gjson.Result)
			tsValueMap[key] = resultArra

			labelsMap := make(map[string][]*labels.Label)
			labelsMap[key] = []*labels.Label{
				{
					Name:  interfaces.LABELS_STR,
					Value: key,
				},
			}
			irateRange := selRange.Milliseconds()
			irateQueryFix.SubIntervalWith30min = getCommonDivisor(irateRange, irateQueryFix.Interval)
			irateQueryFix.FixedStart = int64(math.Floor(float64(irateQueryFix.Start*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			irateQueryFix.FixedEnd = int64(math.Floor(float64(irateQueryFix.End*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			start := irateQueryFix.FixedStart
			end := irateQueryFix.FixedEnd
			j := 1
			calseriesPoint := make(map[int64]static.IratePoint, 1)
			for ts := start; ts <= end; ts += irateQueryFix.Interval {
				if ts > 1653369300000 {
					break
				}
				if j <= 3 {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369020709,
						PreviousV: 171536,
						LastT:     1653369141709,
						LastV:     171011,
					}
				} else {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369141709,
						PreviousV: 171011,
						LastT:     1653369328111,
						LastV:     171722,
					}
				}
				j++
			}
			points := []static.Point{}
			i := 1
			finalPoint := make([]static.Point, 0)
			for ts := start; ts <= end; ts += irateQueryFix.Interval {
				if ts > 1653369000000 {
					break
				}
				if i <= 3 {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				} else {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				}
				i++
			}
			calSeries := []static.IrateSeries{
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
					Points: calseriesPoint,
				},
			}

			expectSeries := []static.Series{
				{
					Metric: calSeries[0].Metric,
					Points: points,
				},
			}
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, irateQueryFix, selRange)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
		Convey("success with instant query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171011.0,\"lastTimestamp\":1653369141709}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:15:00.000Z\",\"key\":1653369300000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171722.0,\"lastTimestamp\":1653369328111}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
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
			query3 := &interfaces.Query{Interval: 300000, Start: 165336900000, End: 165336930000, IsInstantQuery: true}
			irateRange := selRange.Milliseconds()
			query3.SubIntervalWith30min = getCommonDivisor(irateRange, query3.Interval)
			query3.FixedStart = int64(math.Floor(float64(query3.Start*1000)/float64(query3.Interval))) * query3.Interval
			query3.FixedEnd = int64(math.Floor(float64(query3.End*1000)/float64(query3.Interval))) * query3.Interval

			calSeries := []static.IrateSeries{
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
					Points: map[int64]static.IratePoint{
						165336900000: {
							PreviousT: 1653369141709,
							PreviousV: 171011,
							LastT:     1653369328111,
							LastV:     171722,
						},
						165336960000: {
							PreviousT: 1653369020709,
							PreviousV: 171536,
							LastT:     1653369328111,
							LastV:     171722,
						},
					},
				},
			}
			expectSeries := []static.Series{
				{
					Metric: calSeries[0].Metric,
					Points: []static.Point{
						{T: 165336930000, V: 3.8143367560433905},
					},
				},
			}

			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)
			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, query3, selRange)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})

		Convey("success with supplementary points query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171011.0,\"lastTimestamp\":1653369141709}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:25:00.000Z\",\"key\":1653369900000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171722.0,\"lastTimestamp\":1653369328111}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:25:00.000Z\",\"key\":1653369900000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
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
			irateRange := selRange.Milliseconds()
			irateQueryFix = &interfaces.Query{
				Start:      1653369000,
				End:        1653370000,
				Interval:   60000,
				LogGroupId: "a",
			}
			subInterval := getCommonDivisor(irateRange, irateQueryFix.Interval)
			fixedStart := int64(math.Floor(float64(irateQueryFix.Start*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			fixedEnd := int64(math.Floor(float64(irateQueryFix.End*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			start := fixedStart
			end := fixedEnd
			j := 1
			calseriesPoint := make(map[int64]static.IratePoint, 1)
			for ts := start; ts <= end; ts += irateQueryFix.Interval {
				if 1653369600000 <= ts && ts < 1653369960000 {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369020709,
						PreviousV: 171536,
						LastT:     1653369328111,
						LastV:     171722,
					}
				} else if ts == 1653369000000 {
					calseriesPoint[ts] = static.IratePoint{
						PreviousT: 1653369020709,
						PreviousV: 171536,
						LastT:     1653369141709,
						LastV:     171011,
					}
				}
				j++
			}
			points := []static.Point{}
			i := 1
			finalPoint := make([]static.Point, 0)
			for ts := start; ts <= end; ts += irateQueryFix.Interval {

				if ts == 1653369000000 {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				} else if 1653369600000 <= ts && ts < 1653369960000 {
					point := (*irateFunction(&calseriesPoint, ts, &finalPoint))
					points = append(points, point[len(point)-1])
				}

				i++
			}
			calSeries := []static.IrateSeries{
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
					Points: calseriesPoint,
				},
			}

			expectSeries := []static.Series{
				{
					Metric: calSeries[0].Metric,
					Points: points,
				},
			}
			finalPoint = make([]static.Point, 0)
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)

			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 60000, Start: 165336900000, End: 1653370000, IsInstantQuery: false, SubIntervalWith30min: subInterval, FixedStart: fixedStart, FixedEnd: fixedEnd}, selRange)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
		Convey("success with previous points query ", func() {
			resultArra := [][]gjson.Result{
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":171530,\"previousTimestamp\":1653369020710,\"lastValue\":171536.0,\"lastTimestamp\":1653369020712}}",
					},
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:25:00.000Z\",\"key\":1653369900000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171722.0,\"lastTimestamp\":1653369328111}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:25:00.000Z\",\"key\":1653369900000,\"doc_count\":1,\"value\":{\"previousValue\":null,\"previousTimestamp\":null,\"lastValue\":171536.0,\"lastTimestamp\":1653369020709}}",
					},
				},
				{
					{
						Type: gjson.JSON,
						Raw:  "{\"key_as_string\":\"2022-05-24T05:10:00.000Z\",\"key\":1653369000000,\"doc_count\":1,\"value\":{\"previousValue\":171530,\"previousTimestamp\":1653369020718,\"lastValue\":171011.0,\"lastTimestamp\":1653369141709}}",
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

			calSeries := []static.IrateSeries{
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
					Points: map[int64]static.IratePoint{
						1653369000000: {
							PreviousT: 1653369020718,
							PreviousV: 171530,
							LastT:     1653369141709,
							LastV:     171011,
						},
						1653369600000: {
							PreviousT: 1653369020709,
							PreviousV: 171536,
							LastT:     1653369328111,
							LastV:     171722,
						},
						1653369900000: {
							PreviousT: 1653369020709,
							PreviousV: 171536,
							LastT:     1653369328111,
							LastV:     171722,
						},
					},
				},
			}
			keys := make([]int, 0)
			for k := range calSeries[0].Points {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			finalPoint := make([]static.Point, 0)
			expectSeries := []static.Series{
				{
					Metric: calSeries[0].Metric,
					Points: []static.Point{
						(*irateFunction(&(calSeries[0].Points), int64(keys[0]), &finalPoint))[0],
						(*irateFunction(&(calSeries[0].Points), int64(keys[1]), &finalPoint))[1],
						(*irateFunction(&(calSeries[0].Points), int64(keys[2]), &finalPoint))[2],
					},
				},
			}
			finalPoint = make([]static.Point, 0)
			expectMat := make(static.Matrix, 0)
			expectMat = append(expectMat, expectSeries...)
			irateRange := selRange.Milliseconds()
			subInterval := getCommonDivisor(irateRange, query.Interval)
			fixedStart := int64(math.Floor(float64(irateQueryFix.Start*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			fixedEnd := int64(math.Floor(float64(irateQueryFix.End*1000)/float64(irateQueryFix.Interval))) * irateQueryFix.Interval
			mat, err := irateMerge(MapResult{LabelsMap: labelsMap, TsValueMap: tsValueMap}, &interfaces.Query{Interval: 300000, Start: 165336900000, End: 1653370000, IsInstantQuery: false, SubIntervalWith30min: subInterval, FixedStart: fixedStart, FixedEnd: fixedEnd}, selRange)
			So(mat, ShouldResemble, expectMat)
			So(err, ShouldBeNil)
		})
	})
}
