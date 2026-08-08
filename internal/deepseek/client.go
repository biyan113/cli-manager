// Package deepseek 封装 DeepSeek API 访问(OpenAI 兼容格式)。
// 用于"工具说明"功能:根据 GitHub 仓库信息生成中文简介。
package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	baseURL = "https://api.deepseek.com"
	// DefaultModel 是默认使用的 DeepSeek 模型。
	DefaultModel = "deepseek-v4-flash"
)

// Error 是 DeepSeek API 错误响应结构。
type Error struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ChatResponse 是 chat/completions 的非流式响应结构。
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Client 是 DeepSeek API 客户端。
type Client struct {
	token   string
	baseURL string
	model   string
	hc      *http.Client
}

// NewClient 创建客户端。token 可空,会从 DEEPSEEK_API_KEY 环境变量兜底;
// model 为空时用默认模型。
func NewClient(token, model string) *Client {
	if token == "" {
		token = os.Getenv("DEEPSEEK_API_KEY")
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
		model:   model,
		hc:      &http.Client{Timeout: 60 * time.Second},
	}
}

// SetToken 更新 API key。
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetModel 更新模型。
func (c *Client) SetModel(model string) {
	if model != "" {
		c.model = model
	}
}

// HasToken 报告是否已配置 API key。
func (c *Client) HasToken() bool {
	return c.token != ""
}

// Chat 发送一条 system+user 对话请求,返回模型回复文本。
func (c *Client) Chat(ctx context.Context, system, user string) (string, error) {
	if c.token == "" {
		return "", errors.New("未配置 DeepSeek API key,请在设置中填写")
	}

	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"stream": false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 DeepSeek: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		var e Error
		if data, rerr := io.ReadAll(io.LimitReader(resp.Body, 64<<10)); rerr == nil {
			if json.Unmarshal(data, &e) == nil && e.Error.Message != "" {
				msg = e.Error.Message
			}
		}
		return "", fmt.Errorf("DeepSeek API: HTTP %d: %s", resp.StatusCode, msg)
	}

	var out ChatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&out); err != nil {
		return "", fmt.Errorf("解析 DeepSeek 响应: %w", err)
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return "", errors.New("DeepSeek 返回了空结果")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

// ExplainTool 根据仓库信息生成工具说明,返回简体中文与英文两个版本。
// description 是 GitHub 仓库简介;readme 是 README 文本片段。
// 一次调用让模型输出双语,用固定标记行分隔,解析失败时英文为空(向前兼容)。
func (c *Client) ExplainTool(ctx context.Context, name, repo, description, readme string) (zh, en string, err error) {
	const system = "你是 CLI 工具说明助手。根据给定的 GitHub 仓库信息与 README 片段,写两段简介:一段简体中文、一段英文。要求:\n" +
		"1. 说明这个工具是做什么的、主要功能、典型使用场景;\n" +
		"2. 用通俗的话,避免术语堆砌,让普通开发者一看就懂;\n" +
		"3. 中英文各控制在 150 字左右,分 2-4 个短段落或要点,可用简单 Markdown(加粗、短列表);\n" +
		"4. 输出必须严格使用下面两个标记行(标记行本身不要改动),中文在第一个标记后,英文在第二个标记后:\n" +
		"【中文简介】\n<中文内容>\n【English Summary】\n<English content>"

	user := fmt.Sprintf("工具名: %s\n仓库: %s\n仓库简介: %s\n\nREADME 片段(截断):\n%s",
		name, repo, description, truncate(readme, 3000))
	out, err := c.Chat(ctx, system, user)
	if err != nil {
		return "", "", err
	}
	zh, en = splitBilingual(out)
	return zh, en, nil
}

// splitBilingual 从模型输出中切分中文与英文两部分。
// 找不到分隔标记时,整段视为中文、英文为空(兼容模型不按格式输出的情况)。
func splitBilingual(out string) (zh, en string) {
	const sep = "【English Summary】"
	if i := strings.Index(out, sep); i >= 0 {
		zh = strings.TrimSpace(out[:i])
		en = strings.TrimSpace(out[i+len(sep):])
		zh = strings.TrimPrefix(zh, "【中文简介】")
		return strings.TrimSpace(zh), en
	}
	return strings.TrimSpace(out), ""
}

// truncate 把文本截断到 max 字节,按 rune 截断避免切坏 UTF-8。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	runes := []rune(s)
	if len(runes) > max {
		runes = runes[:max]
	}
	return string(runes) + "\n…(截断)"
}
