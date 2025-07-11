package main

import (
	"log"

	"mplus.software/oss/bobbit.go/daemon"
	"mplus.software/oss/bobbit.go/payload"
)

func RouteHandler(d *daemon.DaemonStruct, p payload.JobPayload) {
	if p.Request == payload.EXECUTE_JOB {
		d.HandleJob(p)
		return
	}

	log.Printf("Outbound request database from job %s: requestIota=%v", p.ID, p.Request)
}
