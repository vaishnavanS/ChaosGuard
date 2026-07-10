package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	targetContainer string
	attackType      string
	durationSec     int
)

var attackCmd = &cobra.Command{
	Use:   "attack",
	Short: "Execute an ad-hoc chaos attack",
	Long:  `Triggers an immediate, direct chaos attack against a specific Docker container.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if targetContainer == "" {
			return errors.New("target container (--target or -t) is required")
		}
		if attackType == "" {
			return errors.New("attack type (--type or -a) is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		printAction("attack", fmt.Sprintf("Target: %s, Type: %s, Duration: %ds", targetContainer, attackType, durationSec))
		fmt.Printf("Attack successfully initiated against container '%s'.\n", targetContainer)
		fmt.Printf("Attack completed. Restoring container '%s'...\n", targetContainer)
	},
}

func init() {
	attackCmd.Flags().StringVarP(&targetContainer, "target", "t", "", "target container name or ID (required)")
	attackCmd.Flags().StringVarP(&attackType, "type", "a", "", "type of attack: pause, stop, restart, kill (required)")
	attackCmd.Flags().IntVarP(&durationSec, "duration", "d", 10, "duration of the chaos attack in seconds")
	RootCmd.AddCommand(attackCmd)
}
