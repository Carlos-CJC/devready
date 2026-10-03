package domain

// MetricStatus 描述一次采集的结果状态(见设计文档 §2.1)。
type MetricStatus int

const (
	MetricOK          MetricStatus = iota // 采集成功
	MetricUnavailable                     // 平台不支持 / 无法采集 → 呈现为 SKIP
	MetricError                           // 采集出错 → 呈现为 WARN
)

// Metric 是采集器返回的"值 + 状态"。UI 只渲染它,不判断平台是否支持。
type Metric[T any] struct {
	Value  T
	Status MetricStatus
	Err    error
}

// OK 构造一个成功采集的 Metric。
func OK[T any](v T) Metric[T] { return Metric[T]{Value: v, Status: MetricOK} }

// Unavailable 构造一个"平台无法提供"的 Metric。
func Unavailable[T any]() Metric[T] { return Metric[T]{Status: MetricUnavailable} }

// Failed 构造一个"采集出错"的 Metric。
func Failed[T any](err error) Metric[T] { return Metric[T]{Status: MetricError, Err: err} }

// Usable 表示该指标是否有可展示的值。
func (m Metric[T]) Usable() bool { return m.Status == MetricOK }

// CPUInfo 是 CPU 的概要信息。
type CPUInfo struct {
	Model       string  // 型号,如 "Apple M4"
	Cores       int     // 物理核心数
	Load1       float64 // 1 分钟平均负载
	Load5       float64 // 5 分钟平均负载
	Load15      float64 // 15 分钟平均负载
	Utilization float64 // 总利用率百分比 0-100
}

// MemInfo 是物理内存使用情况,单位字节。
type MemInfo struct {
	Total     uint64
	Used      uint64
	Available uint64 // 可回收(空闲 + 非活跃 + 推测),used + available ≈ total
}

// UsedPercent 返回内存使用率百分比 0-100;总量为 0 时返回 0。
func (m MemInfo) UsedPercent() float64 { return percent(m.Used, m.Total) }

// DiskInfo 是一个挂载点的容量信息,单位字节。
type DiskInfo struct {
	MountPoint string
	Total      uint64
	Used       uint64
}

// UsedPercent 返回磁盘使用率百分比 0-100。
func (d DiskInfo) UsedPercent() float64 { return percent(d.Used, d.Total) }

// NetInfo 是一张网络接口的实时速率,单位字节/秒。
type NetInfo struct {
	Name  string
	RxBps uint64
	TxBps uint64
}

// SystemStatus 是 Machine Scope 一次采集的完整结果。
type SystemStatus struct {
	Hostname string
	CPU      Metric[CPUInfo]
	Memory   Metric[MemInfo]
	Disks    Metric[[]DiskInfo]
	Network  Metric[[]NetInfo]
}

func percent(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
