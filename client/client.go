package client

import (
	"encoding/json"
	"net"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

// DaemonConnectionStruct holds the network connection and configuration for a Bobbit daemon client.
type DaemonConnectionStruct struct {
	Connection net.Conn
	config.BobbitConfig
}

// New creates and returns a new DaemonConnectionStruct initialized with the provided BobbitConfig.
func New(c config.BobbitConfig) *DaemonConnectionStruct {
	return &DaemonConnectionStruct{
		BobbitConfig: c,
	}
}

// BuildPayload establishes a connection to the daemon's Unix socket and marshals the provided metadata
// into the JobPayload. It returns an error if the connection fails or metadata marshalling fails.
func (d *DaemonConnectionStruct) BuildPayload(p *payload.JobPayload, metadata any) (err error) {
	conn, err := net.Dial("unix", d.BobbitConfig.SocketPath)
	if err != nil {
		return err
	}
	d.Connection = conn

	if err := p.MarshalMetadata(metadata); err != nil {
		d.Connection.Close()
		return err
	}
	
	return nil
}

// SendPayload encodes and sends the given target object over the established daemon connection.
// It returns an error if encoding or writing fails.
func (d *DaemonConnectionStruct) SendPayload(target any) error {
	if err := json.NewEncoder(d.Connection).Encode(target); err != nil {
		return err
	}
	return nil
}

// GetPayload decodes the response from the daemon connection into the provided target object.
// It returns an error if decoding fails.
func (d *DaemonConnectionStruct) GetPayload(target any) error {
	if err := json.NewDecoder(d.Connection).Decode(target); err != nil {
		return err
	}
	return nil
}
