package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalSkillAssetStoreWriteReadDelete(t *testing.T) {
	Convey("LocalSkillAssetStore Write Read DeleteFile", t, func() {
		t.Setenv("TMPDIR", t.TempDir())

		store := newSkillAssetStore()
		ctx := context.Background()
		content := []byte("hello skill asset")

		storageKey, checksum, err := store.Write(ctx, "skill-1", "refs/guide.md", content)
		So(err, ShouldBeNil)
		So(checksum, ShouldEqual, checksumSHA256(content))

		_, err = os.Stat(storageKey)
		So(err, ShouldBeNil)

		got, err := store.Read(ctx, storageKey)
		So(err, ShouldBeNil)
		So(string(got), ShouldEqual, string(content))

		err = store.DeleteFile(ctx, storageKey)
		So(err, ShouldBeNil)

		_, err = os.Stat(storageKey)
		So(os.IsNotExist(err), ShouldBeTrue)
	})
}

func TestLocalSkillAssetStoreDeleteSkill(t *testing.T) {
	Convey("LocalSkillAssetStore DeleteSkill removes root", t, func() {
		t.Setenv("TMPDIR", t.TempDir())

		store := newSkillAssetStore()
		ctx := context.Background()
		_, _, err := store.Write(ctx, "skill-2", "refs/guide.md", []byte("guide"))
		So(err, ShouldBeNil)
		_, _, err = store.Write(ctx, "skill-2", "scripts/run.py", []byte("print('ok')"))
		So(err, ShouldBeNil)

		root := filepath.Join(os.TempDir(), "aoi_skill_assets", "skill-2")
		_, err = os.Stat(root)
		So(err, ShouldBeNil)

		err = store.DeleteSkill(ctx, "skill-2")
		So(err, ShouldBeNil)

		_, err = os.Stat(root)
		So(os.IsNotExist(err), ShouldBeTrue)
	})
}

func TestNewSkillAssetStoreWithConfig(t *testing.T) {
	Convey("newSkillAssetStoreWithConfig prefers object storage when configured", t, func() {
		store, err := newSkillAssetStoreWithConfig(config.S3Config{
			Endpoint:        "http://127.0.0.1:9000",
			AccessID:        "ak",
			AccessSecretKey: "sk",
			Bucket:          "skill-assets",
			Region:          "us-east-1",
			StoragePrefix:   "aoi_skill_assets",
		})
		So(err, ShouldBeNil)
		So(store, ShouldHaveSameTypeAs, &s3SkillAssetStore{})
	})
}

func TestNewSkillAssetStoreUsesFormalConfigProvider(t *testing.T) {
	Convey("newSkillAssetStore uses formal config provider before local fallback", t, func() {
		originProvider := loadFormalSkillAssetStoreConfig
		loadFormalSkillAssetStoreConfig = func() (config.S3Config, bool) {
			return config.S3Config{
				Endpoint:        "http://127.0.0.1:9000",
				AccessID:        "ak",
				AccessSecretKey: "sk",
				Bucket:          "skill-assets",
				Region:          "us-east-1",
				StoragePrefix:   "aoi_skill_assets",
			}, true
		}
		defer func() {
			loadFormalSkillAssetStoreConfig = originProvider
		}()

		store := newSkillAssetStore()
		So(store, ShouldHaveSameTypeAs, &s3SkillAssetStore{})
	})
}
