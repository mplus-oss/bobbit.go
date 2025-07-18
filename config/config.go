package config

import (
	"os"

	"github.com/mplus-oss/bobbit.go/internal/lib"
)

type BobbitConfig struct {
	SocketPath string
	DataDir    string
	DebugMode  bool
}

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
