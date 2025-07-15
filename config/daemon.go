package config

import "github.com/mplus-oss/bobbit.go/internal/lib"

type BobbitDaemonConfig struct {
	DataDir string
	BobbitClientConfig
}

func NewDaemon() BobbitDaemonConfig {
	return BobbitDaemonConfig{
		DataDir:            lib.GetDefaultEnv("BOBBIT_DATA_DIR", "/tmp/bobbitd"),
		BobbitClientConfig: NewClient(),
	}
}
