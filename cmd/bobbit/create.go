package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
	"mplus.software/oss/bobbit.go/payload"
)

func RegisterCreateCommand() {
	create := &cobra.Command{
		Use:   "create <job_id> -- <command>",
		Short: "Create new job",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id, command := args[0], args[1:]

			var metadata map[string]any
			metadataStr, err := cmd.Flags().GetString("metadata")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			if metadataStr != "" {
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
					shell.Fatalfln(8, "Metadata given is not valid JSON: %v", err)
				}
			}

			conn, err := client.CreateConnection(c)
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			defer conn.Connection.Close()

			p := payload.JobPayload{Request: payload.REQUEST_EXECUTE_JOB}
			req := payload.JobRequestMetadata{
				ID:       id,
				Command:  command,
				Metadata: metadata,
			}
			if err := p.MarshalMetadata(req); err != nil {
				shell.Fatalfln(3, "Failed to marshal metadata: %v", err)
			}
			if err := conn.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}
			shell.Printfln("Job %s created!", id)
		},
	}
	create.Flags().StringP("metadata", "m", "", "JSON Metadata")
	cmd.AddCommand(create)
}
