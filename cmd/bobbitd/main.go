package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/daemon"
)

var (
	c       = config.New()
	sigChan = make(chan os.Signal, 1)
)

func main() {
	log.Printf("Directory data: %s", c.DataDir)
	log.Printf("Socket Path: %s", c.SocketPath)

	d, err := daemon.CreateDaemon(c)
	if err != nil {
		log.Fatalln(err)
	}

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go d.CleanupDaemon(sigChan)
	log.Println("Daemon started, waiting for response.")

	for {
		conn, err := d.SocketListener.Accept()
		if err != nil {
			log.Printf("Failed to receive connection: %v", err)
			continue
		}
		defer conn.Close()

		go handleConnection(d, conn)
	}
}

func handleConnection(d *daemon.DaemonStruct, conn net.Conn) {
	jobCtx := d.NewJobContext(conn)
	defer jobCtx.Close()

	if err := jobCtx.GetPayload(); err != nil {
		log.Println(err)
		return
	}
	log.Printf("Job received: %+v", jobCtx.Payload)

	RouteHandler(d, jobCtx)
}
