package config

type BobbitClientConfig struct {
	// BobbitConfig holds the configuration parameters for the Bobbit daemon.
	BobbitConfig
}

// NewClient creates and initializes a new BobbitConfig instance for Bobbit client.
func NewClient() BobbitClientConfig {
	return BobbitClientConfig{
		BobbitConfig: BaseConfig(),
	}
}
