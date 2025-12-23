package config

import "github.com/mplus-oss/bobbit.go/internal/lib"

type BobbitDaemonConfig struct {
	// DataPath specifies the root directory where Bobbit Daemon working directory.
	//
	// The directory stores: `metadata.db` that stores job status and metadata; `logs/YYYY/MM/*.log`
	// that stores logfile. Typically the logfile filename is random 64-bit hash pointer in the
	// metadata database.
	DataPath string
	// BobbitConfig holds the configuration parameters for the Bobbit daemon.
	BobbitConfig
}

// NewDaemon creates and initializes a new BobbitConfig instance for Bobbit daemon.
func NewDaemon() BobbitDaemonConfig {
	return BobbitDaemonConfig{
		DataPath:     lib.GetDefaultEnv("BOBBITD_LOGS_DIR", "/tmp/bobbitd/"),
		BobbitConfig: BaseConfig(),
	}
}
