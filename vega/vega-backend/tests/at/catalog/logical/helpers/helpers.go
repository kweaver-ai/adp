// Copyright 2025 KWeaver Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package helpers provides logical catalog specific helpers.
package helpers

import (
	cataloghelpers "vega-backend-tests/at/catalog/helpers"
)

// BuildLogicalCatalogPayload 构建logical catalog创建payload
func BuildLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        cataloghelpers.GenerateUniqueName("logical-catalog"),
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildLogicalCatalogPayloadWithName 构建带指定名称的logical catalog
func BuildLogicalCatalogPayloadWithName(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildFullLogicalCatalogPayload 构建完整字段的logical catalog
func BuildFullLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        cataloghelpers.GenerateUniqueName("full-logical-catalog"),
		"type":        "logical",
		"description": "完整的逻辑Catalog测试，包含所有可选字段",
		"tags":        []string{"test", "logical", "full"},
	}
}
