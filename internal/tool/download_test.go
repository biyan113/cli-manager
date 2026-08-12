package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/biyan113/cli-box/internal/config"
	gh "github.com/biyan113/cli-box/internal/github"
)

// startFakeServer 启动一个模拟 GitHub Releases 的 httptest server。
// 它根据请求路径返回 latest/tags/下载/checksums。
func startFakeServer(t *testing.T, payload []byte, sumsText string) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 下载 asset 或 checksums 走 GET 直接返回 payload
		switch {
		case strings.Contains(r.URL.Path, "/download/"):
			if strings.Contains(r.URL.Path, "checksums") {
				w.Write([]byte(sumsText))
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(payload)
			return
		}
		// API 路径:返回 release JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"tag_name": "3.6.0",
			"name": "3.6.0",
			"assets": [
				{"name": "asc_3.6.0_macOS_arm64", "browser_download_url": "%s/download/asc", "size": %d},
				{"name": "asc_3.6.0_checksums.txt", "browser_download_url": "%s/download/checksums", "size": 100}
			]
		}`, srv.URL, len(payload), srv.URL)))
	}))
	return srv
}

func testSpec() config.ToolSpec {
	return config.ToolSpec{
		ID:               "asc",
		Name:             "asc",
		Repo:             "rorkai/App-Store-Connect-CLI",
		Binary:           "asc",
		AssetPattern:     "{name}_{version}_{os}_{arch}",
		ChecksumsPattern: "{name}_{version}_checksums.txt",
		VersionCmd:       []string{"--version"},
		VersionRegex:     `^([0-9]+\.[0-9]+\.[0-9]+)`,
		OS:               "darwin",
		Arch:             "arm64",
	}
}

func testBinaryPath(dir string) string {
	name := "asc"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

func TestInstallVersion(t *testing.T) {
	// 构造一个真实的二进制 payload(用当前可执行文件内容即可,只测流程)
	payload := []byte("#!/bin/sh\necho 3.6.0\n")
	sum := sha256.Sum256(payload)
	sumText := hex.EncodeToString(sum[:]) + "  asc_3.6.0_macOS_arm64\n"

	srv := startFakeServer(t, payload, sumText)
	defer srv.Close()

	installDir := t.TempDir()
	cfg := &config.Config{InstallDir: installDir, Tools: []config.ToolSpec{testSpec()}}
	st := &config.State{Versions: map[string]config.InstalledInfo{}}

	// 用测试 server 替换 GitHub base
	c := gh.NewClient("")
	// 通过 SetBaseForTest 注入
	gh.SetBaseForTest(c, srv.URL)

	m := NewManager(cfg, st, c)
	m.Ctx = context.Background()

	var progressEvents []string
	m.OnProgress = func(id string, op OpType, d, total int64, phase string) {
		progressEvents = append(progressEvents, phase)
	}
	var statusEvents []string
	m.OnStatus = func(id string, op OpType, status, version, message string) {
		statusEvents = append(statusEvents, status+":"+version)
	}

	res, err := m.installVersion(context.Background(), testSpec(), "", OpInstall)
	if err != nil {
		t.Fatalf("installVersion: %v", err)
	}
	if res.Version != "3.6.0" {
		t.Errorf("version = %q, want 3.6.0", res.Version)
	}

	// 校验文件真实落盘且可执行
	binPath := testBinaryPath(installDir)
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("二进制未安装: %v", err)
	}
	info, _ := os.Stat(binPath)
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o100 == 0 {
		t.Errorf("二进制应可执行, mode = %v", info.Mode())
	}
	// 内容正确
	data, _ := os.ReadFile(binPath)
	if string(data) != string(payload) {
		t.Errorf("二进制内容不符")
	}

	// state 已记录
	rec, ok := st.Versions["asc"]
	if !ok || rec.Version != "3.6.0" || rec.BinPath != binPath {
		t.Errorf("state 未正确记录: %+v", rec)
	}

	// 进度阶段顺序正确
	wantPhases := []string{"resolve", "resolve", "checksum", "download", "download", "checksum", "install"}
	if len(progressEvents) < len(wantPhases) {
		t.Fatalf("进度事件太少: %v", progressEvents)
	}
	for i, want := range wantPhases {
		if progressEvents[i] != want {
			t.Errorf("phase[%d] = %q, want %q", i, progressEvents[i], want)
		}
	}
	// 终态事件
	if len(statusEvents) == 0 || !strings.HasPrefix(statusEvents[len(statusEvents)-1], "done:3.6.0") {
		t.Errorf("期望 done 终态事件, got %v", statusEvents)
	}
}

func TestInstallVersionChecksumMismatch(t *testing.T) {
	payload := []byte("evil binary")
	sum := sha256.Sum256([]byte("something else")) // 错误的 hash
	sumText := hex.EncodeToString(sum[:]) + "  asc_3.6.0_macOS_arm64\n"

	srv := startFakeServer(t, payload, sumText)
	defer srv.Close()

	installDir := t.TempDir()
	cfg := &config.Config{InstallDir: installDir, Tools: []config.ToolSpec{testSpec()}}
	st := &config.State{Versions: map[string]config.InstalledInfo{}}
	c := gh.NewClient("")
	gh.SetBaseForTest(c, srv.URL)
	m := NewManager(cfg, st, c)

	_, err := m.installVersion(context.Background(), testSpec(), "", OpInstall)
	if err == nil {
		t.Fatal("checksums 不匹配应报错")
	}
	if !strings.Contains(err.Error(), "SHA256") {
		t.Errorf("错误应包含 SHA256 提示, got: %v", err)
	}
	// 不应残留临时文件
	entries, _ := os.ReadDir(installDir)
	if len(entries) != 0 {
		t.Errorf("失败后不应残留文件, got %d", len(entries))
	}
}

func TestInstallVersionNoAssetForPlatform(t *testing.T) {
	// 平台映射里用 linux 覆盖,darwin 是 macOS
	spec := testSpec()
	spec.OS = "linux"
	spec.Arch = "amd64"

	payload := []byte("binary")
	sum := sha256.Sum256(payload)
	sumText := hex.EncodeToString(sum[:]) + "  asc_3.6.0_linux_x86_64\n"

	srv := startFakeServer(t, payload, sumText)
	defer srv.Close()

	installDir := t.TempDir()
	cfg := &config.Config{InstallDir: installDir}
	st := &config.State{Versions: map[string]config.InstalledInfo{}}
	c := gh.NewClient("")
	gh.SetBaseForTest(c, srv.URL)
	m := NewManager(cfg, st, c)

	// fake server 只提供 macOS asset,linux 场景应报"找不到资产"
	_, err := m.installVersion(context.Background(), spec, "", OpInstall)
	if err == nil {
		t.Fatal("平台不匹配应报错")
	}
	if !strings.Contains(err.Error(), "找不到资产") {
		t.Errorf("错误应提示找不到资产, got: %v", err)
	}
}
