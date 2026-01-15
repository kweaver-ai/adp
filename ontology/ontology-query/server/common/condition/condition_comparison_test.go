package condition

import (
	"context"
	"testing"

	dtype "ontology-query/interfaces/data_type"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewGtCond(t *testing.T) {
	Convey("Test NewGtCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 创建大于条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewGtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 19},
				},
			}
			cond, err := NewGtCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_GtCond_Convert(t *testing.T) {
	Convey("Test GtCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &GtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `"gt"`)
		})
	})
}

func Test_GtCond_Convert2SQL(t *testing.T) {
	Convey("Test GtCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &GtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `>`)
		})
	})
}

func Test_NewGteCond(t *testing.T) {
	Convey("Test NewGteCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 创建大于等于条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGte,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewGteCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func Test_GteCond_Convert(t *testing.T) {
	Convey("Test GteCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &GteCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"gte"`)
		})
	})
}

func Test_NewLtCond(t *testing.T) {
	Convey("Test NewLtCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 创建小于条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewLtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func Test_LtCond_Convert(t *testing.T) {
	Convey("Test LtCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &LtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 65,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"lt"`)
		})
	})
}

func Test_NewLteCond(t *testing.T) {
	Convey("Test NewLteCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 创建小于等于条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewLteCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func Test_LteCond_Convert(t *testing.T) {
	Convey("Test LteCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &LteCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 65,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"lte"`)
		})
	})
}

func Test_NewNotEqCond(t *testing.T) {
	Convey("Test NewNotEqCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"name": {
				Name: "name",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_name",
				},
			},
		}

		Convey("成功 - 创建不等于条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_NotEqCond_Convert(t *testing.T) {
	Convey("Test NotEqCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"term"`)
		})
	})
}

func Test_NotEqCond_Convert2SQL(t *testing.T) {
	Convey("Test NotEqCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `<>`)
		})
	})
}
