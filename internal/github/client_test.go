package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/rorkai/App-Store-Connect-CLI/releases/latest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != "2022-11-28" {
			t.Errorf("X-GitHub-Api-Version = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "3.6.0",
			"name": "3.6.0",
			"draft": false,
			"prerelease": false,
			"assets": [
				{"name": "asc_3.6.0_macOS_arm64", "browser_download_url": "https://example.com/asc_3.6.0_macOS_arm64", "size": 1234},
				{"name": "asc_3.6.0_checksums.txt", "browser_download_url": "https://example.com/checksums.txt", "size": 99}
			]
		}`))
	}))
	defer srv.Close()

	c := newClientWithBase("", srv.URL)
	rel, err := c.GetLatestRelease(context.Background(), "rorkai", "App-Store-Connect-CLI")
	if err != nil {
		t.Fatalf("GetLatestRelease: %v", err)
	}
	if rel.TagName != "3.6.0" {
		t.Errorf("tag = %q, want 3.6.0", rel.TagName)
	}
	if len(rel.Assets) != 2 {
		t.Errorf("assets = %d, want 2", len(rel.Assets))
	}
	if rel.Assets[0].Name != "asc_3.6.0_macOS_arm64" {
		t.Errorf("asset[0].Name = %q", rel.Assets[0].Name)
	}
}

func TestAuthHeaderSent(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.0.0","assets":[]}`))
	}))
	defer srv.Close()

	c := newClientWithBase("secret-token", srv.URL)
	if _, err := c.GetLatestRelease(context.Background(), "o", "r"); err != nil {
		t.Fatalf("GetLatestRelease: %v", err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Errorf("Authorization = %q, want Bearer secret-token", gotAuth)
	}
}

func TestRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "9999999999")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer srv.Close()

	c := newClientWithBase("", srv.URL)
	_, err := c.GetLatestRelease(context.Background(), "o", "r")
	if err == nil {
		t.Fatal("期望限流错误")
	}
	if !strings.Contains(err.Error(), "限流") {
		t.Errorf("错误应包含限流提示, got: %v", err)
	}
}

func TestNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	c := newClientWithBase("", srv.URL)
	_, err := c.GetLatestRelease(context.Background(), "o", "r")
	if err == nil {
		t.Fatal("期望 404 错误")
	}
	if !strings.Contains(err.Error(), "Not Found") {
		t.Errorf("错误应包含 Not Found, got: %v", err)
	}
}

func TestGetBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("abc123 hash  file.txt\n"))
	}))
	defer srv.Close()

	c := NewClient("")
	body, err := c.GetBody(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("GetBody: %v", err)
	}
	if string(body) != "abc123 hash  file.txt\n" {
		t.Errorf("body = %q", string(body))
	}
}
