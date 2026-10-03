// Package version 提供宽松的语义化版本比较,用于比对项目要求与实际版本。
// 仅解析主.次.修订三段;缺失段按 0 处理,非数字后缀(如 -rc1)被忽略。
package version

import "strings"

// Compare 比较版本 a 与 b,返回 -1(a<b)、0(相等)或 1(a>b)。
func Compare(a, b string) int {
	pa, pb := parse(a), parse(b)
	for i := range pa {
		switch {
		case pa[i] < pb[i]:
			return -1
		case pa[i] > pb[i]:
			return 1
		}
	}
	return 0
}

// AtLeast 判断 v 是否不低于 min。
func AtLeast(v, min string) bool { return Compare(v, min) >= 0 }

func parse(s string) [3]int {
	var out [3]int
	for i, part := range strings.Split(s, ".") {
		if i >= len(out) {
			break
		}
		out[i] = leadingInt(part)
	}
	return out
}

func leadingInt(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
