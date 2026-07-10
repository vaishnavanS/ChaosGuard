package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of ChaosGuard scheduler and targets",
	Long:  `Queries and displays the active daemon status, current running chaos experiments, and monitored Docker containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("status", "Querying scheduler state and active attacks...")
		fmt.Println("ChaosGuard: Idle / Active")
		fmt.Println("Running Experiments: 0")
		fmt.Println("Healthy Containers: 0 / Discovered: 0")
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
