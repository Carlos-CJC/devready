// Package services 侦测本机服务与监听端口(见设计文档 §7)。
//
// 解析逻辑为纯函数,便于单测;命令执行通过 versionmanager.Runner 注入。
package services

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/Carlos-CJC/devready/internal/domain"
	"github.com/Carlos-CJC/devready/internal/versionmanager"
)

// ListeningPorts 列出本机监听中的 TCP 端口。
// lsof 在无匹配结果时以非零码退出且输出为空,此时视为"无监听端口"而非错误。
func ListeningPorts(ctx context.Context, run versionmanager.Runner) ([]domain.PortInfo, error) {
	if _, err := exec.LookPath("lsof"); err != nil {
		return nil, err
	}
	out, err := run(ctx, "lsof", "-nP", "+c", "0", "-iTCP", "-sTCP:LISTEN")
	if err != nil && strings.TrimSpace(out) == "" {
		return nil, nil
	}
	return ParseLsof(out), nil
}

// ParseLsof 解析 `lsof -nP +c 0 -iTCP -sTCP:LISTEN` 输出。
//
//	COMMAND     PID USER   FD   TYPE  ... NODE NAME
//	node      14635  cjc   20u  IPv4  ... TCP 127.0.0.1:52382 (LISTEN)
func ParseLsof(out string) []domain.PortInfo {
	var ports []domain.PortInfo
	seen := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 9 || f[0] == "COMMAND" {
			continue
		}
		name := f[len(f)-1]
		if strings.HasPrefix(name, "(") { // 末尾的 "(LISTEN)" 等状态列
			name = f[len(f)-2]
		}
		host, portStr := splitAddr(name)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		// 同一进程同一地址的 IPv4/IPv6 两条记录合并为一条。
		key := fmt.Sprintf("%s|%s|%d", f[0], host, port)
		if seen[key] {
			continue
		}
		seen[key] = true

		pid, _ := strconv.Atoi(f[1])
		ports = append(ports, domain.PortInfo{
			Port:    port,
			Proto:   "TCP",
			Address: host,
			Process: f[0],
			PID:     pid,
			Docker:  isDocker(f[0]),
		})
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
	return ports
}

// splitAddr 拆分 "127.0.0.1:52382" 形式为地址与端口。IPv6 形如 "[::1]:80"。
func splitAddr(s string) (host, port string) {
	i := strings.LastIndexByte(s, ':')
	if i < 0 {
		return "", ""
	}
	return s[:i], s[i+1:]
}

// isDocker 判断进程是否为 Docker 的端口转发组件。
func isDocker(cmd string) bool {
	c := strings.ToLower(cmd)
	return strings.Contains(c, "docker") || strings.Contains(c, "vpnkit")
}
