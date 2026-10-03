//go:build !darwin

package system

import (
	"context"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

// 非 darwin 平台暂未实现采集,一律诚实降级为 Unavailable(见设计文档 §0、§2.5)。
// Linux 后端落地后,这些函数将按平台拆分到 cpu_linux.go / memory_linux.go 等文件。

// collectCPU 在非 darwin 平台返回不可用。
func collectCPU(context.Context, versionmanager.Runner) domain.Metric[domain.CPUInfo] {
	return domain.Unavailable[domain.CPUInfo]()
}

// collectMemory 在非 darwin 平台返回不可用。
func collectMemory(context.Context, versionmanager.Runner) domain.Metric[domain.MemInfo] {
	return domain.Unavailable[domain.MemInfo]()
}

// collectDisks 在非 darwin 平台返回不可用。
func collectDisks(context.Context, versionmanager.Runner) domain.Metric[[]domain.DiskInfo] {
	return domain.Unavailable[[]domain.DiskInfo]()
}

// collectNetwork 在非 darwin 平台返回不可用。
func collectNetwork(context.Context, versionmanager.Runner) domain.Metric[[]domain.NetInfo] {
	return domain.Unavailable[[]domain.NetInfo]()
}
