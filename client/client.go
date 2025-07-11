package client

import (
	"encoding/json"
	"net"

	"mplus.software/oss/bobbit.go/config"
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

func (d *DaemonConnectionStruct) SendPayload(target any) error {
	if err := json.NewEncoder(d.Connection).Encode(target); err != nil {
		return err
	}
	return nil
}

func (d *DaemonConnectionStruct) GetPayload(target any) error {
	if err := json.NewDecoder(d.Connection).Decode(target); err != nil {
		return err
	}
	return nil
}
