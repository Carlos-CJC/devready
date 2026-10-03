package domain

// ToolVersion 描述一个开发工具的版本与管理状态。
type ToolVersion struct {
	Name      string   // 展示名,如 "Go"
	Command   string   // 命令名,如 "go"
	Manager   string   // 版本管理器:"asdf" / "system" / "unknown"
	Path      string   // 命令解析出的绝对路径,"" 表示未找到
	Available bool     // 命令是否可实际执行
	Active    string   // 当前生效版本,如 "1.22.0";"" 表示未激活
	Source    string   // 激活版本的来源文件(如 .tool-versions 路径)
	Installed []string // 已知已安装的版本列表
}
