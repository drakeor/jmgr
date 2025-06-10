package alert

// Configuration for alert thresholds.
type Thresholds struct {
	CPUPercent      float32
	MemoryUsedBytes uint64
	DiskFreeBytes   uint64
}
