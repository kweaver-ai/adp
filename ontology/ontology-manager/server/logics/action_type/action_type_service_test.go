package action_type

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-manager/common"
	cond "ontology-manager/common/condition"
	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
	dmock "ontology-manager/interfaces/mock"
)

func Test_actionTypeService_CheckActionTypeExistByID(t *testing.T) {
	Convey("Test CheckActionTypeExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
		}

		Convey("Success when action type exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atID := "at1"
			atName := "action_type1"

			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atName, true, nil)

			name, exist, err := service.CheckActionTypeExistByID(ctx, knID, branch, atID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, atName)
		})

		Convey("Success when action type does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atID := "at1"

			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckActionTypeExistByID(ctx, knID, branch, atID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atID := "at1"

			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			name, exist, err := service.CheckActionTypeExistByID(ctx, knID, branch, atID)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_InternalError_CheckActionTypeIfExistFailed)
		})
	})
}

func Test_actionTypeService_CheckActionTypeExistByName(t *testing.T) {
	Convey("Test CheckActionTypeExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
		}

		Convey("Success when action type exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atName := "action_type1"
			atID := "at1"

			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atID, true, nil)

			id, exist, err := service.CheckActionTypeExistByName(ctx, knID, branch, atName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, atID)
		})

		Convey("Success when action type does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atName := "action_type1"

			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckActionTypeExistByName(ctx, knID, branch, atName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atName := "action_type1"

			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			id, exist, err := service.CheckActionTypeExistByName(ctx, knID, branch, atName)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_InternalError_CheckActionTypeIfExistFailed)
		})
	})
}

func Test_actionTypeService_GetActionTypeIDsByKnID(t *testing.T) {
	Convey("Test GetActionTypeIDsByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
		}

		Convey("Success getting action type IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1", "at2"}

			ata.EXPECT().GetActionTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(atIDs, nil)

			result, err := service.GetActionTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, atIDs)
		})

		Convey("Success with empty result\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			ata.EXPECT().GetActionTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.GetActionTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			ata.EXPECT().GetActionTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.GetActionTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_InternalError_GetActionTypesByIDsFailed)
		})
	})
}

func Test_actionTypeService_GetActionTypesByIDs(t *testing.T) {
	Convey("Test GetActionTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
			ps:         ps,
			ots:        ots,
		}

		Convey("Success getting action types by IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1", "at2"}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at2",
						ATName:       "at2",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil).AnyTimes()

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("Failed when action types count mismatch\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1", "at2"}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atArr, nil)

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_ActionTypeNotFound)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetActionTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetObjectTypesMapByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with Affect object type\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
						Affect: &interfaces.ActionAffect{
							ObjectTypeID: "ot2",
						},
					},
				},
			}
			objectTypeMap := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "Object Type 1",
					},
					CommonInfo: interfaces.CommonInfo{
						Icon:  "icon1",
						Color: "color1",
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "Object Type 2",
					},
					CommonInfo: interfaces.CommonInfo{
						Icon:  "icon2",
						Color: "color2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypeMap, nil)

			result, err := service.GetActionTypesByIDs(ctx, knID, branch, atIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].ObjectType.OTID, ShouldEqual, "ot1")
			So(result[0].Affect.ObjectType.OTID, ShouldEqual, "ot2")
		})
	})
}

func Test_actionTypeService_ListActionTypes(t *testing.T) {
	Convey("Test ListActionTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)
		uma := dmock.NewMockUserMgmtAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
			ps:         ps,
			ots:        ots,
			uma:        uma,
		}

		Convey("Success listing action types\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(ats), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, nil)

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ActionType_InternalError))

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Failed when ListActionTypes returns error\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Failed when GetObjectTypesMapByIDs returns error\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(ats), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(ats), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			atArr := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "at1",
						ObjectTypeID: "ot1",
					},
				},
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at2",
						ATName:       "at2",
						ObjectTypeID: "ot1",
					},
				},
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at3",
						ATName:       "at3",
						ObjectTypeID: "ot1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(atArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil).AnyTimes()
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			ats, total, err := service.ListActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(len(ats), ShouldEqual, 2)
			So(ats[0].ATID, ShouldEqual, "at2")
		})
	})
}

func Test_actionTypeService_GetTotal(t *testing.T) {
	Convey("Test GetTotal\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total\n", func() {
			dsl := map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
			}
			countBytes := []byte(`{"count": 10}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countBytes, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 10)
		})

		Convey("Failed when Count returns error\n", func() {
			dsl := map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
			}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when Get returns error\n", func() {
			dsl := map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
			}
			countBytes := []byte(`{"invalid": "json"}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countBytes, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when Int64 conversion returns error\n", func() {
			dsl := map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
			}
			countBytes := []byte(`{"count": "not_a_number"}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countBytes, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_actionTypeService_GetTotalWithATIDs(t *testing.T) {
	Convey("Test GetTotalWithATIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total with ATIDs\n", func() {
			conditionDslStr := `{"match_all": {}}`
			atIDs := []string{"at1", "at2"}
			countBytes := []byte(`{"count": 2}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countBytes, nil)

			total, err := service.GetTotalWithATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
		})

		Convey("Failed when unmarshal dsl fails\n", func() {
			conditionDslStr := `invalid json`
			atIDs := []string{"at1"}

			total, err := service.GetTotalWithATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_InternalError_UnMarshalDataFailed)
		})

		Convey("Failed when GetTotal returns error\n", func() {
			conditionDslStr := `{"match_all": {}}`
			atIDs := []string{"at1"}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			total, err := service.GetTotalWithATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_actionTypeService_GetTotalWithLargeATIDs(t *testing.T) {
	Convey("Test GetTotalWithLargeATIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total with large ATIDs\n", func() {
			conditionDslStr := `{"match_all": {}}`
			atIDs := []string{"at1", "at2", "at3"}
			countBytes := []byte(`{"count": 1}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countBytes, nil).AnyTimes()

			total, err := service.GetTotalWithLargeATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldBeNil)
			So(total, ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("Success with empty ATIDs\n", func() {
			conditionDslStr := `{"match_all": {}}`
			atIDs := []string{}

			total, err := service.GetTotalWithLargeATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when GetTotalWithATIDs returns error\n", func() {
			conditionDslStr := `{"match_all": {}}`
			atIDs := []string{"at1", "at2"}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			total, err := service.GetTotalWithLargeATIDs(ctx, conditionDslStr, atIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_actionTypeService_InsertOpenSearchData(t *testing.T) {
	Convey("Test InsertOpenSearchData\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success inserting OpenSearch data\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success with empty action types\n", func() {
			actionTypes := []*interfaces.ActionType{}

			err := service.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			err := service.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Success inserting OpenSearch data with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			osaWithVector := dmock.NewMockOpenSearchAccess(mockCtrl)
			mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &actionTypeService{
				appSetting: appSettingWithVector,
				osa:        osaWithVector,
				mfa:        mfa,
			}

			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					CommonInfo: interfaces.CommonInfo{
						Tags:    []string{"tag1"},
						Comment: "comment",
						Detail:  "detail",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			osaWithVector.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := serviceWithVector.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldBeNil)
			So(len(actionTypes[0].Vector), ShouldEqual, 3)
		})

		Convey("Failed when GetDefaultModel returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &actionTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			err := serviceWithVector.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &actionTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			err := serviceWithVector.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vectors count mismatch with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &actionTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := serviceWithVector.InsertOpenSearchData(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_actionTypeService_DeleteActionTypesByIDs(t *testing.T) {
	Convey("Test DeleteActionTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
			db:         db,
			ps:         ps,
			osa:        osa,
		}

		Convey("Success deleting action types\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1", "at2"}
			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()
			err := service.DeleteActionTypesByIDs(ctx, nil, knID, branch, atIDs)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ActionType_InternalError))

			err := service.DeleteActionTypesByIDs(ctx, nil, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteActionTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}
			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))
			smock.ExpectCommit()
			err := service.DeleteActionTypesByIDs(ctx, nil, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteData returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1"}
			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))
			smock.ExpectCommit()
			err := service.DeleteActionTypesByIDs(ctx, nil, knID, branch, atIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with rowsAffect != len(atIDs)\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			atIDs := []string{"at1", "at2"}
			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()
			err := service.DeleteActionTypesByIDs(ctx, nil, knID, branch, atIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_actionTypeService_UpdateActionType(t *testing.T) {
	Convey("Test UpdateActionType\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
			db:         db,
			ps:         ps,
			osa:        osa,
		}

		Convey("Success updating action type\n", func() {
			actionType := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "at1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			err := service.UpdateActionType(ctx, nil, actionType)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			actionType := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "at1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ActionType_InternalError))

			err := service.UpdateActionType(ctx, nil, actionType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateActionType returns error\n", func() {
			actionType := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "at1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))
			smock.ExpectCommit()
			err := service.UpdateActionType(ctx, nil, actionType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertOpenSearchData returns error\n", func() {
			actionType := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:   "at1",
					ATName: "at1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))
			smock.ExpectCommit()
			err := service.UpdateActionType(ctx, nil, actionType)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_actionTypeService_CreateActionTypes(t *testing.T) {
	Convey("Test CreateActionTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &actionTypeService{
			appSetting: appSetting,
			ata:        ata,
			db:         db,
			ps:         ps,
			osa:        osa,
		}

		Convey("Success creating action types with normal mode\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CreateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 1)
		})

		Convey("Failed when permission check fails\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ActionType_InternalError))

			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldNotBeNil)
			So(len(atIDs), ShouldEqual, 0)
		})

		Convey("Failed when action type ID already exists in normal mode\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("at1", true, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldNotBeNil)
			So(len(atIDs), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ActionType_ActionTypeIDExisted)
		})

		Convey("Success with empty ATID generates new ID\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CreateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx, tx, at interface{}) {
				atType := at.(*interfaces.ActionType)
				So(atType.ATID, ShouldNotBeEmpty)
			}).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 1)
		})

		Convey("Success with Ignore mode when action type exists\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Ignore

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("at1", true, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 0)
		})

		Convey("Success with Overwrite mode when ID exists\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Overwrite

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("at1", true, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("at1", true, nil)
			ata.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 0)
		})

		Convey("Failed when InsertOpenSearchData returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "at1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ata.EXPECT().CheckActionTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CheckActionTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ata.EXPECT().CreateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))
			smock.ExpectCommit()
			atIDs, err := service.CreateActionTypes(ctx, nil, actionTypes, mode)
			So(err, ShouldNotBeNil)
			So(len(atIDs), ShouldEqual, 0)
		})
	})
}

func Test_actionTypeService_SearchActionTypes(t *testing.T) {
	Convey("Test SearchActionTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		service := &actionTypeService{
			appSetting: appSetting,
			osa:        osa,
			cga:        cga,
		}

		Convey("Success searching action types without concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:      "kn1",
				Branch:    interfaces.MAIN_BRANCH,
				Limit:     10,
				NeedTotal: false,
			}
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"at_id":   "at1",
						"at_name": "at1",
					},
					Sort: []any{"1234567890"},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			result, err := service.SearchActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.Entries), ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("Success searching action types with concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				NeedTotal:     false,
				ConceptGroups: []string{"cg1"},
				ActualCondition: &cond.CondCfg{
					Operation: cond.OperationAnd,
					SubConds: []*cond.CondCfg{
						{
							Name:      "name",
							Operation: cond.OperationEq,
							ValueOptCfg: cond.ValueOptCfg{
								ValueFrom: "const",
								Value:     "at1",
							},
						},
					},
				},
			}
			atIDs := []string{"at1", "at2"}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetActionTypeIDsFromConceptGroupRelation(gomock.Any(), gomock.Any()).Return(atIDs, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{}, nil)

			result, err := service.SearchActionTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("Failed when concept groups not found\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				NeedTotal:     false,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, nil)

			result, err := service.SearchActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptGroupsTotal returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				NeedTotal:     false,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.SearchActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetActionTypeIDsFromConceptGroupRelation returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				NeedTotal:     false,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetActionTypeIDsFromConceptGroupRelation(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ActionType_InternalError))

			result, err := service.SearchActionTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})
	})
}
