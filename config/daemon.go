package config

import (
	"fmt"
	"path/filepath"
	"time"

	"mplus.software/oss/bobbit.go/internal/lib"
	"mplus.software/oss/bobbit.go/payload"
)

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

func (c BobbitDaemonConfig) GetLockfilePath(p payload.JobPayload) string {
	return filepath.Join(c.DataDir, generateFilePath(p, "lock"))
}

func (c BobbitDaemonConfig) GetLogfilePath(p payload.JobPayload) string {
	return filepath.Join(c.DataDir, generateFilePath(p, "log"))
}

func (c BobbitDaemonConfig) GetExitCodePath(p payload.JobPayload) string {
	return filepath.Join(c.DataDir, generateFilePath(p, "exitcode"))
}

func (c BobbitDaemonConfig) GetMetadataPath(p payload.JobPayload) string {
	return filepath.Join(c.DataDir, generateFilePath(p, "metadata"))
}

func generateFilePath(p payload.JobPayload, execPath string) string {
	return fmt.Sprintf("%s-%s.%s", p.Timestamp.Format(time.RFC3339), p.ID, execPath)
}
