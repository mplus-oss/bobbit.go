package main

import (
	"os"

	"github.com/mplus-oss/bobbit.go/client"
	"github.com/mplus-oss/bobbit.go/config"
	"github.com/spf13/cobra"
)

var (
	c   = config.New()
	cli = client.New(c)
	cmd = &cobra.Command{
		Use:   "bobbit",
		Short: "Simply \"yet\" UNIX Socket based job runner",
	}
)

func init() {
	RegisterCreateCommand()
	RegisterDaemonCommand()
	RegisterListCommand()
	RegisterWaitCommand()
	RegisterStatusCommand()
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
