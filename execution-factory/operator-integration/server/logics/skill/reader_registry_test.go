package skill

import (
	"context"
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

		Convey("DeleteSkill rejects cross-domain skill", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-5").Return(&model.SkillRepositoryDB{
				SkillID: "skill-5", OwnerID: "bd-other", Status: model.SkillStatusActive,
			}, nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-5",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})
	})
}
