//go:build darwin

package system

import (
	"context"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// collectMemory 读取物理内存总量与 vm_stat 页计数,计算已用 / 可回收。
func collectMemory(ctx context.Context, run versionmanager.Runner) domain.Metric[domain.MemInfo] {
	totalOut, err := run(ctx, "sysctl", "-n", "hw.memsize")
	if err != nil {
		return domain.Failed[domain.MemInfo](err)
	}
	vmOut, err := run(ctx, "vm_stat")
	if err != nil {
		return domain.Failed[domain.MemInfo](err)
	}
	pageSize, p, ok := ParseVMStat(vmOut)
	if !ok {
		return domain.Unavailable[domain.MemInfo]()
	}

	total := ParseMemTotal(totalOut)
	// 空闲 + 非活跃 + 推测 视为可回收;其余(活跃 + 联动 + 压缩)计为已用。
	avail := (p.Free + p.Inactive + p.Speculative) * pageSize
	if avail > total {
		avail = total
	}
	return domain.OK(domain.MemInfo{Total: total, Used: total - avail, Available: avail})
}
