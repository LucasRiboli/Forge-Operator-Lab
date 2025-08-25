package config

type Config struct {
	HPA struct {
		MinPods         int `yaml:"minPods"`
		MaxPods         int `yaml:"maxPods"`
		CpuThreshold    int `yaml:"cpuThreshold"`
		MemoryThreshold int `yaml:"memoryThreshold"`
	} `yaml:"HPA"`
}
