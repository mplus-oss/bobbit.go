package main

import (
	"encoding/json"
	"log"
	"net"

	"mplus.software/oss/bobbit.go/internal/config"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	var payload config.JobPayload

	if err := json.NewDecoder(conn).Decode(&payload); err != nil {
		log.Printf("ERROR: Failed to decode payload: %v", err)
		return
	}
	if payload.ID == "" || len(payload.Command) < 1 {
		log.Printf("ERROR: Invalid payload: ID or Command not provided.")
		return
	}

	log.Printf("Job received: id=%s command=%s", payload.ID, payload.Command)
	go HandleJob(payload)
}
