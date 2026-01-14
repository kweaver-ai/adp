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
	})
}
