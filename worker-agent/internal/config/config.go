package config

// Config holds worker agent configuration
type Config struct {
	APIServer     string
	WorkerName    string
	MaxConcurrent int
	Labels        map[string]string
	IsolationType string

	// System info (auto-detected)
	CPUCores  int
	MemoryMB  int
	Hostname  string
	IPAddress string
}
