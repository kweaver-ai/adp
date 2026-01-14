package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
)

func Test_ValidateActionType(t *testing.T) {
	Convey("Test ValidateActionType\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid action type\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "action1",
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "_invalid_id",
					ATName: "action1",
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "",
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_NullParameter_Name)
		})

		Convey("Failed with invalid action source type\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "action1",
					ActionSource: interfaces.ActionSource{
						Type: "invalid_type",
					},
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with tool type having mcp data\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "action1",
					ActionSource: interfaces.ActionSource{
						Type:  interfaces.ACTION_TYPE_TOOL,
						McpID: "mcp1",
					},
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty parameter name\n", func() {
			at := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "action1",
					Parameters: []interfaces.Parameter{
						{
							Name: "",
						},
					},
				},
			}
			err := ValidateActionType(ctx, at)
			So(err, ShouldNotBeNil)
		})
	})
}
