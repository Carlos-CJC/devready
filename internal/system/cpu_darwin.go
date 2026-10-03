//go:build darwin

package system

import (
	"context"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

// collectCPU 通过 sysctl 读取型号 / 核心数 / 负载,通过 top 取实时利用率。
func collectCPU(ctx context.Context, run versionmanager.Runner) domain.Metric[domain.CPUInfo] {
	out, err := run(ctx, "sysctl", "-n", "machdep.cpu.brand_string")
	if err != nil {
		return domain.Failed[domain.CPUInfo](err)
	}
	var info domain.CPUInfo
	info.Model = ParseCPUModel(out)

	if out, err := run(ctx, "sysctl", "-n", "hw.physicalcpu"); err == nil {
		info.Cores = ParseCount(out)
	}
	if out, err := run(ctx, "sysctl", "-n", "vm.loadavg"); err == nil {
		info.Load1, info.Load5, info.Load15 = ParseLoadAvg(out)
	}
	// top -l 2 取第二个采样,得到接近实时的利用率(首个采样是自开机以来的均值)。
	if out, err := run(ctx, "top", "-l", "2", "-n", "0", "-s", "1"); err == nil {
		if u, ok := ParseCPUUsage(out); ok {
			info.Utilization = u
		}
	}
	return domain.OK(info)
}
