package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"gopkg.in/yaml.v3"
)

type skillParser struct{}

type skillFrontmatter struct {
	Name         string                 `yaml:"name" validate:"required"`
	Description  string                 `yaml:"description" validate:"required"`
	Dependencies map[string]interface{} `yaml:"dependencies"`
	Metadata     map[string]interface{} `yaml:"metadata"`
}

// 设置SKILL.md统一命名

const SkillMD = "SKILL.md"
const SkillRuntimeYAML = "skill.runtime.yaml"

type skillRuntimeYAML struct {
	Version     int                     `yaml:"version"`
	Entrypoints []*skillRuntimeEntryDef `yaml:"entrypoints"`
}

type skillRuntimeEntryDef struct {
	Name        string         `yaml:"name" validate:"required"`
	Description string         `yaml:"description"`
	RuntimeType string         `yaml:"runtime_type"`
	Command     []string       `yaml:"command" validate:"required,min=1"`
	Inputs      map[string]any `yaml:"inputs"`
	Outputs     map[string]any `yaml:"outputs"`
	Timeout     int            `yaml:"timeout"`
	Status      string         `yaml:"status"`
	ExtendInfo  map[string]any `yaml:"extend_info"`
}

type parsedRuntimeProfile struct {
	Entrypoint      string
	Name            string
	Description     string
	RuntimeType     string
	CommandTemplate []string
	InputSchema     map[string]any
	OutputSchema    map[string]any
	Timeout         int
	Status          string
	ExtendInfo      map[string]any
}

// skillAsset 技能资产
type skillAsset struct {
	RelPath  string
	FileType string
	MimeType string
	Content  []byte
}

var runtimeTemplateVarPattern = regexp.MustCompile(`\{\{([^{}]+)\}\}`)

var allowedRuntimeExecutables = map[string]struct{}{
	"python":  {},
	"python3": {},
	"bash":    {},
	"sh":      {},
	"node":    {},
	"nodejs":  {},
}

func newSkillParser() *skillParser {
	return &skillParser{}
}

func (p *skillParser) parseRegisterReq(req *interfaces.RegisterSkillReq) (skillDB *model.SkillRepositoryDB, files []*interfaces.SkillFileSummary, assets []*skillAsset, runtimeProfiles []*parsedRuntimeProfile, err error) {
	switch req.FileType {
	case "content":
		content, err := decodeContent(req.File)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		skill, err := p.parseSkillContent(content, req)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return skill, nil, nil, nil, nil
	case "zip":
		content, files, assets, runtimeProfiles, err := p.parseSkillZip(req)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		skill, err := p.parseSkillContent(content, req)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return skill, files, assets, runtimeProfiles, nil
	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported file type: %s", req.FileType)
	}
}

func (p *skillParser) parseSkillContent(content string, req *interfaces.RegisterSkillReq) (*model.SkillRepositoryDB, error) {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid SKILL.md format: missing frontmatter")
	}

	fm := &skillFrontmatter{}
	if err := yaml.Unmarshal([]byte(parts[1]), fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skill frontmatter: %w", err)
	}
	if err := validator.New().Struct(fm); err != nil {
		return nil, fmt.Errorf("invalid skill frontmatter: %w", err)
	}

	skill := &model.SkillRepositoryDB{
		Name:         fm.Name,
		Description:  fm.Description,
		SkillContent: strings.TrimSpace(parts[2]),
		Version:      uuid.New().String(),
		Status:       interfaces.BizStatusUnpublish.String(),
		Source:       req.Source,
		Dependencies: utils.ObjectToJSON(fm.Dependencies),
		ExtendInfo:   utils.ObjectToJSON(fm.Metadata),
		CreateUser:   req.UserID,
		UpdateUser:   req.UserID,
	}
	return skill, nil
}

func (p *skillParser) parseSkillZip(req *interfaces.RegisterSkillReq) (string, []*interfaces.SkillFileSummary, []*skillAsset, []*parsedRuntimeProfile, error) {
	reader, err := zip.NewReader(bytes.NewReader(req.File), int64(len(req.File)))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to open zip: %w", err)
	}

	var skillContent string
	var runtimeYAMLContent []byte
	files := make([]*interfaces.SkillFileSummary, 0, len(reader.File))
	assets := make([]*skillAsset, 0, len(reader.File))
	availablePaths := make(map[string]struct{}, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		relPath, err := normalizeZipPath(file.Name)
		if err != nil {
			return "", nil, nil, nil, err
		}

		rc, err := file.Open()
		if err != nil {
			return "", nil, nil, nil, err
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			return "", nil, nil, nil, readErr
		}

		if strings.EqualFold(relPath, SkillMD) {
			skillContent = string(content)
			// 如果refpath为SKILL.md，转换为大写
			relPath = SkillMD
		}
		if strings.EqualFold(relPath, SkillRuntimeYAML) {
			runtimeYAMLContent = content
			relPath = SkillRuntimeYAML
		}
		availablePaths[relPath] = struct{}{}

		files = append(files, &interfaces.SkillFileSummary{
			RelPath:  relPath,
			FileType: detectFileType(relPath),
			Size:     int64(len(content)),
			MimeType: detectMimeType(relPath),
		})
		assets = append(assets, &skillAsset{
			RelPath:  relPath,
			FileType: detectFileType(relPath),
			MimeType: detectMimeType(relPath),
			Content:  content,
		})
	}

	if skillContent == "" {
		return "", nil, nil, nil, fmt.Errorf("SKILL.md not found in zip")
	}
	runtimeProfiles, err := p.parseRuntimeProfiles(runtimeYAMLContent, availablePaths)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return skillContent, files, assets, runtimeProfiles, nil
}

func (p *skillParser) parseRuntimeProfiles(raw []byte, availablePaths map[string]struct{}) ([]*parsedRuntimeProfile, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}

	doc := &skillRuntimeYAML{}
	if err := yaml.Unmarshal(raw, doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skill.runtime.yaml: %w", err)
	}
	if len(doc.Entrypoints) == 0 {
		return nil, fmt.Errorf("skill.runtime.yaml has no entrypoints")
	}

	profiles := make([]*parsedRuntimeProfile, 0, len(doc.Entrypoints))
	entrypoints := make(map[string]struct{}, len(doc.Entrypoints))
	for _, entry := range doc.Entrypoints {
		if entry == nil {
			continue
		}
		if err := validator.New().Struct(entry); err != nil {
			return nil, fmt.Errorf("invalid runtime entrypoint %q: %w", entry.Name, err)
		}
		if _, exists := entrypoints[entry.Name]; exists {
			return nil, fmt.Errorf("duplicate runtime entrypoint: %s", entry.Name)
		}
		entrypoints[entry.Name] = struct{}{}
		if err := validateRuntimeOutputs(entry.Outputs); err != nil {
			return nil, fmt.Errorf("invalid runtime entrypoint %q outputs: %w", entry.Name, err)
		}
		if err := validateRuntimeCommandTemplate(entry, availablePaths); err != nil {
			return nil, fmt.Errorf("invalid runtime entrypoint %q command: %w", entry.Name, err)
		}
		profile := &parsedRuntimeProfile{
			Entrypoint:      entry.Name,
			Name:            entry.Name,
			Description:     entry.Description,
			RuntimeType:     entry.RuntimeType,
			CommandTemplate: entry.Command,
			InputSchema:     entry.Inputs,
			OutputSchema:    entry.Outputs,
			Timeout:         entry.Timeout,
			Status:          entry.Status,
			ExtendInfo:      entry.ExtendInfo,
		}
		if strings.TrimSpace(profile.Description) == "" {
			profile.Description = entry.Name
		}
		profiles = append(profiles, profile)
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("skill.runtime.yaml has no valid entrypoints")
	}
	return profiles, nil
}

func validateRuntimeOutputs(outputs map[string]any) error {
	for name, raw := range outputs {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("output name is empty")
		}
		outputDef, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if pathValue := strings.TrimSpace(stringValue(outputDef["path"])); pathValue != "" {
			if sanitizeRuntimeRelativePath(pathValue) != pathValue && sanitizeRuntimeRelativePath(pathValue) != strings.TrimPrefix(strings.TrimPrefix(pathValue, "./"), "/") {
				return fmt.Errorf("unsafe output path for %s: %s", name, pathValue)
			}
			if sanitizeRuntimeRelativePath(pathValue) == "" {
				return fmt.Errorf("invalid output path for %s: %s", name, pathValue)
			}
		}
	}
	return nil
}

func validateRuntimeCommandTemplate(entry *skillRuntimeEntryDef, availablePaths map[string]struct{}) error {
	if len(entry.Command) == 0 {
		return fmt.Errorf("command is empty")
	}
	if err := validateRuntimeExecutable(entry.Command[0]); err != nil {
		return err
	}
	allowed := map[string]struct{}{
		"skill_id":      {},
		"skill_version": {},
		"entrypoint":    {},
		"runtime_type":  {},
		"package_root":  {},
		"task_root":     {},
		"input_dir":     {},
		"output_dir":    {},
		"tmp_dir":       {},
		"logs_dir":      {},
	}
	for key := range entry.Inputs {
		allowed[key] = struct{}{}
		allowed["inputs."+key] = struct{}{}
	}
	for key := range entry.Outputs {
		allowed[key] = struct{}{}
		allowed["outputs."+key] = struct{}{}
	}
	for _, arg := range entry.Command {
		if err := validateRuntimeCommandArg(arg); err != nil {
			return err
		}
		if candidate := runtimeCommandPackagePathCandidate(arg); candidate != "" && availablePaths != nil {
			if _, ok := availablePaths[candidate]; !ok {
				return fmt.Errorf("command path does not exist in skill package: %s", candidate)
			}
		}
		matches := runtimeTemplateVarPattern.FindAllStringSubmatch(arg, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			name := strings.TrimSpace(match[1])
			if _, ok := allowed[name]; !ok {
				return fmt.Errorf("unsupported placeholder: %s", name)
			}
		}
	}
	return nil
}

func validateRuntimeExecutable(executable string) error {
	name := strings.ToLower(strings.TrimSpace(filepath.Base(executable)))
	if name == "" {
		return fmt.Errorf("runtime executable is empty")
	}
	if _, ok := allowedRuntimeExecutables[name]; !ok {
		return fmt.Errorf("runtime executable is not allowed: %s", executable)
	}
	return nil
}

func runtimeCommandPackagePathCandidate(arg string) string {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" || strings.HasPrefix(trimmed, "-") || strings.Contains(trimmed, "{{") || strings.Contains(trimmed, "}}") {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(trimmed))
	switch ext {
	case ".py", ".sh", ".js", ".ts":
	default:
		return ""
	}
	candidate, err := normalizeZipPath(trimmed)
	if err != nil {
		return ""
	}
	return candidate
}

func validateRuntimeCommandArg(arg string) error {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return nil
	}
	for _, prefix := range []string{"/workspace/", "workspace/", "/runtime/", "runtime/", "/.tasks/", ".tasks/"} {
		if strings.HasPrefix(trimmed, prefix) {
			return fmt.Errorf("sandbox absolute path is not allowed in command: %s", trimmed)
		}
	}
	return nil
}

func decodeContent(raw json.RawMessage) (string, error) {
	var content string
	if err := json.Unmarshal(raw, &content); err == nil {
		return content, nil
	}
	return string(raw), nil
}

func normalizeZipPath(path string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(path))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("invalid skill file path: %s", path)
	}
	return clean, nil
}

func detectFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py", ".js", ".ts", ".sh":
		return "script"
	case ".md", ".txt":
		return "reference"
	case ".yaml", ".yml", ".json", ".toml":
		return "config"
	default:
		return "asset"
	}
}

func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	default:
		return "application/octet-stream"
	}
}
