package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"chaosguard/internal/api/responses"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		printAction("report", fmt.Sprintf("Format: %s, Output: %s, Limit: %d", reportFormat, outputPath, historyLimit))

		port := 8080
		if ActiveConfig != nil {
			port = ActiveConfig.Dashboard.Port
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/experiments", port))
		if err != nil {
			return fmt.Errorf("failed to reach daemon: %w. Is the ChaosGuard daemon running?", err)
		}
		defer resp.Body.Close()

		var expData responses.SuccessResponse
		if err := json.NewDecoder(resp.Body).Decode(&expData); err != nil {
			return err
		}

		var items []interface{}
		if list, ok := expData.Data.([]interface{}); ok {
			items = list
		}

		// Apply limit if specified
		if historyLimit > 0 && len(items) > historyLimit {
			items = items[:historyLimit]
		}

		var outData []byte
		finalPath := outputPath

		switch reportFormat {
		case "json":
			outData, err = json.MarshalIndent(items, "", "  ")
			if err != nil {
				return err
			}
		case "csv":
			var buf bytes.Buffer
			writer := csv.NewWriter(&buf)
			_ = writer.Write([]string{"ID", "Target Container", "Attack Type", "Duration (s)", "Status", "Started At", "Ended At"})
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					id, _ := m["id"].(string)
					cName, _ := m["container_name"].(string)
					cID, _ := m["target_container_id"].(string)
					target := cName
					if target == "" {
						target = cID
					}
					attack, _ := m["attack_type"].(string)
					durVal := m["duration"]
					durStr := fmt.Sprintf("%v", durVal)
					status, _ := m["status"].(string)
					started, _ := m["started_at"].(string)
					ended, _ := m["ended_at"].(string)

					_ = writer.Write([]string{id, target, attack, durStr, status, started, ended})
				}
			}
			writer.Flush()
			outData = buf.Bytes()
		case "html", "pdf":
			// Gracefully fallback to HTML for PDF since it requires massive system binaries
			if reportFormat == "pdf" {
				if filepath.Ext(finalPath) == ".pdf" {
					finalPath = finalPath[:len(finalPath)-4] + ".html"
				}
				fmt.Println("Warning: PDF format requires massive external toolchains. Falling back to HTML format.")
			}

			var htmlBuf bytes.Buffer
			htmlBuf.WriteString("<!DOCTYPE html><html><head><title>ChaosGuard Resilience Report</title>")
			htmlBuf.WriteString("<style>body{font-family:sans-serif;margin:40px;color:#333;}table{border-collapse:collapse;width:100%;}th,td{border:1px solid #ddd;padding:8px;text-align:left;}th{background-color:#f2f2f2;}tr:nth-child(even){background-color:#fafafa;}</style></head><body>")
			htmlBuf.WriteString("<h1>ChaosGuard Resilience Report</h1>")
			htmlBuf.WriteString(fmt.Sprintf("<p>Generated at: %s</p>", time.Now().Format(time.RFC1123)))
			htmlBuf.WriteString("<table><tr><th>ID</th><th>Target</th><th>Attack Type</th><th>Duration (s)</th><th>Status</th><th>Started At</th><th>Ended At</th></tr>")
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					id, _ := m["id"].(string)
					cName, _ := m["container_name"].(string)
					cID, _ := m["target_container_id"].(string)
					target := cName
					if target == "" {
						target = cID
					}
					attack, _ := m["attack_type"].(string)
					durVal := m["duration"]
					durStr := fmt.Sprintf("%v", durVal)
					status, _ := m["status"].(string)
					started, _ := m["started_at"].(string)
					ended, _ := m["ended_at"].(string)

					htmlBuf.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
						id, target, attack, durStr, status, started, ended))
				}
			}
			htmlBuf.WriteString("</table></body></html>")
			outData = htmlBuf.Bytes()
		default:
			return fmt.Errorf("unsupported report format: %s", reportFormat)
		}

		if err := os.WriteFile(finalPath, outData, 0644); err != nil {
			return fmt.Errorf("failed to save report file: %w", err)
		}

		fmt.Printf("Resilience report successfully compiled and saved to %s\n", finalPath)
		return nil
	},
}

func init() {
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "html", "report format (html, csv, json, pdf)")
	reportCmd.Flags().StringVarP(&outputPath, "output", "o", "./report.html", "destination path for the compiled report file")
	reportCmd.Flags().IntVar(&historyLimit, "limit", 50, "limit history records included in the report")
	RootCmd.AddCommand(reportCmd)
}
