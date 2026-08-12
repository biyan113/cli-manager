package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadCreatesDefault 验证首次 Load 会创建含内置 asc 的默认配置。
// 用临时 HOME 隔离,macOS 上 UserConfigDir 会落到 <tmp>/Library/Application Support。
func TestLoadCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantInstallDir := DefaultInstallDir()
	if c.InstallDir != wantInstallDir {
		t.Errorf("InstallDir = %q, want %q", c.InstallDir, wantInstallDir)
	}
	if len(c.Tools) != 5 {
		t.Fatalf("默认配置工具数 = %d, want 5", len(c.Tools))
	}
	if c.Tools[0].ID != "asc" {
		t.Errorf("默认首个工具 = %q, want asc", c.Tools[0].ID)
	}
	if c.Tools[0].Repo != "rorkai/App-Store-Connect-CLI" {
		t.Errorf("asc repo = %q, want rorkai/App-Store-Connect-CLI", c.Tools[0].Repo)
	}
	// 确认文件真实写盘了
	if _, err := os.Stat(ConfigPath()); err != nil {
		t.Errorf("配置文件未落盘: %v", err)
	}
}

func TestLoadMigratesDefaultsOnce(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	legacy := &Config{InstallDir: "~/custom-bin", Tools: []ToolSpec{{ID: "asc", Repo: "custom/asc", Binary: "my-asc"}}}
	if err := legacy.Save(); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SchemaVersion != CurrentSchemaVersion || len(c.Tools) != 5 {
		t.Fatalf("migration = version %d, tools %d", c.SchemaVersion, len(c.Tools))
	}
	if c.Tools[0].Repo != "custom/asc" || c.Tools[0].Binary != "my-asc" {
		t.Fatal("迁移不应覆盖用户修改过的工具")
	}
	if err := c.RemoveTool("gh"); err != nil {
		t.Fatal(err)
	}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	c, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Tools) != 4 {
		t.Fatalf("已迁移配置不应重新添加用户删除的工具，got %d", len(c.Tools))
	}
}

func TestSetLanguage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetLanguage("en"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetLanguage("fr"); err == nil {
		t.Fatal("unsupported language should fail")
	}
	c, _ = Load()
	if c.Language != "en" {
		t.Fatalf("language = %q, want en", c.Language)
	}
}

func TestLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	c.SetToken("abc123")
	c2, err := Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if c2.GithubToken != "abc123" {
		t.Errorf("token 未持久化: %q", c2.GithubToken)
	}
}

func TestAddRemoveTool(t *testing.T) {
	c := &Config{InstallDir: "~/.local/bin", Tools: DefaultTools()}
	spec := ToolSpec{ID: "custom", Name: "custom", Repo: "example/custom", Binary: "custom"}
	if err := c.AddTool(spec); err != nil {
		t.Fatalf("AddTool: %v", err)
	}
	if len(c.Tools) != 6 {
		t.Fatalf("len(Tools) = %d, want 6", len(c.Tools))
	}
	// 重复 id 应报错
	if err := c.AddTool(ToolSpec{ID: "custom", Repo: "x/y"}); err == nil {
		t.Error("重复 id 应报错")
	}
	// 移除
	if err := c.RemoveTool("custom"); err != nil {
		t.Fatalf("RemoveTool: %v", err)
	}
	if err := c.RemoveTool("custom"); err != ErrNotFound {
		t.Errorf("移除不存在的工具应返回 ErrNotFound, got %v", err)
	}
	// 校验必填
	if err := c.AddTool(ToolSpec{ID: "", Repo: "x/y"}); err == nil {
		t.Error("空 id 应报错")
	}
	if err := c.AddTool(ToolSpec{ID: "x", Repo: ""}); err == nil {
		t.Error("空 repo 应报错")
	}
}

func TestFindTool(t *testing.T) {
	c := &Config{InstallDir: "~/.local/bin", Tools: DefaultTools()}
	tool, err := c.FindTool("asc")
	if err != nil {
		t.Fatalf("FindTool: %v", err)
	}
	if tool.Binary != "asc" {
		t.Errorf("binary = %q", tool.Binary)
	}
	if _, err := c.FindTool("nope"); err != ErrNotFound {
		t.Errorf("未找到时应返回 ErrNotFound, got %v", err)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	cases := []struct{ in, want string }{
		{"~/.local/bin", filepath.Join(home, ".local", "bin")},
		{"~/x", filepath.Join(home, "x")},
		{"~", home},
		{"/usr/local/bin", "/usr/local/bin"},
		{"", ""},
	}
	for _, c := range cases {
		if got := ExpandPath(c.in); got != c.want {
			t.Errorf("ExpandPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultPlatformMap(t *testing.T) {
	m := DefaultPlatformMap()
	if m["darwin"] != "macOS" {
		t.Errorf("darwin 映射 = %q, want macOS", m["darwin"])
	}
	if m["linux"] != "linux" {
		t.Errorf("linux 映射 = %q, want linux", m["linux"])
	}
}
