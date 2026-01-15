package object_type

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

func Test_objectTypeService_CheckObjectTypeExistByID(t *testing.T) {
	Convey("Test CheckObjectTypeExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
		}

		Convey("Success when object type exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otID := "ot1"
			otName := "object_type1"

			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otName, true, nil)

			name, exist, err := service.CheckObjectTypeExistByID(ctx, knID, branch, otID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, otName)
		})

		Convey("Success when object type does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otID := "ot1"

			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckObjectTypeExistByID(ctx, knID, branch, otID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otID := "ot1"

			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			name, exist, err := service.CheckObjectTypeExistByID(ctx, knID, branch, otID)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_InternalError_CheckObjectTypeIfExistFailed)
		})
	})
}

func Test_objectTypeService_CheckObjectTypeExistByName(t *testing.T) {
	Convey("Test CheckObjectTypeExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
		}

		Convey("Success when object type exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otName := "object_type1"
			otID := "ot1"

			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otID, true, nil)

			id, exist, err := service.CheckObjectTypeExistByName(ctx, knID, branch, otName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, otID)
		})

		Convey("Success when object type does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otName := "object_type1"

			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckObjectTypeExistByName(ctx, knID, branch, otName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otName := "object_type1"

			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			id, exist, err := service.CheckObjectTypeExistByName(ctx, knID, branch, otName)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_InternalError_CheckObjectTypeIfExistFailed)
		})
	})
}

func Test_objectTypeService_GetObjectTypeIDsByKnID(t *testing.T) {
	Convey("Test GetObjectTypeIDsByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
		}

		Convey("Success getting object type IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1", "ot2"}

			ota.EXPECT().GetObjectTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)

			result, err := service.GetObjectTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, otIDs)
		})

		Convey("Success with empty result\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			ota.EXPECT().GetObjectTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.GetObjectTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			ota.EXPECT().GetObjectTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			result, err := service.GetObjectTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_InternalError_GetObjectTypesByIDsFailed)
		})
	})
}

func Test_objectTypeService_GetObjectTypesByIDs(t *testing.T) {
	Convey("Test GetObjectTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &objectTypeService{
			appSetting: appSetting,
			db:         db,
			ota:        ota,
			ps:         ps,
			cga:        cga,
		}

		Convey("Success getting object types by IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1", "ot2"}
			otArr := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
				},
			}
			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otArr, nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil)
			smock.ExpectCommit()
			result, err := service.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("Failed when object types count mismatch\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1", "ot2"}
			otArr := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otArr, nil)
			smock.ExpectCommit()
			result, err := service.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_ObjectTypeNotFound)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			result, err := service.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetObjectTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptGroupsByOTIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}
			otArr := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otArr, nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			result, err := service.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_GetAllObjectTypesByKnID(t *testing.T) {
	Convey("Test GetAllObjectTypesByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
		}

		Convey("Success getting all object types\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otMap := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			ota.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(otMap, nil)

			result, err := service.GetAllObjectTypesByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			ota.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			result, err := service.GetAllObjectTypesByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_objectTypeService_GetObjectTypeByID(t *testing.T) {
	Convey("Test GetObjectTypeByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
		}

		Convey("Success getting object type by ID\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otID := "ot1"
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   otID,
					OTName: "ot1",
				},
			}

			ota.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ot, nil)

			result, err := service.GetObjectTypeByID(ctx, knID, branch, otID)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.OTID, ShouldEqual, otID)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otID := "ot1"

			ota.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			result, err := service.GetObjectTypeByID(ctx, knID, branch, otID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_objectTypeService_CreateObjectTypes(t *testing.T) {
	Convey("Test CreateObjectTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &objectTypeService{
			appSetting: appSetting,
			db:         db,
			ota:        ota,
			ps:         ps,
			cga:        cga,
			osa:        osa,
		}

		Convey("Success creating object types with normal mode\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CreateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CreateObjectTypeStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "ot1")
		})

		Convey("Failed when permission check fails\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when object type ID already exists in normal mode\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ot1", true, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectRollback()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_ObjectTypeIDExisted)
		})

		Convey("Success with ignore mode when object type exists\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ot1", true, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ot1", true, nil)
			smock.ExpectCommit()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Ignore, false)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with Overwrite mode when ID exists\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil).AnyTimes()
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ot1", true, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ot1", true, nil)
			ota.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Overwrite, false)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with empty OTID generates new ID\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx, knID, branch, otID interface{}) {
				So(otID, ShouldNotBeEmpty)
			}).Return("", false, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CreateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CreateObjectTypeStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldNotBeEmpty)
		})

		Convey("Failed when CreateObjectType returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CreateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when CreateObjectTypeStatus returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CreateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CreateObjectTypeStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when InsertOpenSearchData returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CheckObjectTypeExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ota.EXPECT().CreateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().CreateObjectTypeStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.CreateObjectTypes(ctx, nil, objectTypes, interfaces.ImportMode_Normal, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_ListObjectTypes(t *testing.T) {
	Convey("Test ListObjectTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		uma := dmock.NewMockUserMgmtAccess(mockCtrl)
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &objectTypeService{
			appSetting: appSetting,
			db:         db,
			ota:        ota,
			ps:         ps,
			cga:        cga,
			uma:        uma,
			dva:        dva,
		}

		Convey("Success listing object types\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil)
			smock.ExpectCommit()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, nil)
			smock.ExpectCommit()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when ListObjectTypes returns error\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptGroupsByOTIDs returns error\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil)
			smock.ExpectCommit()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			query := interfaces.ObjectTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			ota.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil).AnyTimes()
			smock.ExpectCommit()

			result, total, err := service.ListObjectTypes(ctx, nil, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_UpdateObjectType(t *testing.T) {
	Convey("Test UpdateObjectType\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		mfa := dmock.NewMockModelFactoryAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &objectTypeService{
			appSetting: appSetting,
			db:         db,
			ota:        ota,
			ps:         ps,
			cga:        cga,
			mfa:        mfa,
			osa:        osa,
		}

		Convey("Success updating object type\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateObjectType(ctx, nil, objectType)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			err := service.UpdateObjectType(ctx, nil, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateObjectType returns error\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			err := service.UpdateObjectType(ctx, nil, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when syncObjectGroups returns error\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.UpdateObjectType(ctx, nil, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertOpenSearchData returns error\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsByOTIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string][]*interfaces.ConceptGroup{}, nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			err := service.UpdateObjectType(ctx, nil, objectType)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_objectTypeService_UpdateDataProperties(t *testing.T) {
	Convey("Test UpdateDataProperties\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		mfa := dmock.NewMockModelFactoryAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
			ps:         ps,
			mfa:        mfa,
			osa:        osa,
		}

		Convey("Success updating data properties\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "prop1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			dataProperties := []*interfaces.DataProperty{
				{
					Name: "prop1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateDataProperties(ctx, objectType, dataProperties)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			dataProperties := []*interfaces.DataProperty{}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			err := service.UpdateDataProperties(ctx, objectType, dataProperties)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateDataProperties returns error\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "prop1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			dataProperties := []*interfaces.DataProperty{
				{
					Name: "prop1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			err := service.UpdateDataProperties(ctx, objectType, dataProperties)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertOpenSearchData returns error\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "prop1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			dataProperties := []*interfaces.DataProperty{
				{
					Name: "prop1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			err := service.UpdateDataProperties(ctx, objectType, dataProperties)
			So(err, ShouldNotBeNil)
		})

		Convey("Success adding new property\n", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object_type1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "prop1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			dataProperties := []*interfaces.DataProperty{
				{
					Name: "prop2",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateDataProperties(ctx, objectType, dataProperties)
			So(err, ShouldBeNil)
			So(len(objectType.DataProperties), ShouldEqual, 2)
		})
	})
}

func Test_objectTypeService_DeleteObjectTypesByIDs(t *testing.T) {
	Convey("Test DeleteObjectTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &objectTypeService{
			appSetting: appSetting,
			db:         db,
			ota:        ota,
			ps:         ps,
			cga:        cga,
			osa:        osa,
		}

		Convey("Success deleting object types\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1", "ot2"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			ota.EXPECT().DeleteObjectTypeStatusByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			smock.ExpectCommit()

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 2)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("Failed when DeleteObjectTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("Failed when DeleteObjectTypeStatusByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ota.EXPECT().DeleteObjectTypeStatusByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("Failed when DeleteData returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ota.EXPECT().DeleteObjectTypeStatusByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("Failed when DeleteObjectTypesFromGroup returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ota.EXPECT().DeleteObjectTypeStatusByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))
			smock.ExpectRollback()

			result, err := service.DeleteObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_GetObjectTypesMapByIDs(t *testing.T) {
	Convey("Test GetObjectTypesMapByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			ota:        ota,
			ps:         ps,
		}

		Convey("Success getting object types map\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1", "ot2"}
			otArr := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
						DataProperties: []*interfaces.DataProperty{
							{
								Name:        "prop1",
								DisplayName: "Property1",
							},
						},
					},
				},
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "object_type2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ota.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otArr, nil)

			result, err := service.GetObjectTypesMapByIDs(ctx, knID, branch, otIDs, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
			So(result["ot1"], ShouldNotBeNil)
			So(result["ot2"], ShouldNotBeNil)
			So(result["ot1"].PropertyMap["prop1"], ShouldEqual, "Property1")
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			otIDs := []string{"ot1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_InternalError_CheckPermissionFailed))

			result, err := service.GetObjectTypesMapByIDs(ctx, knID, branch, otIDs, false)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_InsertOpenSearchData(t *testing.T) {
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

		service := &objectTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success inserting empty list\n", func() {
			objectTypes := []*interfaces.ObjectType{}

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting object types\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_objectTypeService_InsertOpenSearchData_WithVector(t *testing.T) {
	Convey("Test InsertOpenSearchData with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		Convey("Success inserting object types with vector\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
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
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := service.InsertOpenSearchData(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_objectTypeService_GetTotal(t *testing.T) {
	Convey("Test GetTotal\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total\n", func() {
			dsl := map[string]any{
				"query": map[string]any{
					"match_all": map[string]any{},
				},
			}
			countResponse := []byte(`{"count": 10}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countResponse, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 10)
		})

		Convey("Failed when count fails\n", func() {
			dsl := map[string]any{}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when sonic.Get fails\n", func() {
			dsl := map[string]any{}
			countResponse := []byte(`invalid json`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countResponse, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when Int64 conversion fails\n", func() {
			dsl := map[string]any{}
			countResponse := []byte(`{"count": "not a number"}`)

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countResponse, nil)

			total, err := service.GetTotal(ctx, dsl)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_GetTotalWithLargeOTIDs(t *testing.T) {
	Convey("Test GetTotalWithLargeOTIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total with large OTIDs\n", func() {
			conditionDslStr := `{"query":{"match_all":{}}}`
			otIDs := []string{"ot1", "ot2", "ot3"}

			// Mock GetTotalWithOTIDs calls
			countResponse := []byte(`{"count": 5}`)
			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countResponse, nil).Times(1)

			total, err := service.GetTotalWithLargeOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 5)
		})

		Convey("Success with empty OTIDs\n", func() {
			conditionDslStr := `{"query":{"match_all":{}}}`
			otIDs := []string{}

			total, err := service.GetTotalWithLargeOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when GetTotalWithOTIDs returns error\n", func() {
			conditionDslStr := `{"query":{"match_all":{}}}`
			otIDs := []string{"ot1"}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			total, err := service.GetTotalWithLargeOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_GetTotalWithOTIDs(t *testing.T) {
	Convey("Test GetTotalWithOTIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success getting total with OTIDs\n", func() {
			conditionDslStr := `{"query":{"match_all":{}}}`
			otIDs := []string{"ot1", "ot2"}

			countResponse := []byte(`{"count": 2}`)
			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(countResponse, nil)

			total, err := service.GetTotalWithOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
		})

		Convey("Failed when invalid DSL\n", func() {
			conditionDslStr := `invalid json`
			otIDs := []string{"ot1"}

			total, err := service.GetTotalWithOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when GetTotal returns error\n", func() {
			conditionDslStr := `{"query":{"match_all":{}}}`
			otIDs := []string{"ot1"}

			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			total, err := service.GetTotalWithOTIDs(ctx, conditionDslStr, otIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_SearchObjectTypes(t *testing.T) {
	Convey("Test SearchObjectTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dda := dmock.NewMockDataModelAccess(mockCtrl)

		service := &objectTypeService{
			appSetting: appSetting,
			cga:        cga,
			osa:        osa,
			dva:        dva,
			dda:        dda,
		}

		Convey("Success searching object types without concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Limit:  10,
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{}, nil)

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success searching object types with concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
				ActualCondition: &cond.CondCfg{
					Operation: "and",
					SubConds: []*cond.CondCfg{
						{
							Name:      "name",
							Operation: cond.OperationEq,
							ValueOptCfg: cond.ValueOptCfg{
								ValueFrom: "const",
								Value:     "ot1",
							},
						},
					},
				},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{}, nil)

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
		})

		Convey("Failed when concept groups not found\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				NeedTotal:     false,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, nil)

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptGroupsTotal returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptIDsByConceptGroupIDs returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ObjectType_InternalError))

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success with empty concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Limit:  10,
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{}, nil)

			result, err := service.SearchObjectTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})
	})
}
