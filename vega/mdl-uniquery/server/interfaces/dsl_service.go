// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"time"
)

var DslResult = map[string]interface{}{
	"_shards": map[string]interface{}{
		"failed":     0,
		"skipped":    0,
		"successful": 0,
		"total":      0,
	},
	"hits": map[string]interface{}{
		"total": map[string]interface{}{
			"value":    0,
			"relation": "eq",
		},
		"max_score": 0,
		"hits":      []any{},
	},
	"timed_out": false,
	"took":      0,
}

var DslCount = map[string]interface{}{
	"count": 0,
	"_shards": map[string]interface{}{
		"total":      0,
		"successful": 0,
		"skipped":    0,
		"failed":     0,
	},
}

var DslDeleteScrollResult = map[string]interface{}{
	"succeeded": true,
	"num_freed": 1,
}

//go:generate mockgen -source ../interfaces/dsl_service.go -destination ../interfaces/mock/mock_dsl_service.go
type DslService interface {
	Search(ctx context.Context, dsl map[string]interface{}, indiceAlias string, scroll time.Duration) ([]byte, int, error)
	ScrollSearch(ctx context.Context, scroll Scroll) ([]byte, int, error)
	Count(ctx context.Context, dsl map[string]interface{}, indiceAlias string) ([]byte, int, error)
	DeleteScroll(ctx context.Context, deleteScroll DeleteScroll) ([]byte, int, error)
}
