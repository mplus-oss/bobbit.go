package client

import (
	"encoding/json"
	"net"

	"mplus.software/oss/bobbit.go/config"
	"mplus.software/oss/bobbit.go/payload"
)

type DaemonConnectionStruct struct {
	Connection net.Conn
}

func CreateConnection(c config.BobbitClientConfig) (*DaemonConnectionStruct, error) {
	conn, err := net.Dial("unix", c.SocketPath)
	if err != nil {
		return &DaemonConnectionStruct{}, &DaemonNotRunningError{err, c}
	}

	return &DaemonConnectionStruct{Connection: conn}, nil
}

func (d *DaemonConnectionStruct) SendPayload(payload payload.JobPayload) error {
	if err := json.NewEncoder(d.Connection).Encode(payload); err != nil {
		return err
	}
	return nil
}
