package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the ChaosGuard daemon and recover services",
	Long:  `Gracefully shuts down the scheduler, recovers any container currently under attack to its running state, and exits.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printAction("stop", "Sending stop command to ChaosGuard daemon...")

		port := 8080
		if ActiveConfig != nil {
			port = ActiveConfig.Dashboard.Port
		}
		url := fmt.Sprintf("http://localhost:%d/runtime/stop", port)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(url, "application/json", nil)
		if err != nil {
			return fmt.Errorf("failed to reach daemon: %w. Is the ChaosGuard daemon running?", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to stop daemon, server returned status: %d", resp.StatusCode)
		}

		fmt.Println("Graceful shutdown signal sent successfully to the daemon.")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
}
