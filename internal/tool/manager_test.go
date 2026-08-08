package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"cli-manager/internal/config"
)

// mockAsset 实现 AssetLike,便于 FindAsset 单测。
type mockAsset struct{ name string }

func (m mockAsset) Name() string { return m.name }

// ---------- assetname ----------

func TestBuildAssetName(t *testing.T) {
	pm := config.DefaultPlatformMap()
	got := BuildAssetName("{name}_{version}_{os}_{arch}", "asc", "3.6.1", "darwin", "arm64", pm)
	want := "asc_3.6.1_macOS_arm64"
	if got != want {
		t.Errorf("BuildAssetName = %q, want %q", got, want)
	}
}

func TestBuildAssetNameLinuxAmd64(t *testing.T) {
	pm := config.DefaultPlatformMap()
	got := BuildAssetName("{name}_{version}_{os}_{arch}", "asc", "3.6.1", "linux", "amd64", pm)
	want := "asc_3.6.1_linux_x86_64"
	if got != want {
		t.Errorf("BuildAssetName = %q, want %q", got, want)
	}
}

func TestBuildAssetNameCustomMap(t *testing.T) {
	pm := map[string]string{"darwin": "macosx"} // 覆盖默认
	got := BuildAssetName("{os}_{arch}", "", "", "darwin", "arm64", pm)
	if got != "macosx_arm64" {
		t.Errorf("自定义映射 = %q, want macosx_arm64", got)
	}
}

func TestBuildAssetNameUnmappedPassthrough(t *testing.T) {
	pm := config.DefaultPlatformMap()
	// freebsd 不在映射里,原样透传
	got := BuildAssetName("{os}_{arch}", "", "", "freebsd", "riscv64", pm)
	if got != "freebsd_riscv64" {
		t.Errorf("未映射平台应透传, got %q", got)
	}
}

func TestNormalizeVersion(t *testing.T) {
	if got := NormalizeVersion("v3.6.1", "trim_v"); got != "3.6.1" {
		t.Errorf("trim_v = %q, want 3.6.1", got)
	}
	if got := NormalizeVersion("3.6.1", ""); got != "3.6.1" {
		t.Errorf("no transform = %q, want 3.6.1", got)
	}
	if got := NormalizeVersion("3.6.1", "trim_v"); got != "3.6.1" {
		t.Errorf("trim_v on no-v = %q, want 3.6.1", got)
	}
}

// TestCandidateVersions 覆盖 gh 类仓库:tag 带 v,但资产名版本不带 v。
func TestCandidateVersions(t *testing.T) {
	// tag=v2.97.0、无 transform → 先试 v2.97.0,再试去 v 的 2.97.0
	cands := candidateVersions("v2.97.0", "")
	if len(cands) != 2 || cands[0] != "v2.97.0" || cands[1] != "2.97.0" {
		t.Errorf("candidateVersions(v2.97.0) = %v, want [v2.97.0 2.97.0]", cands)
	}
	// transform=trim_v → 直接得到 2.97.0,且不重复
	cands = candidateVersions("v2.97.0", "trim_v")
	if len(cands) != 2 || cands[0] != "2.97.0" || cands[1] != "v2.97.0" {
		t.Errorf("candidateVersions(v2.97.0, trim_v) = %v, want [2.97.0 v2.97.0]", cands)
	}
	// tag 不带 v → 试原样 + 补 v
	cands = candidateVersions("3.6.1", "")
	if len(cands) != 2 || cands[0] != "3.6.1" || cands[1] != "v3.6.1" {
		t.Errorf("candidateVersions(3.6.1) = %v, want [3.6.1 v3.6.1]", cands)
	}
}

// TestInstallAssetGhStyle 模拟 gh 的资产匹配:tag=v2.97.0,资产名不带 v。
// 验证候选版本循环能让 BuildAssetName+FindAsset 命中 gh_2.97.0_macOS_arm64。
func TestInstallAssetGhStyle(t *testing.T) {
	pm := config.DefaultPlatformMap()
	assets := []AssetLike{
		mockAsset{"gh_2.97.0_checksums.txt"},
		mockAsset{"gh_2.97.0_macOS_arm64.zip"},
		mockAsset{"gh_2.97.0_linux_amd64.tar.gz"},
	}
	var foundName string
	effectiveVersion := "v2.97.0"
	for _, cand := range candidateVersions("v2.97.0", "") {
		name := BuildAssetName("{name}_{version}_{os}_{arch}", "gh", cand, "darwin", "arm64", pm)
		if n, ok := FindAsset(assets, name); ok {
			foundName = n
			effectiveVersion = cand
			break
		}
	}
	if foundName != "gh_2.97.0_macOS_arm64.zip" {
		t.Fatalf("gh 风格匹配失败: foundName=%q", foundName)
	}
	if effectiveVersion != "2.97.0" {
		t.Errorf("effectiveVersion = %q, want 2.97.0", effectiveVersion)
	}
	// 校验文件也应以去掉 v 的版本命中
	sumName := BuildChecksumName("{name}_{version}_checksums.txt", "gh", effectiveVersion, "darwin", "arm64", pm)
	if _, ok := FindAsset(assets, sumName); !ok {
		t.Errorf("checksums 文件名 %q 应能命中", sumName)
	}
}

func TestFindAsset(t *testing.T) {
	assets := []AssetLike{
		mockAsset{"asc_3.6.1_linux_amd64"},
		mockAsset{"asc_3.6.1_macOS_arm64"},
		mockAsset{"asc_3.6.1_checksums.txt"},
	}
	// 精确匹配
	if name, ok := FindAsset(assets, "asc_3.6.1_macOS_arm64"); !ok || name != "asc_3.6.1_macOS_arm64" {
		t.Errorf("精确匹配失败: %q %v", name, ok)
	}
	// 大小写不敏感匹配(用户配置里写 macOS 或 MACOS 都行)
	if name, ok := FindAsset(assets, "asc_3.6.1_MACOS_ARM64"); !ok || name != "asc_3.6.1_macOS_arm64" {
		t.Errorf("大小写不敏感匹配失败: %q %v", name, ok)
	}
	// 不存在
	if _, ok := FindAsset(assets, "asc_3.6.1_win32"); ok {
		t.Error("不存在的 asset 不应匹配")
	}
	// nil
	if _, ok := FindAsset(nil, "x"); ok {
		t.Error("nil assets 不应匹配")
	}
}

// ---------- checksums ----------

func TestParseChecksums(t *testing.T) {
	h1 := strings64("aaa")
	h2 := strings64("bbb")
	data := []byte(
		"# comment line\n" +
			h1 + "  asc_3.6.1_macOS_arm64\n" +
			h2 + " *asc_3.6.1_linux_amd64\n" +
			"\n" +
			"badhash  somefile\n")
	sums, err := parseChecksums(data)
	if err != nil {
		t.Fatalf("parseChecksums: %v", err)
	}
	if sums["asc_3.6.1_macOS_arm64"] != h1 {
		t.Errorf("macOS hash = %q, want %q", sums["asc_3.6.1_macOS_arm64"], h1)
	}
	if sums["asc_3.6.1_linux_amd64"] != h2 {
		t.Errorf("linux hash = %q, want %q", sums["asc_3.6.1_linux_amd64"], h2)
	}
	if _, ok := sums["badhash"]; ok {
		t.Error("非法 hash 行应被跳过")
	}
}

func TestVerifySHA256OK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.bin")
	content := []byte("hello cli-manager")
	os.WriteFile(p, content, 0o644)

	sum := sha256.Sum256(content)
	if err := verifySHA256(p, hex.EncodeToString(sum[:])); err != nil {
		t.Errorf("正确 hash 应通过: %v", err)
	}
}

func TestVerifySHA256Mismatch(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.bin")
	os.WriteFile(p, []byte("hello"), 0o644)

	if err := verifySHA256(p, strings64("deadbeef")); err == nil {
		t.Error("错误 hash 应失败")
	}
}

func TestVerifySHA256MissingFile(t *testing.T) {
	if err := verifySHA256("/nonexistent/path", strings64("abc")); err == nil {
		t.Error("缺失文件应报错")
	}
}

// strings64 返回一个 64 字符的 hex 字符串(用 n 填充到 64 位)。
func strings64(fill string) string {
	out := ""
	for len(out) < 64 {
		out += fill
	}
	return out[:64]
}
