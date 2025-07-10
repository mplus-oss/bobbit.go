package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"mplus.software/oss/bobbit.go/internal/config"
)

var (
	c       = config.New()
	sigChan = make(chan os.Signal, 1)
)

func main() {
	log.Printf("Directory data: %s", c.DataDir)
	log.Printf("Socket Path: %s", c.SocketPath)

	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	if err := os.RemoveAll(c.SocketPath); err != nil {
		log.Fatalf("Failed to remove old socket path: %v", err)
	}

	listener, err := net.Listen("unix", c.SocketPath)
	if err != nil {
		log.Fatalf("Failed to listen in socket path: %v", err)
	}
	defer listener.Close()

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go CleanupDaemon()
	log.Println("Daemon started, waiting for response.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to receive connection: %v", err)
			continue
		}
		go HandleConnection(conn)
	}
}

func CleanupDaemon() {
	<-sigChan
	log.Println("Cleanup daemon...")
	os.Remove(c.SocketPath)
	os.Exit(0)
}
