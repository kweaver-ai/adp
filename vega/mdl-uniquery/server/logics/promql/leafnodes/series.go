// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/tidwall/gjson"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/util"
)

// 序列查询。返回与序列选择器集合匹配的序列列表。
func (leafNodes *LeafNodes) Series(matchers *interfaces.Matchers) ([]labels.Labels, int, error) {
	// 按 matcherSet 并发
	mapResultChs := make(chan *MapResult, len(matchers.MatcherSet))
	defer close(mapResultChs)

	var wg sync.WaitGroup
	wg.Add(len(matchers.MatcherSet))
	for _, v := range matchers.MatcherSet {
		// Submit tasks one by one.
		err := util.ExecutePool.Submit(leafNodes.seriesTaskFuncWapper(v, matchers, mapResultChs, &wg))
		if err != nil {
			return nil, http.StatusInternalServerError, uerrors.PromQLError{
				Typ: uerrors.ErrorExec,
				Err: errors.New("ExecutePool.Submit error: " + err.Error()),
			}
		}
	}
	// 等待所有执行结束
	wg.Wait()

	// 按 key 把合并各个 match[] 的 LabelsMap 为一个map，labelsMap 中的 key 是按 __labels_str 去重后得到的，所以最后解析 label_str 为 labels[]
	labelsMap := make(map[string][]*labels.Label)
	for j := 0; j < len(matchers.MatcherSet); j++ {
		mapResulti := <-mapResultChs
		if mapResulti.Err.Err != nil {
			return nil, mapResulti.Status, mapResulti.Err
		}
		for k, v := range mapResulti.LabelsMap {
			labelsMap[k] = v
		}
	}
	labelsArr := make([]labels.Labels, 0, len(labelsMap))
	// 对 map 的 key 排序，使得同一个请求输出的结果不会乱序
	keys := []string{}
	for key := range labelsMap {
		keys = append(keys, key)
	}
	// 对 keys 排序，从小到大
	sort.Strings(keys)
	for _, k := range keys {
		labelsArr = append(labelsArr, parseLabelsStr(k, labelsMap))
	}
	return labelsArr, http.StatusOK, nil
}

// 序列选择器的执行函数。根据序列选择器和请求的时间范围，构造dsl，查询。
func (leafNodes *LeafNodes) seriesTaskFuncWapper(matcher []*labels.Matcher, matchers *interfaces.Matchers,
	mapResultChs chan<- *MapResult, wg *sync.WaitGroup) taskFunc {

	return func() {
		defer wg.Done()

		// 请求 data-manager 获取日志分组的索引信息
		logGroup, status, err := leafNodes.GetLogGroupQueryFilters(context.Background(), matchers.LogGroupId)
		if err != nil {
			mapResultChs <- &MapResult{
				Status: status,
				Err: uerrors.PromQLError{
					Typ: err.(uerrors.PromQLError).Typ,
					Err: err.(uerrors.PromQLError).Err,
				},
			}
			return
		}

		// 构造 dsl
		queryBuffer, status, err := seriesDSL(matcher, matchers, leafNodes.appSetting.PromqlSetting.MaxSearchSeriesSize, logGroup.MustFilter)
		if err != nil {
			mapResultChs <- &MapResult{
				Status: status,
				Err: uerrors.PromQLError{
					Typ: err.(uerrors.PromQLError).Typ,
					Err: err.(uerrors.PromQLError).Err,
				},
			}
			return
		}

		// var dsl map[string]interface{}
		// err = sonic.Unmarshal(queryBuffer.Bytes(), &dsl)
		// if err != nil {
		// 	mapResultChs <- &MapResult{
		// 		Status: http.StatusUnprocessableEntity,
		// 		Err: uerrors.PromQLError{
		// 			Typ: uerrors.ErrorExec,
		// 			Err: errors.New("series.seriesTaskFuncWapper: Unmarshal dsl error: " + err.Error()),
		// 		},
		// 	}
		// 	return
		// }

		// 如果 indexPattern 为空，则直接返回空
		if len(logGroup.IndexPattern) <= 0 {
			mapResultChs <- &MapResult{
				LabelsMap:  make(map[string][]*labels.Label),
				TsValueMap: make(map[string][][]gjson.Result),
			}
			return
		}

		// 执行 dsl
		res, _, err := leafNodes.osAccess.SearchSubmitWithBuffer(context.Background(), *queryBuffer,
			logGroup.IndexPattern, 0, interfaces.DEFAULT_PREFERENCE)
		// 错误处理
		if err != nil {
			mapResultChs <- &MapResult{
				Status: http.StatusInternalServerError,
				Err: uerrors.PromQLError{
					Typ: uerrors.ErrorInternal,
					Err: errors.New(err.Error()),
				},
			}
			return
		}

		// 对各个 match[] 的结果做合并去重
		mapResult := &MapResult{
			LabelsMap:  make(map[string][]*labels.Label),
			TsValueMap: make(map[string][][]gjson.Result),
		}
		json := string(res)
		aggregations := gjson.Get(json, "aggregations")
		iteratorTermsAgg(aggregations, "", make([]*labels.Label, 0), 0,
			[]string{interfaces.LABELS_STR}, mapResult)

		mapResultChs <- mapResult
	}
}

// 序列查询的dsl
func seriesDSL(matcher []*labels.Matcher, matchers *interfaces.Matchers, maxSearchSeriesSize int, mustFilter interface{}) (*bytes.Buffer, int, error) {
	// 根据 matcher 构造 dsl。 服用 makeDSLQuery，利用 vectorSelector 来实现
	vs := parser.VectorSelector{
		LabelMatchers: matcher,
	}

	var queryBuffer bytes.Buffer
	// 构造 query 部分
	queryStr, _ := makeDSLQuery(vs)

	queryBuffer.WriteString(queryStr)

	// 构造must_filter
	mstr, err := sonic.Marshal(mustFilter)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: errors.New("series.seriesDSL Marshal must_filter error: " + err.Error()),
		}
	}

	// 构造 time range && aggs
	timeRangeStr := fmt.Sprintf(`
					{
						"range": {
							"@timestamp": {
								"gte":%d,
								"lte":%d
							}
						}
					}
				],
				"must":
					%s
			}
		},
		"aggs": {
			"%s": {
				"terms": {
					"field": "%s",
					"size": %d
				}
			}
		}
	}`, matchers.Start, matchers.End, string(mstr),
		interfaces.LABELS_STR,
		wrapKeyWordFieldName(interfaces.LABELS_STR),
		maxSearchSeriesSize)
	queryBuffer.WriteString(timeRangeStr)
	return &queryBuffer, http.StatusOK, nil
}
