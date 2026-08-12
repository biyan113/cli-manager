package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/biyan113/cli-box/internal/config"
	"github.com/biyan113/cli-box/internal/deepseek"
	gh "github.com/biyan113/cli-box/internal/github"
	"github.com/biyan113/cli-box/internal/tool"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 是 Wails 绑定层,所有导出方法都会自动生成前端绑定。
type App struct {
	ctx     context.Context
	manager *tool.Manager
	ds      *deepseek.Client
}

// NewApp 创建 App 实例。
func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("配置加载失败:", err)
		cfg = &config.Config{InstallDir: config.DefaultInstallDir(), Tools: config.DefaultTools()}
	}
	st, err := config.LoadState()
	if err != nil {
		fmt.Println("状态加载失败:", err)
		st = &config.State{Versions: map[string]config.InstalledInfo{}}
	}
	client := gh.NewClient(cfg.GithubToken)
	ds := deepseek.NewClient(cfg.DeepSeekToken, cfg.DeepSeekModel)

	m := tool.NewManager(cfg, st, client)
	return &App{manager: m, ds: ds}
}

// startup 在应用启动时调用,保存 context 并挂接事件转发。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 把 Manager 的进度/状态回调转发为 Wails 事件
	a.manager.OnProgress = func(toolID string, op tool.OpType, downloaded, total int64, phase string) {
		var pct float64
		if total > 0 {
			pct = float64(downloaded) / float64(total) * 100
		}
		runtime.EventsEmit(a.ctx, "tool:progress", map[string]any{
			"tool_id":    toolID,
			"op":         string(op),
			"downloaded": downloaded,
			"total":      total,
			"percent":    pct,
			"phase":      phase,
		})
	}
	a.manager.OnStatus = func(toolID string, op tool.OpType, status, version, message string) {
		runtime.EventsEmit(a.ctx, "tool:status", map[string]any{
			"tool_id": toolID,
			"op":      string(op),
			"status":  status,
			"version": version,
			"message": message,
		})
	}
}

// domReady centers the compact initial window after the native WebView exists.
// This avoids oversized or partially off-screen placement on high-DPI Windows displays.
func (a *App) domReady(ctx context.Context) {
	runtime.WindowCenter(ctx)
}

// log 发送一条日志事件到前端。
func (a *App) log(level, message string) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "app:log", map[string]any{
		"level":   level,
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// ---------- 对外绑定方法 ----------

// ListTools 返回所有工具的状态。
func (a *App) ListTools() []tool.ToolStatus {
	tools := a.manager.Config.Tools
	results := make([]tool.ToolStatus, 0, len(tools))
	for _, spec := range tools {
		st, err := a.manager.BuildStatus(a.ctx, spec)
		if err != nil {
			st.Error = err.Error()
		}
		results = append(results, st)
	}
	return results
}

// AddTool 添加一个工具到注册表。
func (a *App) AddTool(spec config.ToolSpec) error {
	if err := a.manager.Config.AddTool(spec); err != nil {
		return err
	}
	if err := a.manager.Config.Save(); err != nil {
		return err
	}
	a.log("info", fmt.Sprintf("已添加工具 %s (%s)", spec.ID, spec.Repo))
	return nil
}

// RemoveTool 从注册表移除一个工具(不影响已装二进制)。
func (a *App) RemoveTool(id string) error {
	if err := a.manager.Config.RemoveTool(id); err != nil {
		return err
	}
	return a.manager.Config.Save()
}

// RefreshTool 重新查询一个工具的状态(latest + 已装)。
func (a *App) RefreshTool(id string) (tool.ToolStatus, error) {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return tool.ToolStatus{}, err
	}
	return a.manager.BuildStatus(a.ctx, *spec)
}

// RefreshAll 重新查询所有工具的状态。
func (a *App) RefreshAll() []tool.ToolStatus {
	return a.ListTools()
}

// InstallTool 安装一个工具到最新版。
func (a *App) InstallTool(id string) (tool.InstallResult, error) {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return tool.InstallResult{}, err
	}
	return a.manager.Install(a.ctx, *spec)
}

// UpdateTool 更新一个工具到最新版。
func (a *App) UpdateTool(id string) (tool.InstallResult, error) {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return tool.InstallResult{}, err
	}
	return a.manager.Update(a.ctx, *spec)
}

// DowngradeTool 降级到指定版本。
func (a *App) DowngradeTool(id, version string) (tool.InstallResult, error) {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return tool.InstallResult{}, err
	}
	return a.manager.Downgrade(a.ctx, *spec, version)
}

// UninstallTool 卸载一个工具。
func (a *App) UninstallTool(id string) error {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return err
	}
	if err := a.manager.Uninstall(*spec); err != nil {
		return err
	}
	a.log("info", fmt.Sprintf("已卸载工具 %s", id))
	return nil
}

// GetAvailableVersions 列出可用于降级的版本 tag。
func (a *App) GetAvailableVersions(id string) ([]string, error) {
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return nil, err
	}
	return a.manager.GetAvailableVersions(a.ctx, *spec)
}

// SetToken 设置 GitHub token。
func (a *App) SetToken(token string) error {
	if err := a.manager.Config.SetToken(token); err != nil {
		return err
	}
	a.manager.Github.SetToken(token)
	a.log("info", "GitHub token 已更新")
	return nil
}

// SetDeepSeekToken 设置 DeepSeek API key。
func (a *App) SetDeepSeekToken(token string) error {
	if err := a.manager.Config.SetDeepSeekToken(token); err != nil {
		return err
	}
	a.ds.SetToken(token)
	a.log("info", "DeepSeek key 已更新")
	return nil
}

// GetConfig 返回当前配置(前端展示;token 不回传明文,只给是否已设置)。
func (a *App) GetConfig() map[string]any {
	cfg := a.manager.Config
	hasToken := cfg.GithubToken != ""
	return map[string]any{
		"install_dir":        cfg.InstallDir,
		"has_token":          hasToken,
		"has_deepseek_token": cfg.DeepSeekToken != "",
		"deepseek_model":     cfg.DeepSeekModel,
		"language":           cfg.Language,
		"tool_count":         len(cfg.Tools),
	}
}

// SetLanguage persists the UI language preference.
func (a *App) SetLanguage(language string) error {
	return a.manager.Config.SetLanguage(language)
}

// SetDeepSeekModel 设置 DeepSeek 模型并落盘。
func (a *App) SetDeepSeekModel(model string) error {
	if err := a.manager.Config.SetDeepSeekModel(model); err != nil {
		return err
	}
	a.ds.SetModel(model)
	// 模型变更后,旧缓存不再适用。
	explainMu.Lock()
	explainCache = map[string]explainEntry{}
	explainMu.Unlock()
	a.log("info", fmt.Sprintf("DeepSeek 模型已更新为 %s", a.manager.Config.DeepSeekModel))
	return nil
}

// ---------- 工具说明(DeepSeek)----------

var (
	explainMu    sync.Mutex
	explainCache = map[string]explainEntry{}
)

type explainEntry struct {
	summary   string
	summaryEN string
	releases  []ReleaseNote
	at        time.Time
}

// ReleaseNote 是某个版本的更新说明。
type ReleaseNote struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
}

// ToolExplanation 是"工具说明"的完整结果:模型生成的中英双语简介 + 最近版本更新说明。
type ToolExplanation struct {
	Summary   string        `json:"summary"`              // 简体中文
	SummaryEN string        `json:"summary_en,omitempty"` // 英文
	Releases  []ReleaseNote `json:"releases"`
}

// GetToolExplanation 生成某个工具的中英双语说明。
// 流程:拉取仓库简介 + README 最新内容 → 交给 DeepSeek 生成双语简介;
// 同时拉取最近几个版本的更新说明。结果按工具缓存 1 小时,避免重复调用。
func (a *App) GetToolExplanation(id string) (ToolExplanation, error) {
	if !a.ds.HasToken() {
		return ToolExplanation{}, errors.New("未配置 DeepSeek API key,请在设置中填写")
	}
	spec, err := a.manager.Config.FindTool(id)
	if err != nil {
		return ToolExplanation{}, err
	}

	explainMu.Lock()
	if e, ok := explainCache[id]; ok && time.Since(e.at) < time.Hour {
		explainMu.Unlock()
		return ToolExplanation{Summary: e.summary, SummaryEN: e.summaryEN, Releases: e.releases}, nil
	}
	explainMu.Unlock()

	owner, repo := splitRepo(spec.Repo)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var description, readme string
	if info, err := a.manager.Github.GetRepoInfo(ctx, owner, repo); err == nil {
		description = info.Description
	}
	if text, err := a.manager.Github.GetReadme(ctx, owner, repo); err == nil {
		readme = text
	}
	// 即使仓库信息拉取失败(如仓库无 README),仍让模型基于已知信息说明。
	summary, summaryEN, err := a.ds.ExplainTool(ctx, spec.Name, spec.Repo, description, readme)
	if err != nil {
		return ToolExplanation{}, err
	}

	// 最近版本更新说明(失败不致命,保持为空)。
	releases := a.recentReleaseNotes(ctx, owner, repo)

	explainMu.Lock()
	explainCache[id] = explainEntry{summary: summary, summaryEN: summaryEN, releases: releases, at: time.Now()}
	explainMu.Unlock()
	return ToolExplanation{Summary: summary, SummaryEN: summaryEN, Releases: releases}, nil
}

// recentReleaseNotes 拉取最近几个正式版本的更新说明。
func (a *App) recentReleaseNotes(ctx context.Context, owner, repo string) []ReleaseNote {
	rs, err := a.manager.Github.ListReleases(ctx, owner, repo, 1, 5)
	if err != nil {
		return nil
	}
	notes := make([]ReleaseNote, 0, len(rs))
	for _, r := range rs {
		if r.Draft {
			continue
		}
		notes = append(notes, ReleaseNote{
			TagName:     r.TagName,
			Name:        r.Name,
			Body:        truncateRunes(r.Body, 4000),
			PublishedAt: r.PublishedAt,
		})
	}
	return notes
}

// truncateRunes 按 rune 截断到 max 字符,避免切坏 UTF-8。
func truncateRunes(s string, max int) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "\n…(截断)"
}

// splitRepo 把 "owner/repo" 拆成两部分。
func splitRepo(repo string) (string, string) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return repo, ""
}

// ---------- 快速添加(GitHub 地址解析)----------

// ParseGithubRepo 从 GitHub 地址(URL 或 owner/repo)解析出仓库信息,
// 并尝试从最新 release 资产推断建议字段,供添加工具表单预填。
// 仓库不可达时不报错,只返回基础建议(owner/repo 正确即可加入,后续可微调)。
func (a *App) ParseGithubRepo(input string) (map[string]any, error) {
	owner, repo := parseRepoInput(input)
	if owner == "" || repo == "" {
		return nil, errors.New("无法识别的 GitHub 地址,请粘贴 https://github.com/owner/repo 或 owner/repo")
	}
	suggest := map[string]any{
		"owner":             owner,
		"repo":              owner + "/" + repo,
		"id":                repo,
		"name":              repo,
		"binary":            repo,
		"asset_pattern":     "{name}_{version}_{os}_{arch}",
		"checksums_pattern": "{name}_{version}_checksums.txt",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	release, err := a.manager.Github.GetLatestRelease(ctx, owner, repo)
	if err != nil || len(release.Assets) == 0 {
		return suggest, nil
	}

	// 选一个可执行/归档资产(排除校验和、签名文件)做命名推断。
	var pick gh.Asset
	for _, as := range release.Assets {
		n := strings.ToLower(as.Name)
		if strings.Contains(n, "checksum") || strings.Contains(n, ".sig") || strings.Contains(n, ".asc") || strings.Contains(n, ".sha") {
			continue
		}
		pick = as
		break
	}
	if pick.Name == "" {
		return suggest, nil
	}

	if binary := inferBinary(release.Assets); binary != "" {
		suggest["binary"] = binary
		suggest["id"] = binary
		suggest["name"] = binary
		if p := inferAssetPattern(pick.Name, binary); p != "" {
			suggest["asset_pattern"] = p
		}
	}
	return suggest, nil
}

// parseRepoInput 从输入提取 owner/repo,支持
// "https://github.com/owner/repo"、"github.com/owner/repo"、"owner/repo"。
func parseRepoInput(input string) (string, string) {
	s := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(input), "/"))
	if s == "" {
		return "", ""
	}
	// 提取 github.com 之后的部分(若包含域名)
	if i := strings.Index(s, "github.com/"); i >= 0 {
		s = s[i+len("github.com/"):]
	} else if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		// 其他域名:取路径前两段
		parts := strings.SplitN(s, "/", 4)
		if len(parts) >= 3 {
			s = parts[2] + "/" + parts[3]
		} else {
			return "", ""
		}
	}
	s = strings.TrimSuffix(s, ".git")
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1]
	}
	return "", ""
}

// inferBinary 从资产名集合推断二进制名:去掉扩展名与校验和/签名文件,
// 先抹掉版本段再求公共前缀,去掉尾部连接符。
// 如 asc_3.6.0_macOS_arm64、asc_3.6.0_linux_arm64 → asc;
// gh_2.97.0_macOS_arm64、gh_2.97.0_linux_amd64 → gh(而非 gh_2.97.0)。
func inferBinary(assets []gh.Asset) string {
	var names []string
	for _, a := range assets {
		n := strings.ToLower(a.Name)
		if strings.Contains(n, "checksum") || strings.Contains(n, ".sig") || strings.Contains(n, ".asc") || strings.Contains(n, ".sha") {
			continue
		}
		names = append(names, stripExt(a.Name))
	}
	if len(names) == 0 {
		return ""
	}
	// 版本段统一抹掉再求公共前缀,避免版本号混入二进制名
	versionRe := regexp.MustCompile(`(?i)(v?\d+\.\d+(?:\.\d+)*(?:-[0-9a-zA-Z.]+)?)`)
	normalized := make([]string, len(names))
	for i, n := range names {
		normalized[i] = versionRe.ReplaceAllString(n, "")
	}
	prefix := normalized[0]
	for _, n := range normalized[1:] {
		for !strings.HasPrefix(n, prefix) {
			if len(prefix) <= 1 {
				prefix = ""
				break
			}
			prefix = prefix[:len(prefix)-1]
		}
		if prefix == "" {
			break
		}
	}
	// 清理版本位被移除后残留的重复分隔符与首尾连接符
	prefix = strings.Trim(prefix, "_.- ")
	prefix = strings.ReplaceAll(prefix, "__", "_")
	return strings.Trim(prefix, "_.- ")
}

// stripExt 去掉常见压缩/可执行扩展名。
func stripExt(name string) string {
	for _, ext := range []string{".tar.gz", ".tar.bz2", ".tar.xz", ".tgz", ".zip", ".tar", ".gz"} {
		if strings.HasSuffix(name, ext) {
			return strings.TrimSuffix(name, ext)
		}
	}
	if i := strings.LastIndex(name, "."); i > 0 {
		return name[:i]
	}
	return name
}

// inferAssetPattern 从单个资产名推断占位符模板。
// 例:asc_3.6.0_macOS_arm64 → {name}_{version}_{os}_{arch}。
func inferAssetPattern(assetName, binary string) string {
	if binary == "" {
		return "{name}_{version}_{os}_{arch}"
	}
	n := stripExt(assetName)
	if strings.HasPrefix(n, binary) {
		n = strings.TrimPrefix(n, binary)
	}
	if n == "" {
		return "{name}"
	}
	// 版本段(如 3.6.0、v1.2.3、2.1.0-beta)替换为 {version}
	versionRe := regexp.MustCompile(`(?i)(v?\d+\.\d+(?:\.\d+)*)(?:-[0-9a-zA-Z.]+)?`)
	n = versionRe.ReplaceAllString(n, "{version}")
	// 平台词替换:先 OS 后 ARCH,避免相互吞并
	for _, kv := range []struct{ val, ph string }{
		{"Windows", "{os}"}, {"macOS", "{os}"}, {"darwin", "{os}"}, {"linux", "{os}"},
	} {
		n = strings.ReplaceAll(n, kv.val, kv.ph)
	}
	for _, kv := range []struct{ val, ph string }{
		{"x86_64", "{arch}"}, {"amd64", "{arch}"}, {"arm64", "{arch}"},
		{"i386", "{arch}"}, {"386", "{arch}"},
	} {
		n = strings.ReplaceAll(n, kv.val, kv.ph)
	}
	n = strings.ReplaceAll(n, "{os}_{os}", "{os}")
	n = strings.ReplaceAll(n, "{arch}_{arch}", "{arch}")
	return "{name}" + n
}
