// Package github 封装 GitHub Releases API 访问。
// 无需第三方依赖,使用 net/http + encoding/json。
package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	baseURL    = "https://api.github.com"
	apiVersion = "2022-11-28"
	// 匿名用户 60 次/时,经认证 5000 次/时。
)

// Client 是 GitHub Releases API 客户端。
type Client struct {
	token string
	base  string
	hc    *http.Client
}

// NewClient 创建一个客户端。token 可空,会从 GITHUB_TOKEN/GH_TOKEN 环境变量兜底。
// baseURLOverride 仅供测试注入 httptest server 地址;生产环境传空字符串。
func NewClient(token string) *Client {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GH_TOKEN")
		}
	}
	return &Client{
		token: token,
		base:  baseURL,
		hc:    &http.Client{Timeout: 30 * time.Second},
	}
}

// newClientWithBase 构造指向自定义 base 的客户端(仅供测试)。
func newClientWithBase(token, base string) *Client {
	c := NewClient(token)
	c.base = base
	return c
}

// SetBaseForTest 重设一个已有客户端的 base URL(仅供跨包测试注入 httptest server)。
func SetBaseForTest(c *Client, base string) {
	c.base = base
}

// SetToken 更新 token。
func (c *Client) SetToken(token string) {
	c.token = token
}

// RepoInfo 是仓库元信息(用于生成工具说明)。
type RepoInfo struct {
	Description string `json:"description"`
}

// GetRepoInfo 返回仓库的简介等信息。
func (c *Client) GetRepoInfo(ctx context.Context, owner, repo string) (RepoInfo, error) {
	var r RepoInfo
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/repos/%s/%s", owner, repo), nil, &r); err != nil {
		return r, err
	}
	return r, nil
}

// readmeResponse 是 GET /repos/{owner}/{repo}/readme 的响应(GitHub 返回 base64 内容)。
type readmeResponse struct {
	Content string `json:"content"`
}

// GetReadme 返回仓库 README 的纯文本内容。
func (c *Client) GetReadme(ctx context.Context, owner, repo string) (string, error) {
	var r readmeResponse
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/repos/%s/%s/readme", owner, repo), nil, &r); err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(r.Content)
	if err != nil {
		return "", fmt.Errorf("解码 README: %w", err)
	}
	return string(data), nil
}

// GetLatestRelease 返回仓库最新的正式 release。
func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	var r Release
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo), nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetReleaseByTag 按 tag 名取 release(用于降级)。
func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*Release, error) {
	var r Release
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tag), nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ListReleases 分页列出 release。
func (c *Client) ListReleases(ctx context.Context, owner, repo string, page, perPage int) ([]Release, error) {
	if perPage == 0 {
		perPage = 30
	}
	path := fmt.Sprintf("/repos/%s/%s/releases?page=%d&per_page=%d", owner, repo, page, perPage)
	var r []Release
	if err := c.doJSON(ctx, "GET", path, nil, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// OpenDownload 打开一个下载流(供大文件流式读取)。
// 注意:调用方负责 resp.Body.Close()。此请求不设总超时。
func (c *Client) OpenDownload(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	// 用独立的、无总超时的 client 下载大文件。
	dc := &http.Client{}
	resp, err := dc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// GetBody 拉取一个小文件(如 checksums.txt)的内容。
func (c *Client) GetBody(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载 %s 失败: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 5<<20)) // 5MB 上限
}

// doJSON 发一个 JSON API 请求并解析响应。path 是完整的 URL(api.github.com 或测试 server)。
func (c *Client) doJSON(ctx context.Context, method, path string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	req.Header.Set("User-Agent", "cli-manager")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if resp.StatusCode == http.StatusForbidden && remaining == "0" {
		return rateLimitErr(resp.Header.Get("X-RateLimit-Reset"))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := resp.Status
		var e Error
		if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&e); err == nil && e.Message != "" {
			msg = e.Message
		}
		return fmt.Errorf("GitHub API %s: HTTP %d: %s", path, resp.StatusCode, msg)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// rateLimitErr 构造限流错误,根据 Reset 时间戳给出建议等待时长。
func rateLimitErr(reset string) error {
	secs, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return &ErrRateLimit{}
	}
	d := time.Until(time.Unix(secs, 0))
	if d < 0 {
		d = 0
	}
	e := &ErrRateLimit{ResetIn: formatDuration(d)}
	return e
}

// formatDuration 把时长格式化成中文可读的"约 X 分钟/秒"。
func formatDuration(d time.Duration) string {
	mins := int(d.Minutes())
	if mins > 0 {
		if mins >= 60 {
			return fmt.Sprintf("%d 小时", mins/60)
		}
		return fmt.Sprintf("%d 分钟", mins)
	}
	secs := int(d.Seconds())
	if secs < 1 {
		secs = 1
	}
	return fmt.Sprintf("%d 秒", secs)
}
