package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/mplus-oss/bobbit.go/payload"
)

func RouteHandler(d *daemon.DaemonStruct, jc *daemon.JobContext) {
	handlers := map[payload.PayloadRequestEnum]struct {
		Name    string
		Handler func(*daemon.JobContext) error
	}{
		payload.REQUEST_VIBE_CHECK:  {"VIBE_CHECK", d.HandleVibeCheck},
		payload.REQUEST_LIST:        {"LIST", d.ListJob},
		payload.REQUEST_EXECUTE_JOB: {"EXECUTE", d.HandleJob},
		payload.REQUEST_WAIT:        {"WAIT", d.WaitJob},
		payload.REQUEST_STATUS:      {"STATUS", d.StatusJob},
		payload.REQUEST_STOP:        {"STOP", d.StopJob},
	}

	var err error

	h, exists := handlers[jc.Payload.Request]
	if exists {
		err = RunJob(d, jc, h.Name, h.Handler)
	} else {
		err = fmt.Errorf("Outbound request: %v", jc.Payload.Request)
	}

	if err != nil {
		log.Printf("Error processing %v: %v", h.Name, err)
		if sendErr := jc.SendPayload(payload.JobErrorResponse{Error: err.Error()}); sendErr != nil {
			log.Println("Failed to send error response:", sendErr)
		}
	}
}

func RunJob(d *daemon.DaemonStruct, jc *daemon.JobContext, name string, handler daemon.HandlerFunc) error {
	// Ignore this route from log if the app is not on DebugMode
	const ignoredRoutes = payload.REQUEST_LIST | payload.REQUEST_VIBE_CHECK
	shouldLog := d.DebugMode || !((ignoredRoutes & jc.Payload.Request) > 0)

	if shouldLog {
		log.Printf("Entering Context: %s: %+v", name, jc.Payload)
	}

	start := time.Now()
	err := handler(jc)

	if shouldLog {
		duration := time.Since(start)
		if err != nil {
			log.Printf("FAILED: %s | Took: %v | Error: %v", name, duration, err)
		} else {
			log.Printf("DONE: %s | Took: %v", name, duration)
		}
	}

	return err
}
