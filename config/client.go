package config

import (
	"os"

	"github.com/mplus-oss/bobbit.go/internal/lib"
)

type BobbitClientConfig struct {
	SocketPath string
	DebugMode  bool
}

func NewClient() BobbitClientConfig {
	debug := false
	if env := os.Getenv("DEBUG"); env != "" {
		debug = true
	}

	return BobbitClientConfig{
		SocketPath: lib.GetDefaultEnv("BOBBIT_SOCKET_PATH", "/tmp/bobbitd.sock"),
		DebugMode:  debug,
	}
}
