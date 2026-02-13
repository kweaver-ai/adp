// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"sync"
)

var (
	httpClientOnce sync.Once
	httpClient     rest.HTTPClient
)

func NewHTTPClient() rest.HTTPClient {
	httpClientOnce.Do(func() {
		httpClient = rest.NewHTTPClient()
	})

	return httpClient
}

func NewHTTPClientWithOptions(opts rest.HttpClientOptions) rest.HTTPClient {
	httpClientOnce.Do(func() {
		httpClient = rest.NewHTTPClientWithOptions(opts)
	})

	return httpClient
}
