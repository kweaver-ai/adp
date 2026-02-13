// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

func TestGetDictInfo(t *testing.T) {
	Convey("Test GetDictInfo", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		ddAccess := &dataDictAccess{httpClient: mockHttpClient}

		Convey("GetDictInfo success ", func() {
			type Entry struct {
				Id         string `json:"id"`
				UpdateTime int64  `json:"update_time"`
			}
			type List struct {
				Entries []Entry `json:"entries"`
			}

			updateTime := time.Now().UnixMilli()
			list := List{
				Entries: []Entry{
					{
						Id:         "1111",
						UpdateTime: updateTime,
					},
				},
			}
			bytes, _ := sonic.Marshal(list)

			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, bytes, nil)

			dict, err := ddAccess.GetDictInfo(testCtx, "123")

			So(dict.DictID, ShouldEqual, "1111")
			So(dict.UpdateTime, ShouldEqual, updateTime)
			So(err, ShouldBeNil)
		})

		Convey("GetDictInfo get method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, nil, fmt.Errorf("method failed"))

			_, err := ddAccess.GetDictInfo(testCtx, "123")

			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("GetDictID not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusNotFound, nil, nil)

			dict, err := ddAccess.GetDictInfo(testCtx, "123")

			So(dict.DictID, ShouldEqual, "")
			So(err, ShouldBeNil)
		})

		Convey("GetDictID failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, nil, nil)

			_, err := ddAccess.GetDictInfo(testCtx, "123")
			So(err, ShouldNotBeNil)
		})

		Convey("GetDictID failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, bytes, nil)

			_, err := ddAccess.GetDictInfo(testCtx, "123")

			So(err.Error(), ShouldEqual, "get data dict failed: failed")
		})

		Convey("GetDictID response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			_, err := ddAccess.GetDictInfo(testCtx, "123")

			So(err, ShouldBeNil)
		})

		Convey("GetDictID response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, []uint8{}, nil)

			_, err := ddAccess.GetDictInfo(testCtx, "123")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetDict(t *testing.T) {
	Convey("Test GetDict", t, func() {

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHttpClient := rmock.NewMockHTTPClient(mockCtrl)
		ddAccess := &dataDictAccess{httpClient: mockHttpClient}

		Convey("success ", func() {
			dict := interfaces.DataDict{
				DictName: "1111",
				Dimension: interfaces.Dimension{
					Keys: []interfaces.DimensionItem{{Name: "key1"}, {Name: "key2"}},
				},
				UniqueKey: true,
				DictItems: []map[string]string{
					{
						"key1":   "key1",
						"key2":   "key2",
						"value1": "value1",
						"value2": "value2",
					},
					{
						"key1":   "k1",
						"key2":   "k2",
						"value1": "v1",
						"value2": "v2",
					},
				},
			}
			dicts := []interfaces.DataDict{}
			dicts = append(dicts, dict)

			bytes, _ := sonic.Marshal(dicts)
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, bytes, nil)

			expectData := map[string][]map[string]string{
				"key1\x00key2": {{
					"key1":   "key1",
					"key2":   "key2",
					"value1": "value1",
					"value2": "value2"}},
				"k1\x00k2": {{
					"key1":   "k1",
					"key2":   "k2",
					"value1": "v1",
					"value2": "v2"}},
			}
			data, err := ddAccess.GetDictIteams(testCtx, "123")
			So(data, ShouldResemble, expectData)
			So(err, ShouldBeNil)
		})

		Convey("GetDict get method failed", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusOK, nil, fmt.Errorf("method failed"))

			_, err := ddAccess.GetDictIteams(testCtx, "123")

			So(err, ShouldResemble, fmt.Errorf("get request method failed: method failed"))
		})

		Convey("GetDict not found", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusNotFound, nil, nil)

			_, err := ddAccess.GetDictIteams(testCtx, "123")

			So(err, ShouldBeNil)
		})

		Convey("GetDict failed because unmarshal", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, nil, nil)

			_, err := ddAccess.GetDictIteams(testCtx, "123")
			So(err, ShouldNotBeNil)
		})

		Convey("GetDict failed because 500", func() {
			bytes, _ := sonic.Marshal(rest.BaseError{ErrorCode: "as",
				Description:  "500",
				ErrorLink:    "",
				Solution:     "",
				ErrorDetails: "failed",
			})
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, bytes, nil)

			_, err := ddAccess.GetDictIteams(testCtx, "123")

			So(err.Error(), ShouldEqual, "get data dict failed: failed")
		})

		Convey("response nil ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			_, err := ddAccess.GetDictIteams(testCtx, "123")

			So(err, ShouldBeNil)
		})

		Convey("response unmarshal failed ", func() {
			mockHttpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, []uint8{}, nil)

			_, err := ddAccess.GetDictIteams(testCtx, "123")
			So(err, ShouldNotBeNil)
		})
	})
}
