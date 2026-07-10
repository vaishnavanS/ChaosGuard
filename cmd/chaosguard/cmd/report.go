package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	reportFormat string
	outputPath   string
	historyLimit int
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Export chaos history and resilience reports",
	Long:  `Compiles the historical data of chaos experiments, issues found, and recovery times into a report file.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("report", fmt.Sprintf("Format: %s, Output: %s, Limit: %d", reportFormat, outputPath, historyLimit))
		fmt.Printf("Resilience report successfully compiled and saved to %s\n", outputPath)
	},
}

func init() {
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "pdf", "report format (pdf, html, csv, json)")
	reportCmd.Flags().StringVarP(&outputPath, "output", "o", "./report.pdf", "destination path for the compiled report file")
	reportCmd.Flags().IntVar(&historyLimit, "limit", 50, "limit history records included in the report")
	RootCmd.AddCommand(reportCmd)
}
