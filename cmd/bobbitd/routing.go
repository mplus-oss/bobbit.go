package main

import (
	"log"

	"mplus.software/oss/bobbit.go/daemon"
	"mplus.software/oss/bobbit.go/payload"
)

func RouteHandler(d *daemon.DaemonStruct, jc *daemon.JobContext) {
	if jc.Payload.Request == payload.REQUEST_EXECUTE_JOB {
		if err := d.HandleJob(jc); err != nil {
			log.Println(err)
		}
		return
	}

	if jc.Payload.Request == payload.REQUEST_LIST {
		if err := d.ListJob(jc); err != nil {
			log.Println(err)
		}
		return
	}

	log.Printf("WARNING: Outbound request: %+v", jc)
}
