package github

// Release 对应 GitHub Releases API 返回的单个 release。
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"` // 更新说明(Markdown)
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset 对应 release 里的一个附件。
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Error 是 GitHub API 错误响应体。
type Error struct {
	Message string `json:"message"`
}

// ErrRateLimit 表示命中 GitHub API 限流。
type ErrRateLimit struct {
	ResetIn string // 建议等待时长描述(如 "60 秒")
}

func (e *ErrRateLimit) Error() string {
	if e.ResetIn != "" {
		return "GitHub API 限流,请约 " + e.ResetIn + " 后重试,或在设置中配置 GitHub token"
	}
	return "GitHub API 限流,请稍后重试,或在设置中配置 GitHub token"
}
