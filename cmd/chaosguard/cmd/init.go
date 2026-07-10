package cmd

import (
	"fmt"

	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ChaosGuard configuration",
	Long:  `Creates a default chaosguard.yaml configuration file in the current directory if it does not exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("init", "Initializing configuration file 'chaosguard.yaml'")
		
		path := cfgFile
		if path == "" {
			path = config.DefaultConfigName
		}
		
		err := config.WriteDefault(path)
		if err != nil {
			logger.Error(err, "Failed to initialize configuration")
			fmt.Printf("Error: %v\n", err)
			return
		}

		logger.Info("Configuration initialized successfully at %s", path)
		fmt.Printf("Initialization complete. Created %s. Run 'chaosguard config' to inspect or 'chaosguard start' to launch.\n", path)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
