package knowledge_network

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

func Test_knowledgeNetworkService_CheckKNExistByID(t *testing.T) {
	Convey("Test CheckKNExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
		}

		Convey("Success when KN exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			knName := "knowledge_network1"

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(knName, true, nil)

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, knName)
		})

		Convey("Success when KN does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_InternalError_CheckKNIfExistFailed)
		})
	})
}

func Test_knowledgeNetworkService_CheckKNExistByName(t *testing.T) {
	Convey("Test CheckKNExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
		}

		Convey("Success when KN exists\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH
			knID := "kn1"

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(knID, true, nil)

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, knID)
		})

		Convey("Success when KN does not exist\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_InternalError_CheckKNIfExistFailed)
		})
	})
}

func Test_knowledgeNetworkService_UpdateKNDetail(t *testing.T) {
	Convey("Test UpdateKNDetail\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
		}

		Convey("Success updating KN detail\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			detail := "updated detail"

			kna.EXPECT().UpdateKNDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateKNDetail(ctx, knID, branch, detail)
			So(err, ShouldBeNil)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			detail := "updated detail"

			kna.EXPECT().UpdateKNDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			err := service.UpdateKNDetail(ctx, knID, branch, detail)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_InternalError)
		})
	})
}

func Test_knowledgeNetworkService_GetStatByKN(t *testing.T) {
	Convey("Test GetStatByKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		rta := dmock.NewMockRelationTypeAccess(mockCtrl)
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			ota:        ota,
			rta:        rta,
			ata:        ata,
			cga:        cga,
		}

		Convey("Success getting statistics\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(2, nil)

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 10)
			So(stats.RtTotal, ShouldEqual, 5)
			So(stats.AtTotal, ShouldEqual, 3)
			So(stats.CgTotal, ShouldEqual, 2)
		})

		Convey("Failed when getting object types total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_InternalError_GetObjectTypesTotalFailed)
		})
	})
}

func Test_knowledgeNetworkService_ListKNs(t *testing.T) {
	Convey("Test ListKNs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		uma := dmock.NewMockUserMgmtAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			uma:        uma,
		}

		Convey("Success listing KNs\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.ResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(kns), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})
	})
}

func Test_knowledgeNetworkService_GetKNByID(t *testing.T) {
	Convey("Test GetKNByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		uma := dmock.NewMockUserMgmtAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			uma:        uma,
		}

		Convey("Success getting KN by ID\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.ResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			uma.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.KNID, ShouldEqual, knID)
		})

		Convey("Failed when KN not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_NotFound)
		})
	})
}

func Test_knowledgeNetworkService_InsertOpenSearchData(t *testing.T) {
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

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success inserting OpenSearch data\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertOpenSearchData(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			err := service.InsertOpenSearchData(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_UpdateKN(t *testing.T) {
	Convey("Test UpdateKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			osa:        osa,
			db:         db,
		}

		Convey("Success updating KN\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateKN(ctx, nil, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			err := service.UpdateKN(ctx, nil, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateKN returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.UpdateKN(ctx, nil, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_DeleteKN(t *testing.T) {
	Convey("Test DeleteKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		bsa := dmock.NewMockBusinessSystemAccess(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)
		rts := dmock.NewMockRelationTypeService(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			osa:        osa,
			bsa:        bsa,
			ots:        ots,
			rts:        rts,
			ats:        ats,
			db:         db,
		}

		Convey("Success deleting KN\n", func() {
			kn := &interfaces.KN{
				KNID:           "kn1",
				KNName:         "kn1",
				Branch:         interfaces.MAIN_BRANCH,
				BusinessDomain: "bd1",
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().GetObjectTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			rts.EXPECT().GetRelationTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			ats.EXPECT().GetActionTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)
			osa.EXPECT().DeleteData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bsa.EXPECT().UnbindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			rts.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			ats.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			smock.ExpectCommit()

			rowsAffected, err := service.DeleteKN(ctx, kn)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			rowsAffected, err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(rowsAffected, ShouldEqual, 0)
		})
	})
}

func Test_knowledgeNetworkService_ListKnSrcs(t *testing.T) {
	Convey("Test ListKnSrcs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
		}

		Convey("Success listing KN sources\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knList := []interfaces.Resource{
				{
					ID:   "kn1",
					Name: "kn1",
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.ResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(resources), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return([]interfaces.Resource{}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(resources), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(resources), ShouldEqual, 0)
		})
	})
}

func Test_knowledgeNetworkService_GetRelationTypePaths(t *testing.T) {
	Convey("Test GetRelationTypePaths\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ots:        ots,
		}

		Convey("Success getting relation type paths\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        1,
			}
			objectType := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}
			neighborPathsMap := map[string][]interfaces.RelationTypePath{
				"ot1": {
					{
						ObjectTypes: []interfaces.ObjectTypeWithKeyField{
							{OTID: "ot1"},
							{OTID: "ot2"},
						},
						TypeEdges: []interfaces.TypeEdge{
							{RelationTypeId: "rt1"},
						},
						Length: 1,
					},
				},
			}

			ots.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, nil).AnyTimes()
			kna.EXPECT().GetNeighborPathsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(neighborPathsMap, nil)

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldBeNil)
			So(len(paths), ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("Success with path length 0\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        0,
			}
			objectType := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			ots.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, nil).AnyTimes()

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldBeNil)
			So(len(paths), ShouldEqual, 1)
		})
	})
}

func Test_knowledgeNetworkService_CreateKN(t *testing.T) {
	Convey("Test CreateKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		kna := dmock.NewMockKNAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		bsa := dmock.NewMockBusinessSystemAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			osa:        osa,
			bsa:        bsa,
			db:         db,
		}

		Convey("Success creating KN with normal mode\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bsa.EXPECT().BindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			knID, err := service.CreateKN(ctx, kn, mode)
			So(err, ShouldBeNil)
			So(knID, ShouldNotBeEmpty)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, oerrors.OntologyManager_KnowledgeNetwork_InternalError))

			knID, err := service.CreateKN(ctx, kn, mode)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when KN ID already exists in normal mode\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", true, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyManager_KnowledgeNetwork_KNIDExisted)
		})
	})
}
