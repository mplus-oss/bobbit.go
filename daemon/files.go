package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

// GenerateJobLogPath will generate full path of log path.
// It automatically creates the parent directories if they do not exist.
func GenerateJobLogPath(c config.BobbitDaemonConfig, p payload.JobDetailMetadata) string {
	fullPath := filepath.Join(
		c.DataPath, "logs",
		strconv.Itoa(p.CreatedAt.Year()),
		fmt.Sprintf("%02d", p.CreatedAt.Month()),
		p.ID,
		string(DAEMON_LOGFILE),
	)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}

	return fullPath
}
