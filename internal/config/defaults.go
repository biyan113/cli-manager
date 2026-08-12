package config

import "runtime"

const (
	CurrentSchemaVersion = 2
	DefaultLanguage      = "auto"
)

// DefaultPlatformMap 是全局默认的 GOOS/GOARCH → GitHub asset 命名映射。
// 工具条目里可通过 ToolSpec.PlatformMap 覆盖。
// 覆盖规则:spec.PlatformMap 的每个 key 覆盖全局同名 key。
func DefaultPlatformMap() map[string]string {
	return map[string]string{
		"darwin":  "macOS",
		"linux":   "linux",
		"windows": "Windows",
		"arm64":   "arm64",
		"amd64":   "x86_64",
	}
}

// DefaultInstallDir 返回默认安装目录(~/.local/bin)。
// 注意:这里返回的仍可能含 ~,由调用方用 ExpandPath 展开。
func DefaultInstallDir() string {
	if runtime.GOOS == "windows" {
		return "~/bin"
	}
	return "~/.local/bin"
}

// DefaultTools returns a curated list verified against official release assets.
func DefaultTools() []ToolSpec {
	return []ToolSpec{
		{
			ID:               "asc",
			Name:             "asc",
			Repo:             "rorkai/App-Store-Connect-CLI",
			Binary:           "asc",
			AssetPattern:     "{name}_{version}_{os}_{arch}",
			ChecksumsPattern: "{name}_{version}_checksums.txt",
			PlatformMap:      nil,
			OS:               "",
			Arch:             "",
			VersionCmd:       []string{"--version"},
			VersionRegex:     `^([0-9]+\.[0-9]+\.[0-9]+)`,
			VersionTransform: "",
			InstallDir:       "",
		},
		{
			ID: "gh", Name: "GitHub CLI", Repo: "cli/cli", Binary: "gh",
			AssetPattern: "{name}_{version}_{os}_{arch}", ChecksumsPattern: "{name}_{version}_checksums.txt",
			PlatformMap: map[string]string{"darwin": "macOS", "linux": "linux", "windows": "windows", "arm64": "arm64", "amd64": "amd64"},
			VersionCmd:  []string{"--version"}, VersionRegex: `gh version ([0-9]+\.[0-9]+\.[0-9]+)`,
		},
		{
			ID: "jq", Name: "jq", Repo: "jqlang/jq", Binary: "jq",
			AssetPattern: "{name}-{os}-{arch}", ChecksumsPattern: "sha256sum.txt",
			PlatformMap: map[string]string{"darwin": "macos", "linux": "linux", "windows": "windows", "arm64": "arm64", "amd64": "amd64"},
			VersionCmd:  []string{"--version"}, VersionRegex: `jq-([0-9]+\.[0-9]+(?:\.[0-9]+)?)`,
		},
		{
			ID: "yq", Name: "yq", Repo: "mikefarah/yq", Binary: "yq",
			AssetPattern: "{name}_{os}_{arch}", ChecksumsPattern: "checksums-bsd",
			PlatformMap: map[string]string{"darwin": "darwin", "linux": "linux", "windows": "windows", "arm64": "arm64", "amd64": "amd64"},
			VersionCmd:  []string{"--version"},
		},
		{
			ID: "fzf", Name: "fzf", Repo: "junegunn/fzf", Binary: "fzf",
			AssetPattern: "{name}-{version}-{os}_{arch}", ChecksumsPattern: "{name}_{version}_checksums.txt",
			PlatformMap: map[string]string{"darwin": "darwin", "linux": "linux", "windows": "windows", "arm64": "arm64", "amd64": "amd64"},
			VersionCmd:  []string{"--version"},
		},
	}
}
