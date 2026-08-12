// Package tool 实现 CLI 工具的安装编排:资产命名、checksums 校验、流式下载与原子安装。
package tool

import (
	"runtime"
	"strings"
)

// PlatformMap 是 GOOS/GOARCH → GitHub asset 命名的映射。
type PlatformMap map[string]string

// resolvePlatform 计算工具实际使用的 GOOS/GOARCH。
// spec.OS / spec.Arch 非空时覆盖 runtime 值(如 Rosetta 下强制 x86_64)。
func resolvePlatform(specOS, specArch string) (string, string) {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	if specOS != "" {
		goos = specOS
	}
	if specArch != "" {
		goarch = specArch
	}
	return goos, goarch
}

// mapValue 用映射把 goos/goarch 转成 asset 命名写法;查不到则原样透传。
func mapValue(m PlatformMap, key string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return key
}

// BuildAssetName 用占位符模板生成目标 asset 名。
// 支持 {name} {version} {os} {arch};{os}/{arch} 先经 platformMap 映射。
// version 应为已归一化(无前导 v)的版本号。
func BuildAssetName(pattern, name, version, goos, goarch string, pm PlatformMap) string {
	os, arch := resolvePlatform(goos, goarch)
	if pm == nil {
		pm = PlatformMap{}
	}
	out := pattern
	out = strings.ReplaceAll(out, "{name}", name)
	out = strings.ReplaceAll(out, "{version}", version)
	out = strings.ReplaceAll(out, "{os}", mapValue(pm, os))
	out = strings.ReplaceAll(out, "{arch}", mapValue(pm, arch))
	return out
}

// BuildChecksumName 生成 checksums 文件名(同样支持占位符)。
func BuildChecksumName(pattern, name, version, goos, goarch string, pm PlatformMap) string {
	return BuildAssetName(pattern, name, version, goos, goarch, pm)
}

// NormalizeVersion 根据 transform 归一化 tag 版本号。
// 支持 "trim_v"(剥前导 v);空 transform 原样返回。
func NormalizeVersion(tag, transform string) string {
	if transform == "trim_v" {
		return strings.TrimPrefix(tag, "v")
	}
	return tag
}

// candidateVersions 返回资产匹配时依次尝试的版本候选:transform 归一化后的版本,
// 以及它的「去前导 v」/「补前导 v」变体。
// 用于 tag 与资产命名 v 前缀不一致的仓库,如 gh:tag=v2.97.0,资产却是 gh_2.97.0_macOS_arm64。
func candidateVersions(tag, transform string) []string {
	base := NormalizeVersion(tag, transform)
	seen := make(map[string]bool)
	var out []string
	add := func(v string) {
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}
	add(base)
	trimmed := strings.TrimPrefix(base, "v")
	add(trimmed)
	if !strings.HasPrefix(base, "v") {
		add("v" + base)
	}
	return out
}

// NormalizeAssetName 去掉常见平台后缀差异,用于大小写不敏感匹配。
// 例如 "asc_3.6.0_macos_arm64" 与 "asc_3.6.0_macOS_arm64" 归一化后相等。
func NormalizeAssetName(name string) string {
	return strings.ToLower(name)
}

// FindAsset locates a release asset. Extension-less templates only match formats
// the installer understands, preventing accidental selection of .deb/.rpm files.
func FindAsset(assets []AssetLike, want string) (found string, ok bool) {
	if assets == nil {
		return "", false
	}
	// 1. 精确匹配
	for _, a := range assets {
		if a.Name() == want {
			return a.Name(), true
		}
	}
	// 2. 归一化(小写)后相等
	low := NormalizeAssetName(want)
	for _, a := range assets {
		if NormalizeAssetName(a.Name()) == low {
			return a.Name(), true
		}
	}
	// 3. Match a supported executable/archive suffix.
	for _, suffix := range []string{".tar.gz", ".tgz", ".zip", ".exe"} {
		candidate := low + suffix
		for _, a := range assets {
			if NormalizeAssetName(a.Name()) == candidate {
				return a.Name(), true
			}
		}
	}
	return "", false
}

// AssetLike 抽象 asset 名,便于 FindAsset 与 github.Asset 解耦、便于单测。
type AssetLike interface {
	Name() string
}
