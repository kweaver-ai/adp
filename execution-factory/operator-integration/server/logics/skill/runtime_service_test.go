package skill

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestBuildRuntimeOutputMappings(t *testing.T) {
	Convey("buildRuntimeOutputMappings derives sandbox output paths from output schema", t, func() {
		taskWorkspace := &interfaces.PrepareTaskWorkspaceResp{
			TaskID:   "task-1",
			TaskRoot: ".tasks/skill/task-1",
			Directories: map[string]string{
				"output": ".tasks/skill/task-1/output",
			},
		}

		outputSchema := utils.ObjectToJSON(map[string]any{
			"output_path": map[string]any{
				"type": "file",
			},
			"images": map[string]any{
				"type": "directory",
				"path": "images",
			},
		})

		mappings := buildRuntimeOutputMappings(taskWorkspace, outputSchema)
		So(mappings["output_path"], ShouldEqual, "/workspace/.tasks/skill/task-1/output/output_path")
		So(mappings["images"], ShouldEqual, "/workspace/.tasks/skill/task-1/output/images")
	})
}

func TestBuildRuntimeOutputRefs(t *testing.T) {
	Convey("buildRuntimeOutputRefs exposes structured sandbox refs for declared outputs", t, func() {
		outputSchema := utils.ObjectToJSON(map[string]any{
			"output_path": map[string]any{
				"type":        "file",
				"description": "Generated pdf",
			},
			"images": map[string]any{
				"type": "directory",
				"path": "images",
			},
		})
		outputMappings := map[string]string{
			"output_path": "/workspace/.tasks/skill/task-1/output/output_path",
			"images":      "/workspace/.tasks/skill/task-1/output/images",
		}

		refs := buildRuntimeOutputRefs("session-1", outputSchema, outputMappings)
		outputRef := refs["output_path"].(map[string]any)
		imageRef := refs["images"].(map[string]any)

		So(outputRef["session_id"], ShouldEqual, "session-1")
		So(outputRef["type"], ShouldEqual, "file")
		So(outputRef["container_path"], ShouldEqual, ".tasks/skill/task-1/output/output_path")
		So(outputRef["description"], ShouldEqual, "Generated pdf")
		So(imageRef["type"], ShouldEqual, "directory")
		So(imageRef["declared_path"], ShouldEqual, "images")
		So(imageRef["container_path"], ShouldEqual, ".tasks/skill/task-1/output/images")
	})
}

func TestBuildRuntimeOutputArtifactKey(t *testing.T) {
	Convey("buildRuntimeOutputArtifactKey derives stable oss keys for output artifacts", t, func() {
		key := buildRuntimeOutputArtifactKey(&model.SkillRepositoryDB{
			SkillID: "docx",
			Version: "v1",
		}, "task-1", "output_path", ".tasks/skill/task-1/output/result.pdf")

		So(key, ShouldEqual, filepath.ToSlash("execution-factory/skill-execution/docx/v1/task-1/output_path/result.pdf"))
	})
}

func TestWorkspaceContainerPath(t *testing.T) {
	Convey("workspaceContainerPath normalizes workspace absolute path to container path", t, func() {
		So(workspaceContainerPath("/workspace/.tasks/skill/task-1/output/result.pdf"), ShouldEqual, ".tasks/skill/task-1/output/result.pdf")
		So(workspaceContainerPath(".tasks/skill/task-1/output/result.pdf"), ShouldEqual, ".tasks/skill/task-1/output/result.pdf")
	})
}

func TestNormalizeSessionFiles(t *testing.T) {
	Convey("normalizeSessionFiles rewrites absolute workspace container paths to relative paths", t, func() {
		files := normalizeSessionFiles([]*interfaces.SessionFileInfo{
			{
				Name:          "result.pdf",
				ContainerPath: "/workspace/.tasks/skill/task-1/output/result.pdf",
				Size:          128,
			},
		})

		So(files, ShouldHaveLength, 1)
		So(files[0].ContainerPath, ShouldEqual, ".tasks/skill/task-1/output/result.pdf")
		So(files[0].Name, ShouldEqual, "result.pdf")
	})
}

func TestAppendRuntimeWarning(t *testing.T) {
	Convey("appendRuntimeWarning accumulates structured warnings in return value", t, func() {
		returnValue := map[string]any{}

		appendRuntimeWarning(returnValue, "output_file_list_failed", errors.New("list failed"))
		appendRuntimeWarning(returnValue, "output_artifact_persist_failed", errors.New("upload failed"))

		warnings, ok := returnValue["warnings"].([]map[string]any)
		So(ok, ShouldBeTrue)
		So(warnings, ShouldHaveLength, 2)
		So(warnings[0]["code"], ShouldEqual, "output_file_list_failed")
		So(warnings[0]["message"], ShouldEqual, "list failed")
		So(warnings[1]["code"], ShouldEqual, "output_artifact_persist_failed")
		So(warnings[1]["message"], ShouldEqual, "upload failed")
	})
}

func TestValidateExecuteSkillInputs(t *testing.T) {
	Convey("validateExecuteSkillInputs validates required and typed inputs", t, func() {
		schema := map[string]any{
			"input_file": map[string]any{
				"type":     "file",
				"required": true,
			},
			"notes": map[string]any{
				"type": "text",
			},
			"output_dir": map[string]any{
				"type": "directory",
			},
		}

		Convey("missing required input returns error", func() {
			err := validateExecuteSkillInputs(schema, map[string]any{
				"notes": "demo",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "missing required input")
		})

		Convey("invalid text type returns error", func() {
			err := validateExecuteSkillInputs(schema, map[string]any{
				"input_file": map[string]any{
					"type":         "inline_file",
					"filename":     "a.txt",
					"content":      "hello",
					"content_type": "text/plain",
				},
				"notes": map[string]any{
					"type": "inline_file",
				},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `input "notes" expects text`)
		})

		Convey("valid file, text and directory inputs pass", func() {
			err := validateExecuteSkillInputs(schema, map[string]any{
				"input_file": map[string]any{
					"type":        "artifact_ref",
					"storage_id":  "s1",
					"storage_key": "docs/a.docx",
				},
				"notes":      "convert this file",
				"output_dir": "/workspace/.tasks/skill/task-1/output/images",
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestApplyRuntimeInputDefaults(t *testing.T) {
	Convey("applyRuntimeInputDefaults injects schema defaults", t, func() {
		schema := map[string]any{
			"mode": map[string]any{
				"type":    "text",
				"default": "fast",
			},
			"format": map[string]any{
				"type":    "text",
				"default": "pdf",
			},
		}

		merged := applyRuntimeInputDefaults(schema, map[string]any{
			"mode": "safe",
		})

		So(merged["mode"], ShouldEqual, "safe")
		So(merged["format"], ShouldEqual, "pdf")
	})
}

func TestValidateExecuteSkillInputsEnum(t *testing.T) {
	Convey("validateExecuteSkillInputs enforces enum constraints for string inputs", t, func() {
		schema := map[string]any{
			"mode": map[string]any{
				"type": "text",
				"enum": []any{"fast", "safe"},
			},
		}

		err := validateExecuteSkillInputs(schema, map[string]any{
			"mode": "turbo",
		})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, `input "mode" must be one of`)

		err = validateExecuteSkillInputs(schema, map[string]any{
			"mode": "fast",
		})
		So(err, ShouldBeNil)
	})
}

func TestValidateRuntimeProfileDefinition(t *testing.T) {
	Convey("validateRuntimeProfileDefinition rejects invalid placeholder references", t, func() {
		err := validateRuntimeProfileDefinition(
			"to_pdf",
			[]string{"python3", "to_pdf.py", "--file", "{{missing}}"},
			map[string]any{
				"input_file": map[string]any{"type": "file"},
			},
			nil,
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unsupported placeholder")
	})

	Convey("validateRuntimeProfileDefinition rejects sandbox absolute command path", t, func() {
		err := validateRuntimeProfileDefinition(
			"to_pdf",
			[]string{"python3", "/workspace/.runtime_packages/pkg/scripts/to_pdf.py"},
			nil,
			nil,
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "sandbox absolute path is not allowed")
	})

	Convey("validateRuntimeProfileDefinition rejects disallowed runtime executable", t, func() {
		err := validateRuntimeProfileDefinition(
			"to_pdf",
			[]string{"perl", "scripts/to_pdf.py"},
			nil,
			nil,
		)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "runtime executable is not allowed")
	})
}

func TestPersistRuntimeOutputArtifactsMatchesNormalizedContainerPath(t *testing.T) {
	Convey("persistRuntimeOutputArtifacts matches normalized output refs against listed session files", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockControlPlane := mocks.NewMockSandBoxControlPlane(ctrl)
		mockOSS := mocks.NewMockOSSGatewayBackendClient(ctrl)
		service := &skillRuntimeService{
			controlPlane: mockControlPlane,
			ossClient:    mockOSS,
		}

		mockOSS.EXPECT().IsReady().Return(true)
		mockOSS.EXPECT().CurrentStorageID(gomock.Any()).Return("storage-1", nil)
		mockControlPlane.EXPECT().DownloadSessionFile(gomock.Any(), "session-1", ".tasks/skill/task-1/output/result.pdf").Return(&interfaces.SessionFileDownloadResp{
			SessionID: "session-1",
			FilePath:  ".tasks/skill/task-1/output/result.pdf",
			Content:   []byte("pdf-bytes"),
			Size:      9,
		}, nil)
		mockOSS.EXPECT().UploadFile(gomock.Any(), gomock.Any(), []byte("pdf-bytes")).Return(nil)

		artifacts, err := service.persistRuntimeOutputArtifacts(
			context.Background(),
			"session-1",
			&model.SkillRepositoryDB{
				SkillID: "docx",
				Version: "v1",
			},
			&interfaces.PrepareTaskWorkspaceResp{
				TaskID: "task-1",
			},
			map[string]any{
				"output_path": map[string]any{
					"name":           "output_path",
					"type":           "file",
					"container_path": ".tasks/skill/task-1/output/result.pdf",
				},
			},
			normalizeSessionFiles([]*interfaces.SessionFileInfo{
				{
					Name:          "result.pdf",
					ContainerPath: "/workspace/.tasks/skill/task-1/output/result.pdf",
					Size:          9,
				},
			}),
		)

		So(err, ShouldBeNil)
		So(artifacts, ShouldContainKey, "output_path")
		artifact := artifacts["output_path"].(map[string]any)
		So(artifact["storage_id"], ShouldEqual, "storage-1")
		So(artifact["storage_key"], ShouldEqual, filepath.ToSlash("execution-factory/skill-execution/docx/v1/task-1/output_path/result.pdf"))
		source := artifact["source"].(map[string]any)
		So(source["container_path"], ShouldEqual, ".tasks/skill/task-1/output/result.pdf")
	})
}

func TestPersistRuntimeOutputArtifactsRejectsPresignedURL(t *testing.T) {
	Convey("persistRuntimeOutputArtifacts returns error when sandbox download responds with presigned url", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockControlPlane := mocks.NewMockSandBoxControlPlane(ctrl)
		mockOSS := mocks.NewMockOSSGatewayBackendClient(ctrl)
		service := &skillRuntimeService{
			controlPlane: mockControlPlane,
			ossClient:    mockOSS,
		}

		mockOSS.EXPECT().IsReady().Return(true)
		mockOSS.EXPECT().CurrentStorageID(gomock.Any()).Return("storage-1", nil)
		mockControlPlane.EXPECT().DownloadSessionFile(gomock.Any(), "session-1", ".tasks/skill/task-1/output/result.pdf").Return(&interfaces.SessionFileDownloadResp{
			SessionID:    "session-1",
			FilePath:     ".tasks/skill/task-1/output/result.pdf",
			PresignedURL: "https://example.com/download/result.pdf",
			Size:         12,
		}, nil)

		artifacts, err := service.persistRuntimeOutputArtifacts(
			context.Background(),
			"session-1",
			&model.SkillRepositoryDB{
				SkillID: "docx",
				Version: "v1",
			},
			&interfaces.PrepareTaskWorkspaceResp{
				TaskID: "task-1",
			},
			map[string]any{
				"output_path": map[string]any{
					"name":           "output_path",
					"type":           "file",
					"container_path": ".tasks/skill/task-1/output/result.pdf",
				},
			},
			normalizeSessionFiles([]*interfaces.SessionFileInfo{
				{
					Name:          ".tasks/skill/task-1/output/result.pdf",
					ContainerPath: "/workspace/.tasks/skill/task-1/output/result.pdf",
					Size:          12,
				},
			}),
		)

		So(artifacts, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "presigned_url")
	})
}

func TestExecuteSkillStatusBehavior(t *testing.T) {
	Convey("ExecuteSkill rejects offline runtime profiles before attempting package build", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockProfileRepo := &skillRuntimeProfileRepoStub{profiles: []*model.SkillRuntimeProfileDB{
			{
				SkillID:         "skill-offline-1",
				SkillVersion:    "v1",
				Entrypoint:      "to_pdf",
				RuntimeType:     "python",
				CommandTemplate: `["python3","scripts/to_pdf.py"]`,
				Status:          interfaces.BizStatusOffline.String(),
			},
		}}
		service := &skillRuntimeService{
			registry: &skillRegistry{
				skillRepo: mockSkillRepo,
			},
			profileRepo: mockProfileRepo,
			AuthService: mockAuthService,
		}

		mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-offline-1").Return(&model.SkillRepositoryDB{
			SkillID: "skill-offline-1",
			Version: "v1",
			Status:  interfaces.BizStatusPublished.String(),
		}, nil)
		mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
		mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), "skill-offline-1", interfaces.AuthResourceTypeSkill).Return(nil)

		resp, err := service.ExecuteSkill(context.Background(), &interfaces.ExecuteSkillReq{
			BusinessDomainID: "bd-1",
			UserID:           "user-1",
			SkillID:          "skill-offline-1",
			Entrypoint:       "to_pdf",
		})

		So(resp, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "runtime profile is offline")
	})

	Convey("ExecuteSkill allows unpublish runtime profiles to proceed past status gate", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		service := &skillRuntimeService{
			registry: &skillRegistry{
				skillRepo: mockSkillRepo,
				fileRepo:  mockFileRepo,
			},
			profileRepo: &skillRuntimeProfileRepoStub{profiles: []*model.SkillRuntimeProfileDB{
				{
					SkillID:         "skill-unpublish-1",
					SkillVersion:    "v1",
					Entrypoint:      "to_pdf",
					RuntimeType:     "python",
					CommandTemplate: `["python3","scripts/to_pdf.py"]`,
					Status:          interfaces.BizStatusUnpublish.String(),
				},
			}},
			AuthService: mockAuthService,
		}

		mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-unpublish-1").Return(&model.SkillRepositoryDB{
			SkillID:      "skill-unpublish-1",
			Version:      "v1",
			Status:       interfaces.BizStatusPublished.String(),
			Name:         "demo",
			Description:  "demo",
			SkillContent: "guide",
		}, nil)
		mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
		mockAuthService.EXPECT().CheckExecutePermission(gomock.Any(), gomock.Any(), "skill-unpublish-1", interfaces.AuthResourceTypeSkill).Return(nil)
		mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-unpublish-1", "v1").Return(nil, errors.New("build package reached"))

		resp, err := service.ExecuteSkill(context.Background(), &interfaces.ExecuteSkillReq{
			BusinessDomainID: "bd-1",
			UserID:           "user-1",
			SkillID:          "skill-unpublish-1",
			Entrypoint:       "to_pdf",
		})

		So(resp, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "build package reached")
		So(err.Error(), ShouldNotContainSubstring, "runtime profile is offline")
	})
}
