package cmd

import (
	"fmt"
	"os/exec"
	sysruntime "runtime"

	"github.com/spf13/cobra"
)

var openBrowser bool

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the ChaosGuard Web Dashboard",
	Long:  `Launches or provides the URL of the ChaosGuard local management dashboard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port := 8080
		if ActiveConfig != nil {
			port = ActiveConfig.Dashboard.Port
		}
		url := fmt.Sprintf("http://localhost:%d/swagger/index.html", port)

		printAction("dashboard", fmt.Sprintf("Open Browser: %v, Port: %d", openBrowser, port))
		fmt.Printf("ChaosGuard dashboard interactive Swagger UI available at: %s\n", url)

		if openBrowser {
			fmt.Printf("Opening default browser to %s...\n", url)
			if err := openBrowserURL(url); err != nil {
				return fmt.Errorf("failed to open default browser: %w", err)
			}
		}
		return nil
	},
}

func openBrowserURL(url string) error {
	var cmd string
	var args []string
	switch sysruntime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

func init() {
	dashboardCmd.Flags().BoolVar(&openBrowser, "open", true, "automatically open the dashboard in the default browser")
	RootCmd.AddCommand(dashboardCmd)
}
