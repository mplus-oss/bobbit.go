package config

import (
	"strconv"

	"github.com/mplus-oss/bobbit.go/internal/lib"
)

type BobbitDaemonConfig struct {
	// DataPath specifies the root directory where Bobbit Daemon working directory.
	//
	// The directory stores: `metadata.db` that stores job status and metadata; `logs/YYYY/MM/*.log`
	// that stores logfile. Typically the logfile filename is random 64-bit hash pointer in the
	// metadata database.
	DataPath string
	// SetMaxOpenConns sets the maximum number of open connections to the database. The default is 1.
	//
	// Check `(sql.DB).SetMaxOpenConns` for more information.
	DBMaxOpenConn int
	// SetMaxOpenConns sets the maximum number of open connections to the database. The default is 1.
	//
	// Check `(sql.DB).SetMaxIdleConns` for more information.
	DBMaxIdleConn int
	// BobbitConfig holds the configuration parameters for the Bobbit daemon.
	BobbitConfig
}

// NewDaemon creates and initializes a new BobbitConfig instance for Bobbit daemon.
func NewDaemon() BobbitDaemonConfig {
	maxOpenConn, err := strconv.Atoi(lib.GetDefaultEnv("BOBBITD_DB_MAX_OPEN_CONN", "1"))
	if err != nil {
		maxOpenConn = 1
	}
	maxIdleConn, err := strconv.Atoi(lib.GetDefaultEnv("BOBBITD_DB_MAX_IDLE_CONN", "1"))
	if err != nil {
		maxIdleConn = 1
	}

	return BobbitDaemonConfig{
		DataPath:      lib.GetDefaultEnv("BOBBITD_LOGS_DIR", "/tmp/bobbitd/"),
		DBMaxOpenConn: maxOpenConn,
		DBMaxIdleConn: maxIdleConn,
		BobbitConfig:  BaseConfig(),
	}
}
