package cmd

import (
	"fmt"

	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	verbose      bool
	ActiveConfig *config.Config
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "chaosguard",
	Short: "ChaosGuard is an Automated Chaos Engineering Platform for Docker",
	Long: `ChaosGuard automatically injects failures into Docker-based microservices,
monitors application behavior, detects resilience issues, and presents detailed
analysis through a modern local dashboard.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize global logger
		logger.Setup(verbose, false)

		// Do not require configuration for init, version, doctor
		if cmd.Name() == "init" || cmd.Name() == "version" || cmd.Name() == "doctor" {
			return nil
		}

		var err error
		ActiveConfig, err = config.Load(cfgFile, func(newCfg *config.Config) {
			logger.Info("Configuration hot-reloaded: %+v", newCfg)
			ActiveConfig = newCfg
		})
		if err != nil {
			logger.Error(err, "Failed to load configuration")
			return err
		}

		logger.Debug("Configuration loaded successfully: %+v", ActiveConfig)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./chaosguard.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) logging")
}

func printAction(cmdName string, details string) {
	fmt.Printf("[ChaosGuard CLI] Executing command: %s\n", cmdName)
	if details != "" {
		fmt.Printf("Details: %s\n", details)
	}
}
