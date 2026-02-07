// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package uniquery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ADMIN_TYPE,
		})
)

func MockNewUniqueryAccess(appSetting *common.AppSetting,
	httpClient rest.HTTPClient) *uniqueryAccess {

	ua := &uniqueryAccess{
		appSetting:  appSetting,
		uniqueryUrl: "http://uniquery-anyrobot:13011/api/uniquery/v1/metric-model",
		httpClient:  httpClient,
	}
	return ua
}

func Test_UniqueryAccess_CheckFormulaByUniquery(t *testing.T) {
	Convey("Test CheckFormulaByUniquery", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		ua := MockNewUniqueryAccess(appSetting, httpClient)

		query := interfaces.MetricModelQuery{
			MetricType: "atomic",
			QueryType:  "promql",
			Formula:    "a",
			DataSource: &interfaces.MetricDataSource{
				Type: interfaces.SOURCE_TYPE_DATA_VIEW,
				ID:   "123",
			},
		}

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			sucess, errStr, err := ua.CheckFormulaByUniquery(testCtx, query)
			So(sucess, ShouldBeFalse)
			So(errStr, ShouldEqual, "")
			So(err, ShouldResemble, fmt.Errorf("post metric model simulate request failed: method failed"))
		})

		Convey("failed, caused by status != 200", func() {
			okResp, _ := sonic.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			sucess, errStr, err := ua.CheckFormulaByUniquery(testCtx, query)
			So(sucess, ShouldBeFalse)
			So(errStr, ShouldEqual, `Formula invalid: {"error_code":"a","description":"a","solution":"","error_link":"","error_details":"a"}`)
			So(err, ShouldBeNil)
		})

		Convey("failed, caused by http result is null", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, nil, nil)

			sucess, errStr, err := ua.CheckFormulaByUniquery(testCtx, query)
			So(sucess, ShouldBeTrue)
			So(errStr, ShouldEqual, "")
			So(err, ShouldBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			okResp, _ := sonic.Marshal(rest.BaseError{ErrorCode: "a", Description: "a", ErrorDetails: "a"})
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).AnyTimes().Return(http.StatusAccepted, okResp, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			sucess, errStr, err := ua.CheckFormulaByUniquery(testCtx, query)
			So(sucess, ShouldBeFalse)
			So(errStr, ShouldEqual, "")
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("success", func() {
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).AnyTimes().Return(http.StatusOK, []byte{}, nil)

			sucess, errStr, err := ua.CheckFormulaByUniquery(testCtx, query)
			So(sucess, ShouldBeTrue)
			So(errStr, ShouldEqual, "")
			So(err, ShouldBeNil)
		})
	})
}
