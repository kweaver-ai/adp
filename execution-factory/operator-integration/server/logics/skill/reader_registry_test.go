package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillReaderAndRegistry(t *testing.T) {
	Convey("SkillReader and SkillRegistry", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("GetSkillContent returns skill content and manifest", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillReader{
				skillRepo:  mockSkillRepo,
				fileRepo:   mockFileRepo,
				assetStore: mockAssetStore,
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-1").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-1",
				Status:       model.SkillStatusActive,
				SkillContent: "demo guide",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference","access_level":"runtime_read","size":5,"mime_type":"text/markdown"}]`,
			}, nil)

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillContent, ShouldEqual, "demo guide")
			So(len(resp.Files), ShouldEqual, 1)
			So(resp.Files[0].RelPath, ShouldEqual, "refs/guide.md")
		})

		Convey("ReadSkillFile rejects restricted files", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillReader{
				skillRepo:  mockSkillRepo,
				fileRepo:   mockFileRepo,
				assetStore: mockAssetStore,
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-2").Return(&model.SkillRepositoryDB{
				SkillID: "skill-2", Status: model.SkillStatusActive,
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-2", "refs/secret.md").Return(&model.SkillFileIndexDB{
				SkillID: "skill-2", RelPath: "refs/secret.md", AccessLevel: string(interfaces.SkillFileAccessLevelRestricted),
			}, nil)

			resp, err := reader.ReadSkillFile(context.Background(), &interfaces.ReadSkillFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-2",
				RelPath:          "refs/secret.md",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "access denied")
		})

		Convey("ReadSkillFile validates checksum", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillReader{
				skillRepo:  mockSkillRepo,
				fileRepo:   mockFileRepo,
				assetStore: mockAssetStore,
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-3").Return(&model.SkillRepositoryDB{
				SkillID: "skill-3", Status: model.SkillStatusActive,
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-3", "refs/guide.md").Return(&model.SkillFileIndexDB{
				SkillID:       "skill-3",
				RelPath:       "refs/guide.md",
				StorageKey:    "/tmp/f1",
				AccessLevel:   string(interfaces.SkillFileAccessLevelRuntimeRead),
				ContentSHA256: checksumSHA256([]byte("original")),
			}, nil)
			mockAssetStore.EXPECT().Read(gomock.Any(), "/tmp/f1").Return([]byte("tampered"), nil)

			resp, err := reader.ReadSkillFile(context.Background(), &interfaces.ReadSkillFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-3",
				RelPath:          "refs/guide.md",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "checksum mismatch")
		})

		Convey("DeleteSkill rejects invalid status", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-4").Return(&model.SkillRepositoryDB{
				SkillID: "skill-4", Status: model.SkillStatusDeleting,
			}, nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-4",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "can not be deleted")
		})

		Convey("DeleteSkill ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-5").Return(&model.SkillRepositoryDB{
				SkillID: "skill-5", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-5", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, errors.New("tx unavailable"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-5",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get tx failed")
		})

		Convey("RegisterSkill checks create permission before registration", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				parser:                newSkillParser(),
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill).Return(errors.New("create forbidden"))

			resp, err := registry.RegisterSkill(context.Background(), &interfaces.RegisterSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				FileType:         "content",
				File:             json.RawMessage(validSkillMarkdown()),
				Source:           "unit-test",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create forbidden")
		})

		Convey("RegisterSkill associates business domain after registration succeeds", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				parser:                newSkillParser(),
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil)
			mockSkillRepo.EXPECT().InsertSkill(gomock.Any(), gomock.Nil(), gomock.Any()).Return("skill-registered", nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-registered", model.SkillStatusActive, "user-1").Return(nil)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), "bd-1", "skill-registered", interfaces.AuthResourceTypeSkill).Return(nil)

			resp, err := registry.RegisterSkill(context.Background(), &interfaces.RegisterSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				FileType:         "content",
				File:             json.RawMessage(validSkillMarkdown()),
				Source:           "unit-test",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-registered")
		})

		Convey("QuerySkillList omits instructions and files from list payload", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(1), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{
					SkillID:      "skill-6",
					Name:         "demo-skill",
					Description:  "demo-desc",
					SkillContent: "full skill markdown",
					FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
					Status:       model.SkillStatusActive,
				},
			}, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), []string{"skill-6"}, interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypeView).Return([]string{"skill-6"}, nil)

			resp, err := registry.QuerySkillList(context.Background(), &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			raw, marshalErr := json.Marshal(resp.Data[0])
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(string(raw), ShouldNotContainSubstring, "owner_id")
			So(string(raw), ShouldNotContainSubstring, "owner_type")
		})

		Convey("QuerySkillList ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}) (int64, error) {
					_, exists := filter["owner_id"]
					So(exists, ShouldBeFalse)
					return int64(1), nil
				},
			)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}, _ interface{}, _ interface{}) ([]*model.SkillRepositoryDB, error) {
					_, exists := filter["owner_id"]
					So(exists, ShouldBeFalse)
					return []*model.SkillRepositoryDB{
						{SkillID: "skill-6b", Name: "demo-skill", Status: model.SkillStatusActive},
					}, nil
				},
			)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), []string{"skill-6b"}, interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypeView).Return([]string{"skill-6b"}, nil)

			resp, err := registry.QuerySkillList(context.Background(), &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-6b")
		})

		Convey("GetSkillDetail omits instructions and files from detail payload", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-7",
				Name:         "demo-skill",
				Description:  "demo-desc",
				SkillContent: "full skill markdown",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
				Status:       model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-7", interfaces.AuthResourceTypeSkill).Return(nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-7",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			raw, marshalErr := json.Marshal(resp)
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(string(raw), ShouldNotContainSubstring, "owner_id")
			So(string(raw), ShouldNotContainSubstring, "owner_type")
		})

		Convey("GetSkillDetail ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-7b", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-7b", interfaces.AuthResourceTypeSkill).Return(nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-7b",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-7b")
		})

		Convey("QuerySkillList hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(2), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-10", Name: "visible", Status: model.SkillStatusActive},
				{SkillID: "skill-11", Name: "hiding", Status: model.SkillStatusDeleting},
			}, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), []string{"skill-10"}, interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypeView).Return([]string{"skill-10"}, nil)

			resp, err := registry.QuerySkillList(context.Background(), &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-10")
		})

		Convey("GetSkillDetail hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-12").Return(&model.SkillRepositoryDB{
				SkillID: "skill-12", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-12",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillDetail checks view permission before returning detail", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-12b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-12b", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-12b", interfaces.AuthResourceTypeSkill).Return(errors.New("view forbidden"))

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-12b",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "view forbidden")
		})

		Convey("QuerySkillList filters non-viewable skills by auth service", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(2), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-12c", Name: "visible", Status: model.SkillStatusActive},
				{SkillID: "skill-12d", Name: "hidden", Status: model.SkillStatusActive},
			}, nil)
			mockAuthService.EXPECT().ResourceFilterIDs(gomock.Any(), gomock.Any(), []string{"skill-12c", "skill-12d"}, interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypeView).Return([]string{"skill-12c"}, nil)

			resp, err := registry.QuerySkillList(context.Background(), &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-12c")
		})

		Convey("GetSkillContent hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-13",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillContent ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13b", Status: model.SkillStatusActive, SkillContent: "demo guide",
			}, nil)

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-13b",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-13b")
		})

		Convey("ReadSkillFile hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-14").Return(&model.SkillRepositoryDB{
				SkillID: "skill-14", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := reader.ReadSkillFile(context.Background(), &interfaces.ReadSkillFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-14",
				RelPath:          "refs/guide.md",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("ReadSkillFile ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo, fileRepo: mockFileRepo, assetStore: mockAssetStore}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-14b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-14b", Status: model.SkillStatusActive,
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-14b", "refs/guide.md").Return(&model.SkillFileIndexDB{
				SkillID: "skill-14b", RelPath: "refs/guide.md", StorageKey: "/tmp/f14b", AccessLevel: string(interfaces.SkillFileAccessLevelRuntimeRead), ContentSHA256: checksumSHA256([]byte("ok")),
			}, nil)
			mockAssetStore.EXPECT().Read(gomock.Any(), "/tmp/f14b").Return([]byte("ok"), nil)

			resp, err := reader.ReadSkillFile(context.Background(), &interfaces.ReadSkillFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-14b",
				RelPath:          "refs/guide.md",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-14b")
		})

		Convey("QuerySkillMarketList filters by public access and business domain visibility", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypePublicAccess).Return([]string{"skill-m1", "skill-m2", "skill-m3"}, nil)
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{
				"skill-m1": "bd-1",
				"skill-m3": "bd-1",
			}, nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-m1", Name: "visible", Status: model.SkillStatusActive},
				{SkillID: "skill-m3", Name: "deleting", Status: model.SkillStatusDeleting},
				{SkillID: "skill-m4", Name: "not-public", Status: model.SkillStatusActive},
			}, nil)

			resp, err := registry.QuerySkillMarketList(context.Background(), &interfaces.QuerySkillMarketListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.TotalCount, ShouldEqual, 1)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-m1")
		})

		Convey("GetSkillMarketDetail checks public access and business domain visibility", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-m-detail").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-m-detail",
				Name:         "market-visible",
				Description:  "demo-desc",
				SkillContent: "full skill markdown",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
				Status:       model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-detail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{
				"skill-m-detail": "bd-1",
			}, nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-detail",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			raw, marshalErr := json.Marshal(resp)
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(resp.SkillID, ShouldEqual, "skill-m-detail")
		})

		Convey("GetSkillMarketDetail hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-m-deleting").Return(&model.SkillRepositoryDB{
				SkillID: "skill-m-deleting", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-deleting",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("DeleteSkill marks deleting before cleanup and hard deletes repository on success", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-8").Return(&model.SkillRepositoryDB{
				SkillID: "skill-8", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-8", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-8", model.SkillStatusDeleting, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-8").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-8", StorageKey: "/tmp/object-1"},
			}, nil)
			mockAssetStore.EXPECT().DeleteFile(gomock.Any(), "/tmp/object-1").Return(nil)
			mockAssetStore.EXPECT().DeleteSkill(gomock.Any(), "skill-8").Return(nil)
			mockFileRepo.EXPECT().DeleteSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-8").Return(nil)
			mockSkillRepo.EXPECT().DeleteSkillByID(gomock.Any(), gomock.Nil(), "skill-8").Return(nil)
			mockBusinessDomainService.EXPECT().BatchDisassociateResource(gomock.Any(), "bd-1", []string{"skill-8"}, interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().DeletePolicy(gomock.Any(), []string{"skill-8"}, interfaces.AuthResourceTypeSkill).Return(nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-8",
			})

			So(err, ShouldBeNil)
		})

		Convey("DeleteSkill keeps deleting status when asset cleanup fails", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-9").Return(&model.SkillRepositoryDB{
				SkillID: "skill-9", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-9", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-9", model.SkillStatusDeleting, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-9").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-9", StorageKey: "/tmp/object-2"},
			}, nil)
			mockAssetStore.EXPECT().DeleteFile(gomock.Any(), "/tmp/object-2").Return(errors.New("delete failed"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-9",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete failed")
		})

		Convey("DeleteSkill checks delete permission before cleanup", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-9b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-9b", Status: model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-9b", interfaces.AuthResourceTypeSkill).Return(errors.New("delete forbidden"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-9b",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete forbidden")
		})

		Convey("DownloadSkill validates visibility and builds zip with skill content and files", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-zip-1").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-zip-1",
				Name:         "demo-skill",
				SkillContent: "Use this skill carefully.",
				Status:       model.SkillStatusActive,
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-zip-1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-zip-1").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-zip-1", RelPath: "refs/guide.md", StorageKey: "obj-1"},
			}, nil)
			mockAssetStore.EXPECT().Read(gomock.Any(), "obj-1").Return([]byte("guide body"), nil)

			resp, err := registry.DownloadSkill(context.Background(), &interfaces.DownloadSkillReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-zip-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.FileName, ShouldEqual, "demo-skill.zip")

			zipReader, zipErr := zip.NewReader(bytes.NewReader(resp.Content), int64(len(resp.Content)))
			So(zipErr, ShouldBeNil)
			entries := map[string]string{}
			for _, file := range zipReader.File {
				rc, openErr := file.Open()
				So(openErr, ShouldBeNil)
				body, readErr := io.ReadAll(rc)
				So(readErr, ShouldBeNil)
				_ = rc.Close()
				entries[file.Name] = string(body)
			}
			So(entries["SKILL.md"], ShouldContainSubstring, "Use this skill carefully.")
			So(entries["refs/guide.md"], ShouldEqual, "guide body")
		})
	})
}
