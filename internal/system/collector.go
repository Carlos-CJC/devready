// Package system 采集本机系统指标(CPU / 内存 / 磁盘 / 网络)。
//
// 解析逻辑集中在 parse.go(纯函数,跨平台可单测);平台相关的命令执行放在
// *_darwin.go,其余平台由 unsupported.go 诚实降级为 Unavailable(见设计文档 §2.2)。
package system

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Carlos-CJC/devscope/internal/domain"
	"github.com/Carlos-CJC/devscope/internal/versionmanager"
)

// collectTimeout 限制单个采集器的最长耗时,避免任一指标卡住整个 Dashboard。
const collectTimeout = 3 * time.Second

// Collect 并发采集本机系统状态。任一采集器失败只影响自身(呈现为 SKIP/WARN),不阻塞整体。
func Collect(ctx context.Context, run versionmanager.Runner) domain.SystemStatus {
	var st domain.SystemStatus
	if h, err := os.Hostname(); err == nil {
		st.Hostname = shortHost(h)
	}

	var wg sync.WaitGroup
	each := func(f func(context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cancel := context.WithTimeout(ctx, collectTimeout)
			defer cancel()
			f(c)
		}()
	}
	each(func(c context.Context) { st.CPU = collectCPU(c, run) })
	each(func(c context.Context) { st.Memory = collectMemory(c, run) })
	each(func(c context.Context) { st.Disks = collectDisks(c, run) })
	each(func(c context.Context) { st.Network = collectNetwork(c, run) })
	wg.Wait()

	return st
}

// shortHost 去掉 "." 之后的域名后缀,便于在标题栏展示。
func shortHost(h string) string {
	if i := strings.IndexByte(h, '.'); i > 0 {
		return h[:i]
	}
	return h
}
