package services

import (
	"os"

	"github.com/Carlos-CJC/devscope/internal/domain"
)

// dockerSock 是 Docker 守护进程的默认套接字;存在即可判定 Docker 正在运行。
const dockerSock = "/var/run/docker.sock"

// Detect 依据监听端口与本地事实,判断若干常见本机服务是否在运行(§7)。
func Detect(ports []domain.PortInfo) []domain.ServiceInfo {
	return []domain.ServiceInfo{
		detectSSHD(ports),
		detectDocker(),
	}
}

// detectSSHD 以 22 端口是否被监听判断本机 sshd 是否在跑(与远程访问无关)。
func detectSSHD(ports []domain.PortInfo) domain.ServiceInfo {
	for _, p := range ports {
		if p.Port == 22 {
			return domain.ServiceInfo{Name: "sshd", Detail: ":22", Running: true}
		}
	}
	return domain.ServiceInfo{Name: "sshd", Detail: "未监听"}
}

// detectDocker 以 Docker 套接字是否存在判断守护进程是否在运行。
func detectDocker() domain.ServiceInfo {
	if _, err := os.Stat(dockerSock); err == nil {
		return domain.ServiceInfo{Name: "Docker", Detail: "running", Running: true}
	}
	return domain.ServiceInfo{Name: "Docker", Detail: "未运行"}
}
