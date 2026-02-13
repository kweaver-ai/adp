// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"bytes"
	"context"
	"time"
)

const (
	DEFAULT_PREFERENCE       = ""
	DEFAULT_TRACK_TOTAL_HITS = false
)

type Scroll struct {
	ScrollId string `json:"scroll_id" binding:"required"`
	Scroll   string `json:"scroll,omitempty"`
}

type DeleteScroll struct {
	ScrollId []string `json:"scroll_id" binding:"required"`
}

type IndexShards struct {
	IndexName string `json:"index"`
	Pri       string `json:"pri"`
	ShardNum  int    `json:"-"`
}

//go:generate mockgen -source ../interfaces/opensearch_access.go -destination ../interfaces/mock/mock_opensearch_access.go
type OpenSearchAccess interface {
	SearchSubmit(ctx context.Context, query map[string]interface{}, indices []string,
		scroll time.Duration, preference string, trackTotalHits bool) ([]byte, int, error)
	SearchSubmitWithBuffer(ctx context.Context, query bytes.Buffer, indices []string,
		scroll time.Duration, preference string) ([]byte, int, error)
	Scroll(ctx context.Context, scroll Scroll) ([]byte, int, error)
	Count(ctx context.Context, query map[string]interface{}, indices []string) ([]byte, int, error)
	LoadIndexShards(ctx context.Context, indices string) ([]byte, int, error)
	DeleteScroll(ctx context.Context, deleteScroll DeleteScroll) ([]byte, int, error)
	CreatePointInTime(ctx context.Context, indices []string, keepAlive time.Duration) ([]byte, string, int, error)
	DeletePointInTime(ctx context.Context, pitIDs []string) (*DeletePitsResp, int, error)
	SearchWithPit(ctx context.Context, query bytes.Buffer) ([]byte, int, error)
}
