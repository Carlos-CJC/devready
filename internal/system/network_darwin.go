//go:build darwin

package system

import (
	"context"
	"sort"
	"time"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// netInterval 是两次采样的间隔,用于把累计字节数换算成速率。
const netInterval = 500 * time.Millisecond

// collectNetwork 对 `netstat -ib` 采样两次求差,得到各接口的实时收发速率。
func collectNetwork(ctx context.Context, run versionmanager.Runner) domain.Metric[[]domain.NetInfo] {
	firstOut, err := run(ctx, "netstat", "-ib")
	if err != nil {
		return domain.Failed[[]domain.NetInfo](err)
	}
	first := ParseNetstat(firstOut)

	select {
	case <-ctx.Done():
		return domain.Failed[[]domain.NetInfo](ctx.Err())
	case <-time.After(netInterval):
	}

	secondOut, err := run(ctx, "netstat", "-ib")
	if err != nil {
		return domain.Failed[[]domain.NetInfo](err)
	}
	second := ParseNetstat(secondOut)

	var nets []domain.NetInfo
	for name, now := range second {
		before, ok := first[name]
		if !ok {
			continue
		}
		rx := byteRate(now[0], before[0])
		tx := byteRate(now[1], before[1])
		if rx == 0 && tx == 0 {
			continue
		}
		nets = append(nets, domain.NetInfo{Name: name, RxBps: rx, TxBps: tx})
	}
	sort.Slice(nets, func(i, j int) bool {
		return nets[i].RxBps+nets[i].TxBps > nets[j].RxBps+nets[j].TxBps
	})
	return domain.OK(nets)
}

// byteRate 把两次采样间的字节差换算为字节/秒;计数器回绕时返回 0。
func byteRate(now, before uint64) uint64 {
	if now < before {
		return 0
	}
	return uint64(float64(now-before) / netInterval.Seconds())
}
