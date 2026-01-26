package main

import (
	"encoding/json"

	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterCreateCommand() {
	create := &cobra.Command{
		Use:   "create <job_name> -- <command>",
		Short: "Create new job",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			jobName, command := args[0], args[1:]

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

			req := payload.JobDetailMetadata{
				JobName:  jobName,
				Command:  command,
				Metadata: metadata,
			}

			job, err := cli.Create(req)
			if err != nil {
				shell.Fatalfln(3, "Failed to create job: %v", err)
			}
			shell.Printfln("Job %s created! [%s]", job.JobName, job.ID)
		},
	}
	create.Flags().StringP("metadata", "m", "", "JSON Metadata")
	cmd.AddCommand(create)
}
