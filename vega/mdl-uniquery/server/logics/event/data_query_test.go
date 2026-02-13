// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
	imock "uniquery/interfaces/mock"
)

func mockNewMetricDataQuery(dataSource []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery,
	mmsMock interfaces.MetricModelService) *MetricDataQuery {
	mdq := &MetricDataQuery{
		DataSource:     dataSource,
		DataSourceType: dataSourceType,
		TimeInterval:   t,
		Start:          query.Start,
		End:            query.End,
		Step:           query.Step,
		Filters:        query.Filters,
		mmService:      mmsMock,
	}
	return mdq
}

func TestMetricFetchSourceRecordsFrom(t *testing.T) {
	Convey("Test FetchSourceRecordsFrom", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mmsMock := imock.NewMockMetricModelService(mockCtrl)

		dataSource := []string{"1"}
		dataSourceType := "metric_model"
		timenterval := interfaces.TimeInterval{
			Interval: 5,
			Unit:     "m",
		}
		query := interfaces.EventQuery{
			Start:   111,
			End:     222,
			Step:    "5m",
			Filters: []interfaces.Filter{},
		}

		mdqMock := mockNewMetricDataQuery(dataSource, dataSourceType, timenterval, query, mmsMock)

		Convey("GetMetricModelData failed", func() {
			expectedErr := errors.New("error")
			mmsMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.MetricModelUniResponse{}, 0, 0, expectedErr)
			_, _, err := mdqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("success", func() {
			stepi := "5m"
			expectRes := interfaces.MetricModelUniResponse{
				Model: interfaces.MetricModel{},
				Datas: []interfaces.MetricModelData{
					{
						Labels: map[string]string{
							"cpu": "1",
						},
						Times: []interface{}{int64(1695003480000), int64(1695003840000), int64(1695004200000),
							int64(1695004560000), int64(1695004920000), int64(1695005280000)},
						Values: []interface{}{1.1, nil, nil, 1.1, nil, nil},
					},
				},
				Step:        &stepi,
				SeriesTotal: 1,
				PointTotal:  2,
			}
			mmsMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(expectRes, 0, 0, nil)
			_, _, err := mdqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldBeNil)
		})
		Convey("success,step is empty", func() {
			query := interfaces.EventQuery{
				Start:   111,
				End:     222,
				Step:    "",
				Filters: []interfaces.Filter{},
			}
			mdqMock := mockNewMetricDataQuery(dataSource, dataSourceType, timenterval, query, mmsMock)
			expectRes := interfaces.MetricModelUniResponse{
				Model: interfaces.MetricModel{},
				Datas: []interfaces.MetricModelData{
					{
						Labels: map[string]string{
							"cpu": "1",
						},
						Times: []interface{}{int64(1695003480000), int64(1695003840000), int64(1695004200000),
							int64(1695004560000), int64(1695004920000), int64(1695005280000)},
						Values: []interface{}{1.1, nil, nil, 1.1, nil, nil},
					},
				},
				Step:        nil,
				SeriesTotal: 1,
				PointTotal:  2,
			}
			mmsMock.EXPECT().Exec(gomock.Any(), gomock.Any()).AnyTimes().Return(expectRes, 0, 0, nil)
			_, _, err := mdqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldBeNil)
		})

	})
}

func mockNewEventDataQuery(dataSource []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery,
	eService interfaces.EventService) *EventDataQuery {
	edq := &EventDataQuery{
		DataSources:    dataSource,
		DataSourceType: dataSourceType,
		TimeInterval:   t,
		Start:          query.Start,
		End:            query.End,
		Filters:        query.Filters,
		eventService:   eService,
	}
	return edq
}

func TestEventFetchSourceRecordsFrom(t *testing.T) {
	Convey("Test EventFetchSourceRecordsFrom", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		esMock := imock.NewMockEventService(mockCtrl)

		dataSource := []string{"1"}
		dataSourceType := "event_model"
		timenterval := interfaces.TimeInterval{
			Interval: 5,
			Unit:     "m",
		}
		query := interfaces.EventQuery{
			Start:   111,
			End:     222,
			Filters: []interfaces.Filter{},
		}
		edqMock := mockNewEventDataQuery(dataSource, dataSourceType, timenterval, query, esMock)
		Convey("success", func() {
			events := []interfaces.IEvents{
				[]interfaces.IEvent{
					interfaces.AtomicEvent{
						BaseEvent: interfaces.BaseEvent{
							Id:             "1",
							Title:          "xxxx_紧急",
							EventModelId:   "1",
							EventModelName: "xxxx",
							EventType:      "atomic",
							Level:          1,
							DefaultTimeWindow: interfaces.TimeInterval{
								Interval: 5,
								Unit:     "m",
							},
							CreateTime: testNow,
							Tags:       []string{},
							Relations:  map[string]any{},
						},
						Message:     "sxxxxxxx",
						TriggerData: interfaces.Records{},
						TriggerTime: testNow,
					},
				},
			}
			// expectRes := interfaces.SourceRecords{
			// 	Records: interfaces.Records{
			// 		map[string]any{
			// 			"id":               "1",
			// 			"type":             "atomic",
			// 			"level":            1,
			// 			"event_model_id":   uint64(1),
			// 			"event_model_name": "xxxx",

			// 			"title":            "xxxx_紧急",
			// 			"@timestamp":       time.Time{},
			// 			"trigger_time":     time.Time{},
			// 			"trigger_data":     interfaces.Records{},
			// 			"tags":             []string{},
			// 			"relations":        map[string]any{},
			// 			"relations_events": []interfaces.RelationEvent{},
			// 		},
			// 	},
			// }
			esMock.EXPECT().Query(gomock.Any(), gomock.Any()).AnyTimes().Return(1, events, nil, nil)
			_, _, err := edqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldBeNil)
			// So(sr, ShouldResemble, expectRes)
		})
		// Convey("success, eventSlicess Is empty", func() {
		// 	events := []interfaces.IEvents{
		// 		[]interfaces.IEvent{
		// 			interfaces.AtomicEvent{
		// 				BaseEvent: interfaces.BaseEvent{
		// 					Id:             uint64(1),
		// 					Title:          "xxxx_紧急",
		// 					EventModelId:   uint64(1),
		// 					EventModelName: "xxxx",
		// 					EventType:      "atomic",
		// 					Level:          1,
		// 					DefaultTimeWindow: interfaces.TimeInterval{
		// 						Interval: 5,
		// 						Unit:     "m",
		// 					},
		// 					CreateTime: time.Time{},
		// 					Tags:       []string{},
		// 				},
		// 				Message:     "sxxxxxxx",
		// 				TriggerData: interfaces.Records{},
		// 				TriggerTime: time.Time{},
		// 			},
		// 		},
		// 	}
		// 	expectRes := interfaces.SourceRecords{
		// 		Records: interfaces.Records{
		// 			map[string]any{
		// 				"id":               "1",
		// 				"type":             "atomic",
		// 				"level":            1,
		// 				"event_model_id":   uint64(1),
		// 				"event_model_name": "xxxx",

		// 				"title":        "xxxx_紧急",
		// 				"@timestamp":   time.Time{},
		// 				"trigger_time": time.Time{},
		// 				"trigger_data": interfaces.Records{},
		// 				"tags":         []string{},
		// 			},
		// 		},
		// 	}
		// 	esMock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(0, []interfaces.IEvents{}, nil)
		// 	esMock.EXPECT().Query(gomock.Any(), gomock.Any()).Return(1, events, nil)

		// 	sr, err := edqMock.FetchSourceRecordsFrom(ctx)
		// 	So(err, ShouldBeNil)
		// 	So(sr, ShouldResemble, expectRes)
		// })
	})
}

func mockNewLogDataQuery(dataSource []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery,
	dvsMock interfaces.DataViewService, oMock interfaces.OpenSearchAccess) *LogDataQuery {
	ldq := &LogDataQuery{
		DataSource:      dataSource,
		DataSourceType:  dataSourceType,
		TimeInterval:    t,
		Start:           query.Start,
		End:             query.End,
		Condition:       convertFiltersToCondition(query.Filters),
		dataViewService: dvsMock,
		osAccess:        oMock,
	}
	return ldq
}

func TestLogFetchSourceRecordsFrom(t *testing.T) {
	Convey("Test LogFetchSourceRecordsFrom", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvsMock := imock.NewMockDataViewService(mockCtrl)
		oMock := imock.NewMockOpenSearchAccess(mockCtrl)

		dataSource := []string{"1"}
		dataSourceType := "data_view"
		timenterval := interfaces.TimeInterval{
			Interval: 5,
			Unit:     "m",
		}
		query := interfaces.EventQuery{
			Start: 111,
			End:   222,
			Filters: []interfaces.Filter{
				{
					Name:      "id",
					Operation: "=",
					Value:     "1",
				},
			},
		}

		ldqMock := mockNewLogDataQuery(dataSource, dataSourceType, timenterval, query, dvsMock, oMock)

		Convey("GetViewData failed", func() {
			expectedErr := errors.New("error")
			dvsMock.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&interfaces.ViewUniResponseV2{}, expectedErr)
			_, _, err := ldqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldResemble, expectedErr)
		})

		Convey("success.value is nil", func() {
			events := interfaces.ViewUniResponseV2{
				PitID:       "",
				SearchAfter: []any{},
				View:        nil,
				Entries:     []map[string]any{},
				TotalCount:  nil,
				ScrollId:    "",
			}
			expectRes := interfaces.SourceRecords{}

			dvsMock.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&events, nil)

			oMock.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return([]byte{}, http.StatusOK, nil)
			sr, _, err := ldqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldBeNil)
			So(sr, ShouldResemble, expectRes)

		})

		Convey("success", func() {
			events := interfaces.ViewUniResponseV2{
				PitID:       "",
				SearchAfter: []any{},
				View:        nil,
				Entries: []map[string]any{
					{
						"id":               "1",
						"type":             "atomic",
						"level":            1,
						"event_model_id":   uint64(1),
						"event_model_name": "xxxx",
						"title":            "xxxx_紧急",
						"@timestamp":       time.Time{},
						"trigger_time":     time.Time{},
						"trigger_data":     interfaces.Records{},
						"tags":             []string{},
					},
				},
				TotalCount: nil,
				ScrollId:   "",
			}

			// expectRes := interfaces.SourceRecords{
			// 	Records: interfaces.Records{map[string]any{
			// 		"@timestamp":       time.Time{},
			// 		"event_model_id":   1,
			// 		"event_model_name": "xxxx",
			// 		"id":               "1",
			// 		"level":            1,
			// 		"tags":             []string{},
			// 		"title":            "xxxx_紧急",
			// 		"trigger_data":     interfaces.Records{},
			// 		"trigger_time":     time.Time{},
			// 		"type":             "atomic",
			// 	}}}

			dvsMock.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&events, nil)
			dvsMock.EXPECT().GetSingleViewData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ViewUniResponseV2{}, nil)
			oMock.EXPECT().DeleteScroll(gomock.Any(), gomock.Any()).AnyTimes().Return([]byte{}, http.StatusOK, nil)

			_, _, err := ldqMock.FetchSourceRecordsFrom(testCtx, "flat")
			So(err, ShouldBeNil)
			// So(sr, ShouldEqual, expectRes)

		})

	})
}
