package domain

// PortInfo 是一个监听中的端口及其归属进程(见设计文档 §7)。
type PortInfo struct {
	Port    int    // 监听端口号
	Proto   string // 传输层协议,如 "TCP"
	Address string // 监听地址,如 "*" / "127.0.0.1"
	Process string // 归属进程名;无法识别时为空
	PID     int    // 进程号
	Docker  bool   // 是否为 Docker 端口映射
}

// ServiceInfo 是一个本机服务的运行状态。
type ServiceInfo struct {
	Name    string // 服务名,如 "sshd"
	Detail  string // 补充说明,如 ":22" / "未运行"
	Running bool
}
