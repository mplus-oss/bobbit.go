package config

import (
	"path/filepath"

	"mplus.software/oss/bobbit.go/internal/lib"
)

type BobbitConfig struct {
	FifoPath   string
	SocketPath string
	DataDir    string
	Separator  string
}

func New() BobbitConfig {
	return BobbitConfig{
		FifoPath:   lib.GetDefaultEnv("BOBBIT_FIFO_PATH", "/tmp/bobbitd.fifo"),
		SocketPath: lib.GetDefaultEnv("BOBBIT_SOCKET_PATH", "/tmp/bobbitd.sock"),
		DataDir:    lib.GetDefaultEnv("BOBBIT_DATA_DIR", "/tmp/bobbitd"),
		Separator:  lib.GetDefaultEnv("BOBBIT_SEPARATOR", "---"),
	}
}

func (c BobbitConfig) GetLockfilePath(id string) string {
	return filepath.Join(c.DataDir, id+".lock")
}
func (c BobbitConfig) GetLogfilePath(id string) string {
	return filepath.Join(c.DataDir, id+".log")
}
func (c BobbitConfig) GetExitCodePath(id string) string {
	return filepath.Join(c.DataDir, id+".exitcode")
}
