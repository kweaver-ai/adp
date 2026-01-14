package concept_group

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-manager/common"
	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
	dmock "ontology-manager/interfaces/mock"
)

func Test_conceptGroupService_CheckConceptGroupExistByID(t *testing.T) {
	Convey("Test CheckConceptGroupExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success when concept group exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cgName, true, nil)

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, cgName)
		})

		Convey("Success when concept group does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed)
		})
	})
}

func Test_conceptGroupService_CheckConceptGroupExistByName(t *testing.T) {
	Convey("Test CheckConceptGroupExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success when concept group exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cgID, true, nil)

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, cgID)
		})

		Convey("Success when concept group does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed)
		})
	})
}

func Test_conceptGroupService_UpdateConceptGroupDetail(t *testing.T) {
	Convey("Test UpdateConceptGroupDetail\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success updating concept group detail\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			detail := "updated detail"

			cga.EXPECT().UpdateConceptGroupDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateConceptGroupDetail(ctx, knID, branch, cgID, detail)
			So(err, ShouldBeNil)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			detail := "updated detail"

			cga.EXPECT().UpdateConceptGroupDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))

			err := service.UpdateConceptGroupDetail(ctx, knID, branch, cgID, detail)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_InternalError)
		})
	})
}

func Test_conceptGroupService_GetStatByConceptGroup(t *testing.T) {
	Convey("Test GetStatByConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		rta := dmock.NewMockRelationTypeAccess(mockCtrl)
		ata := dmock.NewMockActionTypeAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			rta:        rta,
			ata:        ata,
		}

		Convey("Success getting statistics\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			otIDs := []string{"ot1", "ot2"}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 2)
			So(stats.RtTotal, ShouldEqual, 5)
			So(stats.AtTotal, ShouldEqual, 3)
		})

		Convey("Success with empty object types\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 0)
			So(stats.RtTotal, ShouldEqual, 0)
			So(stats.AtTotal, ShouldEqual, 0)
		})

		Convey("Failed when getting concept IDs returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
		})
	})
}

func Test_conceptGroupService_ListConceptGroups(t *testing.T) {
	Convey("Test ListConceptGroups\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		uma := dmock.NewMockUserMgmtAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			uma:        uma,
		}

		Convey("Success listing concept groups\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(cgs), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, nil)

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})
	})
}

func Test_conceptGroupService_GetConceptGroupByID(t *testing.T) {
	Convey("Test GetConceptGroupByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
		}

		Convey("Success getting concept group by ID\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.CGID, ShouldEqual, cgID)
		})

		Convey("Failed when concept group not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_ConceptGroupNotFound)
		})
	})
}

func Test_conceptGroupService_InsertOpenSearchData(t *testing.T) {
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

		service := &conceptGroupService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success inserting OpenSearch data\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertOpenSearchData(ctx, conceptGroup)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))

			err := service.InsertOpenSearchData(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_conceptGroupService_DeleteConceptGroupByID(t *testing.T) {
	Convey("Test DeleteConceptGroupByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			osa:        osa,
			db:         db,
		}

		Convey("Success deleting concept group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			rowsAffected, err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ConceptGroup_InternalError))

			rowsAffected, err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
			So(rowsAffected, ShouldEqual, 0)
		})

		Convey("Failed when DeleteConceptGroupByID returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			rowsAffected, err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
			So(rowsAffected, ShouldEqual, 0)
		})
	})
}

func Test_conceptGroupService_UpdateConceptGroup(t *testing.T) {
	Convey("Test UpdateConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			osa:        osa,
			db:         db,
		}

		Convey("Success updating concept group\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ConceptGroup_InternalError))

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateConceptGroup returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_conceptGroupService_ListConceptGroupRelations(t *testing.T) {
	Convey("Test ListConceptGroupRelations\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			db:         db,
		}

		Convey("Success listing concept group relations\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgrArr := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgrArr, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ConceptGroup_InternalError))

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_conceptGroupService_DeleteObjectTypesFromGroup(t *testing.T) {
	Convey("Test DeleteObjectTypesFromGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			db:         db,
		}

		Convey("Success deleting object types from group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1", "ot2"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			smock.ExpectCommit()

			rowsAffected, err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ConceptGroup_InternalError))

			rowsAffected, err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldNotBeNil)
			So(rowsAffected, ShouldEqual, 0)
		})

		Convey("Failed when DeleteObjectTypesFromGroup returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_ConceptGroup_InternalError))
			smock.ExpectRollback()

			rowsAffected, err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldNotBeNil)
			So(rowsAffected, ShouldEqual, 0)
		})
	})
}

func Test_conceptGroupService_CreateConceptGroup(t *testing.T) {
	Convey("Test CreateConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			osa:        osa,
			db:         db,
		}

		Convey("Success creating concept group with normal mode\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Failed when permission check fails\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_ConceptGroup_InternalError))

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Failed when concept group ID already exists in normal mode\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", true, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_ConceptGroupIDExisted)
		})
	})
}

func Test_conceptGroupService_AddObjectTypesToConceptGroup(t *testing.T) {
	Convey("Test AddObjectTypesToConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ots:        ots,
			db:         db,
		}

		Convey("Success adding object types to concept group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
				{ID: "ot2"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
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
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 2, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil)
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldBeNil)
			So(len(cgrIDs), ShouldEqual, 2)
		})

		Convey("Failed when object types not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 0, nil)
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ObjectType_ObjectTypeNotFound)
		})

		Convey("Failed when relation already exists in normal mode\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}
			cgRelations := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 1, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgRelations, nil)
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_ConceptGroup_ConceptGroupRelationExisted)
		})
	})
}
