// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega

import (
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/common"
	"data-model/interfaces"
)

var (
	vgAccessOnce sync.Once
	vgAccess     interfaces.VegaGatewayAccess
)

type vegaGatewayAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

type VegaError struct {
	Code        string      `json:"code"`        // 错误码
	Description string      `json:"description"` // 错误描述
	Detail      interface{} `json:"detail"`      // 详细内容
}

func NewVegaGatewayAccess(appSetting *common.AppSetting) interfaces.VegaGatewayAccess {
	vgAccessOnce.Do(func() {
		vgAccess = &vegaGatewayAccess{
			appSetting: appSetting,
			//httpClient:  common.NewOAuth2HTTPClient(),
			httpClient: common.NewHTTPClient(),
		}
	})

	return vgAccess
}

// func (vga *vegaGatewayAccess) CreateVegaExcelView(ctx context.Context, req *interfaces.CreateVegaExcelViewReq) (*interfaces.CreateVegaExcelViewRes, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "create vega excel view", trace.WithSpanKind(trace.SpanKindClient))
// 	defer span.End()

// 	urlStr := vga.appSetting.VegaExcelUrl

// 	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
// 		HttpUrl:         urlStr,
// 		HttpMethod:      http.MethodPost,
// 		HttpContentType: rest.ContentTypeJson,
// 	})

// 	accountInfo := interfaces.AccountInfo{}
// 	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
// 		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
// 	}
// 	headers := map[string]string{
// 		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
// 		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
// 		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
// 		"X-Presto-User":                     "admin",
// 	}

// 	respCode, respData, err := vga.httpClient.PostNoUnmarshal(ctx, urlStr, headers, req)
// 	if err != nil {
// 		errDetails := fmt.Sprintf("http.NewRequest create vega excel view %s error, %v", req.FileName, err)
// 		logger.Error(errDetails)

// 		o11y.Error(ctx, errDetails)
// 		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "create vega excel view failed")

// 		return nil, err
// 	}

// 	// 错误码结构是 code, description, detail，需要对错误码做转换
// 	if respCode != http.StatusOK {
// 		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
// 		var vegaError VegaError
// 		if err := sonic.Unmarshal(respData, &vegaError); err != nil {
// 			logger.Errorf("unmalshal VegaError failed: %v", err)

// 			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
// 			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

// 			return nil, err
// 		}
// 		httpErr := &rest.HTTPError{HTTPCode: respCode,
// 			BaseError: rest.BaseError{
// 				ErrorCode:    vegaError.Code,
// 				Description:  vegaError.Description,
// 				ErrorDetails: vegaError.Detail,
// 			}}
// 		logger.Errorf("Create Vega View Error: %v", httpErr.Error())

// 		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
// 		o11y.Error(ctx, fmt.Sprintf("Create Vega Excel View failed: %v", vegaError))

// 		return nil, fmt.Errorf("create vega excel view %s failed, status code: %d, error: %v", req.FileName, respCode, httpErr)
// 	}

// 	// 解码返回值
// 	var resp interfaces.CreateVegaExcelViewRes
// 	if err := sonic.Unmarshal(respData, &resp); err != nil {
// 		logger.Errorf("unmarshal create vega excel view response failed: %v", err)

// 		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "unmarshal create vega excel view response failed")
// 		o11y.Error(ctx, fmt.Sprintf("unmarshal create vega excel view response failed: %v", err))

// 		return nil, fmt.Errorf("unmarshal create vega excel view response failed: %v", err)
// 	}

// 	o11y.AddHttpAttrs4Ok(span, respCode)
// 	return &resp, nil

// }

// func (vga *vegaGatewayAccess) DeleteVegaExcelView(ctx context.Context, req *interfaces.DeleteExcelViewReq) error {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete vega excel view", trace.WithSpanKind(trace.SpanKindClient))
// 	defer span.End()

// 	urlStr := fmt.Sprintf("%s/%s/%s/%s", vga.appSetting.VegaExcelUrl, req.Catalog, req.Schema, req.View)

// 	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
// 		HttpUrl:         urlStr,
// 		HttpMethod:      http.MethodDelete,
// 		HttpContentType: rest.ContentTypeJson,
// 	})

// 	accountInfo := interfaces.AccountInfo{}
// 	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
// 		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
// 	}
// 	headers := map[string]string{
// 		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
// 		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
// 		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
// 		"X-Presto-User":                     "admin",
// 	}

// 	respCode, respData, err := vga.httpClient.DeleteNoUnmarshal(ctx, urlStr, headers)
// 	if err != nil {
// 		errDetails := fmt.Sprintf("http.NewRequest delete vega excel view %s error, %v", req.View, err)
// 		logger.Error(errDetails)

// 		o11y.Error(ctx, errDetails)
// 		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "delete vega view failed")

// 		return err
// 	}

// 	// 错误码结构是 code, description, detail，需要对错误码做转换
// 	if respCode != http.StatusOK {
// 		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
// 		var vegaError VegaError
// 		if err := sonic.Unmarshal(respData, &vegaError); err != nil {
// 			logger.Errorf("unmalshal VegaError failed: %v", err)

// 			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
// 			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

// 			return err
// 		}
// 		httpErr := &rest.HTTPError{HTTPCode: respCode,
// 			BaseError: rest.BaseError{
// 				ErrorCode:    vegaError.Code,
// 				Description:  vegaError.Description,
// 				ErrorDetails: vegaError.Detail,
// 			}}
// 		logger.Errorf("Delete Vega excel View Error: %v", httpErr.Error())

// 		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
// 		o11y.Error(ctx, fmt.Sprintf("Delete Vega excel View failed: %v", vegaError))

// 		return fmt.Errorf("delete vega excel view %s failed, status code: %d, error: %v", req.View, respCode, httpErr)
// 	}

// 	o11y.AddHttpAttrs4Ok(span, respCode)
// 	return nil
// }

// func (vga *vegaGatewayAccess) CreateExcelView(ctx context.Context, req *interfaces.CreateExcelViewReq) (*interfaces.CreateExcelViewRes, error) {
// 	drivenMsg := "DrivenVirtualizationEngine CreateExcelView "
// 	logger.Infof(drivenMsg+"%+v", *req)
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/excel/view", v.BaseURL)

// 	res := &interfaces.CreateExcelViewRes{}
// 	err := base.CallWithTokenUpward(ctx, v.HttpClient, drivenMsg, http.MethodPost, urlStr, req, res, my_errorcode.CreateExcelViewError)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return res, err
// }

// func (vga *vegaGatewayAccess) DeleteExcelView(ctx context.Context, req *interfaces.DeleteExcelViewReq) (*interfaces.DeleteExcelViewRes, error) {
// 	drivenMsg := "DrivenVirtualizationEngine DeleteExcelView "
// 	logger.Infof(drivenMsg+"%+v", *req)
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/excel/view/%s/%s/%s", v.BaseURL, req.VdmCatalog, req.Schema, req.View)

// 	res := &interfaces.DeleteExcelViewRes{}
// 	err := base.CallWithTokenUpward(ctx, v.HttpClient, drivenMsg, http.MethodDelete, urlStr, nil, res, my_errorcode.DeleteExcelViewError)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return res, err
// }

// func (vga *vegaGatewayAccess) GetPreview(ctx context.Context, req *interfaces.ViewEntries) (*interfaces.FetchDataRes, error) {
// 	drivenMsg := "DrivenVirtualizationEngine GetPreview "
// 	logger.Infof(drivenMsg+"%+v", *req)
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/preview/%s/%s/%s?limit=%d&user_id=%s", v.BaseURL, req.CatalogName, req.Schema, req.ViewName, req.Limit, req.UserId)

// 	request, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetPreviewError, err.Error())
// 	}
// 	request.Header.Add("X-Presto-User", "admin")

// 	resp, err := v.HttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetPreviewError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetViewError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		res := &interfaces.FetchDataRes{}
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.GetPreviewError, err.Error())
// 		}
// 		return res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			return nil, Unmarshal(ctx, body, drivenMsg)
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.GetPreviewError, resp.StatusCode)
// 		}
// 	}
// }

// func (vga *vegaGatewayAccess) FetchData(ctx context.Context, statement string) (*interfaces.FetchDataRes, error) {
// 	drivenMsg := "DrivenVirtualizationEngine FetchData "
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/fetch", v.BaseURL)
// 	logger.Infof(drivenMsg+" urlStr:%s \n %+v", urlStr, statement)
// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader([]byte(statement)))
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	request.Header.Add("X-Presto-User", "admin")
// 	//request.Header.Set("Authorization",  util.ObtainToken(ctx))
// 	request.Header.Add("Content-Type", "application/json")

// 	resp, err := v.LongTimeHttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.FetchDataRes
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.GetViewError, err.Error())
// 		}
// 		logger.Infof(drivenMsg+"res : %v ", res)
// 		return &res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			return nil, UnmarshalFetch(ctx, body, drivenMsg)
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.FetchDataError)
// 		}
// 	}
// }

// func (vga *vegaGatewayAccess) FetchAuthorizedData(ctx context.Context, statement string, req *interfaces.FetchReq) (*interfaces.FetchDataRes, error) {
// 	drivenMsg := "DrivenVirtualizationEngine FetchAuthorizedData "
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/fetch", v.BaseURL)
// 	if req != nil {
// 		urlStr = fmt.Sprintf("%s?user_id=%s&action=%s", urlStr, req.UserID, req.Action)
// 	}
// 	logger.Infof(drivenMsg+" urlStr:%s \n %+v", urlStr, statement)
// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader([]byte(statement)))
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	request.Header.Add("X-Presto-User", "admin")
// 	request.Header.Add("Content-Type", "application/json")

// 	resp, err := v.LongTimeHttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.FetchDataRes
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.GetViewError, err.Error())
// 		}
// 		logger.Infof(drivenMsg+"res : %v ", res)
// 		return &res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			return nil, UnmarshalFetch(ctx, body, drivenMsg)
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.FetchDataError)
// 		}
// 	}
// }

// // GetConnectors implements DrivenVirtualizationEngine.
// func (vga *vegaGatewayAccess) GetConnectors(ctx context.Context) (result *interfaces.GetConnectorsRes, err error) {
// 	const drivenMsg = "DrivenVirtualizationEngine GetConnectors"

// 	logger.Info(drivenMsg)
// 	urlStr := fmt.Sprintf("%s/api/virtual_engine_service/v1/connectors", v.BaseURL)
// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" http.NewRequestWithContext error: %v", err)
// 		err = errorcode.Detail(my_errorcode.DrivenGetConnectorsFailed, err.Error())
// 		return
// 	}

// 	if t, ok := ctx.Value(interception.Token).(string); ok {
// 		req.Header.Set("Authorization", t)
// 	}

// 	logger.Info("request", zap.String("method", req.Method), zap.String("urlStr", urlStr))

// 	resp, err := v.HttpClient.Do(req)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" http.DefaultClient.Do error: %v", err)
// 		err = errorcode.Detail(my_errorcode.DrivenGetConnectorsFailed, err.Error())
// 		return
// 	}
// 	defer resp.Body.Close()

// 	logger.Info("response", zap.String("method", req.Method), zap.String("urlStr", urlStr), zap.Int("statusCode", resp.StatusCode))

// 	if resp.StatusCode != http.StatusOK {
// 		err = errorcode.Detail(my_errorcode.DrivenGetConnectorsFailed, resp.Status)
// 		return
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		err = errorcode.Detail(my_errorcode.DrivenGetConnectorsFailed, err.Error())
// 		return
// 	}

// 	logger.Info("response", zap.String("body", string(body)))

// 	result = &interfaces.GetConnectorsRes{}
// 	if err = json.Unmarshal(body, result); err != nil {
// 		logger.Errorf(drivenMsg+" json.Unmarshal error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.DrivenGetConnectorsFailed, err.Error())
// 	}

// 	return result, nil
// }

// func (vga *vegaGatewayAccess) StreamDataFetch(ctx context.Context, urlStr string, statement string) (*interfaces.StreamFetchResp, error) {
// 	var (
// 		request *http.Request
// 		err     error
// 	)

// 	drivenMsg := "DrivenVirtualizationEngine StreamDataFetch "
// 	if len(urlStr) == 0 {
// 		urlStr = fmt.Sprintf("%s/api/virtual_engine_service/v1/fetch?type=1", v.BaseURL)
// 		logger.Infof(drivenMsg+" urlStr:%s \n %+v", urlStr, statement)
// 		request, err = http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(util.StringToBytes(statement)))
// 	} else {
// 		urlStr = fmt.Sprintf("%s/api/virtual_engine_service/v1/statement/executing/%s", v.BaseURL, urlStr)
// 		logger.Infof(drivenMsg+" urlStr:%s \n %+v", urlStr)
// 		request, err = http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)

// 	}
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	request.Header.Add("X-Presto-User", "admin")
// 	request.Header.Add("Content-Type", "application/json")

// 	resp, err := v.HttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.StreamFetchResp
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.FetchDataError, err.Error())
// 		}
// 		logger.Infof(drivenMsg+"res : %v ", res)
// 		return &res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			var res rest.HttpError
// 			if err := sonic.Unmarshal(body, &res); err != nil {
// 				logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 				return nil, errorcode.Detail(my_errorcode.VirtualizationEngineError, err.Error())
// 			}
// 			logger.Errorff("%+v", res)
// 			return nil, errors.New(util.BytesToString(body))
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.FetchDataError)
// 		}
// 	}
// }

// func (vga *vegaGatewayAccess) StreamDataDownload(ctx context.Context, urlStr string,
// 	req *interfaces.StreamDownloadReq) (*interfaces.StreamFetchResp, error) {
// 	var (
// 		request *http.Request
// 		err     error
// 	)

// 	drivenMsg := "DrivenVirtualizationEngine StreamDataDownload "
// 	if len(urlStr) == 0 {
// 		if req == nil {
// 			err = errors.New("req params cannot be nil")
// 			logger.Errorf(drivenMsg+"params invalid", zap.Error(err))
// 			return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 		}
// 		var buf []byte
// 		if buf, err = json.Marshal(req); err != nil {
// 			logger.Errorf(drivenMsg+"json.Marshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 		}
// 		urlStr = fmt.Sprintf("%s/api/virtual_engine_service/v1/download?user_id=%s&action=%s", v.BaseURL, req.UserID, req.Action)
// 		logger.Infof(drivenMsg+" urlStr:%s \n %#v", urlStr, *req)
// 		request, err = http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(buf))
// 		request.Header.Add("user_id", req.UserID)
// 	} else {
// 		logger.Infof(drivenMsg+" urlStr:%s \n %+v", urlStr)
// 		request, err = http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)

// 	}
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 	}

// 	request.Header.Add("X-Presto-User", "admin")
// 	request.Header.Add("Content-Type", "application/json")

// 	resp, err := v.HttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.StreamFetchResp
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.DownloadDataError, err.Error())
// 		}
// 		logger.Infof(drivenMsg+"res : %v ", res)
// 		return &res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			var res rest.HttpError
// 			if err := sonic.Unmarshal(body, &res); err != nil {
// 				logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 				return nil, errorcode.Detail(my_errorcode.VirtualizationEngineError, err.Error())
// 			}
// 			logger.Errorff("%+v", res)
// 			return nil, errors.New(util.BytesToString(body))
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.DownloadDataError)
// 		}
// 	}
// }

// // GetDataTables 获取数据表
// func (m *Metadata) GetDataTables(ctx context.Context, req *interfaces.GetDataTablesReq) ([]*interfaces.GetDataTablesDataRes, error) {
// 	drivenMsg := "DrivenMetadata GetDataTables "
// 	urlStr := fmt.Sprintf("%s/api/metadata-manage/v1/table?ids=%s&offset=%d&limit=1000", vga.baseURL, req.Ids, req.Offset)
// 	request, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTablesError, err.Error())
// 	}
// 	request.Header.Set("Authorization", util.ObtainToken(ctx))
// 	resp, err := vga.HttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTablesError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTablesError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.GetDataTablesRes
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.GetDataTablesError, err.Error())
// 		}
// 		for _, data := range res.Data {
// 			if data.AdvancedParams != "" {
// 				if err := json.Unmarshal([]byte(data.AdvancedParams), &data.AdvancedDataSlice); err != nil {
// 					logger.Errorf(err.Error())
// 				}
// 			}
// 		}
// 		logger.Infof(drivenMsg+"res  msg : %v ,code:%v", res.Description, res.Code)
// 		return res.Data, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			return nil, Unmarshal(ctx, body, drivenMsg)
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.GetDataTablesError, resp.StatusCode)
// 		}
// 	}
// }

// // GetDataTableDetail 表详情，表字段
// func (m *Metadata) GetDataTableDetail(ctx context.Context, req *interfaces.GetDataTableDetailReq) (*interfaces.GetDataTableDetailRes, error) {
// 	drivenMsg := "DrivenMetadata GetDataTableDetail "

// 	urlStr := fmt.Sprintf("%s/api/metadata-manage/v1/datasource/%d/schema/%s/table/%s",
// 		vga.baseURL, req.DataSourceId, req.SchemaId, req.TableId)
// 	request, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"http.NewRequest error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTableDetailError, err.Error())
// 	}
// 	request.Header.Set("Authorization", util.ObtainToken(ctx))
// 	resp, err := vga.HttpClient.Do(request)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+"client.Do error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTableDetailError, err.Error())
// 	}

// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Errorf(drivenMsg+" io.ReadAll error: %v", err)
// 		return nil, errorcode.Detail(my_errorcode.GetDataTableDetailError, err.Error())
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var res interfaces.GetDataTableDetailRes
// 		if err = sonic.Unmarshal(body, &res); err != nil {
// 			logger.Errorf(drivenMsg+" sonic.Unmarshal error: %v", err)
// 			return nil, errorcode.Detail(my_errorcode.GetDataTableDetailError, err.Error())
// 		}
// 		logger.Infof(drivenMsg+"res  msg : %v ,code:%v", res.Description, res.Code)
// 		return &res, nil
// 	} else {
// 		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 			return nil, Unmarshal(ctx, body, drivenMsg)
// 		} else {
// 			logger.Errorf(drivenMsg+"http status error", zap.String("status", resp.Status), zap.String("body", string(body)))
// 			return nil, errorcode.Desc(my_errorcode.GetDataTableDetailError, resp.StatusCode)
// 		}
// 	}
// }
