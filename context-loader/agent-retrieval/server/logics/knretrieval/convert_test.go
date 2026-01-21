// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	. "github.com/smartystreets/goconvey/convey"
)

// TestParseKnOperationType_Success 测试 ParseKnOperationType 成功场景
func TestParseKnOperationType_Success(t *testing.T) {
	Convey("TestParseKnOperationType_Success", t, func() {
		testCases := []struct {
			input    string
			expected interfaces.KnOperationType
		}{
			{"and", interfaces.KnOperationTypeAnd},
			{"or", interfaces.KnOperationTypeOr},
			{"==", interfaces.KnOperationTypeEqual},
			{"!=", interfaces.KnOperationTypeNotEqual},
			{">", interfaces.KnOperationTypeGreater},
			{"<", interfaces.KnOperationTypeLess},
			{">=", interfaces.KnOperationTypeGreaterOrEqual},
			{"<=", interfaces.KnOperationTypeLessOrEqual},
			{"in", interfaces.KnOperationTypeIn},
			{"not_in", interfaces.KnOperationTypeNotIn},
			{"like", interfaces.KnOperationTypeLike},
			{"not_like", interfaces.KnOperationTypeNotLike},
			{"range", interfaces.KnOperationTypeRange},
			{"out_range", interfaces.KnOperationTypeOutRange},
			{"exist", interfaces.KnOperationTypeExist},
			{"not_exist", interfaces.KnOperationTypeNotExist},
			{"regex", interfaces.KnOperationTypeRegex},
			{"match", interfaces.KnOperationTypeMatch},
			{"knn", interfaces.KnOperationTypeKnn},
		}

		for _, tc := range testCases {
			result, err := ParseKnOperationType(tc.input)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, tc.expected)
		}
	})
}

// TestParseKnOperationType_Invalid 测试 ParseKnOperationType 无效输入
func TestParseKnOperationType_Invalid(t *testing.T) {
	Convey("TestParseKnOperationType_Invalid", t, func() {
		invalidInputs := []string{
			"invalid",
			"AND", // 大小写敏感
			"OR",
			"equals",
			"",
		}

		for _, input := range invalidInputs {
			_, err := ParseKnOperationType(input)
			So(err, ShouldNotBeNil)
		}
	})
}
