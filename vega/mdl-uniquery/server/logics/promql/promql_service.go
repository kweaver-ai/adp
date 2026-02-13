// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package promql

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
	"uniquery/logics/data_view"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/leafnodes"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
	"uniquery/logics/promql/util"
)

var (
	pqlServiceOnce sync.Once
	pqlService     interfaces.PromQLService
)

type promQLService struct {
	leafNodes  *leafnodes.LeafNodes
	appSetting *common.AppSetting
	mmService  interfaces.MetricModelService
}

type QueryData struct {
	ResultType parser.ValueType `json:"resultType"`
	Result     parser.Value     `json:"result"`
}

// just for NewPromQLService and UT
func NewPromQLServiceRaw(appSetting *common.AppSetting, ln *leafnodes.LeafNodes, mmService interfaces.MetricModelService) interfaces.PromQLService {
	return &promQLService{
		leafNodes:  ln,
		appSetting: appSetting,
		mmService:  mmService,
	}
}

func NewPromQLService(appSetting *common.AppSetting, mmService interfaces.MetricModelService) interfaces.PromQLService {
	pqlServiceOnce.Do(func() {
		ln := leafnodes.NewLeafNodes(appSetting, logics.OSAccess,
			logics.LGAccess, data_view.NewDataViewService(appSetting))
		pqlService = NewPromQLServiceRaw(appSetting, ln, mmService)

		// init pool
		util.InitAntsPool(appSetting.PoolSetting)

		// 创建 service 的时候触发 refresh 缓存
		go ln.RefreshShards()
	})

	return pqlService
}

// recover is the handler that turns panics into returns from the top level of evaluation.
func (ps *promQLService) recover(status *int, errp *error) {
	e := recover()
	if e == nil {
		return
	}

	*status = http.StatusUnprocessableEntity
	*errp = uerrors.PromQLError{
		Typ: uerrors.ErrorBadData,
		Err: e.(error),
	}

}

// 执行 promql 语句，返回查询结果，状态码，错误信息
func (ps *promQLService) Exec(ctx context.Context, query interfaces.Query) (response interfaces.PromQLResponse, b []byte, status int, err error) {
	var res interfaces.PromQLResponse
	// 解析 query
	parseCtx, span := ar_trace.Tracer.Start(ctx, "解析 promql 表达式")

	expr, err := parser.ParseExpr(ctx, query.QueryStr)
	if err != nil {
		// 设置 trace 的错误信息的 attributes
		span.SetStatus(codes.Error, "Promql parser.ParseExpr error")
		span.End()
		// 记录接口调用参数
		o11y.Error(parseCtx, fmt.Sprintf("Promql [%s] parse error: [%v]", query.QueryStr, err))
		span.End()
		return res, nil, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: errors.New(err.Error()),
		}
	}
	span.SetStatus(codes.Ok, "")
	span.End()

	if expr.Type() != parser.ValueTypeVector && expr.Type() != parser.ValueTypeScalar && expr.Type() != parser.ValueTypeMatrix {
		return res, nil, http.StatusUnprocessableEntity,
			uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("invalid expression type %q for range query, must be Scalar,"+
					"range or instant Vector", parser.DocumentedType(expr.Type())),
			}
	}

	// 对表达式做一个预处理，对于step不变的表达式，封装成 StepInvariantExpr
	expr, err = static.PreprocessExpr(expr, query.Start, query.End)
	if err != nil {
		return res, nil, http.StatusUnprocessableEntity, err
	}

	// 根据请求修正start，end，使得 [start, end] 按step划窗与opensearch的date_histogram的bucket key生成规则吻合
	// 修正逻辑为 Math.floor((value) / interval) * interval
	// 现除以桶，然后按照时区偏移，promql默认使用的是local，不指定时区
	// _, offset := time.Now().In(common.APP_LOCATION).Zone()
	// query.FixedStart = int64(math.Floor(float64(query.Start+int64(offset*1000))/float64(query.Interval)))*query.Interval - int64(offset*1000)
	// query.FixedEnd = int64(math.Floor(float64(query.End+int64(offset*1000))/float64(query.Interval)))*query.Interval - int64(offset*1000)

	// 时间修正, 同期值和本期值不同,同期值需要把时间置到同期的时间轴上
	fixedStart, fixedEnd := static.CorrectingTime(query, common.APP_LOCATION)
	query.FixedStart = fixedStart
	query.FixedEnd = fixedEnd

	//捕获function的panic
	defer ps.recover(&status, &err)

	// 记录 eval 的 query 信息

	// 记录 promql 相关请求信息
	o11y.Info(ctx, fmt.Sprintf("PromQL query info [QueryStr: %s, Start: %d, End: %d, FixedStart:%d,FixedEnd: %d, Interval: %d, DataViewId: %s, "+
		"IsMetricModel: %v, IsInstantQuery:%v]",
		query.QueryStr, query.Start, query.End, query.FixedStart, query.FixedEnd, query.Interval, query.LogGroupId,
		query.IsMetricModel, query.IsInstantQuery))

	// 对于瞬时向量 query 接口的查询，加一个if判断 if s.Start == s.End && s.Interval == 0 {
	if query.Start == query.End || query.IsInstantQuery {
		val, status, err := ps.eval(ctx, expr, &query)
		if err != nil {
			return res, nil, status, err
		}

		var mat static.Matrix
		var seriesTotal int
		switch result := val.(type) {
		case static.Matrix:
			mat = result
		case static.PageMatrix:
			mat = result.Matrix
			seriesTotal = result.TotalSeries
		default:
			return res, nil, http.StatusUnprocessableEntity, fmt.Errorf("invalid expression type %q", val.Type())
		}

		switch expr.Type() {
		case parser.ValueTypeVector:
			// Convert matrix with one value per series into vector.
			vector := make(static.Vector, len(mat))
			for i, s := range mat {
				// Point might have a different timestamp, force it to the evaluation
				// timestamp as that is when we ran the evaluation.
				// 取最后一个数据点，因为在opensearch往前推一个delta之后，对时间分桶，可能会跨两个时间窗，取大。
				vector[i] = static.Sample{Metric: s.Metric, Point: static.Point{V: s.Points[len(s.Points)-1].V, T: query.End}}
			}
			return marshalParseValue(query, vector, seriesTotal, status)

		case parser.ValueTypeScalar:
			return marshalParseValue(query, static.Scalar{V: mat[0].Points[0].V, T: query.Start}, 0, status)
		default:
			return res, nil, http.StatusUnprocessableEntity, fmt.Errorf("unexpected expression type %q", expr.Type())
		}

	}

	// 执行 query
	val, status, err := ps.eval(ctx, expr, &query)
	if err != nil {
		return res, nil, status, err
	}
	var seriesTotal int
	switch result := val.(type) {
	case static.PageMatrix:
		seriesTotal = result.TotalSeries
	}

	// 取值，返回
	return marshalParseValue(query, val, seriesTotal, status)
}

// parser.Value类型的值封装成 Prometheus 的返回结构并序列化返回
func marshalParseValue(query interfaces.Query, v parser.Value, seriesTotal int, status int) (interfaces.PromQLResponse, []byte, int, error) {
	response := interfaces.PromQLResponse{
		Status: "success",
		Data: QueryData{
			ResultType: v.Type(),
			Result:     v,
		},
		SeriesTotal:    seriesTotal,
		VegaDurationMs: query.VegaDurationMs,
	}

	if query.IsMetricModel {
		return response, nil, status, nil
	} else {
		// 标准库的 json 比 jsoniter 稍微快些
		// json := jsoniter.ConfigFastest
		mat, ok := v.(static.PageMatrix)
		if ok {
			v = mat.Matrix
		}

		response := interfaces.PromQLResponse{
			Status: "success",
			Data: QueryData{
				ResultType: v.Type(),
				Result:     v,
			},
		}

		bytes, err := sonic.Marshal(response)
		if err != nil {
			return interfaces.PromQLResponse{}, nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: errors.New("promql_service.marshalParseValue Marshal response error: " + err.Error()),
			}
		}

		return interfaces.PromQLResponse{}, bytes, status, err
	}

}

// eval evaluates the given expression as the given AST expression node requires.
func (ps *promQLService) eval(ctx context.Context, expr parser.Expr, query *interfaces.Query) (parser.Value, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("eval 节点[%T]", expr))
	span.SetAttributes(attribute.Key("promql_expression").String(fmt.Sprintf("%v", expr)))
	defer span.End()

	switch e := expr.(type) {
	case *parser.AggregateExpr:
		span.SetStatus(codes.Ok, "")
		// query加个标识，表示这个query是序列聚合。当叶子节点在做计算时，根据这个标识来确定加载全部还是分页查询
		query.IfNeedAllSeries = true
		return ps.evalAggregateExpr(ctx, e, query)

	case *parser.Call:
		if e.Func == nil {
			span.SetStatus(codes.Error, "FunctionCalls is not defined")
			// 记录异常日志
			o11y.Error(ctx, "FunctionCalls is not defined, please input valid function.")

			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" FunctionCalls is not defined, please input valid function. "),
			}
		}
		// 不允许在query_range请求上使用sort&sort_desc函数
		if !query.IsInstantQuery {
			switch e.Func.Name {
			case "sort", "sort_desc":
				span.SetStatus(codes.Error, fmt.Sprintf("'%s' is not surpport in range query", e.Func.Name))
				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("'%s' can not be used in the query_range requests.", e.Func.Name))

				return nil, http.StatusBadRequest, uerrors.PromQLError{
					Typ: uerrors.ErrorBadData,
					Err: fmt.Errorf(" '%s' can not be used in the query_range requests. ", e.Func.Name),
				}
			}
		}

		// 这里拿到的时call对象
		call := static.FunctionCalls[e.Func.Name].New(e.Args)
		if call == nil {
			span.SetStatus(codes.Error, fmt.Sprintf("'%s' is not currently supported", e.Func.Name))
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("'%s' is not currently supported.", e.Func.Name))
			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" '%s' is not currently supported. ", e.Func.Name),
			}
		}

		var (
			matrixArgIndex int
			matrixArg      bool
			status         int
			err            error
		)
		for i := range e.Args {
			static.UnwrapParenExpr(&e.Args[i])
			a := static.UnwrapStepInvariantExpr(e.Args[i])
			static.UnwrapParenExpr(&a)
			if _, ok := a.(*parser.MatrixSelector); ok {
				matrixArgIndex = i
				matrixArg = true
				break
			}
		}

		if !matrixArg {
			// Does not have a matrix argument.
			span.SetStatus(codes.Ok, "")
			switch e.Func.Name {
			case interfaces.CUMULATIVE_SUM:
				// 通用函数都是在每个step上对v做计算，即对指标在同一个时间点上的各个序列做计算。
				// 而累加和是对每个序列在时间轴上的累计求和的过程。所以，累加和不能用通用function的模式来进行计算，特例处理。
				return ps.evalCumulativeSum(ctx, e, query)
			case interfaces.HISTOGRAM_QUANTILE:
				// histogram_quantile的计算需要全部序列
				query.IfNeedAllSeries = true
			case interfaces.K_MINUTE_DOWNTIME:
				// 通用函数是公用一个query，对于k-minute_downtime来说，其第一个参数表达式应用的step应为1min。
				// 需改变query对象的step和开始时间。
				query.IfNeedAllSeries = true
				return ps.evalKMinuteDowntime(ctx, e, query)

			case interfaces.FUNC_METRIC_MODEL:
				return ps.evalMetricModel(ctx, e, query)
			}
			return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
				return call.Call(v, e.Args, enh), status, err
			}, e.Args...)
		}

		static.UnwrapParenExpr(&e.Args[matrixArgIndex])
		arg := static.UnwrapStepInvariantExpr(e.Args[matrixArgIndex])
		static.UnwrapParenExpr(&arg)
		sel := arg.(*parser.MatrixSelector)

		switch e.Func.Name {
		case interfaces.IRATE_AGG:
			span.SetStatus(codes.Ok, "")
			return ps.leafNodes.IrateEval(ctx, sel, []string{interfaces.LABELS_STR}, interfaces.IRATE_AGG, query)
		case interfaces.RATE_AGG, interfaces.INCREASE_AGG:
			span.SetStatus(codes.Ok, "")
			return ps.leafNodes.RateAggs(ctx, sel, query, call)
		case interfaces.CHANGES_AGG:
			span.SetStatus(codes.Ok, "")
			return ps.leafNodes.ChangesAggs(ctx, sel, query)
		case interfaces.DELTA_AGG:
			span.SetStatus(codes.Ok, "")
			return ps.leafNodes.DeltaAggs(ctx, sel, query, call)
		case interfaces.AVG_OVER_TIME, interfaces.SUM_OVER_TIME, interfaces.MAX_OVER_TIME, interfaces.MIN_OVER_TIME, interfaces.COUNT_OVER_TIME:
			span.SetStatus(codes.Ok, "")
			return ps.leafNodes.AggOverTime(ctx, sel, query, e.Func.Name)
		default:
			span.SetStatus(codes.Error, fmt.Sprintf("unhandled expression of type: %T", e.Func.Name))
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("unhandled expression of type: %T", e.Func.Name))

			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("unhandled expression of type: %T", e.Func.Name),
			}
		}

	case *parser.ParenExpr:
		span.SetStatus(codes.Ok, "")
		// query加个标识，表示这个query是二元运算。当叶子节点在做计算时，根据这个标识来确定加载全部还是分页查询
		query.IfNeedAllSeries = true
		return ps.eval(ctx, e.Expr, query)

	case *parser.UnaryExpr:
		span.SetStatus(codes.Ok, "")
		return ps.evalUnaryExpr(ctx, e, query)

	case *parser.BinaryExpr:
		span.SetStatus(codes.Ok, "")
		// query加个标识，表示这个query是二元运算。当叶子节点在做计算时，根据这个标识来确定加载全部还是分页查询
		query.IfNeedAllSeries = true
		return ps.evalBinaryExpr(ctx, e, query)

	case *parser.NumberLiteral:
		span.SetStatus(codes.Ok, "")
		return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
			return append(enh.Out, static.Sample{Point: static.Point{V: e.Val}}), http.StatusOK, nil
		})

	case *parser.VectorSelector:
		// 对应采样逻辑：termsAgg 用 __labels_str, valueAgg 用 sampling
		// 根据叶子节点 VectorSelector 的属性构建 dsl 查询请求.
		span.SetStatus(codes.Ok, "")
		return ps.leafNodes.EvalVectorSelector(ctx, e, []string{interfaces.LABELS_STR}, interfaces.SAMPLING_AGG, query)

	case *parser.StepInvariantExpr:
		span.SetStatus(codes.Ok, "")
		return ps.evalStepInvariantExpr(ctx, e, query)
	}

	span.SetStatus(codes.Error, fmt.Sprintf("unhandled expression of type: %T", expr))
	// 记录异常日志
	o11y.Error(ctx, fmt.Sprintf("unhandled expression of type: %T", expr))
	return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("unhandled expression of type: %T", expr),
	}
}

// eval 聚合操作
func (ps *promQLService) evalAggregateExpr(ctx context.Context, expr *parser.AggregateExpr, query *interfaces.Query) (parser.Value, int, error) {
	// Grouping labels must be sorted (expected both by generateGroupingKey() and aggregation()).
	sortedGrouping := expr.Grouping
	sort.Strings(sortedGrouping)

	// Prepare a function to initialise series helpers with the grouping key.
	buf := make([]byte, 0, 1024)
	initSeries := func(series labels.Labels, h *static.EvalSeriesHelper) {
		h.GroupingKey, buf = static.GenerateGroupingKey(series, sortedGrouping, expr.Without, buf)
	}

	static.UnwrapParenExpr(&expr.Param)
	param := static.UnwrapStepInvariantExpr(expr.Param)
	static.UnwrapParenExpr(&param)
	// 聚合函数的参数为字符串只用于 count_values 的情况,先去掉.
	// if s, ok := param.(*parser.StringLiteral); ok {
	// 	return ps.rangeEval(query, initSeries, func(v []parser.Value, sh [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
	// 		val, err := static.Aggregation(e.Op, sortedGrouping, e.Without, s.Val, v[0].(static.Vector), sh[0], enh)
	// 		if err != nil {
	// 			return nil, http.StatusUnprocessableEntity, err
	// 		}
	// 		return val, http.StatusOK, nil
	// 	}, e.Expr)
	// }

	return ps.rangeEval(ctx, query, initSeries, func(v []parser.Value, sh [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
		var param float64
		if expr.Param != nil {
			param = v[0].(static.Vector)[0].V
		}

		val, err := static.Aggregation(expr.Op, sortedGrouping, expr.Without, param, v[1].(static.Vector), sh[1], enh)
		if err != nil {
			return nil, http.StatusUnprocessableEntity, err
		}
		return val, http.StatusOK, nil
	}, expr.Param, expr.Expr)
}

// eval 步长不变表达式，例如：1+2
func (ps *promQLService) evalStepInvariantExpr(ctx context.Context, expr *parser.StepInvariantExpr, query *interfaces.Query) (parser.Value, int, error) {
	switch ce := expr.Expr.(type) {
	case *parser.StringLiteral, *parser.NumberLiteral:
		return ps.eval(ctx, ce, query)
	}

	newQuery := interfaces.Query{
		QueryStr:             query.QueryStr,
		Start:                query.Start,
		End:                  query.End,
		FixedStart:           query.FixedStart,
		FixedEnd:             query.FixedStart,
		Interval:             query.Interval,
		IntervalStr:          query.IntervalStr,
		IsInstantQuery:       query.IsInstantQuery,
		SubIntervalWith30min: query.SubIntervalWith30min,
		LogGroupId:           query.LogGroupId,
	}
	res, status, err := ps.eval(ctx, expr.Expr, &newQuery)
	if err != nil {
		return nil, status, err
	}

	// For every evaluation while the value remains same, the timestamp for that
	// value would change for different eval times. Hence we duplicate the result
	// with changed timestamps.
	mat, ok := res.(static.Matrix)
	if !ok {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("unexpected result in StepInvariantExpr evaluation: %T", expr),
		}
	}
	for i := range mat {
		if len(mat[i].Points) != 1 {
			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("unexpected number of samples"),
			}
		}
		// for ts := query.FixedStart + query.Interval; ts <= query.FixedEnd; ts = ts + query.Interval {
		for ts := static.GetNextPointTime(*query, query.FixedStart); ts <= query.FixedEnd; ts = static.GetNextPointTime(*query, ts) {
			mat[i].Points = append(mat[i].Points, static.Point{
				T: ts,
				V: mat[i].Points[0].V,
			})

		}
	}
	return res, http.StatusOK, nil
}

// eval 二元运算表达式，分为四种情况： 标量 op 标量， 标量 op 向量， 向量 op 标量， 向量 op 向量
func (ps *promQLService) evalBinaryExpr(ctx context.Context, expr *parser.BinaryExpr, query *interfaces.Query) (parser.Value, int, error) {
	switch lt, rt := expr.LHS.Type(), expr.RHS.Type(); {
	case lt == parser.ValueTypeScalar && rt == parser.ValueTypeScalar:
		return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
			val, err := static.ScalarBinop(expr.Op, v[0].(static.Vector)[0].Point.V, v[1].(static.Vector)[0].Point.V)
			if err != nil {
				return nil, http.StatusUnprocessableEntity, err
			}
			return append(enh.Out, static.Sample{Point: static.Point{V: val}}), http.StatusOK, nil
		}, expr.LHS, expr.RHS)

	case lt == parser.ValueTypeScalar && rt == parser.ValueTypeVector:
		return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
			val, err := static.VectorscalarBinop(expr.Op, v[1].(static.Vector), static.Scalar{V: v[0].(static.Vector)[0].Point.V}, true, expr.ReturnBool, enh)
			if err != nil {
				return nil, http.StatusUnprocessableEntity, err
			}
			return val, http.StatusOK, nil
		}, expr.LHS, expr.RHS)

	case lt == parser.ValueTypeVector && rt == parser.ValueTypeScalar:
		return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
			val, err := static.VectorscalarBinop(expr.Op, v[0].(static.Vector), static.Scalar{V: v[1].(static.Vector)[0].Point.V}, false, expr.ReturnBool, enh)
			if err != nil {
				return nil, http.StatusUnprocessableEntity, err
			}
			return val, http.StatusOK, nil
		}, expr.LHS, expr.RHS)

	case lt == parser.ValueTypeVector && rt == parser.ValueTypeVector:
		// 逻辑/集合 and、or、unless
		switch expr.Op {
		case parser.LAND:
			return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
				return static.VectorAnd(v[0].(static.Vector), v[1].(static.Vector), expr.VectorMatching, enh), http.StatusOK, nil
			}, expr.LHS, expr.RHS)
		case parser.LOR:
			return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
				return static.VectorOr(v[0].(static.Vector), v[1].(static.Vector), expr.VectorMatching, enh), http.StatusOK, nil
			}, expr.LHS, expr.RHS)
		case parser.LUNLESS:
			return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
				return static.VectorUnless(v[0].(static.Vector), v[1].(static.Vector), expr.VectorMatching, enh), http.StatusOK, nil
			}, expr.LHS, expr.RHS)
		default:
			return ps.rangeEval(ctx, query, nil, func(v []parser.Value, _ [][]static.EvalSeriesHelper, enh *static.EvalNodeHelper) (static.Vector, int, error) {
				val, err := static.VectorBinop(expr.Op, v[0].(static.Vector), v[1].(static.Vector), expr.VectorMatching, expr.ReturnBool, enh)
				if err != nil {
					return nil, http.StatusUnprocessableEntity, err
				}
				return val, http.StatusOK, nil
			}, expr.LHS, expr.RHS)
		}
	}

	return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("unhandled expression of type: %T", expr),
	}
}

// eval 一元运算表达式，例如 -3^1000
func (ps *promQLService) evalUnaryExpr(ctx context.Context, expr *parser.UnaryExpr, query *interfaces.Query) (parser.Value, int, error) {
	val, status, err := ps.eval(ctx, expr.Expr, query)
	if err != nil {
		return nil, status, err
	}
	switch result := val.(type) {
	case static.Matrix:
		return processUnaryMatrix(expr, result)
	case static.PageMatrix:
		mat, status, err := processUnaryMatrix(expr, result.Matrix)
		if err != nil {
			return nil, status, err
		}
		return static.PageMatrix{Matrix: mat, TotalSeries: result.TotalSeries}, status, nil
	default:
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("invalid expression type %q", val.Type())
	}
}

// 处理可分页和不可分页时的一元运算的数据
func processUnaryMatrix(expr *parser.UnaryExpr, mat static.Matrix) (static.Matrix, int, error) {
	if expr.Op == parser.SUB {
		for i := range mat {
			for j := range mat[i].Points {
				mat[i].Points[j].V = -mat[i].Points[j].V
			}
		}
		if mat.ContainsSameLabelset() {
			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("vector cannot contain metrics with the same labelset"),
			}
		}
	}
	return mat, http.StatusOK, nil
}

// eval 累加和函数
func (ps *promQLService) evalCumulativeSum(ctx context.Context, expr *parser.Call, query *interfaces.Query) (parser.Value, int, error) {
	val, status, err := ps.eval(ctx, expr.Args[0], query)
	if err != nil {
		return nil, status, err
	}
	switch result := val.(type) {
	case static.Matrix:
		return processCumulativeSumMatrix(result, query)
	case static.PageMatrix:
		mat, status, err := processCumulativeSumMatrix(result.Matrix, query)
		if err != nil {
			return nil, status, err
		}
		return static.PageMatrix{Matrix: mat, TotalSeries: result.TotalSeries}, status, nil
	default:
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("invalid expression type %q", val.Type())
	}
}

// 处理可分页和不可分页时的累加和的数据
func processCumulativeSumMatrix(mat static.Matrix, query *interfaces.Query) (static.Matrix, int, error) {
	for i := range mat {
		// 遍历序列，对每个step做累加和的计算
		pointIndex := 0
		points := make([]static.Point, 0)
		for ts := query.FixedStart; ts <= query.FixedEnd; {
			// 时间修正可能使得修正的时间大于第一个桶时间
			if ts > mat[i].Points[pointIndex].T {
				ts = mat[i].Points[pointIndex].T
			}

			preVal := 0.0
			if len(points) > 0 {
				preVal = points[len(points)-1].V
			}
			// 如果ts小于当前点的时间，需要一直补点到两者时间相等 ts==mat[i].Points[pointIndex].T
			if ts < mat[i].Points[pointIndex].T {
				for ts < mat[i].Points[pointIndex].T {
					points = append(points, static.Point{
						T: ts,
						V: preVal, // preVal + 0
					})
					// ts += query.Interval
					ts = static.GetNextPointTime(*query, ts)
				}
			} else if ts == mat[i].Points[pointIndex].T {
				points = append(points, static.Point{
					T: ts,
					V: preVal + mat[i].Points[pointIndex].V,
				})
				// ts += query.Interval
				ts = static.GetNextPointTime(*query, ts)
				pointIndex++
			}

			// pointIndex==len(Points) 说明Points已经遍历结束了
			if pointIndex == len(mat[i].Points) {
				// 再读取一次最新的前一个点的值
				preVal = points[len(points)-1].V
				// 如果此时time <= FixedEnd，说明后面缺点
				for ts <= query.FixedEnd {
					points = append(points, static.Point{
						T: ts,
						V: preVal, // preVal + 0
					})
					// ts = ts + query.Interval
					ts = static.GetNextPointTime(*query, ts)
				}
			}
		}
		mat[i].Points = points
	}
	return mat, http.StatusOK, nil
}

// rangeEval evaluates the given expressions, and then for each step calls
// the given funcCall with the values computed for each expression at that
// step. The return value is the combination into time series of all the
// function call results.
// The prepSeries function (if provided) can be used to prepare the helper
// for each series, then passed to each call funcCall.
func (ps *promQLService) rangeEval(ctx context.Context, query *interfaces.Query, prepSeries func(labels.Labels, *static.EvalSeriesHelper), funcCall func([]parser.Value, [][]static.EvalSeriesHelper,
	*static.EvalNodeHelper) (static.Vector, int, error), exprs ...parser.Expr) (static.Matrix, int, error) {

	matrixes := make([]static.Matrix, len(exprs))
	origMatrixes := make([]static.Matrix, len(exprs))

	for i, expr := range exprs {
		// Functions will take string arguments from the expressions, not the values.
		if expr != nil && expr.Type() != parser.ValueTypeString {
			// ev.currentSamples will be updated to the correct value within the ev.eval call.
			val, status, err := ps.eval(ctx, expr, query) // 逐个eval表达式
			if err != nil {
				return nil, status, err
			}
			switch result := val.(type) {
			case static.Matrix:
				matrixes[i] = result
			case static.PageMatrix:
				matrixes[i] = result.Matrix
			}
			// matrixes[i] = val.(static.Matrix) // eval得到的结果放入结果集数组中

			// Keep a copy of the original point slices so that they
			// can be returned to the pool.
			origMatrixes[i] = make(static.Matrix, len(matrixes[i]))
			copy(origMatrixes[i], matrixes[i])
		}
	}

	vectors := make([]static.Vector, len(exprs)) // Input vectors for the function.
	args := make([]parser.Value, len(exprs))     // Argument to function.
	// Create an output vector that is as big as the input matrix with
	// the most time series.
	biggestLen := 1
	for i := range exprs {
		vectors[i] = make(static.Vector, 0, len(matrixes[i]))
		if len(matrixes[i]) > biggestLen {
			biggestLen = len(matrixes[i])
		}
	}
	enh := &static.EvalNodeHelper{Out: make(static.Vector, 0, biggestLen)}
	seriess := make(map[uint64]*static.Series, biggestLen) // Output series by series hash.
	// series 被放在一个map中，最后取出时吐出的 Matrix 的时间序列不是按固定顺序的，用一个数组来记录，先后添加的顺序，遍历map时按这个顺序来把series取出组成matrix
	seriessOrder := make([]*static.Series, 0, biggestLen)

	var (
		seriesHelpers [][]static.EvalSeriesHelper
		bufHelpers    [][]static.EvalSeriesHelper // Buffer updated on each step
	)

	// If the series preparation function is provided, we should run it for
	// every single series in the matrix.
	if prepSeries != nil {
		seriesHelpers = make([][]static.EvalSeriesHelper, len(exprs))
		bufHelpers = make([][]static.EvalSeriesHelper, len(exprs))

		for i := range exprs {
			seriesHelpers[i] = make([]static.EvalSeriesHelper, len(matrixes[i]))
			bufHelpers[i] = make([]static.EvalSeriesHelper, len(matrixes[i]))

			for si, series := range matrixes[i] {
				h := seriesHelpers[i][si]
				prepSeries(series.Metric, &h)
				seriesHelpers[i][si] = h
			}
		}
	}

	// 时间修正, 同期值和本期值不同,同期值需要把时间置到同期的时间轴上
	// fixedStart, fixedEnd := static.CorrectingTime(*query, common.APP_LOCATION)

	// 生成完整的时间点
	allTimes := make([]int64, 0)
	if query.IsInstantQuery {
		// 即时查询时start等于end，只有一个数据点
		allTimes = append(allTimes, query.FixedStart)
	} else {
		for currentTime := query.FixedStart; currentTime <= query.FixedEnd; {
			allTimes = append(allTimes, currentTime)
			currentTime = static.GetNextPointTime(*query, currentTime)
		}
	}
	// numSteps := int((query.FixedEnd-query.FixedStart)/query.Interval) + 1

	// 用fixedstart和fixedend来做遍历
	for _, ts := range allTimes {
		// for ts := fixedStart; ts <= fixedEnd; ts = static.GetNextPointTime(*query, ts) {

		// Gather input vectors for this timestamp.
		for i := range exprs {
			vectors[i] = vectors[i][:0]

			if prepSeries != nil {
				bufHelpers[i] = bufHelpers[i][:0]
			}

			for si, series := range matrixes[i] {
				for _, point := range series.Points {
					if query.IsInstantQuery || point.T == ts {
						vectors[i] = append(vectors[i], static.Sample{Metric: series.Metric, Point: point})
						if prepSeries != nil {
							bufHelpers[i] = append(bufHelpers[i], seriesHelpers[i][si])
						}
						// Move input vectors forward so we don't have to re-scan the same
						// past points at the next step.
						matrixes[i][si].Points = series.Points[1:]
					}
					break
				}
			}
			args[i] = vectors[i]
		}

		// Make the function call.
		enh.Ts = ts
		result, status, err := funcCall(args, bufHelpers, enh)
		if err != nil {
			return nil, status, err
		}
		if result.ContainsSameLabelset() {
			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("vector cannot contain metrics with the same labelset"),
			}
		}
		enh.Out = result[:0] // Reuse result vector.

		// If this could be an instant query, shortcut so as not to change sort order.
		if query.FixedStart == query.FixedEnd {
			mat := make(static.Matrix, len(result))
			for i, s := range result {
				s.Point.T = ts
				mat[i] = static.Series{Metric: s.Metric, Points: []static.Point{s.Point}}
			}
			return mat, http.StatusOK, nil
		}

		// Add samples in output vector to output series.
		for _, sample := range result {
			h := sample.Metric.Hash()
			ss, ok := seriess[h]
			if !ok {
				ss = &static.Series{
					Metric: sample.Metric,
					Points: static.GetPointSlice(len(allTimes)), // 利用sync.pool, 复用已经使用过的对象.
				}
				seriessOrder = append(seriessOrder, ss)
			}
			sample.Point.T = ts
			ss.Points = append(ss.Points, sample.Point)
			seriess[h] = ss
		}
	}

	// Reuse the original point slices.
	for _, m := range origMatrixes {
		for _, s := range m {
			static.PutPointSlice(s.Points)
		}
	}
	// Assemble the output matrix. By the time we get here we know we don't have too many samples.
	mat := make(static.Matrix, 0, len(seriess))
	for _, ss := range seriessOrder {
		mat = append(mat, *ss)
	}
	return mat, http.StatusOK, nil
}

// 通过 match[] 查找对应的序列，返回所有匹配的序列列表，状态码，错误信息
func (ps *promQLService) Series(matchers interfaces.Matchers) ([]byte, int, error) {
	// 根据match构造dsl并查询序列列表
	res, status, err := ps.leafNodes.Series(&matchers)
	if err != nil {
		return nil, status, err
	}

	// 返回结果
	response := &interfaces.PromQLResponse{
		Status: "success",
		Data:   res,
	}
	bytes, err := sonic.Marshal(response)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: errors.New("promql_service.Series: Marshal response error: " + err.Error()),
		}
	}
	return bytes, http.StatusOK, nil
}

// eval 累积连续k分组不可用时长
func (ps *promQLService) evalKMinuteDowntime(ctx context.Context, expr *parser.Call, query *interfaces.Query) (parser.Value, int, error) {
	// 对于k-minute_downtime来说，其第一个参数表达式应用的step应为1min。开始时间往前推移k分钟
	// eval第一个参数表达式，用新的query来计算
	newQuery := *query
	kMinute := int64(expr.Args[0].(*parser.NumberLiteral).Val)

	// 先除以桶，然后按照时区偏移，promql默认使用的是local，不指定时区。
	_, offset := time.Now().In(common.APP_LOCATION).Zone()

	// 步长1min，范围查询
	var start int64
	if query.IsInstantQuery {
		// 计算长度区间
		// lookBackDelta := convert.GetLookBackDelta(query.LookBackDelta, ps.appSetting.PromqlSetting.LookbackDelta)
		lookBackDelta := query.End - query.Start
		start = query.FixedStart - lookBackDelta - kMinute*60*1000 // 再往前挪k-minute
		// 即时查询时的fixedStart按start回退look_bakc_delta之后按1min重新修正
		query.FixedStart = int64(math.Ceil(float64(query.FixedStart-lookBackDelta+int64(offset*1000))/float64(interfaces.KMINUTE_DOWNTTIME_STEP)))*interfaces.KMINUTE_DOWNTTIME_STEP - int64(offset*1000)
	} else {
		// 范围查询，往前再查一个步长或者kminute的数据，取较大者
		start = query.FixedStart - max(query.Interval, kMinute*60*1000)
	}
	newQuery.Interval = interfaces.KMINUTE_DOWNTTIME_STEP
	newQuery.Start = start
	newQuery.IsInstantQuery = false // 范围查询
	newQuery.NotNeedFilling = true  // 不需要补点

	// 一个序列能查的最大桶数是1w，若是超过，报错。
	if (newQuery.End-newQuery.Start)/newQuery.Interval > interfaces.DEFAULT_MAX_QUERY_POINTS {
		return nil, http.StatusUnprocessableEntity,
			errors.New(`when calculating the k-minute downtime of an object, 
				if the requested time range is excessively long, 
				exceeding the maximum number of buckets that can be queried within one sequence, 
				which defaults to 10,000, 
				this model recommends persisting the data by day and then utilizing 
				sum_over_time(<persisted_metric_name>[range_parameter]) to sum the downtime 
				for each day to obtain the total downtime within the calculation period`)
	}

	// newQuery的fixedstart和fixedEnd需重新计算，因为start和interval变化了。
	// newQuery.FixedStart = int64(math.Floor(float64(newQuery.Start+int64(offset*1000))/float64(newQuery.Interval)))*newQuery.Interval - int64(offset*1000)
	// newQuery.FixedEnd = int64(math.Floor(float64(newQuery.End+int64(offset*1000))/float64(newQuery.Interval)))*newQuery.Interval - int64(offset*1000)
	fixedStart, fixedEnd := static.CorrectingTime(newQuery, common.APP_LOCATION)
	newQuery.FixedStart = fixedStart
	newQuery.FixedEnd = fixedEnd

	matrixes := make([]static.Matrix, len(expr.Args)-3)
	// 1. 多个vector需要eval，逐个eval，然后再比较。第四个参数才是vector或者expression
	i := 0
	for j := 3; j < len(expr.Args); j++ {
		// Functions will take string arguments from the expressions, not the values.
		if expr != nil && expr.Type() != parser.ValueTypeString {
			// ev.currentSamples will be updated to the correct value within the ev.eval call.
			val, status, err := ps.eval(ctx, expr.Args[j], &newQuery) // 逐个eval表达式
			if err != nil {
				return nil, status, err
			}
			switch result := val.(type) {
			case static.Matrix:
				matrixes[i] = result
			case static.PageMatrix:
				matrixes[i] = result.Matrix
			}
		}
		i++
	}
	// 2. 先按填充策略对每个指标结果集填充。从newQUery的fixedStart 到 fixedEnd, 按1m的步长
	precedingMissingPolicy := int(expr.Args[1].(*parser.NumberLiteral).Val)
	middleMissingPolicy := int(expr.Args[2].(*parser.NumberLiteral).Val)
	matrixes = static.FillMissingPoint(newQuery, matrixes, precedingMissingPolicy, middleMissingPolicy)

	// 3. 合并（全序列join）计算整体的可用性（认为0是不可用，非0可用），都可用才认为可用。
	mat := static.CombineEvalUsability(matrixes, precedingMissingPolicy, middleMissingPolicy)

	// 4. 计算每个步长点或者是即时查询范围内的不可用时长
	mat = static.CalculateUnavailableTime(mat, *query, kMinute)

	return mat, 200, nil
}

// eval 指标模型 funcMetricModel
func (ps *promQLService) evalMetricModel(ctx context.Context, expr *parser.Call, query *interfaces.Query) (parser.Value, int, error) {
	// 从参数中解析出指标模型id
	modelId := expr.Args[0].(*parser.StringLiteral).Val

	// 获取模型的指标数据, metric_model引用promql，promql不能再引用metric_model了，会循环。promql对metric model的依赖在new metricModelService时注入
	// 获取到模型的统一格式，在此处，需把统一格式转成promql的格式，便于后续的计算
	uniResp, _, _, err := ps.mmService.Exec(ctx, &interfaces.MetricModelQuery{
		QueryTimeParams: interfaces.QueryTimeParams{
			Start:          &query.Start,
			End:            &query.End,
			StepStr:        &query.IntervalStr,
			Step:           &query.Interval,
			IsInstantQuery: query.IsInstantQuery,
		},
		MetricModelID: modelId,
		Filters:       query.Filters,
		MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
			IgnoringHCTS:        query.IgnoringHCTS,
			IgnoringMemoryCache: query.IgnoringMemoryCache,
		},
		IsModelRequest: query.IsModelRequest,
		IsCalendar:     query.IsCalendar,
		AnalysisDims:   query.AnalysisDims,
		OrderByFields:  query.OrderByFields,
	})
	if err != nil {
		return nil, 500, err
	}

	// 把统一结果转成promql的中间节点的结果 Matrix，流入到上层节点
	mat := make(static.Matrix, 0)
	for _, seriesI := range uniResp.Datas {
		// 先处理labels
		var metric labels.Labels
		for k, v := range seriesI.Labels {
			metric = append(metric, &labels.Label{
				Name:  k,
				Value: v,
			})
		}
		metric.Sort()
		// 处理t-v对变成point，如果是nil，则不拼接。节点中计算时不补空，补空是最后的行为
		points := make([]static.Point, 0)
		for i, t := range seriesI.Times {
			if seriesI.Values[i] == nil {
				continue
			}
			val, err := convert.AssertFloat64(seriesI.Values[i])
			if err != nil {
				continue
			}
			points = append(points, static.Point{
				T: t.(int64),
				V: val,
			})
		}

		// 存在点时才append序列
		if len(points) != 0 {
			mat = append(mat, static.Series{
				Metric: metric,
				Points: points,
			})
		}
	}

	// 记录vega耗时
	query.VegaDurationMs += uniResp.VegaDurationMs

	return mat, 200, nil
}

// 获取指标模型中promql使用到的指标的字段集（labels）
func (ps *promQLService) GetFields(ctx context.Context, query interfaces.Query) (result map[string]bool, status int, err error) {
	// 解析 query
	parseCtx, span := ar_trace.Tracer.Start(ctx, "解析 promql 表达式")

	expr, err := parser.ParseExpr(ctx, query.QueryStr)
	if err != nil {
		// 设置 trace 的错误信息的 attributes
		span.SetStatus(codes.Error, "Promql parser.ParseExpr error")
		span.End()
		// 记录接口调用参数
		o11y.Error(parseCtx, fmt.Sprintf("Promql [%s] parse error: [%v]", query.QueryStr, err))
		span.End()
		return nil, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: errors.New(err.Error()),
		}
	}
	span.SetStatus(codes.Ok, "")
	span.End()

	if expr.Type() != parser.ValueTypeVector && expr.Type() != parser.ValueTypeScalar && expr.Type() != parser.ValueTypeMatrix {
		return nil, http.StatusBadRequest,
			uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("invalid expression type %q for range query, must be Scalar,"+
					"range or instant Vector", parser.DocumentedType(expr.Type())),
			}
	}

	// 捕获function的panic
	defer ps.recover(&status, &err)

	// 执行 query
	return ps.evalFieldsInfo(ctx, expr, &query, "")
}

// 获取字段集或者获取指定字段的值集
func (ps *promQLService) evalFieldsInfo(ctx context.Context, expr parser.Expr,
	query *interfaces.Query, fieldName string) (fields map[string]bool, status int, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("eval 节点[%T] 获取字段", expr))
	span.SetAttributes(attribute.Key("promql_expression").String(fmt.Sprintf("%v", expr)))
	defer span.End()

	switch e := expr.(type) {
	case *parser.AggregateExpr:
		span.SetStatus(codes.Ok, "")
		return ps.evalFieldsInfo(ctx, e.Expr, query, fieldName)

	case *parser.Call:
		if e.Func == nil {
			span.SetStatus(codes.Error, "FunctionCalls is not defined")
			// 记录异常日志
			o11y.Error(ctx, "FunctionCalls is not defined, please input valid function.")

			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" FunctionCalls is not defined, please input valid function. "),
			}
		}

		// 这里拿到的时call对象
		if funcCall, exist := static.FunctionCalls[e.Func.Name]; !exist || funcCall.New(e.Args) == nil {
			span.SetStatus(codes.Error, fmt.Sprintf("'%s' is not currently supported", e.Func.Name))
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("'%s' is not currently supported.", e.Func.Name))
			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" '%s' is not currently supported. ", e.Func.Name),
			}
		}

		var (
			matrixArgIndex int
			matrixArg      bool
		)
		for i := range e.Args {
			static.UnwrapParenExpr(&e.Args[i])
			a := static.UnwrapStepInvariantExpr(e.Args[i])
			static.UnwrapParenExpr(&a)
			if _, ok := a.(*parser.MatrixSelector); ok {
				matrixArgIndex = i
				matrixArg = true
				break
			}
		}

		if !matrixArg {
			// Does not have a matrix argument.
			span.SetStatus(codes.Ok, "")
			return ps.evalFuncFields(ctx, query, fieldName, e.Args...)
		}

		static.UnwrapParenExpr(&e.Args[matrixArgIndex])
		arg := static.UnwrapStepInvariantExpr(e.Args[matrixArgIndex])
		static.UnwrapParenExpr(&arg)
		sel := arg.(*parser.MatrixSelector)

		vs, ok := sel.VectorSelector.(*parser.VectorSelector)
		if !ok {
			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("promql.evalFields: invalid expression type %q", sel.VectorSelector.Type()),
			}
		}
		span.SetStatus(codes.Ok, "")
		// eval range算子的vectoeSelector即可
		return ps.evalFieldsInfo(ctx, vs, query, fieldName)

	case *parser.ParenExpr:
		span.SetStatus(codes.Ok, "")
		return ps.evalFieldsInfo(ctx, e.Expr, query, fieldName)

	case *parser.UnaryExpr:
		span.SetStatus(codes.Ok, "")
		return ps.evalFieldsInfo(ctx, e.Expr, query, fieldName)

	case *parser.BinaryExpr:
		span.SetStatus(codes.Ok, "")
		// 二元运算，就把其左右节点往下eval，获取字段集，然后再取并集
		lFields, _, _ := ps.evalFieldsInfo(ctx, e.LHS, query, fieldName)
		rFields, _, _ := ps.evalFieldsInfo(ctx, e.RHS, query, fieldName)
		// 取并集
		return convert.Union(lFields, rFields), http.StatusOK, nil

	case *parser.NumberLiteral:
		span.SetStatus(codes.Ok, "")
		return map[string]bool{}, http.StatusOK, nil

	case *parser.VectorSelector:
		span.SetStatus(codes.Ok, "")
		return ps.leafNodes.EvalVectorSelectorFields(ctx, e, query, fieldName)

	case *parser.StepInvariantExpr:
		span.SetStatus(codes.Ok, "")
		return map[string]bool{}, http.StatusOK, nil
	}

	span.SetStatus(codes.Error, fmt.Sprintf("unhandled expression of type: %T", expr))
	// 记录异常日志
	o11y.Error(ctx, fmt.Sprintf("unhandled expression of type: %T", expr))
	return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("unhandled expression of type: %T", expr),
	}
}

// 获取函数中各参数的字段集或者指定字段的值集
func (ps *promQLService) evalFuncFields(ctx context.Context, query *interfaces.Query, fieldName string, exprs ...parser.Expr) (map[string]bool, int, error) {
	// eval函数的参数，对于dict_labels、dict_valeus，其字段暂时先不作为模型字段列表的一部分。
	// 一般函数时，字段是所有参数的并集
	fieldsArr := make([]map[string]bool, 0)
	for _, expr := range exprs {
		// Functions will take string arguments from the expressions, not the values.
		if expr != nil && expr.Type() != parser.ValueTypeString {
			// ev.currentSamples will be updated to the correct value within the ev.eval call.
			fields, status, err := ps.evalFieldsInfo(ctx, expr, query, fieldName) // 逐个eval表达式
			if err != nil {
				return nil, status, err
			}
			fieldsArr = append(fieldsArr, fields)
		}
	}
	// 取并集操作
	return convert.Union(fieldsArr...), http.StatusOK, nil
}

// 获取字段值，promql的labels字段全是字符串
func (ps *promQLService) GetFieldValues(ctx context.Context, query interfaces.Query, fieldName string) (result map[string]bool, status int, err error) {
	// 解析 query
	parseCtx, span := ar_trace.Tracer.Start(ctx, "解析 promql 表达式")

	expr, err := parser.ParseExpr(ctx, query.QueryStr)
	if err != nil {
		// 设置 trace 的错误信息的 attributes
		span.SetStatus(codes.Error, "Promql parser.ParseExpr error")
		span.End()
		// 记录接口调用参数
		o11y.Error(parseCtx, fmt.Sprintf("Promql [%s] parse error: [%v]", query.QueryStr, err))
		span.End()
		return nil, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: errors.New(err.Error()),
		}
	}
	span.SetStatus(codes.Ok, "")
	span.End()

	if expr.Type() != parser.ValueTypeVector && expr.Type() != parser.ValueTypeScalar && expr.Type() != parser.ValueTypeMatrix {
		return nil, http.StatusBadRequest,
			uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("invalid expression type %q for range query, must be Scalar,"+
					"range or instant Vector", parser.DocumentedType(expr.Type())),
			}
	}

	// 捕获function的panic
	defer ps.recover(&status, &err)

	// 执行 query
	return ps.evalFieldsInfo(ctx, expr, &query, fieldName)
}

// 获取promql表达式的计算结果的维度字段
func (ps *promQLService) GetLabels(ctx context.Context, query interfaces.Query) (result map[string]bool, status int, err error) {
	// 解析 query
	parseCtx, span := ar_trace.Tracer.Start(ctx, "解析 promql 表达式")

	expr, err := parser.ParseExpr(ctx, query.QueryStr)
	if err != nil {
		// 设置 trace 的错误信息的 attributes
		span.SetStatus(codes.Error, "Promql parser.ParseExpr error")
		span.End()
		// 记录接口调用参数
		o11y.Error(parseCtx, fmt.Sprintf("Promql [%s] parse error: [%v]", query.QueryStr, err))
		span.End()
		return nil, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: errors.New(err.Error()),
		}
	}
	span.SetStatus(codes.Ok, "")
	span.End()

	if expr.Type() != parser.ValueTypeVector && expr.Type() != parser.ValueTypeScalar && expr.Type() != parser.ValueTypeMatrix {
		return nil, http.StatusBadRequest,
			uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: fmt.Errorf("invalid expression type %q for range query, must be Scalar,"+
					"range or instant Vector", parser.DocumentedType(expr.Type())),
			}
	}

	// 捕获function的panic
	defer ps.recover(&status, &err)

	// 执行 query
	return ps.evalLabelsInfo(ctx, expr, &query)
}

func (ps *promQLService) evalLabelsInfo(ctx context.Context, expr parser.Expr, query *interfaces.Query) (fields map[string]bool, status int, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("eval 节点[%T] 获取字段", expr))
	span.SetAttributes(attribute.Key("promql_expression").String(fmt.Sprintf("%v", expr)))
	defer span.End()

	switch e := expr.(type) {
	case *parser.AggregateExpr:

		// 对字段做缩减，没有by时是空集，有by是是by里的字段
		aggLables, status, err := ps.evalLabelsInfo(ctx, e.Expr, query)
		if err != nil {
			return aggLables, status, err
		}
		switch e.Op {
		case parser.TOPK, parser.BOTTOMK:
			// 用 by/without 来对输入向量进行分桶，然后桶内按值降序取前 n 个，查询结果中的标签为 <vector expression> 中的原始标签。
			span.SetStatus(codes.Ok, "")
			return aggLables, http.StatusOK, nil
		}

		groupMap := make(map[string]bool)
		for _, group := range e.Grouping {
			groupMap[group] = true
		}
		// grouping字段不存在时，不添加到结果字段中
		labelsMap := make(map[string]bool)
		for label := range aggLables {
			if e.Without {
				// without为true，即排除without的字段进行分组, 不存在 grouping中的字段都保留
				if _, exist := groupMap[label]; !exist {
					labelsMap[label] = true
				}
			} else {
				// without为false，那就是by的字段，存在 grouping中的字段都保留
				if _, exist := groupMap[label]; exist {
					labelsMap[label] = true
				}
			}
		}
		span.SetStatus(codes.Ok, "")
		return labelsMap, http.StatusOK, nil

	case *parser.Call:
		if e.Func == nil {
			span.SetStatus(codes.Error, "FunctionCalls is not defined")
			// 记录异常日志
			o11y.Error(ctx, "FunctionCalls is not defined, please input valid function.")

			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" FunctionCalls is not defined, please input valid function. "),
			}
		}

		// 这里拿到的时call对象
		if funcCall, exist := static.FunctionCalls[e.Func.Name]; !exist || funcCall.New(e.Args) == nil {
			span.SetStatus(codes.Error, fmt.Sprintf("'%s' is not currently supported", e.Func.Name))
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("'%s' is not currently supported.", e.Func.Name))
			return nil, http.StatusBadRequest, uerrors.PromQLError{
				Typ: uerrors.ErrorBadData,
				Err: fmt.Errorf(" '%s' is not currently supported. ", e.Func.Name),
			}
		}

		// 会影响labels的函数：dict_labels:扩展字段，dict_values：作为原始值的字段，
		// label_join：, label_replace：,continuous_k_minute_downtime：参数中多个表达式的并集

		switch e.Func.Name {
		case interfaces.DICT_LABELS:
			// dict_labels函数中把扩展的维度字段写在了expendLables中了
			// 先把eval第一个参数vector，然后再append expendLables即可
			vectorLabels, status, err := ps.evalLabelsInfo(ctx, e.Args[0], query)
			if err != nil {
				return vectorLabels, status, err
			}
			for i := 2; i < len(e.Args); i = i + 2 {
				// 偶数位的参数是维度名称定义
				vectorLabel := string(e.Args[i+1].(*parser.StringLiteral).Val)
				vectorLabels[vectorLabel] = true
			}
			return vectorLabels, http.StatusOK, nil
		case interfaces.DICT_VALUES:
			// dict_values
			labelsMap := make(map[string]bool)
			for i := 2; i < len(e.Args); i = i + 2 {
				vectorLabel := string(e.Args[i+1].(*parser.StringLiteral).Val)
				labelsMap[vectorLabel] = true
			}
			return labelsMap, http.StatusOK, nil

		case interfaces.LABEL_JOIN, interfaces.LABEL_REPLACE:
			// label_join 用于将多个标签的值连接起来，生成一个新的标签。dst_label: 目标标签名称（如果不存在则会创建）
			// dst_label如果不存在则会创建一个新的
			labelsMap, status, err := ps.evalLabelsInfo(ctx, e.Args[0], query)
			if err != nil {
				return nil, status, err
			}
			// 追加dst_labels, 在 args[1]
			labelsMap[static.StringFromArg(e.Args[1])] = true
			return labelsMap, http.StatusOK, nil
		}

		// 其他的函数不影响结果维度，eval各参数取并集即可直接继续往下遍历
		span.SetStatus(codes.Ok, "")
		return ps.evalFuncLabels(ctx, query, e.Args...)

	case *parser.ParenExpr:
		// 用于明确表达式的计算顺序的，比如 (a+b), 不影响计算的结果维度，直接返回
		span.SetStatus(codes.Ok, "")
		return ps.evalLabelsInfo(ctx, e.Expr, query)

	case *parser.UnaryExpr:
		// 一元操作符，例如 -a。不银杏果计算的结果维度，直接返回
		span.SetStatus(codes.Ok, "")
		return ps.evalLabelsInfo(ctx, e.Expr, query)

	case *parser.BinaryExpr:
		span.SetStatus(codes.Ok, "")
		// 二元运算，就把其左右节点往下eval，获取字段集，然后再取交集
		lFields, _, _ := ps.evalLabelsInfo(ctx, e.LHS, query)
		rFields, _, _ := ps.evalLabelsInfo(ctx, e.RHS, query)
		// 根据操作符的 on, ignoring, left_join, out_join来确定要做的操作
		labelsMap := make(map[string]bool)
		if e.VectorMatching != nil {
			if e.VectorMatching.On {
				// on: on的字段需要在两边中全都存在
				for _, lb := range e.VectorMatching.MatchingLabels {
					_, lExist := lFields[lb]
					_, rExist := rFields[lb]
					if lExist && rExist {
						labelsMap[lb] = true
					} else {
						// 如果有一个on字段在其中一边中不存在，则返回空
						return map[string]bool{}, http.StatusOK, nil
					}
				}
				return labelsMap, http.StatusOK, nil
			}

			if e.VectorMatching.LeftJoin {
				// left_join 保留左侧所有的维度，左侧指标按left_join的字段从右侧寻找匹配，如果匹配到多条，按group_left的字段从右侧指标中获取字段扩展到左侧中
				// 所以结果维度字段是：为左侧的字段 + group_left的字段
				// 如果所有的关联字段在右都不存在，则只保留左侧的，忽略group left的字段
				ifNeedGroupLeft := false
				for _, lb := range e.VectorMatching.MatchingLabels {
					_, rExist := rFields[lb]
					if rExist {
						ifNeedGroupLeft = true
						break
					}
				}

				if ifNeedGroupLeft {
					for _, lb := range e.VectorMatching.Include {
						if _, rExist := rFields[lb]; rExist {
							lFields[lb] = true
						}
					}
				}
				return lFields, http.StatusOK, nil
			}
			if e.VectorMatching.OutJoin {
				// out_join中的字段在其中一边存在，select_left字段在左边中存在，select_right字段在右边中存在，都应该输出到结果中
				for _, lb := range e.VectorMatching.MatchingLabels {
					_, lExist := lFields[lb]
					_, rExist := rFields[lb]
					if lExist || rExist {
						labelsMap[lb] = true
					}
				}
				for _, lb := range e.VectorMatching.IncludeLeft {
					if _, lExist := lFields[lb]; lExist {
						labelsMap[lb] = true
					}
				}
				for _, lb := range e.VectorMatching.IncludeRight {
					if _, rExist := rFields[lb]; rExist {
						labelsMap[lb] = true
					}
				}
				return labelsMap, http.StatusOK, nil
			}

			if !e.VectorMatching.On && len(e.VectorMatching.MatchingLabels) > 0 {
				// ignoring: 除了ignoring的字段外，其余字段需要在两边中全都存在
				excludeKeys := make(map[string]bool)
				for _, lb := range e.VectorMatching.MatchingLabels {
					excludeKeys[lb] = true
				}

				// 检查 lFields 和 rFields 的 key 是否一致
				for key := range lFields {
					if !excludeKeys[key] { // 如果 key 不在排除列表中
						if _, exists := rFields[key]; !exists { // 如果 rFields 中没有这个 key
							return nil, http.StatusOK, nil
						} else {
							labelsMap[key] = true
						}
					}
				}

				for key := range rFields {
					if !excludeKeys[key] { // 如果 key 不在排除列表中
						if _, exists := lFields[key]; !exists { // 如果 lFields 中没有这个 key
							return nil, http.StatusOK, nil
						} else {
							labelsMap[key] = true
						}
					}
				}

				return labelsMap, http.StatusOK, nil
			}
		}

		// 未指定关联关系时，如果两边都含有__all,则返回空；如果其中一方还有__all，则返回另一方
		if _, exists := lFields[interfaces.ALL_LABELS_FLAG]; exists {
			if _, exists := rFields[interfaces.ALL_LABELS_FLAG]; exists {
				return nil, http.StatusOK, nil
			}
			return rFields, http.StatusOK, nil
		}
		if _, exists := rFields[interfaces.ALL_LABELS_FLAG]; exists {
			return lFields, http.StatusOK, nil
		}
		// 左右两边的字段都需要在左右两边中都存在，否则返回空
		for key := range lFields {
			if _, exists := rFields[key]; !exists {
				return nil, http.StatusOK, nil
			}
		}
		for key := range rFields {
			if _, exists := lFields[key]; !exists {
				return nil, http.StatusOK, nil
			}
		}
		return convert.Union(lFields, rFields), http.StatusOK, nil

	case *parser.NumberLiteral:
		// 数字，无维度，返回 __all, 当在做交集操作时，把其忽略，作为全集，不改变具体聚合结果
		span.SetStatus(codes.Ok, "")
		return map[string]bool{interfaces.ALL_LABELS_FLAG: true}, http.StatusOK, nil

	case *parser.VectorSelector:
		span.SetStatus(codes.Ok, "")
		// 返回的是包含labels.前缀的原始字段
		fields, status, err := ps.leafNodes.EvalVectorSelectorFields(ctx, e, query, "")
		if err != nil {
			return fields, status, err
		}
		// 维度字段不需要输出labels.
		labelsMap := make(map[string]bool)
		for field := range fields {
			labelsMap[field] = true
		}
		return labelsMap, http.StatusOK, nil

	case *parser.StepInvariantExpr:
		span.SetStatus(codes.Ok, "")
		return map[string]bool{interfaces.ALL_LABELS_FLAG: true}, http.StatusOK, nil
	case *parser.MatrixSelector:
		span.SetStatus(codes.Ok, "")
		return ps.evalLabelsInfo(ctx, e.VectorSelector, query)
	}

	span.SetStatus(codes.Error, fmt.Sprintf("unhandled expression of type: %T", expr))
	// 记录异常日志
	o11y.Error(ctx, fmt.Sprintf("unhandled expression of type: %T", expr))
	return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
		Typ: uerrors.ErrorExec,
		Err: fmt.Errorf("unhandled expression of type: %T", expr),
	}
}

// 获取函数中各参数的维度集
func (ps *promQLService) evalFuncLabels(ctx context.Context, query *interfaces.Query, exprs ...parser.Expr) (map[string]bool, int, error) {
	// eval函数的参数，对于dict_labels、dict_valeus，其字段暂时先不作为模型字段列表的一部分。
	// 一般函数时，字段是所有参数的并集
	fieldsArr := make([]map[string]bool, 0)
	for _, expr := range exprs {
		// Functions will take string arguments from the expressions, not the values.
		if expr != nil && expr.Type() != parser.ValueTypeString {
			// ev.currentSamples will be updated to the correct value within the ev.eval call.
			fields, status, err := ps.evalLabelsInfo(ctx, expr, query) // 逐个eval表达式
			if err != nil {
				return nil, status, err
			}
			fieldsArr = append(fieldsArr, fields)
		}
	}
	// 取并集操作
	return convert.Union(fieldsArr...), http.StatusOK, nil
}
