package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	port     int
	safeMode bool
	daemon   bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ChaosGuard scheduler and dashboard",
	Long:  `Starts the background chaos experiment scheduler, system health monitors, and launches the web interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("start", fmt.Sprintf("Port: %d, Safe Mode: %v, Daemon: %v", port, safeMode, daemon))
		fmt.Println("ChaosGuard daemon started successfully. Point your browser to http://localhost:8080")
	},
}

func init() {
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "port for the dashboard server")
	startCmd.Flags().BoolVar(&safeMode, "safe-mode", true, "prevent attacks on system-critical containers (e.g. database)")
	startCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "run ChaosGuard in background daemon mode")
	RootCmd.AddCommand(startCmd)
}
