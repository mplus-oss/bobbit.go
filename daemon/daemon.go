package daemon

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/metadata"
	"github.com/mplus-oss/bobbit.go/payload"
)

// DaemonFileTypeEnum represents the type of data file associated with a daemon job.
type DaemonFileTypeEnum string

const (
	// DAEMON_LOCKFILE indicates a lock file, signifying a job is running.
	DAEMON_LOCKFILE DaemonFileTypeEnum = "lockfile"
	// DAEMON_LOGFILE indicates a log file for job output.
	DAEMON_LOGFILE DaemonFileTypeEnum = "log"
	// DAEMON_METADATA indicates a metadata file containing job details.
	DAEMON_METADATA DaemonFileTypeEnum = "metadata"
	// DAEMON_EXITCODE indicates a file containing the job's exit code.
	DAEMON_EXITCODE DaemonFileTypeEnum = "exitcode"
)

// DaemonStruct holds the main components of the Bobbit daemon, including
// its socket listener and configuration.
type DaemonStruct struct {
	SocketListener net.Listener
	DB             *sqlx.DB
	config.BobbitDaemonConfig
}

// JobContext holds the context for a single job request handled by the daemon,
// including the network connection and the job payload.
type JobContext struct {
	conn    net.Conn
	daemon  *DaemonStruct
	Payload payload.JobPayload
}

// CreateDaemon initializes and starts the daemon. It checks for existing daemon
// instances, creates necessary data directories, and sets up the Unix socket listener.
// It returns a pointer to a DaemonStruct or an error if initialization fails.
func CreateDaemon(c config.BobbitDaemonConfig) (*DaemonStruct, error) {
	if socket, err := os.Stat(c.SocketPath); err == nil {
		if socket.Mode().Type() == fs.ModeSocket {
			return nil, &DaemonError{"Daemon is already started", fmt.Errorf("Daemon found in %v", c.SocketPath)}
		}
	}

	if err := os.MkdirAll(c.DataPath, 0755); err != nil {
		return nil, &DaemonError{"Failed to create data directory", err}
	}

	db, err := metadata.InitDB(c)
	if err != nil {
		return nil, &DaemonError{"Failed to initialize database", err}
	}

	listener, err := net.Listen("unix", c.SocketPath)
	if err != nil {
		return nil, &DaemonError{"Failed to listen in socket path", err}
	}
	if err := os.RemoveAll(c.SocketPath); err != nil {
		return nil, &DaemonError{"Failed to remove old socket path", err}
	}

	return &DaemonStruct{
		SocketListener:     listener,
		DB:                 db,
		BobbitDaemonConfig: c,
	}, nil
}

// NewJobContext creates and returns a new JobContext for a given network connection.
// This context is used to manage a single job request.
func (d *DaemonStruct) NewJobContext(conn net.Conn) *JobContext {
	return &JobContext{
		conn:   conn,
		daemon: d,
	}
}

// CleanupDaemon listens for an OS signal and performs cleanup operations
// before exiting. It removes the daemon's socket file.
func (d *DaemonStruct) CleanupDaemon(sigChan <-chan os.Signal) {
	<-sigChan
	log.Println("Cleanup daemon...")
	os.Remove(d.SocketPath)
	os.Exit(0)
}

// GetPayload reads and decodes a JobPayload from the JobContext's network connection.
// It populates the Payload field of the JobContext and returns an error if decoding fails.
// If the payload's timestamp is zero, it defaults to the current time.
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

// SendPayload encodes and sends the given target object over the JobContext's
// network connection. It returns an error if encoding or writing fails.
func (jc *JobContext) SendPayload(target any) error {
	if err := json.NewEncoder(jc.conn).Encode(target); err != nil {
		return err
	}
	return nil
}

// Close closes the network connection associated with the JobContext.
func (jc *JobContext) Close() {
	jc.conn.Close()
}
