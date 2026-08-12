// Package config 负责 cli-manager 的注册表/状态读写。
// 配置存放在 os.UserConfigDir()/cli-manager/(macOS 即
// ~/Library/Application Support/cli-manager/),首次启动自动创建默认配置。
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultDeepSeekModel 是"工具说明"功能使用的默认模型。
const DefaultDeepSeekModel = "deepseek-v4-flash"

// ErrNotFound 表示按 id 找不到工具。
var ErrNotFound = errors.New("tool not found")

// ToolSpec 描述一个可被管理的 CLI 工具。
type ToolSpec struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Repo             string            `json:"repo"`   // owner/repo,如 "rorkai/App-Store-Connect-CLI"
	Binary           string            `json:"binary"` // 安装后的可执行文件名
	AssetPattern     string            `json:"asset_pattern"`
	ChecksumsPattern string            `json:"checksums_pattern"`
	PlatformMap      map[string]string `json:"platform_map,omitempty"` // 覆盖默认平台映射
	OS               string            `json:"os,omitempty"`           // 覆盖 runtime.GOOS
	Arch             string            `json:"arch,omitempty"`         // 覆盖 runtime.GOARCH
	VersionCmd       []string          `json:"version_cmd,omitempty"`  // 探测版本命令,默认 ["--version"]
	VersionRegex     string            `json:"version_regex,omitempty"`
	VersionTransform string            `json:"version_transform,omitempty"` // "trim_v" 剥前导 v
	InstallDir       string            `json:"install_dir,omitempty"`       // 覆盖全局 install_dir
}

// Config 是 tools.json 的完整结构。
type Config struct {
	SchemaVersion int        `json:"schema_version"`
	InstallDir    string     `json:"install_dir"`
	GithubToken   string     `json:"github_token"`
	DeepSeekToken string     `json:"deepseek_token"` // DeepSeek API key,用于"工具说明"功能
	DeepSeekModel string     `json:"deepseek_model"` // DeepSeek 模型,用于"工具说明"功能
	Language      string     `json:"language"`       // "auto" | "zh-CN" | "en"
	Tools         []ToolSpec `json:"tools"`

	mu sync.Mutex `json:"-"`
}

// State 记录每个工具的安装元信息,作为版本探测的兜底。
type State struct {
	Versions map[string]InstalledInfo `json:"versions"`
}

// InstalledInfo 是某个工具已安装版本的信息。
type InstalledInfo struct {
	Version string `json:"version"`
	BinPath string `json:"bin_path"`
	At      string `json:"at"`
}

// ConfigPath 返回 tools.json 的完整路径。
func ConfigPath() string {
	return filepath.Join(configDir(), "tools.json")
}

// StatePath 返回 state.json 的完整路径。
func StatePath() string {
	return filepath.Join(configDir(), "state.json")
}

// configDir 返回配置目录,并确保其存在(0700)。
func configDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		// 极少数环境拿不到 UserConfigDir,退回 home。
		home, herr := os.UserHomeDir()
		if herr != nil {
			home = "."
		}
		base = home
	}
	dir := filepath.Join(base, "cli-manager")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		// MkdirAll 失败不致命,读写时再报错。
		return dir
	}
	return dir
}

// Load 读取配置;文件不存在时写入默认配置(含内置 asc)。
func Load() (*Config, error) {
	c := &Config{}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			c.SchemaVersion = CurrentSchemaVersion
			c.InstallDir = DefaultInstallDir()
			c.Language = DefaultLanguage
			c.Tools = DefaultTools()
			if err := c.Save(); err != nil {
				return nil, fmt.Errorf("写入默认配置: %w", err)
			}
			return c, nil
		}
		return nil, fmt.Errorf("读取配置: %w", err)
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("解析配置 %s: %w", ConfigPath(), err)
	}
	if c.normalize() {
		if err := c.Save(); err != nil {
			return nil, fmt.Errorf("迁移配置: %w", err)
		}
	}
	return c, nil
}

// normalize 补齐缺失字段，并按 schema version 执行一次性默认工具迁移。
// 返回配置是否发生变化。
func (c *Config) normalize() bool {
	changed := false
	if c.InstallDir == "" {
		c.InstallDir = DefaultInstallDir()
		changed = true
	}
	if c.DeepSeekModel == "" {
		c.DeepSeekModel = DefaultDeepSeekModel
		changed = true
	}
	if !validLanguage(c.Language) {
		c.Language = DefaultLanguage
		changed = true
	}
	if c.SchemaVersion < CurrentSchemaVersion {
		existing := make(map[string]bool, len(c.Tools))
		for _, spec := range c.Tools {
			existing[spec.ID] = true
		}
		for _, spec := range DefaultTools() {
			if !existing[spec.ID] {
				c.Tools = append(c.Tools, spec)
			}
		}
		c.SchemaVersion = CurrentSchemaVersion
		changed = true
	}
	for i := range c.Tools {
		t := &c.Tools[i]
		if t.VersionCmd == nil || len(t.VersionCmd) == 0 {
			t.VersionCmd = []string{"--version"}
			changed = true
		}
	}
	return changed
}

// Save 原子写配置:先写临时文件再 os.Rename 覆盖。
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return atomicWriteJSON(ConfigPath(), c)
}

// FindTool 按 id 查找工具,返回拷贝。
func (c *Config) FindTool(id string) (*ToolSpec, error) {
	for i := range c.Tools {
		if c.Tools[i].ID == id {
			cp := c.Tools[i]
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

// AddTool 追加一个工具;id 或 repo 为空时报错;id 重复时报错。
func (c *Config) AddTool(spec ToolSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	spec.ID = strings.TrimSpace(spec.ID)
	spec.Repo = strings.TrimSpace(spec.Repo)
	if spec.ID == "" || spec.Repo == "" {
		return errors.New("id 和 repo 不能为空")
	}
	if spec.Binary == "" {
		spec.Binary = spec.ID
	}
	if spec.AssetPattern == "" {
		spec.AssetPattern = "{name}_{version}_{os}_{arch}"
	}
	if spec.ChecksumsPattern == "" {
		spec.ChecksumsPattern = "{name}_{version}_checksums.txt"
	}
	if spec.VersionCmd == nil || len(spec.VersionCmd) == 0 {
		spec.VersionCmd = []string{"--version"}
	}
	for i := range c.Tools {
		if c.Tools[i].ID == spec.ID {
			return fmt.Errorf("工具 %q 已存在", spec.ID)
		}
	}
	c.Tools = append(c.Tools, spec)
	return nil
}

// RemoveTool 移除一个工具注册(不影响已安装的二进制)。
func (c *Config) RemoveTool(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Tools {
		if c.Tools[i].ID == id {
			c.Tools = append(c.Tools[:i], c.Tools[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// SetToken 更新 GitHub token 并落盘。
func (c *Config) SetToken(token string) error {
	c.GithubToken = strings.TrimSpace(token)
	return c.Save()
}

// SetDeepSeekToken 更新 DeepSeek API key 并落盘。
func (c *Config) SetDeepSeekToken(token string) error {
	c.DeepSeekToken = strings.TrimSpace(token)
	return c.Save()
}

// SetDeepSeekModel 更新 DeepSeek 模型并落盘;空值回退默认模型。
func (c *Config) SetDeepSeekModel(model string) error {
	c.DeepSeekModel = strings.TrimSpace(model)
	if c.DeepSeekModel == "" {
		c.DeepSeekModel = DefaultDeepSeekModel
	}
	return c.Save()
}

// SetLanguage updates the UI language preference.
func (c *Config) SetLanguage(language string) error {
	language = strings.TrimSpace(language)
	if !validLanguage(language) {
		return fmt.Errorf("不支持的语言 %q", language)
	}
	c.Language = language
	return c.Save()
}

func validLanguage(language string) bool {
	return language == "auto" || language == "zh-CN" || language == "en"
}

// LoadState 读取安装状态;不存在时返回空状态。
func LoadState() (*State, error) {
	s := &State{}
	data, err := os.ReadFile(StatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("读取状态: %w", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("解析状态 %s: %w", StatePath(), err)
	}
	if s.Versions == nil {
		s.Versions = map[string]InstalledInfo{}
	}
	return s, nil
}

// Save 原子写状态文件。
func (s *State) Save() error {
	if s.Versions == nil {
		s.Versions = map[string]InstalledInfo{}
	}
	return atomicWriteJSON(StatePath(), s)
}

// ExpandPath 展开路径中的 ~ 前缀。
func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if p == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}

// atomicWriteJSON 以 0600 权限原子写一个 JSON 文件。
func atomicWriteJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename 成功后是 no-op

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
