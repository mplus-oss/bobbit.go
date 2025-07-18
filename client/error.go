package client

import (
	"fmt"

	"github.com/mplus-oss/bobbit.go/config"
)

type DaemonNotRunningError struct {
	NetError error
	Config   config.BobbitConfig
}

func (d *DaemonNotRunningError) Error() string {
	return fmt.Sprintf("Cannot connect to %s. Is bobbitd running? Error: %v", d.Config.SocketPath, d.NetError)
}
