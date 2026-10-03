package services

import (
	"reflect"
	"testing"

	"github.com/Carlos-CJC/devready/internal/domain"
)

func TestParseLsof(t *testing.T) {
	out := `COMMAND                           PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
rapportd                          608  cjc    8u  IPv4 0x1c319432a046ad96      0t0  TCP *:49154 (LISTEN)
rapportd                          608  cjc    9u  IPv6 0x80b4d43bf83b281c      0t0  TCP *:49154 (LISTEN)
node                            14635  cjc   20u  IPv4 0x528eb140d0932445      0t0  TCP 127.0.0.1:52382 (LISTEN)
com.docker.backend               900  cjc   31u  IPv6 0x0000000000000000      0t0  TCP *:5432 (LISTEN)
`
	got := ParseLsof(out)
	want := []domain.PortInfo{
		{Port: 5432, Proto: "TCP", Address: "*", Process: "com.docker.backend", PID: 900, Docker: true},
		{Port: 49154, Proto: "TCP", Address: "*", Process: "rapportd", PID: 608},
		{Port: 52382, Proto: "TCP", Address: "127.0.0.1", Process: "node", PID: 14635},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseLsof =\n%+v\nwant\n%+v", got, want)
	}
}

func TestSplitAddr(t *testing.T) {
	cases := []struct{ in, host, port string }{
		{"127.0.0.1:52382", "127.0.0.1", "52382"},
		{"*:22", "*", "22"},
		{"[::1]:8080", "[::1]", "8080"},
	}
	for _, c := range cases {
		h, p := splitAddr(c.in)
		if h != c.host || p != c.port {
			t.Errorf("splitAddr(%q) = (%q, %q), want (%q, %q)", c.in, h, p, c.host, c.port)
		}
	}
}

func TestDetectSSHD(t *testing.T) {
	if s := detectSSHD(nil); s.Running {
		t.Error("detectSSHD(空) Running=true, want false")
	}
	s := detectSSHD([]domain.PortInfo{{Port: 22}})
	if !s.Running || s.Detail != ":22" {
		t.Errorf("detectSSHD = %+v, want running :22", s)
	}
}
