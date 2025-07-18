package daemon

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

type DaemonFileTypeEnum string

const (
	DAEMON_LOCKFILE DaemonFileTypeEnum = "lockfile"
	DAEMON_LOGFILE  DaemonFileTypeEnum = "log"
	DAEMON_METADATA DaemonFileTypeEnum = "metadata"
	DAEMON_EXITCODE DaemonFileTypeEnum = "exitcode"
)

type DaemonStruct struct {
	SocketListener net.Listener
	config.BobbitConfig
}

type JobContext struct {
	conn    net.Conn
	daemon  *DaemonStruct
	Payload payload.JobPayload
}

func CreateDaemon(c config.BobbitConfig) (*DaemonStruct, error) {
	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		return nil, &DaemonError{"Failed to create data directory: %v", err}
	}

	if err := os.RemoveAll(c.SocketPath); err != nil {
		return nil, &DaemonError{"Failed to remove old socket path: %v", err}
	}

	listener, err := net.Listen("unix", c.SocketPath)
	if err != nil {
		return nil, &DaemonError{"Failed to listen in socket path: %v", err}
	}

	return &DaemonStruct{
		SocketListener: listener,
		BobbitConfig:   c,
	}, nil
}

func (d *DaemonStruct) NewJobContext(conn net.Conn) *JobContext {
	return &JobContext{
		conn:   conn,
		daemon: d,
	}
}

func (d *DaemonStruct) CleanupDaemon(sigChan <-chan os.Signal) {
	<-sigChan
	log.Println("Cleanup daemon...")
	os.Remove(d.SocketPath)
	os.Exit(0)
}

func (jc *JobContext) GetPayload() error {
	var p payload.JobPayload
	if err := json.NewDecoder(jc.conn).Decode(&p); err != nil {
		return &DaemonError{"Failed to decode payload.", err}
	}
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now()
	}

	jc.Payload = p
	return nil
}

func (jc *JobContext) SendPayload(target any) error {
	if err := json.NewEncoder(jc.conn).Encode(target); err != nil {
		return err
	}
	return nil
}

func (jc *JobContext) Close() {
	jc.conn.Close()
}
