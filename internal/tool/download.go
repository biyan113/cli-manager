package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/biyan113/cli-box/internal/config"
	gh "github.com/biyan113/cli-box/internal/github"
)

// progressReader 包裹一个 ReadCloser,在 Read 时累计字节数并回调进度。
type progressReader struct {
	rc         io.ReadCloser
	downloaded int64
	onProgress func(downloaded int64)
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.rc.Read(buf)
	if n > 0 {
		d := atomic.AddInt64(&p.downloaded, int64(n))
		p.onProgress(d)
	}
	return n, err
}

func (p *progressReader) Close() error { return p.rc.Close() }

// installVersion 是安装/更新/降级的核心流程。
// version 为空表示安装 latest;否则安装指定版本(降级)。
func (m *Manager) installVersion(ctx context.Context, spec config.ToolSpec, version string, op OpType) (InstallResult, error) {
	// 阶段 1:解析 release
	m.emitProgress(spec.ID, op, 0, 0, "resolve")
	var release *gh.Release
	var err error
	owner, repo := splitRepo(spec.Repo)
	if version == "" {
		release, err = m.getLatestReleaseCached(ctx, spec)
	} else {
		release, err = m.Github.GetReleaseByTag(ctx, owner, repo, version)
	}
	if err != nil {
		m.emitStatus(spec.ID, op, "error", "", err.Error())
		return InstallResult{}, wrapErr(spec.ID, err)
	}

	// 版本归一化(v 前缀等)
	normVersion := NormalizeVersion(release.TagName, spec.VersionTransform)

	// 阶段 2:定位 asset
	m.emitProgress(spec.ID, op, 0, 0, "resolve")
	pm := mergePlatformMap(spec.PlatformMap)
	// tag 与资产命名可能存在 v 前缀差异(如 gh:tag=v2.97.0,资产 gh_2.97.0_macOS_arm64),
	// 依次尝试候选版本生成资产名;命中后统一用该版本做后续校验与状态记录。
	var foundName string
	effectiveVersion := normVersion
	for _, cand := range candidateVersions(release.TagName, spec.VersionTransform) {
		name := BuildAssetName(spec.AssetPattern, spec.Binary, cand, spec.OS, spec.Arch, pm)
		if n, ok := FindAsset(toAssetLikes(release.Assets), name); ok {
			foundName = n
			effectiveVersion = cand
			break
		}
	}
	if foundName == "" {
		assetName := BuildAssetName(spec.AssetPattern, spec.Binary, normVersion, spec.OS, spec.Arch, pm)
		msg := fmt.Sprintf("在 %s %s 中找不到资产 %q(当前平台 %s/%s)", spec.Repo, release.TagName, assetName, platformFor(spec.OS, spec.Arch, "goos"), platformFor(spec.OS, spec.Arch, "goarch"))
		m.emitStatus(spec.ID, op, "error", "", msg)
		return InstallResult{}, fmt.Errorf("%s", msg)
	}

	// 定位下载 URL
	var downloadURL string
	var assetSize int64
	for _, a := range release.Assets {
		if a.Name == foundName {
			downloadURL = a.BrowserDownloadURL
			assetSize = a.Size
			break
		}
	}
	if downloadURL == "" {
		return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("资产 %q 缺少下载 URL", foundName))
	}

	// 阶段 3:拉 checksums 并取期望 hash
	m.emitProgress(spec.ID, op, 0, 0, "checksum")
	wantHash, err := m.fetchChecksum(ctx, spec, release, effectiveVersion, foundName)
	if err != nil {
		m.emitStatus(spec.ID, op, "error", "", err.Error())
		return InstallResult{}, wrapErr(spec.ID, err)
	}

	// 阶段 4:下载到同目录临时文件
	dir := m.installDir(spec)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("创建安装目录: %w", err))
	}
	target := m.binPath(spec)

	m.emitProgress(spec.ID, op, 0, assetSize, "download")
	resp, err := m.Github.OpenDownload(ctx, downloadURL)
	if err != nil {
		m.emitStatus(spec.ID, op, "error", "", err.Error())
		return InstallResult{}, wrapErr(spec.ID, err)
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp(dir, "."+spec.Binary+".tmp-*")
	if err != nil {
		return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("创建临时文件: %w", err))
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename 成功后是 no-op;失败时清理残留

	total := resp.ContentLength
	if assetSize > total {
		total = assetSize
	}

	// 进度节流:≥256KB 或 ≥100ms
	var lastEmit atomic.Int64
	lastEmit.Store(time.Now().UnixMilli())
	pr := &progressReader{rc: resp.Body, onProgress: func(d int64) {
		now := time.Now().UnixMilli()
		if d-lastEmit.Load() >= 256<<10 || now-lastEmit.Load() >= 100 {
			lastEmit.Store(now)
			m.emitProgress(spec.ID, op, d, total, "download")
		}
	}}
	if _, err := io.Copy(tmp, pr); err != nil {
		return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("下载失败: %w", err))
	}
	if err := tmp.Close(); err != nil {
		return InstallResult{}, wrapErr(spec.ID, err)
	}
	m.emitProgress(spec.ID, op, total, total, "download")

	// 阶段 5:校验 SHA256
	m.emitProgress(spec.ID, op, 0, 0, "checksum")
	if err := verifySHA256(tmpName, wantHash); err != nil {
		return InstallResult{}, wrapErr(spec.ID, err)
	}

	// 阶段 6:若为 zip/tar.gz 压缩包,解出真正的二进制
	binTmp, err := extractBinary(tmpName, spec.Binary)
	if err != nil {
		m.emitStatus(spec.ID, op, "error", "", err.Error())
		return InstallResult{}, wrapErr(spec.ID, err)
	}
	if binTmp != tmpName {
		defer os.Remove(binTmp) // 解压出的临时文件,安装完成后清理
	}

	// 阶段 7:设权限 + 原子替换(独占锁保护)
	m.installMu.Lock()
	defer m.installMu.Unlock()
	m.emitProgress(spec.ID, op, 0, 0, "install")

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binTmp, 0o755); err != nil {
			return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("设置权限: %w", err))
		}
	}
	if err := os.Rename(binTmp, target); err != nil {
		// Windows 下 rename 覆盖已存在文件可能失败,退化为删旧再 rename(非原子)。
		if runtime.GOOS == "windows" {
			if rmErr := os.Remove(target); rmErr == nil || os.IsNotExist(rmErr) {
				err = os.Rename(binTmp, target)
			}
		}
		if err != nil {
			return InstallResult{}, wrapErr(spec.ID, fmt.Errorf("原子替换失败: %w", err))
		}
	}

	// 阶段 8:写 state
	m.State.Versions[spec.ID] = config.InstalledInfo{
		Version: effectiveVersion,
		BinPath: target,
		At:      time.Now().Format(time.RFC3339),
	}
	if err := m.State.Save(); err != nil {
		// state 写失败不阻断安装成功,仅告警。
		m.emitStatus(spec.ID, op, "done", effectiveVersion, "已安装,但状态记录失败: "+err.Error())
	} else {
		m.emitStatus(spec.ID, op, "done", effectiveVersion, "")
	}

	return InstallResult{
		ToolID:    spec.ID,
		Version:   effectiveVersion,
		BinPath:   target,
		Operation: string(op),
	}, nil
}

// fetchChecksum 拉取 checksums 文件并返回目标 asset 的期望 sha256。
func (m *Manager) fetchChecksum(ctx context.Context, spec config.ToolSpec, release *gh.Release, version, assetName string) (string, error) {
	pm := mergePlatformMap(spec.PlatformMap)
	sumName := BuildChecksumName(spec.ChecksumsPattern, spec.Binary, version, spec.OS, spec.Arch, pm)
	if sumName == "" {
		return "", fmt.Errorf("未配置 checksums 模板")
	}

	var sumURL string
	for _, a := range release.Assets {
		if a.Name == sumName {
			sumURL = a.BrowserDownloadURL
			break
		}
	}
	if sumURL == "" {
		return "", fmt.Errorf("找不到校验文件 %q", sumName)
	}

	data, err := m.Github.GetBody(ctx, sumURL)
	if err != nil {
		return "", fmt.Errorf("下载校验文件: %w", err)
	}
	sums, err := parseChecksums(data)
	if err != nil {
		return "", err
	}
	want, ok := sums[assetName]
	if !ok {
		return "", fmt.Errorf("校验文件中没有资产 %q 的条目", assetName)
	}
	return want, nil
}

// mergePlatformMap 合并工具级覆盖到默认平台映射。
func mergePlatformMap(overrides map[string]string) PlatformMap {
	m := PlatformMap(config.DefaultPlatformMap())
	for k, v := range overrides {
		m[k] = v
	}
	return m
}

// toAssetLikes 把 github.Asset 转成 AssetLike。
func toAssetLikes(assets []gh.Asset) []AssetLike {
	out := make([]AssetLike, 0, len(assets))
	for i := range assets {
		a := assets[i]
		out = append(out, assetAdapter{a})
	}
	return out
}

type assetAdapter struct{ gh.Asset }

func (a assetAdapter) Name() string { return a.Asset.Name }

// platformFor 返回平台描述(用于错误信息)。
func platformFor(specOS, specArch, which string) string {
	if which == "goos" {
		if specOS != "" {
			return specOS
		}
		return runtime.GOOS
	}
	if specArch != "" {
		return specArch
	}
	return runtime.GOARCH
}
