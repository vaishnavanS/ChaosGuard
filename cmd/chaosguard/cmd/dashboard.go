package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var openBrowser bool

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the ChaosGuard Web Dashboard",
	Long:  `Launches or provides the URL of the ChaosGuard local management dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("dashboard", fmt.Sprintf("Open Browser: %v", openBrowser))
		fmt.Println("Dashboard available at: http://localhost:8080")
	},
}

func init() {
	dashboardCmd.Flags().BoolVar(&openBrowser, "open", true, "automatically open the dashboard in the default browser")
	RootCmd.AddCommand(dashboardCmd)
}
