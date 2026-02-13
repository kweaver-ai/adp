// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dsl

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	dslServiceOnce sync.Once
	dService       interfaces.DslService
)

type dslService struct {
	appSetting *common.AppSetting
	lgAccess   interfaces.LogGroupAccess
	osClient   interfaces.OpenSearchAccess
}

func NewDslService(appSetting *common.AppSetting) interfaces.DslService {
	dslServiceOnce.Do(func() {
		dService = &dslService{
			appSetting: appSetting,
			lgAccess:   logics.LGAccess,
			osClient:   logics.OSAccess,
		}
	})
	return dService
}

// Search dsl查询
func (ds *dslService) Search(ctx context.Context, dsl map[string]interface{}, indicesAlias string, scroll time.Duration) ([]byte, int, error) {
	var library []string
	var err error
	dsl, library, err = ds.getLibrary(dsl, indicesAlias)
	if err != nil {
		return nil, http.StatusBadRequest, err
	} else if library == nil {
		// 如果 path 和 body 中的日志库交集为空，那么直接返回空。
		_, exit := dsl["x_library"]
		if indicesAlias != "" || exit {
			return convert.MapToByte(interfaces.DslResult), http.StatusOK, nil
		}
		// 日志库参数为空，则取日志分组参数
		var status int
		library, status, err = ds.spliceMustFilters(dsl)
		if err != nil {
			return nil, status, err
		} else if len(library) == 0 {
			return convert.MapToByte(interfaces.DslResult), http.StatusOK, nil
		}
	}

	// 把 dsl 中的 ar_dataview 删除
	delete(dsl, "ar_dataview")
	res, status, err := ds.osClient.SearchSubmit(ctx, dsl, library, scroll,
		interfaces.DEFAULT_PREFERENCE, true)
	return res, status, err
}

// ScrollSearch scroll查询
func (ds *dslService) ScrollSearch(ctx context.Context, scroll interfaces.Scroll) ([]byte, int, error) {
	res, status, err := ds.osClient.Scroll(ctx, scroll)
	return res, status, err
}

// Count 获取查询数据的总数
func (ds *dslService) Count(ctx context.Context, dsl map[string]interface{}, indicesAlias string) ([]byte, int, error) {
	var library []string
	var err error
	dsl, library, err = ds.getLibrary(dsl, indicesAlias)
	if err != nil {
		return nil, http.StatusBadRequest, err
	} else if library == nil {
		// 如果 path 和 body 中的日志库交集为空，那么直接返回空。
		_, exit := dsl["x_library"]
		if indicesAlias != "" || exit {
			return convert.MapToByte(interfaces.DslCount), http.StatusOK, nil
		}
		// 日志库参数为空，则取日志分组参数
		var status int
		library, status, err = ds.spliceMustFilters(dsl)
		if err != nil {
			return nil, status, err
		} else if len(library) == 0 {
			return convert.MapToByte(interfaces.DslCount), http.StatusOK, nil
		}
	}
	// 把 dsl 中的 ar_dataview 删除
	delete(dsl, "ar_dataview")
	res, status, err := ds.osClient.Count(ctx, dsl, library)
	return res, status, err
}

// getLibrary 处理pathlibrary和x_library
func (ds *dslService) getLibrary(dsl map[string]interface{}, indicesAlias string) (map[string]interface{}, []string, error) {
	var library []string
	pathLib := strings.Split(indicesAlias, ",")
	tempStr, exit := dsl["x_library"]
	if indicesAlias == "" && !exit {
		return dsl, library, nil
	} else if indicesAlias != "" && exit {
		bodyLib := convert.InterToArray(tempStr)
		if bodyLib == nil {
			oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("x_library must be array")
			return dsl, library, oerr
		}
		library = convert.IntersectArray(pathLib, bodyLib)
	} else if exit {
		// indicesAlias == ""
		bodyLib := convert.InterToArray(tempStr)
		if bodyLib == nil {
			oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("x_library must be array")
			return dsl, library, oerr
		}
		library = bodyLib
	} else {
		// indicesAlias != "" and exit == false
		library = pathLib
	}

	if len(library) == 0 {
		return dsl, nil, nil
	}
	for index := range library {
		library[index] += "-*"
		library = append(library, "mdl-"+library[index])

	}

	delete(dsl, "x_library")
	return dsl, library, nil
}

// DeleteScroll 删除scroll查询
func (ds *dslService) DeleteScroll(ctx context.Context, deleteScroll interfaces.DeleteScroll) ([]byte, int, error) {
	res, status, err := ds.osClient.DeleteScroll(ctx, deleteScroll)
	return res, status, err
}

// spliceMustFilters 拼接日志分组的过滤条件到 dsl 中，以及根据日志分组获取索引列表
func (ds *dslService) spliceMustFilters(dsl map[string]interface{}) ([]string, int, error) {
	// 如果 library 为空，则从 body 中读取 ar_dataview 参数。如果 ar_dataview 也为空，则返回空集。
	var library []string
	arDataview, exist := dsl["ar_dataview"]
	if !exist {
		// library 为空 && ar_dataview 不存在，则返回空集
		return library, http.StatusOK, nil
	} else {
		// ar_dataview 存在，根据 id 请求 data-manager 获取日志分组的过滤条件
		arDataviewId, ok := arDataview.(string)
		if !ok {
			return library, http.StatusBadRequest, uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).
				WithReason("ar_dataview must be string")
		}
		// 如果日志分组为空，则返回空
		if arDataviewId == "" {
			return library, http.StatusOK, nil
		}

		dataview, _, err := ds.lgAccess.GetLogGroupQueryFilters(arDataviewId)
		if err != nil {
			return library, http.StatusInternalServerError, uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason(err.Error())
		}
		library = dataview.IndexPattern

		// 把 dsl 中的 ar_dataview 删除
		delete(dsl, "ar_dataview")
		// 拼接 mustFilters 到 dsl 中
		must := dataview.MustFilter.([]interface{})
		query, exists := dsl["query"].(map[string]interface{})
		if !exists {
			// 如果没有 query，则把当前的直接拼接上
			dsl["query"] = map[string]interface{}{
				"bool": map[string]interface{}{
					"must": must,
				},
			}
		} else {
			// 如果存在，则判断bool是否存在，如果不存在，则说明是单个过滤条件的情况。此时应把其和mustfilters合并到bool中
			boolStr, exists := query["bool"].(map[string]interface{})
			if !exists {
				// 把 mustfilters 和 dsl 的请求一起拼到 must 数组中
				must = append(must, query)
				// 清空query，然后重新组装
				delete(dsl, "query")
				dsl["query"] = map[string]interface{}{
					"bool": map[string]interface{}{
						"must": must,
					},
				}
			} else {
				// 如果 bool 存在，判断是否有filters，有就往must里append, 还需要判断 must 不是数组的场景，因为一个可以不写成数组
				dslMust, exists := boolStr["must"]
				if !exists {
					// 如果不存在，塞一个新的must进去
					boolStr["must"] = must
				} else {
					// must 是数组的话，就往数组里append；如果不是数组，则需要把must转成数组
					dslMustArr, ok := dslMust.([]interface{})
					if ok {
						must = append(must, dslMustArr...)
						boolStr["must"] = must
					} else {
						must = append(must, dslMust)
						// 清空query，然后重新组装
						boolStr["must"] = must
					}
				}
			}
		}
		return library, http.StatusOK, nil
	}
}
