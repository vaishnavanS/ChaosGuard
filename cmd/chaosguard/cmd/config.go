package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var printJson bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect the parsed ChaosGuard configuration",
	Long:  `Reads the local configuration file (and environmental variable overrides) and displays active settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		if ActiveConfig == nil {
			fmt.Println("No configuration loaded. Use 'chaosguard init' to create a default configuration.")
			return
		}

		if printJson {
			data, err := json.MarshalIndent(ActiveConfig, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				return
			}
			fmt.Println(string(data))
		} else {
			data, err := yaml.Marshal(ActiveConfig)
			if err != nil {
				fmt.Printf("Error formatting YAML: %v\n", err)
				return
			}
			fmt.Println(string(data))
		}
	},
}

func init() {
	configCmd.Flags().BoolVar(&printJson, "json", false, "format output in JSON instead of YAML")
	RootCmd.AddCommand(configCmd)
}
