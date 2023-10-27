package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Integration Agent",
	Run: func(cmd *cobra.Command, args []string) {
		// Start a listener to handle events from the LaunchboxHQ SSE Stream

		// CRD watcher for propagating events back to LaunchboxHQ

		fmt.Println("Running the agent process")
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
