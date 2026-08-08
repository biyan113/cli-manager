package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"cli-manager/internal/config"
	gh "cli-manager/internal/github"
)

// OpType 表示安装操作的种类,用于前端文案与事件标记。
type OpType string

const (
	OpInstall   OpType = "install"
	OpUpdate    OpType = "update"
	OpDowngrade OpType = "downgrade"
)

// ProgressHandler 接收进度事件。Wails 绑定层会把它们转发为 EventsEmit。
type ProgressHandler func(toolID string, op OpType, downloaded, total int64, phase string)

// StatusHandler 接收操作终态事件。
type StatusHandler func(toolID string, op OpType, status string, version, message string)

// Manager 是 CLI 工具安装/更新的编排层。
type Manager struct {
	Config  *config.Config
	State   *config.State
	Github  *gh.Client
	Ctx     context.Context
	OnProgress ProgressHandler
	OnStatus   StatusHandler

	// installMu 只保护「rename 安装」这一小段,下载不占锁。
	installMu sync.Mutex

	// latestCache 缓存 latest 查询结果,TTL 5 分钟,降低 GitHub 限流消耗。
	latestCache sync.Map // key: "owner/repo" → *cacheEntry
}

type cacheEntry struct {
	release *gh.Release
	at      time.Time
}

// ToolStatus 是前端展示的完整工具状态。
type ToolStatus struct {
	Spec             config.ToolSpec `json:"spec"`
	Installed        bool            `json:"installed"`
	InstalledVersion string          `json:"installed_version"`
	InstalledFrom    string          `json:"installed_from"` // "detected" | "recorded"
	LatestVersion    string          `json:"latest_version"`
	UpdateAvailable  bool            `json:"update_available"`
	Error            string          `json:"error"`
}

// InstallResult 是一次安装/更新操作的结果。
type InstallResult struct {
	ToolID   string `json:"tool_id"`
	Version  string `json:"version"`
	BinPath  string `json:"bin_path"`
	Operation string `json:"operation"`
}

// NewManager 构造 Manager。
func NewManager(cfg *config.Config, st *config.State, client *gh.Client) *Manager {
	if st.Versions == nil {
		st.Versions = map[string]config.InstalledInfo{}
	}
	return &Manager{
		Config: cfg,
		State:  st,
		Github: client,
	}
}

// ---------- 版本探测 ----------

// DetectInstalled 探测某个工具当前是否已安装及版本。
// 优先真实执行二进制探测;失败回退 state 记录。
func (m *Manager) DetectInstalled(spec config.ToolSpec) (version string, from string, ok bool) {
	binPath := m.binPath(spec)
	if binPath == "" {
		return "", "", false
	}
	if _, err := os.Stat(binPath); err != nil {
		return "", "", false
	}

	if v, found := m.detectViaCommand(spec, binPath); found {
		return v, "detected", true
	}
	if info, found := m.State.Versions[spec.ID]; found {
		return info.Version, "recorded", true
	}
	return "", "", false
}

// detectViaCommand 执行 bin --version 并用 regex 提取版本。
func (m *Manager) detectViaCommand(spec config.ToolSpec, binPath string) (string, bool) {
	args := spec.VersionCmd
	if len(args) == 0 {
		args = []string{"--version"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", false
	}
	re := spec.VersionRegex
	if re == "" {
		// 默认匹配输出里的第一个语义化版本。
		re = `[0-9]+\.[0-9]+\.[0-9]+`
	}
	expr, err := regexp.Compile(re)
	if err != nil {
		return "", false
	}
	match := expr.FindStringSubmatch(string(out))
	if len(match) >= 2 && match[1] != "" {
		return match[1], true
	}
	if len(match) == 1 && match[0] != "" {
		return match[0], true
	}
	return "", false
}

// binPath 返回工具二进制的安装路径。
func (m *Manager) binPath(spec config.ToolSpec) string {
	dir := m.installDir(spec)
	if dir == "" {
		return ""
	}
	bin := spec.Binary
	if bin == "" {
		bin = spec.ID
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(bin, ".exe") {
		bin += ".exe"
	}
	return filepath.Join(dir, bin)
}

// installDir 解析安装目录(spec.InstallDir 覆盖全局)。
func (m *Manager) installDir(spec config.ToolSpec) string {
	dir := spec.InstallDir
	if dir == "" {
		dir = m.Config.InstallDir
	}
	return config.ExpandPath(dir)
}

// ---------- latest 查询(带 TTL 缓存) ----------

// GetLatestVersion 返回工具的最新版本号(tag),带 5 分钟缓存。
func (m *Manager) GetLatestVersion(ctx context.Context, spec config.ToolSpec) (string, error) {
	rel, err := m.getLatestReleaseCached(ctx, spec)
	if err != nil {
		return "", err
	}
	return rel.TagName, nil
}

// getLatestReleaseCached 取最新 release,命中缓存则直接返回。
func (m *Manager) getLatestReleaseCached(ctx context.Context, spec config.ToolSpec) (*gh.Release, error) {
	key := spec.Repo
	if v, ok := m.latestCache.Load(key); ok {
		entry := v.(*cacheEntry)
		if time.Since(entry.at) < 5*time.Minute {
			return entry.release, nil
		}
	}
	owner, repo := splitRepo(spec.Repo)
	rel, err := m.Github.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	m.latestCache.Store(key, &cacheEntry{release: rel, at: time.Now()})
	return rel, nil
}

// GetAvailableVersions 列出可用于降级的版本 tag(取最近 30 个)。
func (m *Manager) GetAvailableVersions(ctx context.Context, spec config.ToolSpec) ([]string, error) {
	owner, repo := splitRepo(spec.Repo)
	releases, err := m.Github.ListReleases(ctx, owner, repo, 1, 30)
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		tags = append(tags, r.TagName)
	}
	return tags, nil
}

// splitRepo 把 "owner/repo" 拆开。
func splitRepo(spec string) (string, string) {
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 {
		return spec, ""
	}
	return parts[0], parts[1]
}

// ---------- 状态组装 ----------

// BuildStatus 组装一个工具的完整状态(含 latest 查询)。
func (m *Manager) BuildStatus(ctx context.Context, spec config.ToolSpec) (ToolStatus, error) {
	st := ToolStatus{Spec: spec}

	installedVersion, from, installed := m.DetectInstalled(spec)
	st.Installed = installed
	st.InstalledVersion = installedVersion
	st.InstalledFrom = from

	latest, err := m.GetLatestVersion(ctx, spec)
	if err != nil {
		st.Error = err.Error()
		// 即使 latest 查询失败,仍返回已探测的部分状态。
		return st, nil
	}
	st.LatestVersion = latest
	// tag 带 v、探测出的版本不带 v,须按语义比较,而非字符串相等。
	st.UpdateAvailable = installed && latest != "" && installedVersion != "" && CompareVersions(latest, installedVersion) > 0
	return st, nil
}

// emitProgress 转发进度事件(由绑定层接住)。
func (m *Manager) emitProgress(toolID string, op OpType, downloaded, total int64, phase string) {
	if m.OnProgress != nil {
		m.OnProgress(toolID, op, downloaded, total, phase)
	}
}

// emitStatus 转发终态事件。
func (m *Manager) emitStatus(toolID string, op OpType, status, version, message string) {
	if m.OnStatus != nil {
		m.OnStatus(toolID, op, status, version, message)
	}
}

// ---------- 高层操作 ----------

// Install 安装一个工具到最新版本。
func (m *Manager) Install(ctx context.Context, spec config.ToolSpec) (InstallResult, error) {
	return m.installVersion(ctx, spec, "", OpInstall)
}

// Update 更新一个工具到最新版本(与 Install 等价,操作标记不同)。
func (m *Manager) Update(ctx context.Context, spec config.ToolSpec) (InstallResult, error) {
	return m.installVersion(ctx, spec, "", OpUpdate)
}

// Downgrade 降级到指定版本(tag)。
func (m *Manager) Downgrade(ctx context.Context, spec config.ToolSpec, version string) (InstallResult, error) {
	if version == "" {
		return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("降级需要指定版本号"))
	}
	return m.installVersion(ctx, spec, version, OpDowngrade)
}

// Uninstall 卸载一个工具:删除二进制并清 state。
// 只移除注册;卸载失败但文件已删时仍算成功。
func (m *Manager) Uninstall(spec config.ToolSpec) error {
	target := m.binPath(spec)
	if target != "" {
		if _, err := os.Stat(target); err == nil {
			if err := os.Remove(target); err != nil {
				return wrapErr(spec.ID, fmt.Errorf("删除二进制: %w", err))
			}
		}
	}
	delete(m.State.Versions, spec.ID)
	return m.State.Save()
}

// wrapErr 让错误带工具上下文。
func wrapErr(toolID string, err error) error {
	return fmt.Errorf("[%s] %w", toolID, err)
}
