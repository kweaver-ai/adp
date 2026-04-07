package skill

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/sandbox"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

const (
	defaultRuntimeType = "python"
)

// skillRuntimeService Skill Runtime Profile + Execution 编排服务
type skillRuntimeService struct {
	registry     *skillRegistry
	profileRepo  model.ISkillRuntimeProfile
	sessionPool  sandbox.SessionPool
	controlPlane interfaces.SandBoxControlPlane
	ossClient    interfaces.OSSGatewayBackendClient
	AuthService  interfaces.IAuthorizationService
	Logger       interfaces.Logger
}

var (
	runtimeServiceOnce sync.Once
	runtimeServiceInst *skillRuntimeService
)

// NewSkillRuntimeService 创建 Skill 执行编排服务
func NewSkillRuntimeService() interfaces.SkillRuntimeProfileService {
	return newSkillRuntimeService()
}

// NewSkillExecutionService 创建 Skill 执行服务
func NewSkillExecutionService() interfaces.SkillExecutionService {
	return newSkillRuntimeService()
}

func newSkillRuntimeService() *skillRuntimeService {
	runtimeServiceOnce.Do(func() {
		conf := config.NewConfigLoader()
		runtimeServiceInst = &skillRuntimeService{
			registry:     NewSkillRegistry().(*skillRegistry),
			profileRepo:  dbaccess.NewSkillRuntimeProfileDB(),
			sessionPool:  sandbox.GetSessionPool(),
			controlPlane: drivenadapters.NewSandBoxControlPlaneClient(),
			ossClient:    drivenadapters.NewOSSGatewayBackendClient(),
			AuthService:  auth.NewAuthServiceImpl(),
			Logger:       conf.GetLogger(),
		}
	})
	return runtimeServiceInst
}

// UpsertSkillRuntimeProfile 新增/更新 Skill Runtime Profile
func (s *skillRuntimeService) UpsertSkillRuntimeProfile(ctx context.Context, req *interfaces.UpsertSkillRuntimeProfileReq) (resp *interfaces.UpsertSkillRuntimeProfileResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id":   req.SkillID,
		"entrypoint": req.Entrypoint,
		"user_id":    req.UserID,
		"bd_id":      req.BusinessDomainID,
	})

	skill, err := s.resolveCurrentSkill(ctx, req.SkillID)
	if err != nil {
		return nil, err
	}
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = s.AuthService.CheckModifyPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	if err = validateRuntimeProfileDefinition(req.Entrypoint, req.CommandTemplate, req.InputSchema, req.OutputSchema); err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	profile := &model.SkillRuntimeProfileDB{
		SkillID:         skill.SkillID,
		SkillVersion:    skill.Version,
		Entrypoint:      req.Entrypoint,
		Name:            req.Name,
		Description:     req.Description,
		RuntimeType:     req.RuntimeType,
		CommandTemplate: utils.ObjectToJSON(req.CommandTemplate),
		InputSchema:     utils.ObjectToJSON(req.InputSchema),
		OutputSchema:    utils.ObjectToJSON(req.OutputSchema),
		Timeout:         req.Timeout,
		Status:          req.Status.String(),
		ExtendInfo:      utils.ObjectToJSON(req.ExtendInfo),
		CreateUser:      req.UserID,
		UpdateUser:      req.UserID,
	}
	if profile.RuntimeType == "" {
		profile.RuntimeType = defaultRuntimeType
	}
	if profile.Status == "" {
		profile.Status = interfaces.BizStatusPublished.String()
	}

	existing, err := s.profileRepo.SelectSkillRuntimeProfileBySkillIDAndEntrypoint(ctx, nil, skill.SkillID, skill.Version, req.Entrypoint)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		profile.ID = existing.ID
		profile.CreateTime = existing.CreateTime
		profile.CreateUser = existing.CreateUser
		if err = s.profileRepo.UpdateSkillRuntimeProfile(ctx, nil, profile); err != nil {
			return nil, err
		}
	} else {
		if err = s.profileRepo.InsertSkillRuntimeProfile(ctx, nil, profile); err != nil {
			return nil, err
		}
	}
	resp = &interfaces.UpsertSkillRuntimeProfileResp{
		Profile: convertRuntimeProfile(profile),
	}
	return resp, nil
}

// GetSkillRuntimeProfile 查询 Skill Runtime Profile
func (s *skillRuntimeService) GetSkillRuntimeProfile(ctx context.Context, req *interfaces.GetSkillRuntimeProfileReq) (resp *interfaces.GetSkillRuntimeProfileResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id":   req.SkillID,
		"entrypoint": req.Entrypoint,
		"user_id":    req.UserID,
		"bd_id":      req.BusinessDomainID,
	})

	skill, err := s.resolveCurrentSkill(ctx, req.SkillID)
	if err != nil {
		return nil, err
	}
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = s.AuthService.CheckViewPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.SelectSkillRuntimeProfileBySkillIDAndEntrypoint(ctx, nil, skill.SkillID, skill.Version, req.Entrypoint)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("runtime profile not found: %s/%s", req.SkillID, req.Entrypoint))
	}
	resp = &interfaces.GetSkillRuntimeProfileResp{
		Profile: convertRuntimeProfile(profile),
	}
	return resp, nil
}

// ExecuteSkill 执行 Skill Runtime Profile
func (s *skillRuntimeService) ExecuteSkill(ctx context.Context, req *interfaces.ExecuteSkillReq) (resp *interfaces.ExecuteSkillResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id":   req.SkillID,
		"entrypoint": req.Entrypoint,
		"user_id":    req.UserID,
		"bd_id":      req.BusinessDomainID,
	})

	skill, err := s.resolveCurrentSkill(ctx, req.SkillID)
	if err != nil {
		return nil, err
	}
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = s.AuthService.CheckExecutePermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.SelectSkillRuntimeProfileBySkillIDAndEntrypoint(ctx, nil, skill.SkillID, skill.Version, req.Entrypoint)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("runtime profile not found: %s/%s", req.SkillID, req.Entrypoint))
	}
	if strings.EqualFold(profile.Status, interfaces.BizStatusOffline.String()) {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("runtime profile is offline: %s/%s", req.SkillID, req.Entrypoint))
	}
	inputSchema := utils.JSONToObject[map[string]any](profile.InputSchema)
	outputSchema := utils.JSONToObject[map[string]any](profile.OutputSchema)
	if err = validateRuntimeProfileDefinition(profile.Entrypoint, utils.JSONToObject[[]string](profile.CommandTemplate), inputSchema, outputSchema); err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}
	effectiveInputs := applyRuntimeInputDefaults(inputSchema, req.Inputs)
	if err = validateExecuteSkillInputs(inputSchema, effectiveInputs); err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	packageContent, err := s.registry.buildSkillPackage(ctx, skill)
	if err != nil {
		return nil, err
	}
	packageHash := checksumBytes(packageContent)
	sessionID, release, err := s.sessionPool.BorrowSession(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	packagePath := filepath.ToSlash(filepath.Join(".packages", skill.SkillID, skill.Version, packageHash+".zip"))
	if _, err = s.controlPlane.UploadSessionFile(ctx, sessionID, packagePath, packageContent, "application/zip"); err != nil {
		return nil, err
	}
	materialized, err := s.controlPlane.MaterializePackage(ctx, sessionID, &interfaces.MaterializePackageReq{
		PackagePath: packagePath,
		PackageHash: packageHash,
	})
	if err != nil {
		return nil, err
	}
	taskID := buildSkillTaskID(skill.SkillID, profile.Entrypoint)
	taskWorkspace, err := s.controlPlane.PrepareTaskWorkspace(ctx, sessionID, &interfaces.PrepareTaskWorkspaceReq{
		TaskID:     taskID,
		TaskType:   "skill",
		CreateDirs: []string{"input", "output", "tmp", "logs"},
		Reset:      true,
	})
	if err != nil {
		return nil, err
	}
	inputMappings, err := s.materializeRuntimeInputs(ctx, sessionID, taskWorkspace, effectiveInputs)
	if err != nil {
		return nil, err
	}

	commandTemplate := utils.JSONToObject[[]string](profile.CommandTemplate)
	if len(commandTemplate) == 0 {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("runtime profile command template is empty: %s/%s", req.SkillID, req.Entrypoint))
	}
	renderVars := flattenRuntimeVariables(skill, profile, effectiveInputs)
	for key, value := range inputMappings {
		renderVars[key] = value
		renderVars["inputs."+key] = value
	}
	for key, rel := range taskWorkspace.Directories {
		renderVars[key+"_dir"] = workspaceAbsPath(rel)
	}
	outputMappings := buildRuntimeOutputMappings(taskWorkspace, profile.OutputSchema)
	outputRefs := buildRuntimeOutputRefs(sessionID, profile.OutputSchema, outputMappings)
	for key, value := range outputMappings {
		renderVars[key] = value
		renderVars["outputs."+key] = value
	}
	renderVars["task_root"] = workspaceAbsPath(taskWorkspace.TaskRoot)
	renderVars["package_root"] = workspaceAbsPath(filepath.ToSlash(filepath.Join(materialized.TargetDir, "package")))
	renderedCommand, err := renderCommandTemplate(commandTemplate, renderVars)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}
	if len(renderedCommand) == 0 {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("runtime profile command template render failed: %s/%s", req.SkillID, req.Entrypoint))
	}

	timeout := profile.Timeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}
	if timeout <= 0 {
		timeout = 300
	}
	execCode := buildSkillShellCode(materialized, taskWorkspace, renderedCommand)
	execReq := &interfaces.ExecuteCodeReq{
		Code:     execCode,
		Event:    map[string]any{},
		Language: "shell",
		Timeout:  timeout,
	}

	execResp, err := s.controlPlane.ExecuteCodeSync(ctx, sessionID, execReq)
	if err != nil {
		return nil, err
	}
	resp = &interfaces.ExecuteSkillResp{
		SkillID:       skill.SkillID,
		SkillVersion:  skill.Version,
		Entrypoint:    profile.Entrypoint,
		SessionID:     sessionID,
		RuntimeType:   profile.RuntimeType,
		ExitCode:      execResp.ExitCode,
		ErrorMessage:  execResp.ErrorMessage,
		ExecutionTime: execResp.ExecutionTime,
		Stdout:        execResp.Stdout,
		Stderr:        execResp.Stderr,
		ReturnValue: map[string]any{
			"task_id":          taskWorkspace.TaskID,
			"task_root":        workspaceAbsPath(taskWorkspace.TaskRoot),
			"package_target":   workspaceAbsPath(materialized.TargetDir),
			"package_checksum": materialized.Checksum,
			"directories":      taskWorkspace.Directories,
			"effective_inputs": effectiveInputs,
			"input_mappings":   inputMappings,
			"output_mappings":  outputMappings,
			"output_refs":      outputRefs,
		},
		Profile: convertRuntimeProfile(profile),
	}
	if outputDir, ok := taskWorkspace.Directories["output"]; ok && outputDir != "" {
		returnValue := resp.ReturnValue.(map[string]any)
		if files, fileErr := s.controlPlane.ListSessionFiles(ctx, sessionID, outputDir, 1000); fileErr == nil && files != nil {
			files.Files = normalizeSessionFiles(files.Files)
			returnValue["output_files"] = files.Files
			returnValue["output_file_count"] = files.Count
			if artifactRefs, artifactErr := s.persistRuntimeOutputArtifacts(ctx, sessionID, skill, taskWorkspace, outputRefs, files.Files); artifactErr == nil && len(artifactRefs) > 0 {
				returnValue["output_artifacts"] = artifactRefs
			} else if artifactErr != nil {
				appendRuntimeWarning(returnValue, "output_artifact_persist_failed", artifactErr)
			}
		} else if fileErr != nil {
			appendRuntimeWarning(returnValue, "output_file_list_failed", fileErr)
		}
	}
	if resp.ExitCode != 0 && resp.ErrorMessage == "" {
		resp.ErrorMessage = fmt.Sprintf("skill execution failed with exit code %d", resp.ExitCode)
	}
	return resp, nil
}

func (s *skillRuntimeService) resolveCurrentSkill(ctx context.Context, skillID string) (*model.SkillRepositoryDB, error) {
	skill, err := s.registry.skillRepo.SelectSkillByID(ctx, nil, skillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", skillID))
	}
	return skill, nil
}

func convertRuntimeProfile(profile *model.SkillRuntimeProfileDB) *interfaces.SkillRuntimeProfileInfo {
	if profile == nil {
		return nil
	}
	return &interfaces.SkillRuntimeProfileInfo{
		SkillID:         profile.SkillID,
		SkillVersion:    profile.SkillVersion,
		Entrypoint:      profile.Entrypoint,
		Name:            profile.Name,
		Description:     profile.Description,
		RuntimeType:     profile.RuntimeType,
		CommandTemplate: utils.JSONToObject[[]string](profile.CommandTemplate),
		InputSchema:     utils.JSONToObject[map[string]any](profile.InputSchema),
		OutputSchema:    utils.JSONToObject[map[string]any](profile.OutputSchema),
		Timeout:         profile.Timeout,
		Status:          profile.Status,
		ExtendInfo:      utils.JSONToObject[map[string]any](profile.ExtendInfo),
		CreateUser:      profile.CreateUser,
		CreateTime:      profile.CreateTime,
		UpdateUser:      profile.UpdateUser,
		UpdateTime:      profile.UpdateTime,
	}
}

func flattenRuntimeVariables(skill *model.SkillRepositoryDB, profile *model.SkillRuntimeProfileDB, inputs map[string]any) map[string]string {
	vars := map[string]string{
		"skill_id":      skill.SkillID,
		"skill_version": skill.Version,
		"entrypoint":    profile.Entrypoint,
		"runtime_type":  profile.RuntimeType,
	}
	for key, value := range inputs {
		switch v := value.(type) {
		case string:
			vars[key] = v
			vars["inputs."+key] = v
		default:
			vars[key] = utils.ObjectToJSON(v)
			vars["inputs."+key] = vars[key]
		}
	}
	return vars
}

func (s *skillRuntimeService) materializeRuntimeInputs(ctx context.Context, sessionID string, taskWorkspace *interfaces.PrepareTaskWorkspaceResp, inputs map[string]any) (map[string]string, error) {
	if len(inputs) == 0 {
		return map[string]string{}, nil
	}
	inputDir, ok := taskWorkspace.Directories["input"]
	if !ok || inputDir == "" {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, "task workspace missing input directory")
	}

	mappings := make(map[string]string, len(inputs))
	for key, raw := range inputs {
		switch value := raw.(type) {
		case string:
			mappings[key] = value
		case map[string]any:
			resolved, err := s.materializeRuntimeInputObject(ctx, sessionID, inputDir, key, value)
			if err != nil {
				return nil, err
			}
			mappings[key] = resolved
		default:
			mappings[key] = utils.ObjectToJSON(value)
		}
	}
	return mappings, nil
}

func (s *skillRuntimeService) materializeRuntimeInputObject(ctx context.Context, sessionID, inputDir, key string, value map[string]any) (string, error) {
	inputType := strings.ToLower(strings.TrimSpace(stringValue(value["type"])))
	if inputType == "" {
		inputType = strings.ToLower(strings.TrimSpace(stringValue(value["kind"])))
	}
	switch inputType {
	case "", "text", "path":
		return stringValue(value["value"]), nil
	case "inline_file":
		filename := sanitizeRuntimeFileName(stringValue(value["filename"]))
		if filename == "" {
			filename = sanitizeRuntimeFileName(key)
		}
		content := stringValue(value["content"])
		if content == "" {
			return "", errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("inline_file input %q missing content", key))
		}
		targetRel := filepath.ToSlash(filepath.Join(inputDir, filename))
		if _, err := s.controlPlane.UploadSessionFile(ctx, sessionID, targetRel, []byte(content), stringValueDefault(value["content_type"], "text/plain")); err != nil {
			return "", err
		}
		return workspaceAbsPath(targetRel), nil
	case "inline_file_base64":
		filename := sanitizeRuntimeFileName(stringValue(value["filename"]))
		if filename == "" {
			filename = sanitizeRuntimeFileName(key)
		}
		encoded := stringValue(value["content_base64"])
		if encoded == "" {
			return "", errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("inline_file_base64 input %q missing content_base64", key))
		}
		content, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid base64 for input %q: %v", key, err))
		}
		targetRel := filepath.ToSlash(filepath.Join(inputDir, filename))
		if _, err = s.controlPlane.UploadSessionFile(ctx, sessionID, targetRel, content, stringValueDefault(value["content_type"], "application/octet-stream")); err != nil {
			return "", err
		}
		return workspaceAbsPath(targetRel), nil
	case "artifact_ref", "resource_ref", "oss_object":
		filename := sanitizeRuntimeFileName(stringValue(value["filename"]))
		object, err := s.resolveRuntimeInputObjectRef(ctx, value)
		if err != nil {
			return "", err
		}
		if filename == "" {
			filename = sanitizeRuntimeFileName(filepath.Base(object.StorageKey))
		}
		if filename == "" {
			filename = sanitizeRuntimeFileName(key)
		}
		if !s.ossClient.IsReady() {
			return "", errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "oss gateway backend is not ready")
		}
		content, err := s.ossClient.DownloadFile(ctx, object)
		if err != nil {
			return "", err
		}
		targetRel := filepath.ToSlash(filepath.Join(inputDir, filename))
		if _, err = s.controlPlane.UploadSessionFile(ctx, sessionID, targetRel, content, stringValueDefault(value["content_type"], "application/octet-stream")); err != nil {
			return "", err
		}
		return workspaceAbsPath(targetRel), nil
	default:
		return "", errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("unsupported input type for %q: %s", key, inputType))
	}
}

func (s *skillRuntimeService) resolveRuntimeInputObjectRef(ctx context.Context, value map[string]any) (*interfaces.OssObject, error) {
	storageID := strings.TrimSpace(stringValue(value["storage_id"]))
	storageKey := strings.TrimSpace(stringValue(value["storage_key"]))
	if storageID == "" {
		storageID = strings.TrimSpace(stringValue(value["storageId"]))
	}
	if storageKey == "" {
		storageKey = strings.TrimSpace(stringValue(value["storageKey"]))
	}
	if storageID != "" && storageKey != "" {
		return &interfaces.OssObject{
			StorageID:  storageID,
			StorageKey: storageKey,
		}, nil
	}

	if rawObject, ok := value["object"].(map[string]any); ok {
		storageID = strings.TrimSpace(stringValue(rawObject["storage_id"]))
		storageKey = strings.TrimSpace(stringValue(rawObject["storage_key"]))
		if storageID == "" {
			storageID = strings.TrimSpace(stringValue(rawObject["storageId"]))
		}
		if storageKey == "" {
			storageKey = strings.TrimSpace(stringValue(rawObject["storageKey"]))
		}
		if storageID != "" && storageKey != "" {
			return &interfaces.OssObject{
				StorageID:  storageID,
				StorageKey: storageKey,
			}, nil
		}
	}

	return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, "resource/artifact ref requires storage_id and storage_key")
}

func stringValue(v any) string {
	switch value := v.(type) {
	case string:
		return value
	default:
		return ""
	}
}

func stringValueDefault(v any, fallback string) string {
	if value := stringValue(v); value != "" {
		return value
	}
	return fallback
}

func sanitizeRuntimeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "_")
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	return name
}

func sanitizeRuntimeRelativePath(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	clean := filepath.ToSlash(filepath.Clean(name))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	clean = strings.ReplaceAll(clean, "..", "_")
	if clean == "" || clean == "." {
		return ""
	}
	return clean
}

func buildRuntimeOutputMappings(taskWorkspace *interfaces.PrepareTaskWorkspaceResp, rawOutputSchema string) map[string]string {
	outputDir, ok := taskWorkspace.Directories["output"]
	if !ok || outputDir == "" {
		return map[string]string{}
	}
	outputSchema := utils.JSONToObject[map[string]any](rawOutputSchema)
	if len(outputSchema) == 0 {
		return map[string]string{}
	}
	mappings := make(map[string]string, len(outputSchema))
	for name, raw := range outputSchema {
		key := strings.TrimSpace(name)
		if key == "" {
			continue
		}
		targetRel := filepath.ToSlash(filepath.Join(outputDir, sanitizeRuntimeRelativePath(key)))
		if outputDef, ok := raw.(map[string]any); ok {
			outputType := strings.ToLower(strings.TrimSpace(stringValue(outputDef["type"])))
			if outputType == "" {
				outputType = strings.ToLower(strings.TrimSpace(stringValue(outputDef["kind"])))
			}
			if pathValue := sanitizeRuntimeRelativePath(stringValue(outputDef["path"])); pathValue != "" {
				targetRel = filepath.ToSlash(filepath.Join(outputDir, pathValue))
			} else if outputType == "file" {
				targetRel = filepath.ToSlash(filepath.Join(outputDir, sanitizeRuntimeFileName(key)))
			}
		}
		mappings[key] = workspaceAbsPath(targetRel)
	}
	return mappings
}

func buildRuntimeOutputRefs(sessionID, rawOutputSchema string, outputMappings map[string]string) map[string]any {
	outputSchema := utils.JSONToObject[map[string]any](rawOutputSchema)
	if len(outputSchema) == 0 || len(outputMappings) == 0 {
		return map[string]any{}
	}
	refs := make(map[string]any, len(outputMappings))
	for name, absPath := range outputMappings {
		ref := map[string]any{
			"name":           name,
			"session_id":     sessionID,
			"container_path": workspaceContainerPath(absPath),
			"abs_path":       absPath,
			"ref_type":       "sandbox_path",
		}
		if outputDef, ok := outputSchema[name].(map[string]any); ok {
			if outputType := normalizeRuntimeSchemaType(outputDef); outputType != "" {
				ref["type"] = outputType
			}
			if description := strings.TrimSpace(stringValue(outputDef["description"])); description != "" {
				ref["description"] = description
			}
			if pathValue := strings.TrimSpace(stringValue(outputDef["path"])); pathValue != "" {
				ref["declared_path"] = pathValue
			}
		}
		if _, exists := ref["type"]; !exists {
			ref["type"] = "file"
		}
		refs[name] = ref
	}
	return refs
}

func normalizeSessionFiles(files []*interfaces.SessionFileInfo) []*interfaces.SessionFileInfo {
	if len(files) == 0 {
		return files
	}
	normalized := make([]*interfaces.SessionFileInfo, 0, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		cloned := *file
		cloned.ContainerPath = workspaceContainerPath(file.ContainerPath)
		normalized = append(normalized, &cloned)
	}
	return normalized
}

func appendRuntimeWarning(returnValue map[string]any, code string, err error) {
	if returnValue == nil || strings.TrimSpace(code) == "" || err == nil {
		return
	}
	warning := map[string]any{
		"code":    strings.TrimSpace(code),
		"message": err.Error(),
	}
	existing, _ := returnValue["warnings"].([]map[string]any)
	returnValue["warnings"] = append(existing, warning)
}

func (s *skillRuntimeService) persistRuntimeOutputArtifacts(ctx context.Context, sessionID string, skill *model.SkillRepositoryDB, taskWorkspace *interfaces.PrepareTaskWorkspaceResp, outputRefs map[string]any, files []*interfaces.SessionFileInfo) (map[string]any, error) {
	if len(outputRefs) == 0 || len(files) == 0 {
		return map[string]any{}, nil
	}
	if s.ossClient == nil || !s.ossClient.IsReady() {
		return map[string]any{}, nil
	}
	storageID, err := s.ossClient.CurrentStorageID(ctx)
	if err != nil {
		return nil, err
	}
	filesByPath := make(map[string]*interfaces.SessionFileInfo, len(files))
	for _, file := range files {
		if file == nil || strings.TrimSpace(file.ContainerPath) == "" {
			continue
		}
		filesByPath[filepath.ToSlash(file.ContainerPath)] = file
	}
	artifacts := make(map[string]any)
	for name, rawRef := range outputRefs {
		ref, ok := rawRef.(map[string]any)
		if !ok {
			continue
		}
		outputType := strings.ToLower(strings.TrimSpace(stringValue(ref["type"])))
		if outputType != "" && outputType != "file" {
			continue
		}
		containerPath := filepath.ToSlash(strings.TrimSpace(stringValue(ref["container_path"])))
		if containerPath == "" {
			continue
		}
		if _, ok := filesByPath[containerPath]; !ok {
			continue
		}
		downloaded, err := s.controlPlane.DownloadSessionFile(ctx, sessionID, containerPath)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(downloaded.PresignedURL) != "" {
			return nil, fmt.Errorf("sandbox returned presigned_url for output %s; direct artifact persistence is not supported yet", containerPath)
		}
		object := &interfaces.OssObject{
			StorageID:  storageID,
			StorageKey: buildRuntimeOutputArtifactKey(skill, taskWorkspace.TaskID, name, containerPath),
		}
		if err = s.ossClient.UploadFile(ctx, object, downloaded.Content); err != nil {
			return nil, err
		}
		artifacts[name] = map[string]any{
			"name":        name,
			"type":        "file",
			"storage_id":  object.StorageID,
			"storage_key": object.StorageKey,
			"ref_type":    "artifact_ref",
			"source": map[string]any{
				"session_id":     sessionID,
				"container_path": containerPath,
			},
		}
	}
	return artifacts, nil
}

func buildRuntimeOutputArtifactKey(skill *model.SkillRepositoryDB, taskID, outputName, containerPath string) string {
	fileName := sanitizeRuntimeFileName(filepath.Base(containerPath))
	if fileName == "" {
		fileName = sanitizeRuntimeFileName(outputName)
	}
	return filepath.ToSlash(filepath.Join(
		interfaces.OSSGatewayPrefix,
		"skill-execution",
		skill.SkillID,
		skill.Version,
		taskID,
		sanitizeRuntimeRelativePath(outputName),
		fileName,
	))
}

func validateRuntimeProfileDefinition(entrypoint string, commandTemplate []string, inputSchema, outputSchema map[string]any) error {
	if strings.TrimSpace(entrypoint) == "" {
		return fmt.Errorf("runtime profile entrypoint is empty")
	}
	if err := validateRuntimeOutputs(outputSchema); err != nil {
		return fmt.Errorf("invalid runtime profile outputs: %w", err)
	}
	entry := &skillRuntimeEntryDef{
		Name:    entrypoint,
		Command: commandTemplate,
		Inputs:  inputSchema,
		Outputs: outputSchema,
	}
	if len(commandTemplate) == 0 {
		return fmt.Errorf("runtime profile command template is empty")
	}
	if err := validateRuntimeCommandTemplate(entry, nil); err != nil {
		return fmt.Errorf("invalid runtime profile command: %w", err)
	}
	return nil
}

func validateExecuteSkillInputs(schema map[string]any, inputs map[string]any) error {
	if len(schema) == 0 {
		return nil
	}
	for name, rawSchema := range schema {
		key := strings.TrimSpace(name)
		if key == "" {
			continue
		}
		schemaDef, _ := rawSchema.(map[string]any)
		if isRuntimeInputRequired(schemaDef) {
			value, exists := inputs[key]
			if !exists || isRuntimeInputEmpty(value) {
				return fmt.Errorf("missing required input: %s", key)
			}
		}
		if value, exists := inputs[key]; exists {
			if err := validateExecuteSkillInputValue(key, schemaDef, value); err != nil {
				return err
			}
			if err := validateExecuteSkillInputEnum(key, schemaDef, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateExecuteSkillInputValue(name string, schema map[string]any, value any) error {
	inputType := normalizeRuntimeSchemaType(schema)
	switch inputType {
	case "", "any":
		return nil
	case "text", "string":
		if _, ok := value.(string); ok {
			return nil
		}
		if inputObject, ok := value.(map[string]any); ok {
			switch normalizeRuntimeInputObjectType(inputObject) {
			case "text":
				return nil
			}
		}
		return fmt.Errorf("input %q expects text", name)
	case "path":
		if _, ok := value.(string); ok {
			return nil
		}
		if inputObject, ok := value.(map[string]any); ok {
			switch normalizeRuntimeInputObjectType(inputObject) {
			case "path", "artifact_ref", "resource_ref", "oss_object", "inline_file", "inline_file_base64":
				return nil
			}
		}
		return fmt.Errorf("input %q expects a path-like value", name)
	case "file":
		if inputObject, ok := value.(map[string]any); ok {
			switch normalizeRuntimeInputObjectType(inputObject) {
			case "path", "artifact_ref", "resource_ref", "oss_object", "inline_file", "inline_file_base64":
				return nil
			}
		}
		if _, ok := value.(string); ok {
			return nil
		}
		return fmt.Errorf("input %q expects a file reference", name)
	case "directory", "dir":
		if _, ok := value.(string); ok {
			return nil
		}
		if inputObject, ok := value.(map[string]any); ok && normalizeRuntimeInputObjectType(inputObject) == "path" {
			return nil
		}
		return fmt.Errorf("input %q expects a directory path", name)
	default:
		return nil
	}
}

func validateExecuteSkillInputEnum(name string, schema map[string]any, value any) error {
	enumValues, ok := schema["enum"].([]any)
	if !ok || len(enumValues) == 0 {
		return nil
	}
	switch current := value.(type) {
	case string:
		for _, candidate := range enumValues {
			if candidateStr, ok := candidate.(string); ok && current == candidateStr {
				return nil
			}
		}
		return fmt.Errorf("input %q must be one of %v", name, enumValues)
	default:
		return nil
	}
}

func applyRuntimeInputDefaults(schema, inputs map[string]any) map[string]any {
	if len(schema) == 0 && len(inputs) == 0 {
		return map[string]any{}
	}
	merged := make(map[string]any, len(schema)+len(inputs))
	for key, value := range inputs {
		merged[key] = value
	}
	for name, rawSchema := range schema {
		key := strings.TrimSpace(name)
		if key == "" {
			continue
		}
		if _, exists := merged[key]; exists {
			continue
		}
		schemaDef, _ := rawSchema.(map[string]any)
		if defaultValue, ok := schemaDef["default"]; ok {
			merged[key] = defaultValue
		}
	}
	return merged
}

func normalizeRuntimeInputSchemaType(schema map[string]any) string {
	if len(schema) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(stringValueDefault(schema["type"], stringValue(schema["kind"]))))
}

func normalizeRuntimeSchemaType(schema map[string]any) string {
	return normalizeRuntimeInputSchemaType(schema)
}

func normalizeRuntimeInputObjectType(value map[string]any) string {
	if len(value) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(stringValueDefault(value["type"], stringValue(value["kind"]))))
}

func isRuntimeInputRequired(schema map[string]any) bool {
	if len(schema) == 0 {
		return false
	}
	if required, ok := schema["required"].(bool); ok {
		return required
	}
	return false
}

func isRuntimeInputEmpty(value any) bool {
	switch current := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(current) == ""
	case map[string]any:
		return len(current) == 0
	default:
		return false
	}
}

func renderCommandTemplate(commandTemplate []string, vars map[string]string) ([]string, error) {
	rendered := make([]string, 0, len(commandTemplate))
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})
	for _, arg := range commandTemplate {
		renderedArg := arg
		for _, key := range keys {
			renderedArg = strings.ReplaceAll(renderedArg, fmt.Sprintf("{{%s}}", key), vars[key])
		}
		if strings.Contains(renderedArg, "{{") || strings.Contains(renderedArg, "}}") {
			return nil, fmt.Errorf("unresolved command template placeholder in %q", renderedArg)
		}
		rendered = append(rendered, renderedArg)
	}
	return rendered, nil
}

func checksumBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func workspaceAbsPath(rel string) string {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "/"))
	return "/workspace/" + rel
}

func workspaceContainerPath(absOrRel string) string {
	if strings.TrimSpace(absOrRel) == "" {
		return ""
	}
	normalized := filepath.ToSlash(strings.TrimSpace(absOrRel))
	normalized = strings.TrimPrefix(normalized, "/workspace/")
	normalized = strings.TrimPrefix(normalized, "workspace/")
	return strings.TrimPrefix(normalized, "/")
}

func buildSkillTaskID(skillID, entrypoint string) string {
	cleanSkill := strings.ReplaceAll(skillID, "/", "_")
	cleanEntrypoint := strings.ReplaceAll(entrypoint, "/", "_")
	return fmt.Sprintf("%s_%s_%d", cleanSkill, cleanEntrypoint, time.Now().UnixNano())
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func shellJoinCommand(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func buildSkillShellCode(materialized *interfaces.MaterializePackageResp, taskWorkspace *interfaces.PrepareTaskWorkspaceResp, command []string) string {
	packageRoot := workspaceAbsPath(filepath.ToSlash(filepath.Join(materialized.TargetDir, "package")))

	lines := []string{
		"set -euo pipefail",
		"PACKAGE_ROOT=" + shellQuote(packageRoot),
		"TASK_ROOT=" + shellQuote(workspaceAbsPath(taskWorkspace.TaskRoot)),
		"mkdir -p \"$TASK_ROOT\"",
	}
	for name, rel := range taskWorkspace.Directories {
		lines = append(lines, strings.ToUpper(name)+"_DIR="+shellQuote(workspaceAbsPath(rel)))
	}
	lines = append(lines,
		"cd "+shellQuote(packageRoot),
		shellJoinCommand(command),
	)
	return strings.Join(lines, "\n")
}
