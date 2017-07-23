package types

// ContainerStatsRaw is the object that Docker exposes in the stats stream of a container
type ContainerStatsRaw struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PreCPUStats struct {
		CPUUsage struct {
			PerCPUUsage       []int64 `json:"percpu_usage"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelMode int64   `json:"usage_in_kernelmode"`
			UsageInUserMode   int64   `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		OnlineCPUs     int64 `json:"online_cpus"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	CPUStats struct {
		CPUUsage struct {
			PerCPUUsage       []int64 `json:"percpu_usage"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelMode int64   `json:"usage_in_kernelmode"`
			UsageInUserMode   int64   `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		OnlineCPUs     int64 `json:"online_cpus"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	MemoryStats struct {
		Limit    int64 `json:"limit"`
		MaxUsage int64 `json:"max_usage"`
		Stats    struct {
			RSS int64 `json:"rss"`
		} `json:"stats"`
	} `json:"memory_stats"`
}

// ContainerResourceUsage represents the current container resource usage on a running swarm cluster node
type ContainerResourceUsage struct {
	CPU    float64
	Memory float64
}

// ContainerStats is the object that Docker exposes in the stats stream of a container
type ContainerStats struct {
	Raw   ContainerStatsRaw
	Usage ContainerResourceUsage
}
