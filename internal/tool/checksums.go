package tool

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// parseChecksums 解析 sha256sum 格式的 checksums 文本,返回 filename → sha256 hex。
// 兼容行首 * 的 binary 模式标记、空格/多空格分隔、空行和 # 注释。
// 非 64 位 hex 的行会被跳过。
func parseChecksums(data []byte) (map[string]string, error) {
	sums := make(map[string]string)
	for lineNo, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue // 缺 hash 或文件名,跳过
		}
		hash := fields[0]
		if len(hash) != 64 {
			continue
		}
		if _, err := hex.DecodeString(hash); err != nil {
			continue
		}
		fname := fields[len(fields)-1]
		fname = strings.TrimPrefix(fname, "*")
		sums[fname] = strings.ToLower(hash)
		_ = lineNo
	}
	return sums, nil
}

// verifySHA256 计算文件的 sha256 并与期望的 hex 比对(常量时间比较)。
func verifySHA256(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开临时文件: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("计算 sha256: %w", err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	want = strings.ToLower(strings.TrimSpace(want))

	if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
		return fmt.Errorf("SHA256 校验失败:期望 %s,实际 %s", want, got)
	}
	return nil
}
