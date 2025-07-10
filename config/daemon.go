package config

import (
	"path/filepath"

	"mplus.software/oss/bobbit.go/internal/lib"
)

type BobbitDaemonConfig struct {
	DataDir    string
	BobbitClientConfig
}

func NewDaemon() BobbitDaemonConfig {
	return BobbitDaemonConfig{
		DataDir:            lib.GetDefaultEnv("BOBBIT_DATA_DIR", "/tmp/bobbitd"),
		BobbitClientConfig: NewClient(),
	}
}

func (c BobbitDaemonConfig) GetLockfilePath(id string) string {
	return filepath.Join(c.DataDir, id+".lock")
}

func (c BobbitDaemonConfig) GetLogfilePath(id string) string {
	return filepath.Join(c.DataDir, id+".log")
}

func (c BobbitDaemonConfig) GetExitCodePath(id string) string {
	return filepath.Join(c.DataDir, id+".exitcode")
}
