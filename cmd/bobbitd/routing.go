package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/payload"
)

// RouteHandlerMap handling job route for (daemon.JobContext).Handler
//
// Some job don't need to verbose every single action. This map help the middleware to make
// route discovery more easier.
//
// See: const main.RunJob:ignoredRoutes
type RouteHandlerMap map[payload.PayloadRequestEnum]func(*daemon.JobContext) error

func RouteHandler(d *daemon.DaemonStruct, jc *daemon.JobContext) {
	handlers := RouteHandlerMap{
		payload.REQUEST_VIBE_CHECK:  d.HandleVibeCheck,
		payload.REQUEST_LIST:        d.ListJob,
		payload.REQUEST_EXECUTE_JOB: d.HandleJob,
		payload.REQUEST_WAIT:        d.WaitJob,
		payload.REQUEST_STATUS:      d.StatusJob,
		payload.REQUEST_STOP:        d.StopJob,
	}

	var (
		err   error
		hName string
	)

	h, exists := handlers[jc.Payload.Request]
	if exists {
		hName = payload.ParsePayloadRequest(jc.Payload.Request)
		err = RunJob(d, jc, hName, h)
	} else {
		err = fmt.Errorf("Outbound request: %v", jc.Payload.Request)
	}

	if err != nil {
		log.Printf("Error processing %v: %v", hName, err)
		if sendErr := jc.SendPayload(payload.JobErrorResponse{Error: err.Error()}); sendErr != nil {
			log.Println("Failed to send error response:", sendErr)
		}
	}
}

func RunJob(d *daemon.DaemonStruct, jc *daemon.JobContext, name string, handler daemon.HandlerFunc) error {
	// Ignore this route from log if the app is not on DebugMode
	const ignoredRoutes = payload.REQUEST_LIST | payload.REQUEST_VIBE_CHECK
	shouldLog := d.DebugMode || !((ignoredRoutes & jc.Payload.Request) > 0)

	// Add hash for logging
	hash, _ := lib.GenerateRandomHash(8)

	if shouldLog {
		log.Printf("[%s] Entering Context: %s: %+v", hash, name, jc.Payload)
	}

	start := time.Now()
	err := handler(jc)

	if shouldLog {
		duration := time.Since(start)
		if err != nil {
			log.Printf("[%s] FAILED: %s | Took: %v | Error: %v", hash, name, duration, err)
		} else {
			log.Printf("[%s] DONE: %s | Took: %v", hash, name, duration)
		}
	}

	return err
}
