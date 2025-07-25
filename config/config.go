package config

import (
	"os"

	"github.com/mplus-oss/bobbit.go/internal/lib"
)

// BobbitConfig holds the configuration parameters for the Bobbit daemon.
type BobbitConfig struct {
	// SocketPath specifies the file system path for the Unix domain socket.
	SocketPath string
	// DataDir specifies the directory where job-related data (logs, metadata, lock files) are stored.
	DataDir string
	// DebugMode indicates whether the daemon should run in debug mode, enabling verbose logging.
	DebugMode bool
}

// New creates and initializes a new BobbitConfig instance.
// It retrieves configuration values from environment variables or uses default paths.
// Specifically, it looks for "BOBBIT_DATA_DIR" for the data directory
// and "BOBBIT_SOCKET_PATH" for the socket path.
// DebugMode is enabled if the "DEBUG" environment variable is set to any non-empty value.
func New() BobbitConfig {
	debug := false
	if env := os.Getenv("DEBUG"); env != "" {
		debug = true
	}

	return BobbitConfig{
		DataDir:    lib.GetDefaultEnv("BOBBIT_DATA_DIR", "/tmp/bobbitd"),
		SocketPath: lib.GetDefaultEnv("BOBBIT_SOCKET_PATH", "/tmp/bobbitd.sock"),
		DebugMode:  debug,
	}
}
