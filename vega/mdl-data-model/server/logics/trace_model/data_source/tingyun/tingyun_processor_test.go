// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package tingyun

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewTingYunTraceProcessor(appSetting *common.AppSetting) *tingYunTraceProcessor {
	return &tingYunTraceProcessor{
		appSetting: appSetting,
	}
}

func Test_TingYunTraceProcessor_GetSpanFieldInfo(t *testing.T) {
	Convey("Test GetSpanFieldInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tytp := MockNewTingYunTraceProcessor(appSetting)

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{}
			_, err := tytp.GetSpanFieldInfo(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunTraceProcessor_GetRelatedLogFieldInfo(t *testing.T) {
	Convey("Test GetRelatedLogFieldInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tytp := MockNewTingYunTraceProcessor(appSetting)

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{}
			_, err := tytp.GetRelatedLogFieldInfo(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}
