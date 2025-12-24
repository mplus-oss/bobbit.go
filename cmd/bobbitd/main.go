package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/daemon"
	"github.com/spf13/cobra"
)

var (
	c       = config.NewDaemon()
	sigChan = make(chan os.Signal, 1)
	cmd     = &cobra.Command{
		Use:   "bobbitd",
		Short: "Daemon worker for bobbit.",
	}
)

func init() {
	cmd.Run = func(cmd *cobra.Command, args []string) {
		log.Printf("Directory data: %s", c.DataPath)
		log.Printf("Socket Path: %s", c.SocketPath)

		startDaemon()
	}
}

func startDaemon() {
	d, err := daemon.CreateDaemon(c)
	if err != nil {
		log.Fatalln(err)
	}

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go d.CleanupDaemon(sigChan)
	log.Println("Daemon started, waiting for response.")

	for {
		conn, err := d.SocketListener.Accept()
		if err != nil {
			log.Printf("Failed to receive connection: %v", err)
			continue
		}

		go handleConnection(d, conn)
	}
}

func handleConnection(d *daemon.DaemonStruct, conn net.Conn) {
	defer conn.Close()

	jobCtx := d.NewJobContext(conn)
	defer jobCtx.Close()

	if err := jobCtx.GetPayload(); err != nil {
		log.Println(err)
		return
	}

	RouteHandler(d, jobCtx)
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(100)
	}
}
