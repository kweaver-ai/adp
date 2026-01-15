package condition

import (
	"context"
	"fmt"
	"testing"

	dtype "ontology-query/interfaces/data_type"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewMatchCond(t *testing.T) {
	Convey("Test NewMatchCond", t, func() {
		ctx := context.Background()
		textFieldsMap := map[string]*DataProperty{
			"text_field": {
				Name: "text_field",
				Type: dtype.DATATYPE_TEXT,
				MappedField: Field{
					Name: "mapped_text",
				},
			},
			"string_field": {
				Name: "string_field",
				Type: dtype.DATATYPE_STRING,
				IndexConfig: &IndexConfig{
					FulltextConfig: FulltextConfig{
						Enabled: true,
					},
				},
				MappedField: Field{
					Name: "mapped_string",
				},
			},
			"string_field_no_fulltext": {
				Name: "string_field_no_fulltext",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_string",
				},
			},
		}

		Convey("成功 - TEXT字段", func() {
			cfg := &CondCfg{
				Name:      "text_field",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - STRING字段（配置了全文索引）", func() {
			cfg := &CondCfg{
				Name:      "string_field",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - AllField", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - STRING字段（未配置全文索引）", func() {
			cfg := &CondCfg{
				Name:      "string_field_no_fulltext",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_MatchCond_Convert(t *testing.T) {
	Convey("Test MatchCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &MatchCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldNames: []string{"text_field"},
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"multi_match"`)
			So(result, ShouldContainSubstring, `"query"`)
			So(result, ShouldContainSubstring, `"fields"`)
		})
	})
}

func Test_MatchCond_Convert2SQL(t *testing.T) {
	Convey("Test MatchCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL（返回空）", func() {
			cond := &MatchCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldNames: []string{"text_field"},
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "")
		})
	})
}

func Test_rewriteMatchCond(t *testing.T) {
	Convey("Test rewriteMatchCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "text_field",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: &DataProperty{
					Name: "text_field",
					MappedField: Field{
						Name: "mapped_text",
					},
				},
			}
			result, err := rewriteMatchCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_text")
		})

		Convey("成功 - AllField", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			result, err := rewriteMatchCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, AllField)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "text_field",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteMatchCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_NewMatchPhraseCond(t *testing.T) {
	Convey("Test NewMatchPhraseCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"string_field": {
				Name: "string_field",
				Type: dtype.DATATYPE_STRING,
				IndexConfig: &IndexConfig{
					FulltextConfig: FulltextConfig{
						Enabled: true,
					},
				},
				MappedField: Field{
					Name: "mapped_string",
				},
			},
			"string_field_no_fulltext": {
				Name: "string_field_no_fulltext",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_string",
				},
			},
		}

		Convey("成功 - 配置了全文索引的字段", func() {
			cfg := &CondCfg{
				Name:      "string_field",
				Operation: OperationMatchPhrase,
				ValueOptCfg: ValueOptCfg{
					Value: "test phrase",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - AllField", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMatchPhrase,
				ValueOptCfg: ValueOptCfg{
					Value: "test phrase",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 未配置全文索引", func() {
			cfg := &CondCfg{
				Name:      "string_field_no_fulltext",
				Operation: OperationMatchPhrase,
				ValueOptCfg: ValueOptCfg{
					Value: "test phrase",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_MatchPhraseCond_Convert(t *testing.T) {
	Convey("Test MatchPhraseCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &MatchPhraseCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test phrase",
					},
				},
				mFilterFieldNames: []string{"string_field"},
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"multi_match"`)
			So(result, ShouldContainSubstring, `"type"`)
			So(result, ShouldContainSubstring, `"phrase"`)
		})
	})
}

func Test_NewMultiMatchCond(t *testing.T) {
	Convey("Test NewMultiMatchCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"text_field": {
				Name: "text_field",
				Type: dtype.DATATYPE_TEXT,
				MappedField: Field{
					Name: "mapped_text",
				},
			},
			"string_field": {
				Name: "string_field",
				Type: dtype.DATATYPE_STRING,
				IndexConfig: &IndexConfig{
					FulltextConfig: FulltextConfig{
						Enabled: true,
					},
				},
				MappedField: Field{
					Name: "mapped_string",
				},
			},
		}

		Convey("成功 - 指定字段", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields": []any{"text_field"},
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - AllField", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields": []any{AllField},
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 指定match_type", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields":     []any{"text_field"},
					"match_type": "most_fields",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - fields不是数组", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields": "text_field",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - fields包含非字符串元素", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields": []any{"text_field", 123},
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 无效的match_type", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"fields":     []any{"text_field"},
					"match_type": "invalid_type",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_MultiMatchCond_Convert(t *testing.T) {
	Convey("Test MultiMatchCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL（有fields）", func() {
			cond := &MultiMatchCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
					RemainCfg: map[string]any{
						"match_type": "best_fields",
					},
				},
				mFilterFieldNames: []string{"text_field"},
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"multi_match"`)
			So(result, ShouldContainSubstring, `"type"`)
			So(result, ShouldContainSubstring, `"fields"`)
		})

		Convey("成功 - 转换DSL（无fields）", func() {
			cond := &MultiMatchCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldNames: []string{},
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"multi_match"`)
			So(result, ShouldNotContainSubstring, `"fields"`)
		})
	})
}

func Test_NewKnnCond(t *testing.T) {
	Convey("Test NewKnnCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"vector_field": {
				Name: "vector_field",
				Type: dtype.DATATYPE_VECTOR,
				IndexConfig: &IndexConfig{
					VectorConfig: VectorConfig{
						Enabled: true,
						ModelID: "model1",
					},
				},
				MappedField: Field{
					Name: "mapped_vector",
				},
			},
			"vector_field_no_config": {
				Name: "vector_field_no_config",
				Type: dtype.DATATYPE_VECTOR,
				MappedField: Field{
					Name: "mapped_vector",
				},
			},
		}

		Convey("成功 - 配置了向量索引的字段", func() {
			cfg := &CondCfg{
				Name:      "vector_field",
				Operation: OperationKNN,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - AllField", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationKNN,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 未配置向量索引", func() {
			cfg := &CondCfg{
				Name:      "vector_field_no_config",
				Operation: OperationKNN,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_KnnCond_Convert(t *testing.T) {
	Convey("Test KnnCond Convert", t, func() {
		ctx := context.Background()
		vectorizer := func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error) {
			return []VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}, nil
		}

		Convey("成功 - 转换DSL（无子条件）", func() {
			cond := &KnnCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
					NameField: &DataProperty{
						Name: "vector_field",
					},
					RemainCfg: map[string]any{
						"limit_key":   "k",
						"limit_value": 10,
					},
				},
				mFilterFieldName: "_vector_vector_field",
				mSubConds:        []Condition{},
			}
			result, err := cond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"knn"`)
			So(result, ShouldContainSubstring, `"_vector_vector_field"`)
		})

		Convey("成功 - 转换DSL（有子条件）", func() {
			fieldsMap := map[string]*DataProperty{
				"name": {
					Name: "name",
					Type: dtype.DATATYPE_STRING,
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			subCfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			subCond, _ := NewEqCond(ctx, subCfg, fieldsMap)
			cond := &KnnCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
					NameField: &DataProperty{
						Name: "vector_field",
					},
					RemainCfg: map[string]any{
						"limit_key":   "k",
						"limit_value": 10,
					},
				},
				mFilterFieldName: "_vector_vector_field",
				mSubConds:        []Condition{subCond},
			}
			result, err := cond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"knn"`)
			So(result, ShouldContainSubstring, `"filter"`)
		})

		Convey("失败 - vectorizer错误", func() {
			failingVectorizer := func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error) {
				return nil, fmt.Errorf("vectorizer error")
			}
			cond := &KnnCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
					NameField: &DataProperty{
						Name: "vector_field",
					},
					RemainCfg: map[string]any{
						"limit_key":   "k",
						"limit_value": 10,
					},
				},
				mFilterFieldName: "_vector_vector_field",
				mSubConds:        []Condition{},
			}
			result, err := cond.Convert(ctx, failingVectorizer)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, "")
		})
	})
}
