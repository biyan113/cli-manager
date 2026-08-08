package tool

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"cli-manager/internal/config"
	gh "cli-manager/internal/github"
)

func writeTestZip(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "asset.zip")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeTestTarGz(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "asset.tar.gz")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestExtractBinaryFromZip 模拟 gh 的 zip:二进制藏在 gh_2.97.0_macOS_arm64/gh。
func TestExtractBinaryFromZip(t *testing.T) {
	binary := []byte("fake mach-o gh binary")
	p := writeTestZip(t, map[string][]byte{
		"gh_2.97.0_macOS_arm64/LICENSE": []byte("license text"),
		"gh_2.97.0_macOS_arm64/gh":      binary,
		"gh_2.97.0_macOS_arm64/man/gh.1": []byte("man page"),
	})
	out, err := extractBinary(p, "gh")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	defer os.Remove(out)
	data, _ := os.ReadFile(out)
	if !bytes.Equal(data, binary) {
		t.Errorf("提取内容不符: got %q, want %q", data, binary)
	}
}

// TestExtractBinaryFromTarGz 模拟 lazygit 的 tar.gz:lazygit_0.64.0_darwin_arm64/lazygit。
func TestExtractBinaryFromTarGz(t *testing.T) {
	binary := []byte("fake mach-o lazygit binary")
	p := writeTestTarGz(t, map[string][]byte{
		"lazygit_0.64.0_darwin_arm64/LICENSE": []byte("license"),
		"lazygit_0.64.0_darwin_arm64/lazygit":  binary,
	})
	out, err := extractBinary(p, "lazygit")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	defer os.Remove(out)
	data, _ := os.ReadFile(out)
	if !bytes.Equal(data, binary) {
		t.Errorf("提取内容不符: got %q, want %q", data, binary)
	}
}

// TestExtractBinaryPlainFile 单文件二进制应原样返回,不复制。
func TestExtractBinaryPlainFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "asc")
	os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755)
	out, err := extractBinary(p, "asc")
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if out != p {
		t.Errorf("单文件应返回原路径, got %q want %q", out, p)
	}
}

// TestExtractBinaryZipMissingBinary zip 里没有目标二进制时应报错。
func TestExtractBinaryZipMissingBinary(t *testing.T) {
	p := writeTestZip(t, map[string][]byte{"README.txt": []byte("hi")})
	if _, err := extractBinary(p, "gh"); err == nil {
		t.Fatal("zip 中缺少二进制应报错")
	}
}

// TestInstallVersionFromZip 完整安装流程:资产是 zip,安装后 target 应为解压出的二进制。
func TestInstallVersionFromZip(t *testing.T) {
	// zip 内含 asc 二进制
	realBin := []byte("real asc binary content")
	zipPayload := buildZipBytes(t, map[string][]byte{
		"asc_3.6.0_macOS_arm64/asc": realBin,
		"asc_3.6.0_macOS_arm64/README": []byte("readme"),
	})
	sum := sha256.Sum256(zipPayload)
	sumText := hex.EncodeToString(sum[:]) + "  asc_3.6.0_macOS_arm64\n"

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/download/") {
			if strings.Contains(r.URL.Path, "checksums") {
				w.Write([]byte(sumText))
				return
			}
			w.Write(zipPayload)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"3.6.0","assets":[
			{"name":"asc_3.6.0_macOS_arm64","browser_download_url":"` + srv.URL + `/download/asc","size":` + strconv.Itoa(len(zipPayload)) + `},
			{"name":"asc_3.6.0_checksums.txt","browser_download_url":"` + srv.URL + `/download/checksums","size":100}]}`))
	}))
	defer srv.Close()

	installDir := t.TempDir()
	cfg := &config.Config{InstallDir: installDir, Tools: []config.ToolSpec{testSpec()}}
	st := &config.State{Versions: map[string]config.InstalledInfo{}}
	c := gh.NewClient("")
	gh.SetBaseForTest(c, srv.URL)
	m := NewManager(cfg, st, c)

	res, err := m.installVersion(context.Background(), testSpec(), "", OpInstall)
	if err != nil {
		t.Fatalf("installVersion: %v", err)
	}
	if res.Version != "3.6.0" {
		t.Errorf("version = %q, want 3.6.0", res.Version)
	}

	binPath := filepath.Join(installDir, "asc")
	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("二进制未安装: %v", err)
	}
	if !bytes.Equal(data, realBin) {
		t.Errorf("安装的应为解压后的二进制, got %q want %q", data, realBin)
	}
	info, _ := os.Stat(binPath)
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("二进制应可执行, mode = %v", info.Mode())
	}
}

// buildZipBytes 返回一个含给定条目的 zip 字节。
func buildZipBytes(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	p := writeTestZip(t, entries)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

