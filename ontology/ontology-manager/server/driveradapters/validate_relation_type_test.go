package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
)

func Test_ValidateRelationType(t *testing.T) {
	Convey("Test ValidateRelationType\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid relation type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "_invalid_id",
					RTName: "relation1",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_RelationType_NullParameter_Name)
		})

		Convey("Failed with invalid type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "invalid_type",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with mapping rules but empty type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "",
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with direct mapping rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldBeNil)
		})

		Convey("Failed with direct mapping rules empty source prop\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": ""},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with direct mapping rules empty target prop\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": ""},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with direct mapping rules invalid format\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:         "rt1",
					RTName:       "relation1",
					Type:         interfaces.RELATION_TYPE_DIRECT,
					MappingRules: "invalid_format",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with data_view mapping rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "bridge1"},
								"target_property": map[string]string{"name": "prop2"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source.type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": "",
							"id":   "dv1",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules invalid backing_data_source.type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": "invalid_type",
							"id":   "dv1",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source.id\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source prop in source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": ""},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target prop in source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": ""},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source prop in target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": ""},
								"target_property": map[string]string{"name": "prop2"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target prop in target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "bridge1"},
								"target_property": map[string]string{"name": ""},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules invalid format\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:         "rt1",
					RTName:       "relation1",
					Type:         interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: "invalid_format",
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid type in validateMappingRules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "invalid_type",
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})
	})
}
