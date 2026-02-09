// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

var (
	osAccessOnce sync.Once
	osAccess     interfaces.OpenSearchAccess
	osAddress    string
)

type openSearchAccess struct {
	appSetting *common.AppSetting
	client     *opensearch.Client
}

func NewOpenSearchAccess(appSetting *common.AppSetting) interfaces.OpenSearchAccess {
	osAccessOnce.Do(func() {
		osAccess = &openSearchAccess{
			appSetting: appSetting,
			client:     rest.NewOpenSearchClient(appSetting.OpenSearchSetting),
		}
		osAddress = fmt.Sprintf("%s://%s:%d", appSetting.OpenSearchSetting.Protocol,
			appSetting.OpenSearchSetting.Host, appSetting.OpenSearchSetting.Port)
	})

	return osAccess
}

// SearchSubmit Search 查询
func (osa *openSearchAccess) SearchSubmit(ctx context.Context, query map[string]interface{},
	indices []string, scroll time.Duration, preference string, trackTotalHits bool) ([]byte, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Opensearch search", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attr.Key("url").String(osAddress),
		attr.Key("indices").StringSlice(indices),
		attr.Key("preference").String(preference),
		attr.Key("api").String("_search"),
	)
	defer span.End()

	// 对请求体进行编码
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		span.SetStatus(codes.Error, "Encode DSL query error")
		o11y.Error(ctx, fmt.Sprintf("Encode query %v error: %v", query, oerr))

		return nil, http.StatusInternalServerError, oerr
	}

	res, err := osa.execute(ctx, buf, indices, scroll, preference, trackTotalHits)
	if err != nil {
		logger.Errorf("Error getting _search response from opensearch: %v", err)

		span.SetStatus(codes.Error, "Get _search response error")
		o11y.Error(ctx, fmt.Sprintf("Get _search response from opensearch error: %v", err))

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	result, code, err := commonProcess(ctx, res)
	if err != nil {
		span.SetStatus(codes.Error, "Process _search response error")
		o11y.Error(ctx, fmt.Sprintf("Process _search response from opensearch error: %v", err))
		return nil, code, err
	}

	span.SetStatus(codes.Ok, "")
	return result, code, nil
}

// execute 执行不同的search查询
func (osa *openSearchAccess) execute(ctx context.Context, buf bytes.Buffer, indices []string,
	scroll time.Duration, preference string, trackTotalHits bool) (*opensearchapi.Response, error) {

	ignoreUnavailable := true
	req := opensearchapi.SearchRequest{
		Body:              &buf,
		Index:             indices,
		IgnoreUnavailable: &ignoreUnavailable,
		Preference:        preference,
		TrackTotalHits:    trackTotalHits,
	}

	if scroll != 0 {
		req.Scroll = scroll
	}

	res, err := req.Do(ctx, osa.client)
	return res, err
}

// 验证，少一次marsharl
func (osa *openSearchAccess) SearchSubmitWithBuffer(ctx context.Context, query bytes.Buffer, indices []string,
	scroll time.Duration, preference string) ([]byte, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Opensearch search", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attr.Key("url").String(osAddress),
		attr.Key("indices").StringSlice(indices),
		attr.Key("preference").String(preference),
		attr.Key("api").String("_search"),
	)
	defer span.End()

	ignoreUnavailable := true
	req := opensearchapi.SearchRequest{
		Body:              &query,
		Index:             indices,
		IgnoreUnavailable: &ignoreUnavailable,
		Preference:        preference,
	}

	if scroll != 0 {
		req.Scroll = scroll
	}

	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting _search response from opensearch: %v", err)

		span.SetStatus(codes.Error, "Get _search response error")
		o11y.Error(ctx, fmt.Sprintf("Get _search response from opensearch error: %v", err))

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	result, code, err := commonProcess(ctx, res)
	if err != nil {
		span.SetStatus(codes.Error, "Process _search response error")
		o11y.Error(ctx, fmt.Sprintf("Process _search response from opensearch error: %v", err))
		return nil, code, err
	}

	span.SetStatus(codes.Ok, "")
	return result, code, nil
}

// Search with pit 查询
func (osa *openSearchAccess) SearchWithPit(ctx context.Context, query bytes.Buffer) ([]byte, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Opensearch search with pit", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attr.Key("url").String(osAddress),
		attr.Key("api").String("_search"),
	)
	defer span.End()

	// 不使用 ignoreUnavailable 参数，因为使用 pit 时，不能传递索引
	req := opensearchapi.SearchRequest{
		Body: &query,
		// IgnoreUnavailable: &ignoreUnavailable,
	}

	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting _search with pit response from opensearch: %v", err)

		span.SetStatus(codes.Error, "Get _search with pit response error")
		o11y.Error(ctx, fmt.Sprintf("Get _search with pit response from opensearch error: %v", err))

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	result, code, err := commonProcess(ctx, res)
	if err != nil {
		span.SetStatus(codes.Error, "Process _search with pit response error")
		o11y.Error(ctx, fmt.Sprintf("Process _search with pit response from opensearch error: %v", err))
		return nil, code, err
	}

	span.SetStatus(codes.Ok, "")
	return result, code, nil
}

// Scroll scroll分页查询获取数据
func (osa *openSearchAccess) Scroll(ctx context.Context, scroll interfaces.Scroll) ([]byte, int, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(scroll); err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	req := opensearchapi.ScrollRequest{
		Body: &buf,
	}
	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting scroll response from opensearch: %v", err)

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	return commonProcess(context.Background(), res)
}

// Count count获取查询数据的总数
func (osa *openSearchAccess) Count(ctx context.Context,
	query map[string]interface{}, indices []string) ([]byte, int, error) {
	// 对请求体进行编码
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	ignoreUnavailable := true
	req := opensearchapi.CountRequest{
		Body:              &buf,
		Index:             indices,
		IgnoreUnavailable: &ignoreUnavailable,
	}
	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting count response from opensearch: %v", err)

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}
	return commonProcess(context.Background(), res)
}

// 从 opensearch 中获取索引库下的索引分片数
func (osa *openSearchAccess) LoadIndexShards(ctx context.Context, indexBase string) ([]byte, int, error) {
	// GET _cat/indices/metricbeat*?v&h=index,pri&format=json
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get index[%s] shards", indexBase), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attr.Key("url").String(osAddress),
		attr.Key("api").String(fmt.Sprintf("_cat/indices/%s?v&h=index,pri&format=json", indexBase)),
	)
	defer span.End()

	withV := true
	req := opensearchapi.CatIndicesRequest{
		Index:  []string{indexBase},
		V:      &withV,
		H:      []string{"index", "pri"},
		Format: "json",
	}
	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting _cat/indices response from opensearch: %v", err)
		span.SetStatus(codes.Error, "Get _cat/indices response error")
		o11y.Error(ctx, fmt.Sprintf("Get _cat/indices response from opensearch error: %v", err))
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	result, code, err := commonProcess(ctx, res)
	if err != nil {
		span.SetStatus(codes.Error, "Process _cat/indices response error")
		o11y.Error(ctx, fmt.Sprintf("Process _cat/indices response from opensearch error: %v", err))
		return nil, code, err
	}

	span.SetStatus(codes.Ok, "")
	return result, code, nil
}

// 对 opensearch 返回结果做统一的处理
func commonProcess(ctx context.Context, res *opensearchapi.Response) ([]byte, int, error) {

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		o11y.Error(ctx, fmt.Sprintf("Read opensearch response error: %v", err))
		return nil, http.StatusInternalServerError, oerr
	}
	if res.IsError() {
		o11y.Error(ctx, fmt.Sprintf("Opensearch response is error: %v", err))
		return nil, res.StatusCode, errors.New(string(resBytes))
	}

	return resBytes, res.StatusCode, nil
}

// DeleteScroll 删除scroll查询
// deleteScroll.ScrollId=[]string{"_all"}时, 会删除所有scroll查询
func (osa *openSearchAccess) DeleteScroll(ctx context.Context,
	deleteScroll interfaces.DeleteScroll) ([]byte, int, error) {

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(deleteScroll); err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}

	req := opensearchapi.ClearScrollRequest{
		Body: &buf,
	}
	res, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting delete scroll response from opensearch: %v", err)

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}
	return commonProcess(context.Background(), res)
}

func (osa *openSearchAccess) CreatePointInTime(ctx context.Context, indices []string, keepAlive time.Duration) ([]byte, string, int, error) {
	req := opensearchapi.PointInTimeCreateRequest{
		Index:     indices,
		KeepAlive: keepAlive,
	}
	res, pointInTimeCreateResp, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting delete scroll response from opensearch: %v", err)

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, "", http.StatusInternalServerError, oerr
	}

	resBytes, code, err := commonProcess(context.Background(), res)
	if err != nil {
		return nil, "", code, err
	}

	logger.Debugf("Create point in time response: %s", string(resBytes))

	return resBytes, pointInTimeCreateResp.PitID, code, nil
}

// 批量删除 pit，如果是删除全部，pit_id 传递  "_all"
func (osa *openSearchAccess) DeletePointInTime(ctx context.Context, pitIDs []string) (*opensearchapi.PointInTimeDeleteResp, int, error) {
	req := opensearchapi.PointInTimeDeleteRequest{
		PitID: pitIDs,
	}
	res, pitDeleteResp, err := req.Do(ctx, osa.client)
	if err != nil {
		logger.Errorf("Error getting delete point in time response from opensearch: %v", err)

		oerr := uerrors.NewOpenSearchError(uerrors.InternalServerError).WithReason(err.Error())
		return nil, http.StatusInternalServerError, oerr
	}
	resBytes, code, err := commonProcess(context.Background(), res)
	if err != nil {
		return nil, code, err
	}

	if code >= http.StatusBadRequest {
		return nil, code, errors.New(string(resBytes))
	}

	return pitDeleteResp, code, nil
}
