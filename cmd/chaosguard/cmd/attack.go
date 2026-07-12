package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"chaosguard/internal/api/requests"
	"chaosguard/internal/api/responses"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		printAction("attack", fmt.Sprintf("Target: %s, Type: %s, Duration: %ds", targetContainer, attackType, durationSec))

		port := 8080
		if ActiveConfig != nil {
			port = ActiveConfig.Dashboard.Port
		}

		client := &http.Client{Timeout: 5 * time.Second}

		// 1. Fetch containers to resolve name to ID
		respCont, err := client.Get(fmt.Sprintf("http://localhost:%d/containers", port))
		if err != nil {
			return fmt.Errorf("failed to reach daemon: %w. Is the ChaosGuard daemon running?", err)
		}
		defer respCont.Body.Close()

		var contData responses.SuccessResponse
		if err := json.NewDecoder(respCont.Body).Decode(&contData); err != nil {
			return err
		}

		var targetID string
		var actualName string
		if items, ok := contData.Data.([]interface{}); ok {
			for _, it := range items {
				if m, ok := it.(map[string]interface{}); ok {
					id, _ := m["id"].(string)
					name, _ := m["name"].(string)
					if id == targetContainer || name == targetContainer {
						targetID = id
						actualName = name
						break
					}
				}
			}
		}

		if targetID == "" {
			return fmt.Errorf("container '%s' not discovered by ChaosGuard", targetContainer)
		}

		// 2. Trigger experiment via POST
		reqBody := requests.CreateExperimentRequest{
			TargetContainerID: targetID,
			AttackType:        attackType,
			DurationSec:       durationSec,
		}

		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}

		respExp, err := client.Post(fmt.Sprintf("http://localhost:%d/experiments", port), "application/json", bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("failed to execute attack request: %w", err)
		}
		defer respExp.Body.Close()

		if respExp.StatusCode != http.StatusCreated {
			var errResp responses.ErrorResponse
			_ = json.NewDecoder(respExp.Body).Decode(&errResp)
			if errResp.Error != "" {
				return fmt.Errorf("attack request failed: %s", errResp.Error)
			}
			return fmt.Errorf("attack request returned status: %d", respExp.StatusCode)
		}

		fmt.Printf("Attack successfully initiated against container '%s' (resolved ID: %s).\n", actualName, targetID)
		fmt.Printf("Check 'chaosguard status' to track current failure injection.\n")
		return nil
	},
}

func init() {
	attackCmd.Flags().StringVarP(&targetContainer, "target", "t", "", "target container name or ID (required)")
	attackCmd.Flags().StringVarP(&attackType, "type", "a", "", "type of attack: pause, stop, restart, kill (required)")
	attackCmd.Flags().IntVarP(&durationSec, "duration", "d", 10, "duration of the chaos attack in seconds")
	RootCmd.AddCommand(attackCmd)
}
