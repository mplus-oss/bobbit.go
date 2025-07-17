package client

import (
	"encoding/json"
	"net"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

type DaemonConnectionStruct struct {
	Connection net.Conn
	config.BobbitClientConfig
}

func New(c config.BobbitClientConfig) *DaemonConnectionStruct {
	return &DaemonConnectionStruct{
		BobbitClientConfig: c,
	}
}

func (d *DaemonConnectionStruct) BuildPayload(p *payload.JobPayload, metadata any) error {
	conn, err := net.Dial("unix", d.BobbitClientConfig.SocketPath)
	if err != nil {
		conn.Close()
		return err
	}
	d.Connection = conn

	if err := p.MarshalMetadata(metadata); err != nil {
		d.Connection.Close()
		return err
	}
	
	return nil
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
