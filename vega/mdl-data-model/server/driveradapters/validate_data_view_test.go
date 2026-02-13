// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
)

func Test_ValidateDataView_ValidateViewFiltersType(t *testing.T) {
	Convey("Test ValidateViewFiltersType", t, func() {

		Convey("nil configuration", func() {
			cfg := (any)(nil)
			expected := (*interfaces.CondCfg)(nil)

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldEqual, expected)
		})

		Convey("CondCfg configuration", func() {
			cfg := &interfaces.CondCfg{}
			expected := cfg

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldEqual, expected)
		})

		Convey("Filter slice configuration", func() {
			cfg := []interfaces.Filter{{}, {}}
			expected := common.ConvertFiltersToCondition(cfg)

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldEqual, expected)
		})

		Convey("empty map configuration", func() {
			cfg := map[string]any{}
			expected := (*interfaces.CondCfg)(nil)

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldEqual, expected)
		})

		Convey("valid map configuration", func() {
			cfg := map[string]any{
				"field":     "f1",
				"operation": "exist",
			}
			expected := &interfaces.CondCfg{
				Name:      "f1",
				Operation: dcond.OperationExist,
			}

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, expected)
		})

		Convey("invalid map configuration", func() {
			cfg := map[string]any{
				"field":     "value1",
				"operation": 123,
			}

			_, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldNotBeNil)
		})
		Convey("valid slice configuration", func() {
			cfg := []any{
				interfaces.Filter{
					Name:      "key1",
					Operation: dcond.Operation_EQ,
					Value:     "value1",
				},
				interfaces.Filter{
					Name:      "key2",
					Value:     "value2",
					Operation: dcond.Operation_EQ,
				},
			}

			expected := &interfaces.CondCfg{
				Name:      "",
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Name:      "key1",
						Operation: dcond.OperationEq,
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: dcond.ValueFrom_Const,
							Value:     "value1",
						},
					},
					{
						Name:      "key2",
						Operation: dcond.OperationEq,
						ValueOptCfg: interfaces.ValueOptCfg{
							ValueFrom: dcond.ValueFrom_Const,
							Value:     "value2",
						},
					},
				},
			}

			result, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, expected)
		})

		Convey("invalid slice configuration", func() {
			cfg := []any{
				"invalid_filter",
			}

			_, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldNotBeNil)
		})

		Convey("invalid configuration type", func() {
			cfg := "invalid_type"

			_, err := validateViewFiltersType(testCtx, cfg)

			So(err, ShouldNotBeNil)
		})
	})
}

// func Test_ValidateDataView_ValidateDataView(t *testing.T) {
// 	Convey("Test ValidateDataView", t, func() {

// 		Convey("Validate failed, because view name is null", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "",
// 				},
// 			}
// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_NullParameter_ViewName)
// 		})

// 		Convey("Validate failed, because the length of view name exceeds the limit", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRST",
// 				},
// 			}
// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_LengthExceeded_ViewName)
// 		})

// 		Convey("Validate failed, because tag count > 5", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					Tags:     []string{"a", "b", "c", "d", "e", "f"},
// 				},
// 			}
// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_CountExceeded_Tags)
// 		})

// 		Convey("Validate failed, because comment is > 255", func() {
// 			str := ""
// 			for i := 0; i < interfaces.COMMENT_MAX_LENGTH+10; i++ {
// 				str += "a"
// 			}

// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					Comment:  str,
// 				},
// 			}
// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_LengthExceeded_Comment)
// 		})

// 		Convey("Validate failed, because datasource type is null", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName:   "xxx",
// 					DataSource: map[string]any{},
// 				},
// 			}
// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because datasource type is not string", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName:   "xxx",
// 					DataSource: map[string]any{"type": 2},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because no dataSource index_base", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 					},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because dataSource index_base count is 0", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type":       "index_base",
// 						"index_base": []any{},
// 					},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because field_scope is 1 and only support one index base", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because index base names is not a list", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type":       "index_base",
// 						"index_base": map[string]string{},
// 					},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_InvalidParameter_DataSource)
// 		})

// 		Convey("Validate failed, because unsupport datasource type", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName:   "xxx",
// 					DataSource: map[string]any{"type": "af"},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_UnsupportDataSourceType)
// 		})

// 		Convey("Validate failed, because field scope is 0 and fields are []", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 				Fields: []*interfaces.ViewField{},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_NullParameter_Fields)
// 		})

// 		Convey("Validate failed, because missing required fields", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{Name: "xxx"},
// 				},
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res, ShouldBeNil)

// 			// So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			// So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_MissingRequiredField)
// 		})

// 		Convey("Validate failed, because subConds count exceeds", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{Name: "@timestamp"},
// 					{Name: "xxx"},
// 					{Name: "__data_type"},
// 					{Name: "__index_base"},
// 					{Name: "__write_time"},
// 					{Name: "category"},
// 					{Name: "tags"},
// 					{Name: "__id"},
// 					{Name: "__routing"},
// 					{Name: "__tsid"},
// 					{Name: "__pipeline_id"},
// 				},
// 				// Condition: &interfaces.CondCfg{
// 				// 	Operation: "and",
// 				// 	SubConds: []*interfaces.CondCfg{
// 				// 		{Name: "xxx", Operation: dcond.OperationEq, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationNotEq, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationEq, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationNotLike, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationLike, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 		{Name: "xxx", Operation: dcond.OperationRegex, ValueOptCfg: interfaces.ValueOptCfg{Value: 2, ValueFrom: dcond.ValueFrom_Const}},
// 				// 	},
// 				// },
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_CountExceeded_Filters)
// 		})

// 		Convey("Validate failed, because validate filters error", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{Name: "@timestamp"},
// 					{Name: "xxx"},
// 					{Name: "__data_type"},
// 					{Name: "__index_base"},
// 					{Name: "__write_time"},
// 					{Name: "category"},
// 					{Name: "tags"},
// 					{Name: "__id"},
// 					{Name: "__routing"},
// 					{Name: "__tsid"},
// 					{Name: "__pipeline_id"},
// 				},
// 				// Condition: &interfaces.CondCfg{
// 				// 	Operation: "and",
// 				// 	SubConds: []*interfaces.CondCfg{
// 				// 		{
// 				// 			Name:      "",
// 				// 			Operation: "like",
// 				// 		},
// 				// 	},
// 				// },
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(res.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_NullParameter_FilterName)
// 		})

// 		Convey("Validate success", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewName: "xxx",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{Name: "x"},
// 							interfaces.SimpleIndexBase{Name: "y"},
// 						},
// 					},
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{Name: "@timestamp"},
// 					{Name: "message"},
// 					{Name: "__data_type"},
// 					{Name: "__index_base"},
// 					{Name: "__write_time"},
// 					{Name: "category"},
// 					{Name: "tags"},
// 					{Name: "__id"},
// 					{Name: "__routing"},
// 					{Name: "__tsid"},
// 					{Name: "__pipeline_id"},
// 				},
// 				// Condition: &interfaces.CondCfg{
// 				// 	Name: "type", Operation: "==", ValueOptCfg: interfaces.ValueOptCfg{Value: "as", ValueFrom: dcond.ValueFrom_Const},
// 				// },
// 			}

// 			res := ValidateDataView(testCtx, view)

// 			So(res, ShouldBeNil)
// 		})

// 	})
// }

func Test_ValidateDataView_ValidateGroupNameOnCreateGroup(t *testing.T) {
	Convey("Test ValidateDataViewGroupName", t, func() {

		Convey("Validate failed, because group name is empty", func() {
			groupName := ""
			err := validateGroupName(testCtx, groupName)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because group name is too long", func() {
			groupName := strings.Repeat("x", 100)
			err := validateGroupName(testCtx, groupName)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because group name contains invalid characters", func() {
			groupName := "\\/xxx"
			err := validateGroupName(testCtx, groupName)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success", func() {
			groupName := "xxx"
			err := validateGroupName(testCtx, groupName)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateDataView_ValidateViewID(t *testing.T) {
	Convey("Test ValidateViewID", t, func() {

		Convey("Validate failed, because view id contains invalid characters and is built-in", func() {
			viewID := "xxx&^xxx"
			builtin := true
			err := validateViewID(testCtx, viewID, builtin)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because view id contains invalid characters and is not built-in", func() {
			viewID := "_xxxxxx"
			builtin := false
			err := validateViewID(testCtx, viewID, builtin)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success, and view is built-in", func() {
			viewID := "__xx-lxx"
			builtin := true
			err := validateViewID(testCtx, viewID, builtin)
			So(err, ShouldBeNil)
		})

		Convey("Validate success, and view is not built-in", func() {
			viewID := "8xx-l34xx"
			builtin := false
			err := validateViewID(testCtx, viewID, builtin)
			So(err, ShouldBeNil)
		})
	})
}
