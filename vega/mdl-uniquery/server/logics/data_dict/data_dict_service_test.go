// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	"uniquery/interfaces"
	mock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewDataDictService(datadictAccess interfaces.DataDictAccess, appSetting *common.AppSetting) *DataDictService {
	return &DataDictService{
		appSetting: appSetting,
		ddAccess:   datadictAccess,
	}
}

func TestLoadDict(t *testing.T) {
	Convey("Test Load", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDataDictAccess := mock.NewMockDataDictAccess(mockCtrl)
		appSetting := &common.AppSetting{}

		dd := MockNewDataDictService(mockDataDictAccess, appSetting)

		Convey("Load success", func() {
			mockDataDictAccess.EXPECT().GetDictInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, nil)
			mockDataDictAccess.EXPECT().GetDictIteams(gomock.Any(), gomock.Any()).AnyTimes().Return(map[string][]map[string]string{}, nil)
			err := dd.LoadDict(testCtx, "name")
			time.Sleep(time.Microsecond)
			So(err, ShouldBeNil)
		})

		Convey("Load GetDict failed", func() {
			mockDataDictAccess.EXPECT().GetDictInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, fmt.Errorf("failed to find dict"))
			err := dd.LoadDict(testCtx, "name")
			time.Sleep(time.Microsecond)
			So(err.Error(), ShouldResemble, "failed to find dict")
		})
	})
}

func TestGetDictByName(t *testing.T) {
	Convey("Test GetDictByName", t, func() {

		Convey("get dict success", func() {
			dict := interfaces.DataDict{
				DictID:   "1",
				DictName: "dict",
			}
			dictCache["dict"] = dict
			res, _ := GetDictByName("dict")
			So(res, ShouldResemble, dict)
		})

		Convey("get dict failed", func() {
			res, _ := GetDictByName("name1")
			So(res, ShouldResemble, interfaces.DataDict{})
		})
	})
}

func TestGetRecordsByKey(t *testing.T) {
	Convey("Test GetDictByName", t, func() {

		dict := interfaces.DataDict{
			DictID:   "1",
			DictName: "dict",
			DictRecords: map[string][]map[string]string{
				"key1": {
					{
						"key":   "key1",
						"value": "value1",
					},
				},
			},
		}

		res, ok := GetRecordsByKey(dict, []string{"key1"})
		So(res, ShouldResemble, []map[string]string{
			{
				"key":   "key1",
				"value": "value1",
			},
		})
		So(ok, ShouldBeTrue)
	})

}
