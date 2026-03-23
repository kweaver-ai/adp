package skill

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	Convey("newSkillAssetStoreWithConfig prefers oss gateway when configured", t, func() {
		store, err := newSkillAssetStoreWithConfig(config.OSSGatewayConfig{
			BaseURL:   "http://127.0.0.1:9000",
			StorageID: "storage-1",
		})
		So(err, ShouldBeNil)
		So(store, ShouldHaveSameTypeAs, &ossGatewaySkillAssetStore{})
	})
}

func TestNewSkillAssetStoreUsesFormalConfigProvider(t *testing.T) {
	Convey("newSkillAssetStore uses formal config provider before local fallback", t, func() {
		originProvider := loadFormalSkillAssetStoreConfig
		loadFormalSkillAssetStoreConfig = func() (config.OSSGatewayConfig, bool) {
			return config.OSSGatewayConfig{
				BaseURL:   "http://127.0.0.1:9000",
				StorageID: "storage-1",
			}, true
		}
		defer func() {
			loadFormalSkillAssetStoreConfig = originProvider
		}()

		store := newSkillAssetStore()
		So(store, ShouldHaveSameTypeAs, &ossGatewaySkillAssetStore{})
	})
}

func TestOSSGatewaySkillAssetStoreWriteReadDeleteFile(t *testing.T) {
	Convey("ossGatewaySkillAssetStore writes, reads and deletes via gateway signed URLs", t, func() {
		var stored []byte
		var deletedPath string
		objectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPut:
				body, err := io.ReadAll(r.Body)
				So(err, ShouldBeNil)
				stored = body
				w.WriteHeader(http.StatusOK)
			case http.MethodGet:
				_, err := w.Write(stored)
				So(err, ShouldBeNil)
			case http.MethodDelete:
				deletedPath = r.URL.Path
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))
		defer objectServer.Close()

		gatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.URL.Path, ShouldContainSubstring, "/api/v1/")
			var method string
			switch {
			case strings.Contains(r.URL.Path, "/upload/"):
				method = http.MethodPut
			case strings.Contains(r.URL.Path, "/download/"):
				method = http.MethodGet
			case strings.Contains(r.URL.Path, "/delete/"):
				method = http.MethodDelete
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, err := fmt.Fprintf(w, `{"code":0,"message":"success","data":{"method":"%s","url":"%s/object","headers":{}}}`, method, objectServer.URL)
			So(err, ShouldBeNil)
		}))
		defer gatewayServer.Close()

		store, err := newSkillAssetStoreWithConfig(config.OSSGatewayConfig{
			BaseURL:   gatewayServer.URL,
			StorageID: "storage-1",
		})
		So(err, ShouldBeNil)
		So(store, ShouldNotBeNil)

		ctx := context.Background()
		content := []byte("gateway-asset")
		storageKey, checksum, err := store.Write(ctx, "skill-3", "refs/guide.md", content)
		So(err, ShouldBeNil)
		So(checksum, ShouldEqual, checksumSHA256(content))
		So(storageKey, ShouldEqual, "aoi_skill_assets/skill-3/refs/guide.md")
		So(string(stored), ShouldEqual, string(content))

		readContent, err := store.Read(ctx, storageKey)
		So(err, ShouldBeNil)
		So(string(readContent), ShouldEqual, string(content))

		err = store.DeleteFile(ctx, storageKey)
		So(err, ShouldBeNil)
		So(deletedPath, ShouldEqual, "/object")
	})
}
