package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/payload"
)

// BobbitConfig holds the configuration parameters for the Bobbit.
type BobbitConfig struct {
	// SocketPath specifies the file system path for the Unix domain socket.
	//
	// Default: `/tmp/bobbitd.sock`
	SocketPath string
	// DebugMode indicates whether the daemon should run in debug mode, enabling verbose logging.
	DebugMode bool
	// DataPath specifies the root directory for data storage.
	//
	// The directory stores: `metadata.db` that stores job status and metadata; `logs/YYYY/MM/*.log`
	// that stores logfile. Typically the logfile filename is random 64-bit hash pointer in the
	// metadata database.
	//
	// - For daemon: REQUIRED. Stores metadata.db and logs/
	//
	// - For client: OPTIONAL. Enables local log access if client shares filesystem with daemon.
	//
	// Default: `/tmp/bobbitd/`
	DataPath string
}

// New creates and initializes a new BobbitConfig instance.
// It retrieves configuration values from environment variables or uses default paths.
// DebugMode is enabled if the "DEBUG" environment variable is set to any non-empty value.
func BaseConfig() BobbitConfig {
	debug := false
	if env := os.Getenv("DEBUG"); env != "" {
		debug = true
	}

	return BobbitConfig{
		DataPath:   lib.GetDefaultEnv("BOBBIT_DATA_DIR", "/tmp/bobbitd/"),
		SocketPath: lib.GetDefaultEnv("BOBBIT_SOCKET_PATH", "/tmp/bobbitd.sock"),
		DebugMode:  debug,
	}
}

// GenerateJobLogPath will generate full path of log path.
// It automatically creates the parent directories if they do not exist.
func GenerateJobLogPath(c BobbitConfig, p payload.JobDetailMetadata) string {
	fullPath := filepath.Join(
		c.DataPath, "logs",
		strconv.Itoa(p.CreatedAt.Year()),
		fmt.Sprintf("%02d", p.CreatedAt.Month()),
		p.ID,
	)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}

	return fullPath
}
