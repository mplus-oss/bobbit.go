package config

import (
	"os"

	"github.com/mplus-oss/bobbit.go/internal/lib"
)

// BobbitConfig holds the configuration parameters for the Bobbit.
type BobbitConfig struct {
	// SocketPath specifies the file system path for the Unix domain socket.
	SocketPath string
	// DebugMode indicates whether the daemon should run in debug mode, enabling verbose logging.
	DebugMode bool
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
		SocketPath: lib.GetDefaultEnv("BOBBIT_SOCKET_PATH", "/tmp/bobbitd.sock"),
		DebugMode:  debug,
	}
}
