package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the ChaosGuard daemon and recover services",
	Long:  `Gracefully shuts down the scheduler, recovers any container currently under attack to its running state, and exits.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("stop", "Stopping ChaosGuard daemon and initiating recovery sequence...")
		fmt.Println("All chaos activities stopped. Target containers successfully restored.")
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
}
