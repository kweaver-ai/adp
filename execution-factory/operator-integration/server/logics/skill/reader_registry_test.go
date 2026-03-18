package skill

import (
	"context"
	"encoding/json"
	"errors"
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

		Convey("GetSkillGuide returns guide and manifest", func() {
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
				OwnerID:      "bd-1",
				Status:       model.SkillStatusActive,
				Instructions: "demo guide",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference","access_level":"runtime_read","size":5,"mime_type":"text/markdown"}]`,
			}, nil)

			resp, err := reader.GetSkillGuide(context.Background(), &interfaces.GetSkillGuideReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Content, ShouldEqual, "demo guide")
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
				SkillID: "skill-2", OwnerID: "bd-1", Status: model.SkillStatusActive,
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
				SkillID: "skill-3", OwnerID: "bd-1", Status: model.SkillStatusActive,
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
				SkillID: "skill-4", OwnerID: "bd-1", Status: model.SkillStatusDeleting,
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
			registry := &skillRegistry{skillRepo: mockSkillRepo, dbTx: mockDBTx}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-5").Return(&model.SkillRepositoryDB{
				SkillID: "skill-5", OwnerID: "bd-other", Status: model.SkillStatusActive,
			}, nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, errors.New("tx unavailable"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-5",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get tx failed")
		})

		Convey("QuerySkillList omits instructions and files from list payload", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(1), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{
					SkillID:      "skill-6",
					Name:         "demo-skill",
					Description:  "demo-desc",
					Instructions: "full skill markdown",
					FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
					Status:       model.SkillStatusActive,
					OwnerID:      "bd-1",
				},
			}, nil)

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
		})

		Convey("QuerySkillList ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
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
						{SkillID: "skill-6b", Name: "demo-skill", Status: model.SkillStatusActive, OwnerID: "bd-other"},
					}, nil
				},
			)

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
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-7",
				Name:         "demo-skill",
				Description:  "demo-desc",
				Instructions: "full skill markdown",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
				Status:       model.SkillStatusActive,
				OwnerID:      "bd-1",
			}, nil)

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
		})

		Convey("GetSkillDetail ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-7b", OwnerID: "bd-other", Status: model.SkillStatusActive,
			}, nil)

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
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(2), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-10", Name: "visible", Status: model.SkillStatusActive, OwnerID: "bd-1"},
				{SkillID: "skill-11", Name: "hiding", Status: model.SkillStatusDeleting, OwnerID: "bd-1"},
			}, nil)

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
				SkillID: "skill-12", OwnerID: "bd-1", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-12",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillGuide hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13", OwnerID: "bd-1", Status: model.SkillStatusDeleting,
			}, nil)

			resp, err := reader.GetSkillGuide(context.Background(), &interfaces.GetSkillGuideReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-13",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillGuide ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13b", OwnerID: "bd-other", Status: model.SkillStatusActive, Instructions: "demo guide",
			}, nil)

			resp, err := reader.GetSkillGuide(context.Background(), &interfaces.GetSkillGuideReq{
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
				SkillID: "skill-14", OwnerID: "bd-1", Status: model.SkillStatusDeleting,
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
				SkillID: "skill-14b", OwnerID: "bd-other", Status: model.SkillStatusActive,
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

		Convey("DeleteSkill marks deleting before cleanup and hard deletes repository on success", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			registry := &skillRegistry{
				skillRepo:  mockSkillRepo,
				fileRepo:   mockFileRepo,
				assetStore: mockAssetStore,
				dbTx:       mockDBTx,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-8").Return(&model.SkillRepositoryDB{
				SkillID: "skill-8", OwnerID: "bd-1", Status: model.SkillStatusActive,
			}, nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-8", model.SkillStatusDeleting, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-8").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-8", StorageKey: "/tmp/object-1"},
			}, nil)
			mockAssetStore.EXPECT().DeleteFile(gomock.Any(), "/tmp/object-1").Return(nil)
			mockAssetStore.EXPECT().DeleteSkill(gomock.Any(), "skill-8").Return(nil)
			mockFileRepo.EXPECT().DeleteSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-8").Return(nil)
			mockSkillRepo.EXPECT().DeleteSkillByID(gomock.Any(), gomock.Nil(), "skill-8").Return(nil)

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
			registry := &skillRegistry{
				skillRepo:  mockSkillRepo,
				fileRepo:   mockFileRepo,
				assetStore: mockAssetStore,
				dbTx:       mockDBTx,
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-9").Return(&model.SkillRepositoryDB{
				SkillID: "skill-9", OwnerID: "bd-1", Status: model.SkillStatusActive,
			}, nil)
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
	})
}
