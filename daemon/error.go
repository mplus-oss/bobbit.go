package daemon

import (
	"fmt"
	"log"
)

type DaemonError struct{
	Message     string
	ParentError error
}

func (d *DaemonError) Error() string {
	return fmt.Sprintf("ERROR: %s: %v", d.Message, d.ParentError)
}

type DaemonPayloadError struct {
	Message     string
	JobID       string
	ParentError error
}

func (d *DaemonPayloadError) Error() string {
	return fmt.Sprintf("ERROR: [%s] %s: %v", d.JobID, d.Message, d.ParentError)
}

func (d *DaemonPayloadError) Warning() {
	log.Printf("WARNING: [%s] %s: %v", d.JobID, d.Message, d.ParentError)
}
