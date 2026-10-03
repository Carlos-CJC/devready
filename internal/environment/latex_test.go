package environment

import (
	"context"
	"testing"
)

func TestParseLatexEngine(t *testing.T) {
	cases := []struct{ in, want string }{
		{"XeTeX 3.141592653-2.6-0.999998 (TeX Live 2026)\nkpathsea version 6.4.2", "XeLaTeX"},
		{"pdfTeX 3.141592653", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseLatexEngine(c.in); got != c.want {
			t.Errorf("parseLatexEngine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectLaTeX(t *testing.T) {
	run := fakeRunner(map[string]string{
		"xelatex --version": "XeTeX 3.141592653-2.6-0.999998 (TeX Live 2026)",
	})
	got := detect(context.Background(), run, latexSpec, "/Library/TeX/texbin/xelatex")
	if got.Active != "XeLaTeX" || !got.Available {
		t.Errorf("Active/Available = (%q, %v), want (XeLaTeX, true)", got.Active, got.Available)
	}
}
