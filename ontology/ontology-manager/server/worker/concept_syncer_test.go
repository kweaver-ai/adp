package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-manager/common"
	"ontology-manager/interfaces"
	dmock "ontology-manager/interfaces/mock"
)

func TestNewConceptSyncer(t *testing.T) {
	Convey("Test NewConceptSyncer", t, func() {
		appSetting := &common.AppSetting{}

		syncer1 := NewConceptSyncer(appSetting)
		syncer2 := NewConceptSyncer(appSetting)

		Convey("Should return singleton instance", func() {
			So(syncer1, ShouldNotBeNil)
			So(syncer2, ShouldEqual, syncer1)
		})
	})
}

func TestConceptSyncer_handleKNs(t *testing.T) {
	Convey("Test handleKNs", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		kna := dmock.NewMockKNAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			kna:        kna,
			osa:        osa,
		}

		Convey("Success with no knowledge networks", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{}, nil)

			err := cs.handleKNs()
			So(err, ShouldBeNil)
		})

		Convey("Success with knowledge networks needing update", func() {
			knID := "kn1"
			branch := "main"
			kn := &interfaces.KN{
				KNID:       knID,
				KNName:     "test_kn",
				Branch:     branch,
				UpdateTime: time.Now().UnixMilli(),
			}

			ota := dmock.NewMockObjectTypeAccess(mockCtrl)
			rta := dmock.NewMockRelationTypeAccess(mockCtrl)
			ata := dmock.NewMockActionTypeAccess(mockCtrl)
			cga := dmock.NewMockConceptGroupAccess(mockCtrl)

			cs.ota = ota
			cs.rta = rta
			cs.ata = ata
			cs.cga = cga

			// handleKNs 调用顺序：
			// 1. GetAllKNs
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{knID: kn}, nil)
			// 2. getAllKNsFromOpenSearch (内部调用 SearchData)
			osa.EXPECT().SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			// 3. handleKnowledgeNetwork 会调用多个 getAllXXXFromOpenSearchByKnID
			// 每个都会调用 SearchData
			osa.EXPECT().SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(4)

			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.handleKNs()
			So(err, ShouldBeNil)
		})

		Convey("Failed to get knowledge networks", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(nil, errors.New("db error"))

			err := cs.handleKNs()
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to get knowledge networks from OpenSearch", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			err := cs.handleKNs()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleKnowledgeNetwork(t *testing.T) {
	Convey("Test handleKnowledgeNetwork", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		kna := dmock.NewMockKNAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		rta := dmock.NewMockRelationTypeAccess(mockCtrl)
		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		cga := dmock.NewMockConceptGroupAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			kna:        kna,
			osa:        osa,
			ota:        ota,
			rta:        rta,
			ata:        ata,
			cga:        cga,
		}

		knID := "kn1"
		branch := "main"
		kn := &interfaces.KN{
			KNID:       knID,
			KNName:     "test_kn",
			Branch:     branch,
			UpdateTime: time.Now().UnixMilli(),
		}

		Convey("Success handling knowledge network", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(4)

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldBeNil)
		})

		Convey("No update needed", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(4)

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			err := cs.handleKnowledgeNetwork(ctx, kn, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed to handle object types", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleObjectTypes(t *testing.T) {
	Convey("Test handleObjectTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ota := dmock.NewMockObjectTypeAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ota:        ota,
			osa:        osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling object types", func() {
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(objectTypes, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].ID, ShouldEqual, "ot1")
			So(simpleItems[0].Name, ShouldEqual, "object_type1")
		})

		Convey("Failed to get object types", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleRelationTypes(t *testing.T) {
	Convey("Test handleRelationTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		rta := dmock.NewMockRelationTypeAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			rta:        rta,
			osa:        osa,
		}

		knID := "kn1"
		branch := "main"
		objectTypesMap := map[string]string{
			"ot1": "object_type1",
			"ot2": "object_type2",
		}

		Convey("Success handling relation types", func() {
			relationTypes := map[string]*interfaces.RelationType{
				"rt1": {
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(relationTypes, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleRelationTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].ID, ShouldEqual, "rt1")
			So(simpleItems[0].SourceObjectTypeName, ShouldEqual, "object_type1")
			So(simpleItems[0].TargetObjectTypeName, ShouldEqual, "object_type2")
		})

		Convey("Failed to get relation types", func() {
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleActionTypes(t *testing.T) {
	Convey("Test handleActionTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ata := dmock.NewMockActionTypeAccess(mockCtrl)
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ata:        ata,
			osa:        osa,
		}

		knID := "kn1"
		branch := "main"
		objectTypesMap := map[string]string{
			"ot1": "object_type1",
		}

		Convey("Success handling action types", func() {
			actionTypes := map[string]*interfaces.ActionType{
				"at1": {
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "action_type1",
						ObjectTypeID: "ot1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(actionTypes, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleActionTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].ID, ShouldEqual, "at1")
			So(simpleItems[0].ObjectTypeName, ShouldEqual, "object_type1")
		})

		Convey("Failed to get action types", func() {
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleConceptGroups(t *testing.T) {
	Convey("Test handleConceptGroups", t, func() {
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

		cs := &ConceptSyncer{
			appSetting: appSetting,
			cga:        cga,
			osa:        osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling concept groups", func() {
			conceptGroups := map[string]*interfaces.ConceptGroup{
				"cg1": {
					CGID:       "cg1",
					CGName:     "concept_group1",
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(conceptGroups, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].ID, ShouldEqual, "cg1")
			So(simpleItems[0].Name, ShouldEqual, "concept_group1")
		})

		Convey("Failed to get concept groups", func() {
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForKN(t *testing.T) {
	Convey("Test insertOpenSearchDataForKN", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)
		mfa := dmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		kn := &interfaces.KN{
			KNID:   "kn1",
			KNName: "test_kn",
			Branch: "main",
		}

		Convey("Success inserting KN data", func() {
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed to insert KN data", func() {
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllKNsFromOpenSearch(t *testing.T) {
	Convey("Test getAllKNsFromOpenSearch", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			osa: osa,
		}

		Convey("Success getting KNs from OpenSearch", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"kn_id":   "kn1",
						"kn_name": "test_kn",
						"branch":  "main",
					},
				},
			}

			osa.EXPECT().SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(hits, nil)

			kns, err := cs.getAllKNsFromOpenSearch(ctx)
			So(err, ShouldBeNil)
			So(len(kns), ShouldEqual, 1)
		})

		Convey("Failed to search KNs", func() {
			osa.EXPECT().SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, err := cs.getAllKNsFromOpenSearch(ctx)
			So(err, ShouldNotBeNil)
		})
	})
}
