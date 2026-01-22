package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-manager/common"
	cond "ontology-manager/common/condition"
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

		Convey("Failed to decode KN from hit", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"name": make(chan int), // 无法解码的类型
					},
				},
			}

			osa.EXPECT().SearchData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(hits, nil)

			_, err := cs.getAllKNsFromOpenSearch(ctx)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForKN_WithVector(t *testing.T) {
	Convey("Test insertOpenSearchDataForKN with vector enabled\n", t, func() {
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

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		kn := &interfaces.KN{
			KNID:    "kn1",
			KNName:  "test_kn",
			Branch:  "main",
			Tags:    []string{"tag1"},
			Comment: "comment",
			Detail:  "detail",
		}
		vectors := []*cond.VectorResp{
			{
				Vector: []float32{0.1, 0.2, 0.3},
			},
		}

		Convey("Success inserting KN data with vector\n", func() {
			mfa.EXPECT().GetDefaultModel(ctx).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error")).AnyTimes()

			err := cs.insertOpenSearchDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForObjectTypes(t *testing.T) {
	Convey("Test insertOpenSearchDataForObjectTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success with empty list\n", func() {
			objectTypes := []*interfaces.ObjectType{}

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
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

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
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

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForObjectTypes_WithVector(t *testing.T) {
	Convey("Test insertOpenSearchDataForObjectTypes with vector enabled\n", t, func() {
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

		cs := &ConceptSyncer{
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
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
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

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
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
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
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

			err := cs.insertOpenSearchDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForRelationTypes(t *testing.T) {
	Convey("Test insertOpenSearchDataForRelationTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success with empty list\n", func() {
			relationTypes := []*interfaces.RelationType{}

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting relation types\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForRelationTypes_WithVector(t *testing.T) {
	Convey("Test insertOpenSearchDataForRelationTypes with vector enabled\n", t, func() {
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

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		Convey("Success inserting relation types with vector\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
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
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()

			err := cs.insertOpenSearchDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForActionTypes(t *testing.T) {
	Convey("Test insertOpenSearchDataForActionTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success with empty list\n", func() {
			actionTypes := []*interfaces.ActionType{}

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting action types\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForActionTypes_WithVector(t *testing.T) {
	Convey("Test insertOpenSearchDataForActionTypes with vector enabled\n", t, func() {
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

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		Convey("Success inserting action types with vector\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
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
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := cs.insertOpenSearchDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForConceptGroups(t *testing.T) {
	Convey("Test insertOpenSearchDataForConceptGroups\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
		}

		Convey("Success with empty list\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{}

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting concept groups\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertOpenSearchDataForConceptGroups_WithVector(t *testing.T) {
	Convey("Test insertOpenSearchDataForConceptGroups with vector enabled\n", t, func() {
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

		cs := &ConceptSyncer{
			appSetting: appSetting,
			osa:        osa,
			mfa:        mfa,
		}

		Convey("Success inserting concept groups with vector\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
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
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(nil)

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := cs.insertOpenSearchDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllObjectTypesFromOpenSearchByKnID(t *testing.T) {
	Convey("Test getAllObjectTypesFromOpenSearchByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			osa: osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting object types from OpenSearch\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"ot_id":   "ot1",
						"ot_name": "object_type1",
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			objectTypes, err := cs.getAllObjectTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(objectTypes), ShouldEqual, 1)
		})

		Convey("Failed to search object types\n", func() {
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, err := cs.getAllObjectTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode object type from hit\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"name": make(chan int), // 无法解码的类型
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			_, err := cs.getAllObjectTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllRelationTypesFromOpenSearchByKnID(t *testing.T) {
	Convey("Test getAllRelationTypesFromOpenSearchByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			osa: osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting relation types from OpenSearch\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"rt_id":   "rt1",
						"rt_name": "relation_type1",
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			relationTypes, err := cs.getAllRelationTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(relationTypes), ShouldEqual, 1)
		})

		Convey("Failed to search relation types\n", func() {
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, err := cs.getAllRelationTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode relation type from hit\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"name": make(chan int), // 无法解码的类型
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			_, err := cs.getAllRelationTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllActionTypesFromOpenSearchByKnID(t *testing.T) {
	Convey("Test getAllActionTypesFromOpenSearchByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			osa: osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting action types from OpenSearch\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"at_id":   "at1",
						"at_name": "action_type1",
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			actionTypes, err := cs.getAllActionTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 1)
		})

		Convey("Failed to search action types\n", func() {
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, err := cs.getAllActionTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode action type from hit\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"name": make(chan int), // 无法解码的类型
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			_, err := cs.getAllActionTypesFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllConceptGroupsFromOpenSearchByKnID(t *testing.T) {
	Convey("Test getAllConceptGroupsFromOpenSearchByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := dmock.NewMockOpenSearchAccess(mockCtrl)

		cs := &ConceptSyncer{
			osa: osa,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting concept groups from OpenSearch\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"cg_id":   "cg1",
						"cg_name": "concept_group1",
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			conceptGroups, err := cs.getAllConceptGroupsFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(conceptGroups), ShouldEqual, 1)
		})

		Convey("Failed to search concept groups\n", func() {
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, err := cs.getAllConceptGroupsFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode concept group from hit\n", func() {
			hits := []interfaces.Hit{
				{
					Source: map[string]any{
						"name": make(chan int), // 无法解码的类型
					},
				},
			}

			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(hits, nil)

			_, err := cs.getAllConceptGroupsFromOpenSearchByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleKnowledgeNetwork_Errors(t *testing.T) {
	Convey("Test handleKnowledgeNetwork error cases\n", t, func() {
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

		Convey("Failed to handle relation types\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to handle action types\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(2)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to handle concept groups\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(3)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to update KN detail\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(4)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert open search data for KN\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil).Times(4)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleObjectTypes_Errors(t *testing.T) {
	Convey("Test handleObjectTypes error cases\n", t, func() {
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

		Convey("Failed to get object types from OpenSearch\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert open search data\n", func() {
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
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleRelationTypes_Errors(t *testing.T) {
	Convey("Test handleRelationTypes error cases\n", t, func() {
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
		}

		Convey("Failed to get relation types from OpenSearch\n", func() {
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert open search data\n", func() {
			relationTypes := map[string]*interfaces.RelationType{
				"rt1": {
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(relationTypes, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleActionTypes_Errors(t *testing.T) {
	Convey("Test handleActionTypes error cases\n", t, func() {
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

		Convey("Failed to get action types from OpenSearch\n", func() {
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert open search data\n", func() {
			actionTypes := map[string]*interfaces.ActionType{
				"at1": {
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(actionTypes, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch, objectTypesMap)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleConceptGroups_Errors(t *testing.T) {
	Convey("Test handleConceptGroups error cases\n", t, func() {
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

		Convey("Failed to get concept groups from OpenSearch\n", func() {
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return(nil, errors.New("opensearch error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert open search data\n", func() {
			conceptGroups := map[string]*interfaces.ConceptGroup{
				"cg1": {
					CGID:       "cg1",
					CGName:     "concept_group1",
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(conceptGroups, nil)
			osa.EXPECT().SearchData(gomock.Any(), interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any()).Return([]interfaces.Hit{}, nil)
			osa.EXPECT().InsertData(ctx, interfaces.KN_CONCEPT_INDEX_NAME, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}
