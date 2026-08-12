package tool

import (
	"strconv"
	"strings"
)

// CompareVersions 比较两个版本字符串,返回 <0 / 0 / >0。
// 忽略 GitHub tag 在第一个数字版本前的前缀(如 v、jq-、release-);
// 优先按 semver 数字分段比较,无法解析的分段回退字符串比较。
func CompareVersions(a, b string) int {
	as := splitVersion(a)
	bs := splitVersion(b)
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		x := ""
		if i < len(as) {
			x = as[i]
		}
		y := ""
		if i < len(bs) {
			y = bs[i]
		}
		if x == y {
			continue
		}
		// 空分段按 0 参与数值比较(如 1.0 == 1.0.0)。
		xi, xe := 0, error(nil)
		yi, ye := 0, error(nil)
		if x != "" {
			xi, xe = strconv.Atoi(x)
		}
		if y != "" {
			yi, ye = strconv.Atoi(y)
		}
		if xe == nil && ye == nil {
			if xi < yi {
				return -1
			}
			if xi > yi {
				return 1
			}
			continue
		}
		// 非纯数字段(如 0-beta)按字符串比较。
		if x < y {
			return -1
		}
		return 1
	}
	return 0
}

func splitVersion(s string) []string {
	s = strings.TrimSpace(s)
	// GitHub release tags are not consistently semver-only. For example jq
	// publishes "jq-1.8.2", while `jq --version` reports "jq-1.8.2" and our
	// configured detector extracts "1.8.2". Compare from the first numeric
	// component so a repository-specific tag prefix cannot create a false
	// update notification.
	if start := strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' }); start >= 0 {
		s = s[start:]
	}
	return strings.Split(s, ".")
}
