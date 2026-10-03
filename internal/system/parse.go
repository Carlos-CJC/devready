package system

import (
	"strconv"
	"strings"

	"github.com/Carlos-CJC/devscope/internal/domain"
)

// 本文件只做纯解析:输入是系统命令的标准输出,输出是领域类型。
// 不涉及平台判断或命令执行,因此可在任意平台单测(见设计文档 §2.2、§17)。

// ParseCPUModel 解析 `sysctl -n machdep.cpu.brand_string`,如 "Apple M4"。
func ParseCPUModel(out string) string { return strings.TrimSpace(out) }

// ParseCount 解析单个整数输出,如 `sysctl -n hw.physicalcpu`。
func ParseCount(out string) int {
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0
	}
	return n
}

// ParseLoadAvg 解析 `sysctl -n vm.loadavg`,输入形如 "{ 1.64 1.45 1.43 }"。
func ParseLoadAvg(out string) (l1, l5, l15 float64) {
	f := strings.Fields(strings.NewReplacer("{", " ", "}", " ").Replace(out))
	if len(f) < 3 {
		return 0, 0, 0
	}
	l1, _ = strconv.ParseFloat(f[0], 64)
	l5, _ = strconv.ParseFloat(f[1], 64)
	l15, _ = strconv.ParseFloat(f[2], 64)
	return l1, l5, l15
}

// ParseCPUUsage 解析 top 输出中 "CPU usage: ... % idle" 行,返回利用率(0-100)与是否解析成功。
// 多行匹配时取最后一行——对应 `top -l 2` 的第二个采样,更接近实时(首个采样是自开机以来的均值)。
func ParseCPUUsage(out string) (float64, bool) {
	idle, found := 0.0, false
	for _, line := range strings.Split(out, "\n") {
		i := strings.Index(line, "CPU usage:")
		if i < 0 {
			continue
		}
		for _, part := range strings.Split(line[i:], ",") {
			f := strings.Fields(part)
			if len(f) == 2 && strings.HasSuffix(f[1], "idle") {
				if v, err := strconv.ParseFloat(strings.TrimSuffix(f[0], "%"), 64); err == nil {
					idle, found = v, true
				}
			}
		}
	}
	if !found {
		return 0, false
	}
	if idle < 0 {
		idle = 0
	}
	return 100 - idle, true
}

// ParseMemTotal 解析 `sysctl -n hw.memsize`(字节)。
func ParseMemTotal(out string) uint64 {
	n, err := strconv.ParseUint(strings.TrimSpace(out), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// VMPages 是 vm_stat 中与内存使用量相关的页计数。
type VMPages struct {
	Free        uint64
	Active      uint64
	Inactive    uint64
	Speculative uint64
	Wired       uint64
	Compressor  uint64
}

// ParseVMStat 解析 `vm_stat` 输出,返回页大小与相关页计数。
func ParseVMStat(out string) (pageSize uint64, p VMPages, ok bool) {
	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.Contains(line, "page size of"):
			pageSize = leadingUint(line)
			ok = pageSize > 0
		case strings.HasPrefix(line, "Pages "):
			key, val := splitVMStat(line)
			switch key {
			case "free":
				p.Free = val
			case "active":
				p.Active = val
			case "inactive":
				p.Inactive = val
			case "speculative":
				p.Speculative = val
			case "wired down":
				p.Wired = val
			case "occupied by compressor":
				p.Compressor = val
			}
		}
	}
	return pageSize, p, ok
}

// ParseDF 解析 `df -k -P` 输出,返回本地磁盘容量。容量单位由 KB 换算为字节。
func ParseDF(out string) []domain.DiskInfo {
	var disks []domain.DiskInfo
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 6 || !strings.HasPrefix(f[0], "/dev/") {
			continue
		}
		mount := strings.Join(f[5:], " ")
		if !keepMount(mount) {
			continue
		}
		total, err1 := strconv.ParseUint(f[1], 10, 64)
		used, err2 := strconv.ParseUint(f[2], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		disks = append(disks, domain.DiskInfo{
			MountPoint: mount,
			Total:      total * 1024,
			Used:       used * 1024,
		})
	}
	return disks
}

// keepMount 过滤掉 macOS 的 APFS 伪卷与内存交换卷,只保留用户关心的挂载点。
func keepMount(mount string) bool {
	if mount == "/System/Volumes/Data" {
		return true
	}
	return !strings.HasPrefix(mount, "/System/Volumes/") &&
		!strings.HasPrefix(mount, "/private/var/vm")
}

// ParseNetstat 解析 `netstat -ib`,返回各接口的累计字节数:name -> {收, 发}。
// 只取 <Link#> 行(承载真实的接口计数器),并跳过回环接口。
func ParseNetstat(out string) map[string][2]uint64 {
	res := make(map[string][2]uint64)
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 10 || !strings.HasPrefix(f[2], "<Link#") || f[0] == "lo0" {
			continue
		}
		rx, err1 := strconv.ParseUint(f[6], 10, 64)
		tx, err2 := strconv.ParseUint(f[9], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		res[f[0]] = [2]uint64{rx, tx}
	}
	return res
}

// splitVMStat 拆分一行 "Pages <key>: <n>." 为 key 与数值。
func splitVMStat(line string) (key string, val uint64) {
	body := strings.TrimPrefix(line, "Pages ")
	i := strings.IndexByte(body, ':')
	if i < 0 {
		return "", 0
	}
	key = strings.TrimSpace(body[:i])
	rest := strings.TrimRight(strings.TrimSpace(body[i+1:]), ".")
	val, _ = strconv.ParseUint(rest, 10, 64)
	return key, val
}

// leadingUint 取字符串中出现的第一个连续整数。
func leadingUint(s string) uint64 {
	start := -1
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			start = i
			break
		}
	}
	if start < 0 {
		return 0
	}
	end := start
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	n, _ := strconv.ParseUint(s[start:end], 10, 64)
	return n
}
