// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

var (
	onlyIndexPatternFilters = map[string]interface{}{
		"indices": map[string]interface{}{
			"index_pattern": []interface{}{
				"hahaha",
				"test_topic_r_0gx_e_0000",
			},
			"ware_house_id": []interface{}{
				"101a6ec9f938885df0a44f20458d2eb4",
				"fab03138b29cc52cbb573ef634f10c1c",
			},
			"manual_index": []interface{}{},
		},
		"must_filter": []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
					"analyze_wildcard": true,
				},
			},
		},
	}

	onlyManualIndexFilters = map[string]interface{}{
		"indices": map[string]interface{}{
			"index_pattern": []interface{}{},
			"ware_house_id": []interface{}{
				"101a6ec9f938885df0a44f20458d2eb4",
				"fab03138b29cc52cbb573ef634f10c1c",
			},
			"manual_index": []interface{}{"a", "b"},
		},
		"must_filter": []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
					"analyze_wildcard": true,
				},
			},
		},
	}

	bothFilters = map[string]interface{}{
		"indices": map[string]interface{}{
			"index_pattern": []interface{}{
				"hahaha",
				"test_topic_r_0gx_e_0000",
			},
			"ware_house_id": []interface{}{
				"101a6ec9f938885df0a44f20458d2eb4",
				"fab03138b29cc52cbb573ef634f10c1c",
			},
			"manual_index": []interface{}{"a", "b"},
		},
		"must_filter": []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "(type.keyword: \"metricbeat_node_exporter\" OR type.keyword: \"system_metric_default\") AND (tags.keyword: \"metricbeat\" OR tags.keyword: \"node_exporter\")",
					"analyze_wildcard": true,
				},
			},
		},
	}

	onlyIndexPatternFiltersExpect = interfaces.LogGroup{
		IndexPattern: []string{
			"hahaha-*",
			"mdl-hahaha-*",
			"test_topic_r_0gx_e_0000-*",
			"mdl-test_topic_r_0gx_e_0000-*",
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

	onlyManualIndexFiltersExpect = interfaces.LogGroup{
		IndexPattern: []string{
			"a", "b",
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

	bothFiltersExpect = interfaces.LogGroup{
		IndexPattern: []string{
			"hahaha-*",
			"mdl-hahaha-*",
			"test_topic_r_0gx_e_0000-*",
			"mdl-test_topic_r_0gx_e_0000-*",
			"a", "b",
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
)

// func TestGetLogGroupQueryFilters(t *testing.T) {
// 	Convey("Test opensearch submit", t, func() {

// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
// 		daAccess := &logGroupAccess{httpClient: mockHttpClient}

// 		Convey("get request method failed", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusOK, nil, fmt.Errorf("method failed"))

// 			logGroup, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(logGroup, ShouldResemble, interfaces.DataView{})
// 			So(status, ShouldEqual, http.StatusInternalServerError)
// 			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
// 		})

// 		Convey("get queryfilters failed", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusCreated, nil, nil)

// 			logGroup, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(logGroup, ShouldResemble, interfaces.DataView{})
// 			So(status, ShouldEqual, http.StatusInternalServerError)
// 			So(err, ShouldResemble, fmt.Errorf("get queryfilters failed: <nil>"))
// 		})

// 		Convey("response nil ", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusOK, nil, nil)

// 			logGroup, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(logGroup, ShouldResemble, interfaces.DataView{})
// 			So(status, ShouldEqual, http.StatusOK)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("response with only index_pattern ", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusOK, onlyIndexPatternFilters, nil)

// 			logGroup, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(logGroup, ShouldResemble, onlyIndexPatternFiltersExpect)
// 			So(status, ShouldEqual, http.StatusOK)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("response with only manual_index ", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusOK, onlyManualIndexFilters, nil)

// 			resByte, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(resByte, ShouldResemble, onlyManualIndexFiltersExpect)
// 			So(status, ShouldEqual, http.StatusOK)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("response with index_pattern and manual_index ", func() {
// 			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
// 				http.StatusOK, bothFilters, nil)

// 			resByte, status, err := daAccess.GetLogGroupQueryFilters("a")

// 			So(resByte, ShouldResemble, bothFiltersExpect)
// 			So(status, ShouldEqual, http.StatusOK)
// 			So(err, ShouldBeNil)
// 		})

// 	})
// }

func TestGetLogGroupQueryFilters2(t *testing.T) {
	Convey("Test opensearch submit", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		daAccess := &logGroupAccess{httpClient: mockHttpClient}

		Convey("get request method failed", func() {
			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, fmt.Errorf("method failed"))

			logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

			So(logGroup, ShouldResemble, interfaces.LogGroup{})
			So(isExist, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("get queryfilters failed", func() {
			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusCreated, nil, nil)

			logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

			So(logGroup, ShouldResemble, interfaces.LogGroup{})
			So(isExist, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf("get queryfilters failed: <nil>"))
		})

		// Convey("response nil ", func() {
		// 	mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
		// 		http.StatusOK, nil, nil)

		// 	logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

		// 	So(logGroup, ShouldResemble, interfaces.DataView{})
		// 	So(isExist, ShouldEqual, true)
		// 	So(err, ShouldBeNil)
		// })

		Convey("response with only index_pattern ", func() {
			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, onlyIndexPatternFilters, nil)

			logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

			So(logGroup, ShouldResemble, onlyIndexPatternFiltersExpect)
			So(isExist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("response with only manual_index ", func() {
			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, onlyManualIndexFilters, nil)

			logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

			So(logGroup, ShouldResemble, onlyManualIndexFiltersExpect)
			So(isExist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("response with index_pattern and manual_index ", func() {
			mockHttpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, bothFilters, nil)

			logGroup, isExist, err := daAccess.GetLogGroupQueryFilters("a")

			So(logGroup, ShouldResemble, bothFiltersExpect)
			So(isExist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetLogGroupByName(t *testing.T) {
	Convey("Test GetLogGroupByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		da := &logGroupAccess{
			dataManagerUrl: "http://data-manager-anyrobot:13001",
			httpClient:     mockHttpClient,
		}

		Convey("failed, caused by http error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			logGroup, err := da.GetLogGroupByName("a")
			So(logGroup, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := sonic.Marshal([]interfaces.LogGroupInfo{{Id: "1", Name: "a"}})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusAccepted, okResp, nil)

			logGroup, err := da.GetLogGroupByName("a")
			So(logGroup, ShouldBeEmpty)
			So(err, ShouldResemble, fmt.Errorf("get log group by name failed: <nil>"))
		})

		Convey("failed, caused by http result is null", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			logGroup, err := da.GetLogGroupByName("a")
			So(logGroup, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := sonic.Marshal([]interfaces.LogGroupInfo{{Id: "1", Name: "a"}})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			logGroup, err := da.GetLogGroupByName("a")
			So(logGroup, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			okResp, _ := sonic.Marshal([]interfaces.LogGroupInfo{{Id: "1", Name: "a"}})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			logGroup, err := da.GetLogGroupByName("a")
			So(logGroup, ShouldResemble, []interfaces.LogGroupInfo{{Id: "1", Name: "a"}})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetLogGroupQueryFiltersAndFields(t *testing.T) {
	Convey("Test GetLogGroupQueryFiltersAndFields", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)

		da := &logGroupAccess{
			dataManagerUrl: "http://data-manager-anyrobot:13001",
			httpClient:     mockHttpClient,
		}

		emptyInfo := interfaces.LogGroup{}
		Convey("failed, caused by http error", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			dataViewFilters, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(dataViewFilters, ShouldResemble, emptyInfo)
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusAccepted, nil, nil)

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(logGroup, ShouldResemble, emptyInfo)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("failed, caused by http result is null", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, nil, nil)

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(logGroup, ShouldResemble, emptyInfo)
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := sonic.Marshal(interfaces.LogGroup{IndexPattern: []string{"a"}})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFunc(sonic.Unmarshal,
				func(data []byte, v any) error {
					return errors.New("error")
				},
			)
			defer patch.Reset()

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(logGroup, ShouldResemble, emptyInfo)
			So(exists, ShouldBeFalse)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by loggroup's field type is wrong", func() {
			dataViewExp := interfaces.QueryFilters{
				Indices: interfaces.Indices{
					IndexPattern: []string{"a"},
					ManualIndex:  []string{"b"},
				},
				MustFilter: "",
				ArrayFields: map[string][]interfaces.LogGroupField{
					"ac": {
						{Name: "f1", Type: 1},
					},
				},
			}
			okResp, _ := sonic.Marshal(dataViewExp)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(logGroup, ShouldResemble, emptyInfo)
			So(exists, ShouldBeFalse)
			So(err.Error(), ShouldEqual, "loggroup field type is wrong")
		})

		Convey("success with field", func() {
			dataViewExp := interfaces.QueryFilters{
				Indices: interfaces.Indices{
					IndexPattern: []string{"a"},
					ManualIndex:  []string{"b"},
				},
				MustFilter: "",
				ArrayFields: map[string][]interfaces.LogGroupField{
					"ac": {
						{Name: "f1", Type: "text"},
						{
							Name: "labels",
							Type: []interfaces.LogGroupField{
								{Name: "cpu", Type: "text"},
							},
						},
					},
				},
			}
			okResp, _ := sonic.Marshal(dataViewExp)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(exists, ShouldBeTrue)
			So(logGroup, ShouldResemble, interfaces.LogGroup{IndexPattern: []string{"a-*", "mdl-a-*", "b"}, MustFilter: "",
				Fields: map[string]string{"f1": "text", "labels.cpu": "text"}})
			So(err, ShouldBeNil)
		})

		Convey("success", func() {
			dataViewExp := interfaces.QueryFilters{
				Indices: interfaces.Indices{
					IndexPattern: []string{"a"},
					ManualIndex:  []string{"b"},
				},
				MustFilter: "",
			}
			okResp, _ := sonic.Marshal(dataViewExp)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, okResp, nil)

			logGroup, exists, err := da.GetLogGroupQueryFiltersAndFields("a")
			So(exists, ShouldBeTrue)
			So(logGroup, ShouldResemble, interfaces.LogGroup{IndexPattern: []string{"a-*", "mdl-a-*", "b"}, MustFilter: "",
				Fields: map[string]string{}})
			So(err, ShouldBeNil)
		})
	})
}
