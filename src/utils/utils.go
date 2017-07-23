package utils

import (
	"../types"
)

// PanicOnError panics if err is not nil
func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// ExtractContainerResourceUsage parses a ContainerStatsRaw object and extracts a ContainerResourceUsage object
func ExtractContainerResourceUsage(stats types.ContainerStatsRaw) types.ContainerResourceUsage {
	cpu := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemCPUDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)

	if cpuDelta > 0.0 && systemCPUDelta > 0.0 {
		cpu = (cpuDelta / systemCPUDelta) * float64(len(stats.CPUStats.CPUUsage.PerCPUUsage)) * 100.0
	}

	return types.ContainerResourceUsage{
		CPU: cpu,
	}
}
