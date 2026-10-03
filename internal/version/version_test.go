package version

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.22.0", "1.22.0", 0},
		{"1.22.1", "1.22.0", 1},
		{"1.21.9", "1.22.0", -1},
		{"1.22", "1.22.0", 0},       // 缺失段补 0
		{"1.22.0-rc1", "1.22.0", 0}, // 非数字后缀忽略
		{"2.0.0", "1.99.99", 1},
		{"1.9.0", "1.10.0", -1}, // 数字比较而非字典序
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestAtLeast(t *testing.T) {
	cases := []struct {
		v, min string
		want   bool
	}{
		{"1.22.0", "1.22.0", true},
		{"1.22.8", "1.22.0", true},
		{"1.21.0", "1.22.0", false},
		{"1.22", "1.22.0", true},
	}
	for _, c := range cases {
		if got := AtLeast(c.v, c.min); got != c.want {
			t.Errorf("AtLeast(%q, %q) = %v, want %v", c.v, c.min, got, c.want)
		}
	}
}
