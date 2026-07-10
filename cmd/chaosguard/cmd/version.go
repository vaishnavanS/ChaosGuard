package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "v0.1.0-dev"
	GitCommit = "none"
	BuildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the ChaosGuard version information",
	Long:  `Displays the current build version, Git commit hash, and compiler build time.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ChaosGuard Version: %s\n", Version)
		fmt.Printf("Git Commit:         %s\n", GitCommit)
		fmt.Printf("Build Time:         %s\n", BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
