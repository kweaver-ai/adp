// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dsl

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common/convert"
	"uniquery/interfaces"
	umock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func mockNewDslService(
	openSearchAccess interfaces.OpenSearchAccess,
	LogGroupAccess interfaces.LogGroupAccess,
) (ds *dslService) {
	ds = &dslService{
		osClient: openSearchAccess,
		lgAccess: LogGroupAccess,
	}
	return ds
}

func TestDslServiceScrollSearch(t *testing.T) {

	Convey("Test scroll search", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockOpenSearchService := umock.NewMockOpenSearchAccess(mockCtrl)
		ds := mockNewDslService(mockOpenSearchService, nil)

		Convey("scroll", func() {
			mockOpenSearchService.EXPECT().Scroll(gomock.Any(), gomock.Any()).Return(
				convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			scroll := interfaces.Scroll{
				ScrollId: "12324",
			}

			_, status, err := ds.ScrollSearch(testCtx, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

func TestDslServiceSearch(t *testing.T) {

	Convey("Test log service search", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockOpenSearchService := umock.NewMockOpenSearchAccess(mockCtrl)
		mockLogGroupAccess := umock.NewMockLogGroupAccess(mockCtrl)
		ds := mockNewDslService(mockOpenSearchService, mockLogGroupAccess)

		Convey("x_library illegal,param_library exist", func() {
			dsl := map[string]interface{}{
				"x_library": "",
				"size":      10,
			}
			indicesAlias := "kc,123"
			scroll := time.Duration(0)

			res, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})
		Convey("x_library exist but illegal,param_library not exist", func() {
			dsl := map[string]interface{}{
				"x_library": "234",
				"size":      10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			res, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("x_library,param_library,ar_dataview all not exist", func() {
			dsl := map[string]interface{}{
				"size": 10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("x_library exist && param_library,ar_dataview not exist", func() {
			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
				"size":      10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library exist && x_library,ar_dataview not exist", func() {
			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"size": 10,
			}
			indicesAlias := "kc"
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview exist && x_library,param_library not exist", func() {
			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)
			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"size":        10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library and ar_dataview exists && x_library not exist", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"size":        10,
			}
			indicesAlias := "kc"
			scroll := time.Duration(0)

			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library and x_library exists && ar_dataview not exist", func() {
			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
				"size":      10,
			}
			indicesAlias := "kc"
			scroll := time.Duration(0)

			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview and x_library exists && param_library not exist", func() {
			dsl := map[string]interface{}{
				"x_library":   []interface{}{"kc"},
				"ar_dataview": "123",
				"size":        10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview, x_library, param_library all exists", func() {
			dsl := map[string]interface{}{
				"x_library":   []interface{}{"kc"},
				"ar_dataview": "123",
				"size":        10,
			}
			indicesAlias := "kc"
			scroll := time.Duration(0)

			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("x_library,param_library intersectArray empty", func() {
			dsl := map[string]interface{}{
				"x_library": []interface{}{"234"},
				"size":      10,
			}
			indicesAlias := "kc,123"
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("x_library,param_library exist,but not be equal", func() {
			mockOpenSearchService.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(convert.MapToByte(interfaces.DslResult), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc", "ar_audit_log", "123"},
				"size":      10,
			}
			indicesAlias := "kc,123"
			scroll := time.Duration(0)

			_, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("x_library, param_library not exists && GetLogGroupQueryFilters error", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"size":        10,
			}
			indicesAlias := ""
			scroll := time.Duration(0)

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get queryfilters failed"))

			res, status, err := ds.Search(testCtx, dsl, indicesAlias, scroll)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":400,"error":{"type":"UniQuery.IllegalArgumentException","reason":"get queryfilters failed"}}`)
		})
	})
}

func TestDslServiceCount(t *testing.T) {

	Convey("Test log service count", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockOpenSearchService := umock.NewMockOpenSearchAccess(mockCtrl)
		mockLogGroupAccess := umock.NewMockLogGroupAccess(mockCtrl)
		ds := mockNewDslService(mockOpenSearchService, mockLogGroupAccess)

		Convey("x_library illegal,param_library exist", func() {
			dsl := map[string]interface{}{
				"x_library": "",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			res, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})
		Convey("x_library exist but illegal,param_library not exist", func() {
			dsl := map[string]interface{}{
				"x_library": "234",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := ""

			res, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
		})

		Convey("x_library,param_library,ar_dataview all not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := ""

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("x_library exist && param_library,ar_dataview not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"x_library": []interface{}{"kc"},
			}
			indicesAlias := ""
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library exist && x_library,ar_dataview not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc"
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview exist && x_library,param_library not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"ar_dataview": "123",
			}
			indicesAlias := ""
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)
			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library and ar_dataview exists && x_library not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"ar_dataview": "123",
			}
			indicesAlias := "kc"
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("param_library and x_library exists && ar_dataview not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"x_library": []interface{}{"kc"},
			}
			indicesAlias := "kc"
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview and x_library exists && param_library not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"x_library":   []interface{}{"kc"},
				"ar_dataview": "123",
			}
			indicesAlias := ""
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("ar_dataview, x_library, param_library all exists", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"x_library":   []interface{}{"kc"},
				"ar_dataview": "123",
			}
			indicesAlias := "kc"
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("x_library,param_library intersectArray empty", func() {
			dsl := map[string]interface{}{
				"x_library": []interface{}{"234"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("x_library,param_library exist,but not be equal", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc", "ar_audit_log", "123"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("x_library exist,param_library not exist,and dsl only contain x_library", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)
			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
			}
			indicesAlias := ""

			_, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("x_library, param_library not exists && GetLogGroupQueryFilters error", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"ar_dataview": "123",
			}
			indicesAlias := ""

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get queryfilters failed"))

			res, status, err := ds.Count(testCtx, dsl, indicesAlias)
			So(res, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":400,"error":{"type":"UniQuery.IllegalArgumentException","reason":"get queryfilters failed"}}`)
		})
	})
}

func TestDslServiceGetLibrary(t *testing.T) {
	Convey("Test log service getlibrary", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockOpenSearchService := umock.NewMockOpenSearchAccess(mockCtrl)
		ds := mockNewDslService(mockOpenSearchService, nil)

		Convey("x_library,param_library both not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := ""

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library illegal,param_library exist", func() {
			dsl := map[string]interface{}{
				"x_library": "",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("x_library exist but illegal,param_library not exist", func() {
			dsl := map[string]interface{}{
				"x_library": "234",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := ""

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("param_library exist", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc"

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library,param_library exist", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc"

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library exist,param_library not exist", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := ""

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library,param_library intersectArray empty", func() {
			dsl := map[string]interface{}{
				"x_library": []interface{}{"234"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library,param_library exist,but not be equal", func() {
			mockOpenSearchService.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				convert.MapToByte(interfaces.DslCount), http.StatusOK, nil)

			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc", "ar_audit_log", "123"},
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			indicesAlias := "kc,123"

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldNotBeNil)
			So(library, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
		Convey("x_library exist,param_library not exist,and dsl only contain x_library", func() {
			dsl := map[string]interface{}{
				"x_library": []interface{}{"kc"},
			}
			indicesAlias := ""

			res, library, err := ds.getLibrary(dsl, indicesAlias)
			So(res, ShouldResemble, map[string]interface{}{})
			So(library, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestDslServiceDeleteScroll(t *testing.T) {

	Convey("Test delete scroll", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockOpenSearchService := umock.NewMockOpenSearchAccess(mockCtrl)
		ds := mockNewDslService(mockOpenSearchService, nil)

		Convey("delete scroll success", func() {
			mockOpenSearchService.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).Return(
				convert.MapToByte(interfaces.DslDeleteScrollResult), http.StatusOK, nil)

			deleteScroll := interfaces.DeleteScroll{
				ScrollId: []string{"_all"},
			}

			_, status, err := ds.DeleteScroll(testCtx, deleteScroll)
			//So(res, ShouldResemble, convert.MapToByte(interfaces.DslDeleteScrollResult))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
	})
}

var (
	dataViewQueryFilters = interfaces.LogGroup{
		IndexPattern: []string{
			"indexPattern",
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

	expectDsl = map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"query_string": map[string]interface{}{
							"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
							"analyze_wildcard": true,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"labels.instance.keyword": map[string]interface{}{
								"value": "12",
							},
						},
					},
				},
			},
		},
	}
)

func TestDslServiceSpliceMustFilters(t *testing.T) {
	Convey("Test dsl service spliceMustFilters", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockLogGroupAccess := umock.NewMockLogGroupAccess(mockCtrl)

		ds := mockNewDslService(nil, mockLogGroupAccess)

		Convey("ar_dataview not exist", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			var expect []string

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, expect)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview exist but GetLogGroupQueryFilters error", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"ar_dataview": "123",
			}
			var expect []string

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				interfaces.LogGroup{}, false, fmt.Errorf("get queryfilters failed"))

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, expect)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `{"status":400,"error":{"type":"UniQuery.IllegalArgumentException","reason":"get queryfilters failed"}}`)
		})

		Convey("ar_dataview is int ", func() {
			dsl := map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"ar_dataview": 123,
			}
			var expect []string

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, expect)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `{"status":400,"error":{"type":"UniQuery.IllegalArgumentException","reason":"ar_dataview must be string"}}`)
		})

		Convey("ar_dataview exist but query not exists", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
			}

			expectDsl := map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{
								"query_string": map[string]interface{}{
									"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
									"analyze_wildcard": true,
								},
							},
						},
					},
				},
			}

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, []string{"indexPattern"})
			So(dsl, ShouldResemble, expectDsl)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview exist but bool not exists", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"query": map[string]interface{}{
					"term": map[string]interface{}{
						"labels.instance.keyword": map[string]interface{}{
							"value": "12",
						},
					},
				},
			}

			expectDsl := map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{
								"query_string": map[string]interface{}{
									"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
									"analyze_wildcard": true,
								},
							},
							map[string]interface{}{
								"term": map[string]interface{}{
									"labels.instance.keyword": map[string]interface{}{
										"value": "12",
									},
								},
							},
						},
					},
				},
			}

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, []string{"indexPattern"})
			So(dsl, ShouldResemble, expectDsl)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview exist but must not exists", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"filter": []interface{}{
							map[string]interface{}{
								"term": map[string]interface{}{
									"labels.instance.keyword": map[string]interface{}{
										"value": "12",
									},
								},
							},
						},
					},
				},
			}

			expectDslc := map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{
								"query_string": map[string]interface{}{
									"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
									"analyze_wildcard": true,
								},
							},
						},
						"filter": []interface{}{
							map[string]interface{}{
								"term": map[string]interface{}{
									"labels.instance.keyword": map[string]interface{}{
										"value": "12",
									},
								},
							},
						},
					},
				},
			}

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, []string{"indexPattern"})
			So(dsl, ShouldResemble, expectDslc)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview exist but must is a array", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{
								"term": map[string]interface{}{
									"labels.instance.keyword": map[string]interface{}{
										"value": "12",
									},
								},
							},
						},
					},
				},
			}

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, []string{"indexPattern"})
			So(dsl, ShouldResemble, expectDsl)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview exist but must is a object", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "123",
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": map[string]interface{}{
							"term": map[string]interface{}{
								"labels.instance.keyword": map[string]interface{}{
									"value": "12",
								},
							},
						},
					},
				},
			}

			mockLogGroupAccess.EXPECT().GetLogGroupQueryFilters(gomock.Any()).AnyTimes().Return(
				dataViewQueryFilters, true, nil)

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, []string{"indexPattern"})
			So(dsl, ShouldResemble, expectDsl)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("ar_dataview = \"\" ", func() {
			dsl := map[string]interface{}{
				"ar_dataview": "",
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			}
			var expect []string

			library, status, err := ds.spliceMustFilters(dsl)

			So(library, ShouldResemble, expect)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

	})
}
