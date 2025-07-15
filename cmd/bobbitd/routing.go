package main

import (
	"log"

	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/mplus-oss/bobbit.go/payload"
)

func RouteHandler(d *daemon.DaemonStruct, jc *daemon.JobContext) {
	switch r := jc.Payload.Request; r {
	case payload.REQUEST_EXECUTE_JOB:
		if err := d.HandleJob(jc); err != nil {
			log.Println(err)
		}
	case payload.REQUEST_LIST:
		if err := d.ListJob(jc); err != nil {
			log.Println(err)
		}
	case payload.REQUEST_WAIT:
		if err := d.WaitJob(jc); err != nil {
			log.Println(err)
		}
	default:
		log.Printf("WARNING: Outbound request: %+v", jc)
	}
}
