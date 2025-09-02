package containerd

// Config holds containerD client configuration
type Config struct {
	Socket    string `mapstructure:"socket"`
	Namespace string `mapstructure:"namespace"`
}

// DefaultConfig returns the default containerD configuration
func DefaultConfig() Config {
	return Config{
		Socket:    "/run/containerd/containerd.sock",
		Namespace: "scheduler",
	}
}
