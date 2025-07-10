package daemon

import (
	"encoding/json"
	"net"

	"mplus.software/oss/bobbit.go/internal/config"
)

type DaemonConnectionStruct struct {
	Connection net.Conn
}

func CreateConnection(c config.BobbitConfig) (*DaemonConnectionStruct, error) {
	conn, err := net.Dial("unix", c.SocketPath)
	if err != nil {
		return &DaemonConnectionStruct{}, &DaemonNotRunningError{err, c}
	}

	return &DaemonConnectionStruct{Connection: conn}, nil
}

func (d *DaemonConnectionStruct) SendPayload(payload config.JobPayload) error {
	if err := json.NewEncoder(d.Connection).Encode(payload); err != nil {
		return err
	}
	return nil
}
