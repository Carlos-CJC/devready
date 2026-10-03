package system

import (
	"reflect"
	"testing"

	"github.com/Carlos-CJC/devscope/internal/domain"
)

func TestParseCPUModel(t *testing.T) {
	if got := ParseCPUModel("Apple M4\n"); got != "Apple M4" {
		t.Errorf("ParseCPUModel = %q, want Apple M4", got)
	}
}

func TestParseCount(t *testing.T) {
	if got := ParseCount("10\n"); got != 10 {
		t.Errorf("ParseCount = %d, want 10", got)
	}
	if got := ParseCount("abc"); got != 0 {
		t.Errorf("ParseCount(非法) = %d, want 0", got)
	}
}

func TestParseLoadAvg(t *testing.T) {
	l1, l5, l15 := ParseLoadAvg("{ 1.64 1.45 1.43 }")
	if l1 != 1.64 || l5 != 1.45 || l15 != 1.43 {
		t.Errorf("ParseLoadAvg = (%v, %v, %v), want (1.64, 1.45, 1.43)", l1, l5, l15)
	}
}

func TestParseCPUUsage(t *testing.T) {
	// top -l 2:两行匹配,应取最后一行(更接近实时)。
	twoSamples := `Load Avg: 1.74, 1.48, 1.44
CPU usage: 3.61% user, 8.43% sys, 87.95% idle
Load Avg: 1.74, 1.48, 1.44
CPU usage: 3.52% user, 3.52% sys, 92.94% idle
`
	got, ok := ParseCPUUsage(twoSamples)
	if !ok {
		t.Fatal("ParseCPUUsage ok=false, want true")
	}
	if diff := got - 7.06; diff > 0.001 || diff < -0.001 {
		t.Errorf("ParseCPUUsage = %v, want 7.06", got)
	}

	if _, ok := ParseCPUUsage("no cpu line here"); ok {
		t.Error("ParseCPUUsage ok=true, want false")
	}
}

func TestParseMemTotal(t *testing.T) {
	if got := ParseMemTotal("17179869184\n"); got != 17179869184 {
		t.Errorf("ParseMemTotal = %d, want 17179869184", got)
	}
}

func TestParseVMStat(t *testing.T) {
	out := `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                               10277.
Pages active:                            436466.
Pages inactive:                          430109.
Pages speculative:                         5405.
Pages wired down:                        114000.
Pages purgeable:                           8937.
"Translation faults":                  26509390.
Pages occupied by compressor:             17436.
Swapins:                                      0.
`
	pageSize, p, ok := ParseVMStat(out)
	if !ok {
		t.Fatal("ParseVMStat ok=false, want true")
	}
	if pageSize != 16384 {
		t.Errorf("pageSize = %d, want 16384", pageSize)
	}
	want := VMPages{Free: 10277, Active: 436466, Inactive: 430109, Speculative: 5405, Wired: 114000, Compressor: 17436}
	if p != want {
		t.Errorf("pages = %+v, want %+v", p, want)
	}
}

func TestParseDF(t *testing.T) {
	out := `Filesystem     1024-blocks     Used Available Capacity  Mounted on
/dev/disk3s3s1   239362496 17951812 165173268    10%    /
devfs                  208      208         0   100%    /dev
/dev/disk3s6     239362496       24 165173268     1%    /System/Volumes/VM
/dev/disk3s1     239362496 37357364 165173268    19%    /System/Volumes/Data
map auto_home            0        0         0   100%    /System/Volumes/Data/home
/dev/disk7s1     976746180  2213656 974330220     1%    /Volumes/Data
`
	got := ParseDF(out)
	want := []domain.DiskInfo{
		{MountPoint: "/", Total: 239362496 * 1024, Used: 17951812 * 1024},
		{MountPoint: "/System/Volumes/Data", Total: 239362496 * 1024, Used: 37357364 * 1024},
		{MountPoint: "/Volumes/Data", Total: 976746180 * 1024, Used: 2213656 * 1024},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseDF = %+v, want %+v", got, want)
	}
}

func TestParseNetstat(t *testing.T) {
	out := `Name       Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
lo0        16384 <Link#1>                        247413     0  130927637   247413     0  130927637     0
en0        1500  <Link#7>    d0:11:e5:80:00:bd        0     0          0        0     0          0     0
en1        1500  <Link#15>   da:1e:9d:ac:86:93  1896043     0 1055592410  2476944     0 1081464712     0
`
	got := ParseNetstat(out)
	want := map[string][2]uint64{
		"en0": {0, 0},
		"en1": {1055592410, 1081464712},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseNetstat = %+v, want %+v", got, want)
	}
}

func TestKeepMount(t *testing.T) {
	cases := []struct {
		mount string
		want  bool
	}{
		{"/", true},
		{"/Volumes/Data", true},
		{"/System/Volumes/Data", true},
		{"/System/Volumes/VM", false},
		{"/System/Volumes/Preboot", false},
		{"/private/var/vm", false},
	}
	for _, c := range cases {
		if got := keepMount(c.mount); got != c.want {
			t.Errorf("keepMount(%q) = %v, want %v", c.mount, got, c.want)
		}
	}
}

func TestShortHost(t *testing.T) {
	if got := shortHost("caojiechengdeMac-mini.local"); got != "caojiechengdeMac-mini" {
		t.Errorf("shortHost = %q, want caojiechengdeMac-mini", got)
	}
	if got := shortHost("localhost"); got != "localhost" {
		t.Errorf("shortHost = %q, want localhost", got)
	}
}
