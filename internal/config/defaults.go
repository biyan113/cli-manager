package config

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
	return "~/.local/bin"
}

// DefaultTools 返回内置的工具清单。
// 目前内置 asc(App Store Connect CLI,基于实测的 3.6.0 数据):
//   - repo: rorkai/App-Store-Connect-CLI
//   - GitHub tag 本身不带 v(如 3.6.0),asset 名 asc_3.6.0_macOS_arm64
//   - asc --version 输出 "3.6.0 (commit: ...)",regex 锚定行首数字
func DefaultTools() []ToolSpec {
	return []ToolSpec{
		{
			ID:                "asc",
			Name:              "asc",
			Repo:              "rorkai/App-Store-Connect-CLI",
			Binary:            "asc",
			AssetPattern:      "{name}_{version}_{os}_{arch}",
			ChecksumsPattern:  "{name}_{version}_checksums.txt",
			PlatformMap:       nil,
			OS:                "",
			Arch:              "",
			VersionCmd:        []string{"--version"},
			VersionRegex:      `^([0-9]+\.[0-9]+\.[0-9]+)`,
			VersionTransform:  "",
			InstallDir:        "",
		},
	}
}
