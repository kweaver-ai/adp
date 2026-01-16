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
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
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
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
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
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: ""},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
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
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: ""},
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
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "prop1"},
								TargetProp: interfaces.SimpleProperty{Name: "bridge1"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "bridge1"},
								TargetProp: interfaces.SimpleProperty{Name: "prop2"},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: nil,
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: "",
							ID:   "dv1",
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: "invalid_type",
							ID:   "dv1",
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "",
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: ""},
								TargetProp: interfaces.SimpleProperty{Name: "bridge1"},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "prop1"},
								TargetProp: interfaces.SimpleProperty{Name: ""},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "prop1"},
								TargetProp: interfaces.SimpleProperty{Name: "bridge1"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "prop1"},
								TargetProp: interfaces.SimpleProperty{Name: "bridge1"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: ""},
								TargetProp: interfaces.SimpleProperty{Name: "prop2"},
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
					MappingRules: interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "prop1"},
								TargetProp: interfaces.SimpleProperty{Name: "bridge1"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "bridge1"},
								TargetProp: interfaces.SimpleProperty{Name: ""},
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
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
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
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt)
			So(err, ShouldNotBeNil)
		})
	})
}
