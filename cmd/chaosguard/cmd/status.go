package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chaosguard/internal/api/responses"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of ChaosGuard scheduler and targets",
	Long:  `Queries and displays the active daemon status, current running chaos experiments, and monitored Docker containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printAction("status", "Querying ChaosGuard daemon status...")

		port := 8080
		if ActiveConfig != nil {
			port = ActiveConfig.Dashboard.Port
		}

		client := &http.Client{Timeout: 5 * time.Second}

		// 1. Get Runtime Status
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/runtime", port))
		if err != nil {
			fmt.Println("==================================================")
			fmt.Println("ChaosGuard System Status")
			fmt.Println("==================================================")
			fmt.Println("Daemon Status:       Inactive (unreachable)")
			fmt.Println("==================================================")
			return nil
		}
		defer resp.Body.Close()

		var runtimeData responses.SuccessResponse
		if err := json.NewDecoder(resp.Body).Decode(&runtimeData); err != nil {
			return err
		}

		// 2. Get Scheduler Status
		var schedRunning bool
		var schedMode string
		if respSched, err := client.Get(fmt.Sprintf("http://localhost:%d/scheduler/status", port)); err == nil {
			defer respSched.Body.Close()
			var schedData responses.SuccessResponse
			if err := json.NewDecoder(respSched.Body).Decode(&schedData); err == nil {
				if m, ok := schedData.Data.(map[string]interface{}); ok {
					schedRunning, _ = m["running"].(bool)
					schedMode, _ = m["mode"].(string)
				}
			}
		}

		// 3. Get Containers
		var totalContainers, monitoredContainers, runningContainers, pausedContainers, stoppedContainers int
		if respCont, err := client.Get(fmt.Sprintf("http://localhost:%d/containers", port)); err == nil {
			defer respCont.Body.Close()
			var contData responses.SuccessResponse
			if err := json.NewDecoder(respCont.Body).Decode(&contData); err == nil {
				if items, ok := contData.Data.([]interface{}); ok {
					totalContainers = len(items)
					for _, it := range items {
						if m, ok := it.(map[string]interface{}); ok {
							isMon, _ := m["is_monitored"].(bool)
							if isMon {
								monitoredContainers++
							}
							status, _ := m["status"].(string)
							switch status {
							case "running":
								runningContainers++
							case "paused":
								pausedContainers++
							default:
								stoppedContainers++
							}
						}
					}
				}
			}
		}

		// 4. Get Experiments
		var runningExperiments int
		if respExp, err := client.Get(fmt.Sprintf("http://localhost:%d/experiments", port)); err == nil {
			defer respExp.Body.Close()
			var expData responses.SuccessResponse
			if err := json.NewDecoder(respExp.Body).Decode(&expData); err == nil {
				if items, ok := expData.Data.([]interface{}); ok {
					for _, it := range items {
						if m, ok := it.(map[string]interface{}); ok {
							status, _ := m["status"].(string)
							if status == "running" || status == "pending" {
								runningExperiments++
							}
						}
					}
				}
			}
		}

		fmt.Println("==================================================")
		fmt.Println("ChaosGuard System Status")
		fmt.Println("==================================================")
		fmt.Printf("Daemon Status:       Active\n")

		stateStr := "unknown"
		if m, ok := runtimeData.Data.(map[string]interface{}); ok {
			stateStr, _ = m["state"].(string)
		}
		fmt.Printf("Runtime State:       %s\n", stateStr)

		schedStatus := "Stopped"
		if schedRunning {
			schedStatus = fmt.Sprintf("Running (%s mode)", schedMode)
		}
		fmt.Printf("Scheduler:           %s\n", schedStatus)
		fmt.Printf("Running Experiments: %d\n", runningExperiments)
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Containers:\n")
		fmt.Printf("  Discovered:        %d\n", totalContainers)
		fmt.Printf("  Monitored:         %d\n", monitoredContainers)
		fmt.Printf("  State - Running:   %d\n", runningContainers)
		fmt.Printf("  State - Paused:    %d\n", pausedContainers)
		fmt.Printf("  State - Stopped:   %d\n", stoppedContainers)
		fmt.Println("==================================================")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
