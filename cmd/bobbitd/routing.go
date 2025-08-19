package main

import (
	"fmt"
	"log"

	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/mplus-oss/bobbit.go/payload"
)

func RouteHandler(d *daemon.DaemonStruct, jc *daemon.JobContext) {
	var err error

	switch r := jc.Payload.Request; r {
	case payload.REQUEST_VIBE_CHECK:
		err = d.HandleVibeCheck(jc)

	case payload.REQUEST_EXECUTE_JOB:
		err = d.HandleJob(jc)

	case payload.REQUEST_LIST:
		err = d.ListJob(jc)

	case payload.REQUEST_WAIT:
		err = d.WaitJob(jc)

	case payload.REQUEST_STATUS:
		err = d.StatusJob(jc)

	case payload.REQUEST_STOP:
		err = d.StopJob(jc)

	default:
		err = fmt.Errorf("Outbound request: %v", jc)
	}

	if err == nil {
		return
	}
	if err := jc.SendPayload(payload.JobErrorResponse{Error: err.Error()}); err != nil {
		log.Println(err)
	}
}
