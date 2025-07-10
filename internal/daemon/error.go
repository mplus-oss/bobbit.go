package daemon

import (
	"fmt"

	"mplus.software/oss/bobbit.go/internal/config"
)

type DaemonNotRunningError struct{
	NetError error
	config.BobbitConfig
}

func (d *DaemonNotRunningError) Error() string {
	return fmt.Sprintf("Cannot connect to %s. Is bobbitd running? Error: %v", d.BobbitConfig.SocketPath, d.NetError)
}
