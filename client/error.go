package client

import (
	"fmt"

	"mplus.software/oss/bobbit.go/config"
)

type DaemonNotRunningError struct {
	NetError error
	Config   config.BobbitClientConfig
}

func (d *DaemonNotRunningError) Error() string {
	return fmt.Sprintf("Cannot connect to %s. Is bobbitd running? Error: %v", d.Config.SocketPath, d.NetError)
}
