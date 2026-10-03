//go:build darwin

package system

import (
	"context"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// collectDisks 通过 `df -k -P` 列出本地磁盘容量。
func collectDisks(ctx context.Context, run versionmanager.Runner) domain.Metric[[]domain.DiskInfo] {
	out, err := run(ctx, "df", "-k", "-P")
	if err != nil {
		return domain.Failed[[]domain.DiskInfo](err)
	}
	disks := ParseDF(out)
	if len(disks) == 0 {
		return domain.Unavailable[[]domain.DiskInfo]()
	}
	return domain.OK(disks)
}
