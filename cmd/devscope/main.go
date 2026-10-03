package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cjc/devscope/internal/environment"
	"github.com/cjc/devscope/internal/versionmanager"
)

func main() {
	ctx := context.Background()
	dir, _ := os.Getwd()

	info := environment.DetectGo(ctx, versionmanager.ExecRunner, dir)

	fmt.Println("DevScope · 开发环境")
	fmt.Println()
	fmt.Println("Go")
	printField("管理器", dash(info.Manager))
	printField("路径", dash(info.Path))
	printField("激活版本", dash(info.Active))
	printField("来源", dash(info.Source))
	printField("已安装", dash(strings.Join(info.Installed, ", ")))
	printField("可执行", yesNo(info.Available))
	if info.Required != "" {
		printField("项目要求", info.Required+" (来自 go.mod)")
	}
}

const labelWidth = 10 // 以终端显示宽度计(中文按 2 列)

func printField(label, value string) {
	pad := labelWidth - displayWidth(label)
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("  %s%s  %s\n", label, strings.Repeat(" ", pad), value)
}

// displayWidth 估算字符串的终端显示宽度,全角/CJK 字符按 2 列计。
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x2E80 {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
