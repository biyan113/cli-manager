package tool

import "testing"

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		// 带 v 与不带 v 数值相等 → 相等
		{"2.97.0", "v2.97.0", 0},
		{"0.64.0", "v0.64.0", 0},
		{"v3.6.1", "3.6.1", 0},
		{"v2.97.0", "v2.97.0", 0},
		// 仓库自定义 tag 前缀不应造成误报更新。
		{"jq-1.8.2", "1.8.2", 0},
		{"release-1.2.3", "v1.2.3", 0},
		// 真正的版本差异
		{"0.65.0", "0.64.0", 1},
		{"0.64.0", "0.65.0", -1},
		{"v2.98.0", "v2.97.0", 1},
		{"3.6.1", "3.6.0", 1},
		// 多段/少段
		{"1.0", "1.0.0", 0},
		{"1.0.1", "1.0", 1},
		// 数值比较而非字典序(2.10 > 2.9)
		{"2.10.0", "2.9.0", 1},
		{"2.9.0", "2.10.0", -1},
	}
	for _, c := range cases {
		got := CompareVersions(c.a, c.b)
		if got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
		// 对称性
		rev := CompareVersions(c.b, c.a)
		if rev != -c.want {
			t.Errorf("对称性 CompareVersions(%q, %q) = %d, want %d", c.b, c.a, rev, -c.want)
		}
	}
}
