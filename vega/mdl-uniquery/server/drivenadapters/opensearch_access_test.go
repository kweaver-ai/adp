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
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/interfaces"
)

func TestOpenSearchServiceSearchSubmit(t *testing.T) {
	Convey("Test opensearch submit", t, func() {

		okResp := `{
			"took" : 122,
			"timed_out" : false,
			"_shards" : {
				"total" : 6,
				"successful" : 5,
				"failed" : 1,
				"failures" : [{
					"shard" : 0,
					"index" : ".kibana",
					"node" : "jucBX9QkQIini9dLG9tZIw",
					"reason" : {
						"type" : "search_parse_exception",
						"reason" : "No mapping found for [offset] in order to sort on"
					}
				}]
			},
			"hits" : {
				"total" : 10,
				"max_score" : null,
				"hits" : [ {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPU",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 10",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 144,
						"type" : "log",
						"input_type" : "log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 144 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPT",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 9",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 128,
						"input_type" : "log",
						"count" : 1,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"type" : "log",
						"fields" : null,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 128 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPR",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 8",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"type" : "log",
						"input_type" : "log",
						"source" : "/Users/yury/logs/alice.log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"offset" : 112,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 112 ]
				} ]
			}
		}`

		errResp := `{
			"status": 400,
			"error": {
				"root_cause": [{
					"type": "illegal_argument_exception",
					"reason": "this node does not have the remote_cluster_client role"
				}],
				"type": "illegal_argument_exception",
				"reason": "this node does not have the remote_cluster_client role"
			}
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}

		req := opensearchapi.SearchRequest{}

		query := map[string]interface{}{"size": 10}
		indices := []string{"kc"}
		scroll_0 := time.Duration(0)
		scroll_1 := time.Duration(1)

		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.SearchSubmit(testCtx, query, indices, scroll_1,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})
		Convey("Encode error", func() {
			patch := ApplyMethodReturn(&json.Encoder{}, "Encode", errors.New("error"))
			defer patch.Reset()

			resByte, status, err := osAccess.SearchSubmit(testCtx, query, indices, scroll_0,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("search failed", func() {
			patches1 := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches1.Reset()

			resByte, status, err := osAccess.SearchSubmit(testCtx, query, indices, scroll_0,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.SearchSubmit(testCtx, query, indices, scroll_0,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}
			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.SearchSubmit(testCtx, query, indices, scroll_0,
				interfaces.DEFAULT_PREFERENCE, interfaces.DEFAULT_TRACK_TOTAL_HITS)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceSearchSubmitWithBuffer(t *testing.T) {
	Convey("Test opensearch submit with buffer", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `{
			"took" : 122,
			"timed_out" : false,
			"_shards" : {
				"total" : 6,
				"successful" : 5,
				"failed" : 1,
				"failures" : [{
					"shard" : 0,
					"index" : ".kibana",
					"node" : "jucBX9QkQIini9dLG9tZIw",
					"reason" : {
						"type" : "search_parse_exception",
						"reason" : "No mapping found for [offset] in order to sort on"
					}
				}]
			},
			"hits" : {
				"total" : 10,
				"max_score" : null,
				"hits" : [ {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPU",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 10",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 144,
						"type" : "log",
						"input_type" : "log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 144 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPT",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 9",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 128,
						"input_type" : "log",
						"count" : 1,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"type" : "log",
						"fields" : null,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 128 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPR",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 8",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"type" : "log",
						"input_type" : "log",
						"source" : "/Users/yury/logs/alice.log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"offset" : 112,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 112 ]
				} ]
			}
		}`

		errResp := `{
			"status": 400,
			"error": {
				"root_cause": [{
					"type": "illegal_argument_exception",
					"reason": "this node does not have the remote_cluster_client role"
				}],
				"type": "illegal_argument_exception",
				"reason": "this node does not have the remote_cluster_client role"
			}
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}

		req := opensearchapi.SearchRequest{}

		var queryBuffer bytes.Buffer
		queryBuffer.WriteString(`{"size":10}`)
		indices := []string{"kc"}
		scroll_0 := time.Duration(0)
		scroll_1 := time.Duration(1)

		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.SearchSubmitWithBuffer(testCtx, queryBuffer, indices,
				scroll_1, interfaces.DEFAULT_PREFERENCE)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("search failed", func() {
			patches1 := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches1.Reset()

			resByte, status, err := osAccess.SearchSubmitWithBuffer(testCtx, queryBuffer, indices,
				scroll_0, interfaces.DEFAULT_PREFERENCE)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.SearchSubmitWithBuffer(testCtx, queryBuffer, indices,
				scroll_0, interfaces.DEFAULT_PREFERENCE)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}
			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.SearchSubmitWithBuffer(testCtx, queryBuffer, indices,
				scroll_0, interfaces.DEFAULT_PREFERENCE)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceScroll(t *testing.T) {
	Convey("Test opensearch scroll", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `{
			"_scroll_id":"FGluY2x1ZGVfY29udGV4dF91dWlkDnF1ZXJ5VGhlbkZldGNoCRQ5MUQ4d25vQlJ1T0oyaFNpLVM5eQAAAAAAAALcFlNoWHlMdDJJUXFHcmZjQ19hNnFEbkEUOVZEOHdub0JSdU9KMmhTaS1TOXQAAAAAAAAC2RZTaFh5THQySVFxR3JmY0NfYTZxRG5BFDJwdjh3bm9CdzdnSXA3cmotYTV0AAAAAAAAAtwWSU10ZWV2dE9SWG00cWstOEZ3V2kyURQtRkQ4d25vQlJ1T0oyaFNpLVM5eQAAAAAAAALbFlNoWHlMdDJJUXFHcmZjQ19hNnFEbkEUM0p2OHdub0J3N2dJcDdyai1hNXUAAAAAAAAC3hZJTXRlZXZ0T1JYbTRxay04RndXaTJRFDJKdjh3bm9CdzdnSXA3cmotYTVzAAAAAAAAAtsWSU10ZWV2dE9SWG00cWstOEZ3V2kyURQ5bEQ4d25vQlJ1T0oyaFNpLVM5dAAAAAAAAALaFlNoWHlMdDJJUXFHcmZjQ19hNnFEbkEUMjV2OHdub0J3N2dJcDdyai1hNXQAAAAAAAAC3RZJTXRlZXZ0T1JYbTRxay04RndXaTJRFDJadjh3bm9CdzdnSXA3cmotYTVzAAAAAAAAAtoWSU10ZWV2dE9SWG00cWstOEZ3V2kyUQ=="
			"took" : 122,
			"timed_out" : false,
			"_shards" : {
				"total" : 6,
				"successful" : 5,
				"failed" : 1,
				"failures" : [{
					"shard" : 0,
					"index" : ".kibana",
					"node" : "jucBX9QkQIini9dLG9tZIw",
					"reason" : {
						"type" : "search_parse_exception",
						"reason" : "No mapping found for [offset] in order to sort on"
					}
				}]
			},
			"hits": {
				"total" : 10,
				"max_score" : null,
				"hits" : [{
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPU",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 10",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 144,
						"type" : "log",
						"input_type" : "log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 144 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPT",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 9",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"source" : "/Users/yury/logs/alice.log",
						"offset" : 128,
						"input_type" : "log",
						"count" : 1,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"type" : "log",
						"fields" : null,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 128 ]
				}, {
					"_index" : "logstash-2016.07.25",
					"_type" : "log",
					"_id" : "AVYkNv542Gim_t2htKPR",
					"_score" : null,
					"_source" : {
						"message" : "Alice message 8",
						"@version" : "1",
						"@timestamp" : "2016-07-25T22:39:55.760Z",
						"type" : "log",
						"input_type" : "log",
						"source" : "/Users/yury/logs/alice.log",
						"count" : 1,
						"fields" : null,
						"beat" : {
							"hostname" : "Yurys-MacBook-Pro.local",
							"name" : "Yurys-MacBook-Pro.local"
						},
						"offset" : 112,
						"host" : "Yurys-MacBook-Pro.local",
						"tags" : [ "beats_input_codec_plain_applied" ],
						"app" : "alice"
					},
					"sort" : [ 112 ]
				}]
			}
		}`

		errResp := `{
			"status": 400,
			"error": {
				"root_cause": [{
					"type": "illegal_argument_exception",
					"reason": "this node does not have the remote_cluster_client role"
				}],
				"type": "illegal_argument_exception",
				"reason": "this node does not have the remote_cluster_client role"
			}
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}

		req := opensearchapi.ScrollRequest{}

		scroll := interfaces.Scroll{
			Scroll:   "scroll",
			ScrollId: "2343",
		}

		Convey("response 200", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.Scroll(testCtx, scroll)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Encode error", func() {
			patch := ApplyMethodReturn(&json.Encoder{}, "Encode", errors.New("error"))
			defer patch.Reset()

			resByte, status, err := osAccess.Scroll(testCtx, scroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("search failed", func() {
			patches := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches.Reset()

			resByte, status, err := osAccess.Scroll(testCtx, scroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.Scroll(testCtx, scroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			errScroll := interfaces.Scroll{}
			resByte, status, err := osAccess.Scroll(testCtx, errScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceCount(t *testing.T) {
	Convey("Test opensearch count", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `{
			"count": 0,
			"_shards": {
				"total": 4,
				"successful": 4,
				"skipped": 0,
				"failed": 0
			}
		}`

		errResp := `{
			"status": 400,
			"error": {
				"root_cause": [{
					"type": "illegal_argument_exception",
					"reason": "this node does not have the remote_cluster_client role"
				}],
				"type": "illegal_argument_exception",
				"reason": "this node does not have the remote_cluster_client role"
			}
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}
		req := opensearchapi.CountRequest{}

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		indices := []string{"kc"}

		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.Count(testCtx, query, indices)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Encode error", func() {
			patch := ApplyMethodReturn(&json.Encoder{}, "Encode", errors.New("error"))
			defer patch.Reset()

			resByte, status, err := osAccess.Count(testCtx, query, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("count failed", func() {
			patches := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches.Reset()

			resByte, status, err := osAccess.Count(testCtx, query, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.Count(testCtx, query, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.Count(testCtx, query, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceLoadIndexShards(t *testing.T) {
	Convey("Test opensearch load index shards", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `[
			{
			  "index" : ".kibana_92668751_admin_1",
			  "pri" : "1"
			}
		  ]
		  `

		errResp := `{
			"error" : {
			  "root_cause" : [
				{
				  "type" : "index_not_found_exception",
				  "reason" : "no such index [txy]",
				  "index" : "txy",
				  "resource.id" : "txy",
				  "resource.type" : "index_or_alias",
				  "index_uuid" : "_na_"
				}
			  ],
			  "type" : "index_not_found_exception",
			  "reason" : "no such index [txy]",
			  "index" : "txy",
			  "resource.id" : "txy",
			  "resource.type" : "index_or_alias",
			  "index_uuid" : "_na_"
			},
			"status" : 404
		  }`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}
		req := opensearchapi.CatIndicesRequest{}

		indices := ".kibana"

		Convey("get _cat/indices response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.LoadIndexShards(testCtx, indices)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("LoadIndexShards failed", func() {
			patches := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches.Reset()

			resByte, status, err := osAccess.LoadIndexShards(testCtx, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("LoadIndexShards ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.LoadIndexShards(testCtx, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})

		Convey("LoadIndexShards response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.LoadIndexShards(testCtx, indices)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceDeleteScroll(t *testing.T) {
	Convey("Test opensearch delete scroll", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `{
			"succeeded": true,
			"num_freed": 1
		}`

		errResp := `{
			"error" : {
				"root_cause" : [
				  {
					"type" : "illegal_argument_exception",
					"reason" : "Cannot parse scroll id"
				  }
				],
				"type" : "illegal_argument_exception",
				"reason" : "Cannot parse scroll id",
				"caused_by" : {
				  "type" : "array_index_out_of_bounds_exception",
				  "reason" : "arraycopy: last source index 14041 out of bounds for byte[2]"
				}
			  },
			  "status" : 400
		}`

		notFoundResp := `{
			"succeeded": true,
			"num_freed": 0
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}
		req := opensearchapi.ClearScrollRequest{}

		deleteScroll := interfaces.DeleteScroll{
			ScrollId: []string{"_all"},
		}
		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldResemble, []byte(okResp))
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Encode error", func() {
			patch := ApplyMethodReturn(&json.Encoder{}, "Encode", errors.New("error"))
			defer patch.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("delete scroll failed", func() {
			patches := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			patches1 := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, `{"status":500,"error":{`+
				`"type":"UniQuery.InternalServerError","reason":"error"}}`)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
		Convey("response 404", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusNotFound,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(notFoundResp)),
			}

			patches := ApplyMethodReturn(req, "Do", resp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.DeleteScroll(testCtx, deleteScroll)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusNotFound)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, notFoundResp)
		})
	})
}

func TestOpenSearchServiceCreatePointInTime(t *testing.T) {
	Convey("Test create pit", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `{
			"pit_id": "17qEQQMhYXItYXJfYXVkaXRfbG9nLTIwMjQuMTEuMTUtMDAwMDAwFlFmTDJiWTBCUmgtMkFrSU1NUkoyQkECFlMyT0h5cTlYVGdPVThIeUE2cWZ4VFEAAAAAAAAKWuEWUmdVc01GN1pUR2UwZU1PSXYtMTJQZyFhci1hcl9hdWRpdF9sb2ctMjAyNC4xMS4xNS0wMDAwMDAWUWZMMmJZMEJSaC0yQWtJTU1SSjJCQQEWUzJPSHlxOVhUZ09VOEh5QTZxZnhUUQAAAAAAAApa4BZSZ1VzTUY3WlRHZTBlTU9Jdi0xMlBnIWFyLWFyX2F1ZGl0X2xvZy0yMDI0LjExLjE1LTAwMDAwMBZRZkwyYlkwQlJoLTJBa0lNTVJKMkJBABZTMk9IeXE5WFRnT1U4SHlBNnFmeFRRAAAAAAAAClrfFlJnVXNNRjdaVEdlMGVNT0l2LTEyUGcBFlFmTDJiWTBCUmgtMkFrSU1NUkoyQkEAAA==",
			"_shards": {
				"total": 3,
				"successful": 3,
				"skipped": 0,
				"failed": 0
			},
			"creation_time": 1732499935335
		}`
		errResp := `{
			"error": {
				"root_cause": [
				{
					"type": "illegal_argument_exception",
					"reason": "Keep alive for request (2d) is too large. It must be less than (1d). This limit can be set by changing the [point_in_time.max_keep_alive] cluster level setting."
				}
				],
				"type": "illegal_argument_exception",
				"reason": "Keep alive for request (2d) is too large. It must be less than (1d). This limit can be set by changing the [point_in_time.max_keep_alive] cluster level setting."
			},
			"status": 400
		}`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}
		req := opensearchapi.PointInTimeCreateRequest{}

		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			pointInTimeCreateResp := &opensearchapi.PointInTimeCreateResp{
				PitID: "1123==",
			}

			patches := ApplyMethodReturn(req, "Do", resp, pointInTimeCreateResp, nil)
			defer patches.Reset()

			resByte, pitID, status, err := osAccess.CreatePointInTime(testCtx, []string{"aaaa"}, time.Second)
			So(resByte, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
			So(pitID, ShouldNotBeEmpty)
		})

		Convey("create pit failed", func() {
			pointInTimeCreateResp := &opensearchapi.PointInTimeCreateResp{}
			patches := ApplyMethodReturn(req, "Do", nil, pointInTimeCreateResp, errors.New("error"))
			defer patches.Reset()

			resByte, _, status, err := osAccess.CreatePointInTime(testCtx, []string{"aaaa"}, time.Second)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}
			pointInTimeCreateResp := &opensearchapi.PointInTimeCreateResp{}
			patches1 := ApplyMethodReturn(req, "Do", resp, pointInTimeCreateResp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, _, status, err := osAccess.CreatePointInTime(testCtx, []string{"aaaa"}, time.Second)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}
			pointInTimeCreateResp := &opensearchapi.PointInTimeCreateResp{}
			patches := ApplyMethodReturn(req, "Do", resp, pointInTimeCreateResp, nil)
			defer patches.Reset()

			resByte, _, status, err := osAccess.CreatePointInTime(testCtx, []string{"aaaa"}, time.Second)
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}

func TestOpenSearchServiceDeletePointInTime(t *testing.T) {
	Convey("Test delete pit", t, func() {
		testCtx := context.WithValue(context.Background(), rest.XLangKey, "zh-CN")

		okResp := `
		{
			"pits": [
				{
					"successful": true,
					"pit_id": "o463QQEPbXktaW5kZXgtMDAwMDAxFkhGN09fMVlPUkVPLXh6MUExZ1hpaEEAFjBGbmVEZHdGU1EtaFhhUFc4ZkR5cWcAAAAAAAAAAAEWaXBPNVJtZEhTZDZXTWFFR05waXdWZwEWSEY3T18xWU9SRU8teHoxQTFnWGloQQAA"
				},
				{
					"successful": false,
					"pit_id": "o463QQEPbXktaW5kZXgtMDAwMDAxFkhGN09fMVlPUkVPLXh6MUExZ1hpaEEAFjBGbmVEZHdGU1EtaFhhUFc4ZkR5cWcAAAAAAAAAAAIWaXBPNVJtZEhTZDZXTWFFR05waXdWZwEWSEY3T18xWU9SRU8teHoxQTFnWGloQQAA"
				}
			]
		}
		`

		errResp := `
		{
			"error": {
				"root_cause": [
				{
					"type": "security_exception",
					"reason": "Unexpected exception indices:data/read/point_in_time/delete"
				}
				],
				"type": "security_exception",
				"reason": "Unexpected exception indices:data/read/point_in_time/delete"
			},
			"status": 500
		}
		`

		client, err := opensearch.NewClient(opensearch.Config{})
		So(err, ShouldBeNil)
		osAccess := &openSearchAccess{client: client}
		req := opensearchapi.PointInTimeDeleteRequest{}

		Convey("response ok", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}

			pointInTimeDeleteResp := &opensearchapi.PointInTimeDeleteResp{}

			patches := ApplyMethodReturn(req, "Do", resp, pointInTimeDeleteResp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.DeletePointInTime(testCtx, []string{"aaaa"})
			So(resByte, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("Encode error", func() {
			patch := ApplyMethodReturn(&json.Encoder{}, "Encode", errors.New("error"))
			defer patch.Reset()

			resByte, status, err := osAccess.DeletePointInTime(testCtx, []string{"aaaa"})
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})
		Convey("delete pit failed", func() {
			pointInTimeDeleteResp := &opensearchapi.PointInTimeDeleteResp{}
			patches := ApplyMethodReturn(req, "Do", nil, pointInTimeDeleteResp, errors.New("error"))
			defer patches.Reset()

			resByte, status, err := osAccess.DeletePointInTime(testCtx, []string{"aaaa"})
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})
		Convey("ioutil.ReadAll failed", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(okResp)),
			}
			pointInTimeDeleteResp := &opensearchapi.PointInTimeDeleteResp{}
			patches1 := ApplyMethodReturn(req, "Do", resp, pointInTimeDeleteResp, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(io.ReadAll, nil, errors.New("error"))
			defer patches2.Reset()

			resByte, status, err := osAccess.DeletePointInTime(testCtx, []string{"aaaa"})
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})
		Convey("response 400", func() {
			resp := &opensearchapi.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(errResp)),
			}
			pointInTimeDeleteResp := &opensearchapi.PointInTimeDeleteResp{}
			patches := ApplyMethodReturn(req, "Do", resp, pointInTimeDeleteResp, nil)
			defer patches.Reset()

			resByte, status, err := osAccess.DeletePointInTime(testCtx, []string{"aaaa"})
			So(resByte, ShouldBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, errResp)
		})
	})
}
