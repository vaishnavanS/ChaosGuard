package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"chaosguard/pkg/config"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate local environment prerequisites",
	Long:  `Performs system diagnostic checks to ensure Docker is running, sockets are reachable, configuration is valid, and ports are free.`,
	Run: func(cmd *cobra.Command, args []string) {
		printAction("doctor", "Running environment dependency checks...")

		failed := false

		// 1. Docker Daemon Check
		fmt.Print("Checking Docker daemon connectivity... ")
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("[✗] (Initialization failed: %v)\n", err)
			failed = true
		} else {
			defer cli.Close()
			_, err = cli.Ping(context.Background())
			if err != nil {
				fmt.Printf("[✗] (Daemon unreachable. Is Docker running? Error: %v)\n", err)
				failed = true
			} else {
				fmt.Println("[✓] Reachable")
			}
		}

		// 2. Ports check (8080 for Web, 2112 for Metrics)
		ports := []int{8080, 2112}
		if ActiveConfig != nil {
			ports = []int{ActiveConfig.Dashboard.Port, ActiveConfig.Metrics.Port}
		} else {
			// Try to load default config just in case
			if cfg, err := config.Load(cfgFile, nil); err == nil {
				ports = []int{cfg.Dashboard.Port, cfg.Metrics.Port}
			}
		}

		for _, port := range ports {
			fmt.Printf("Checking port %d availability... ", port)
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				fmt.Printf("[✗] (Port %d already in use: %v)\n", port, err)
				failed = true
			} else {
				ln.Close()
				fmt.Println("[✓] Available")
			}
		}

		// 3. SQLite Database Write Access Check
		dbPath := "./chaosguard.db"
		if ActiveConfig != nil {
			dbPath = ActiveConfig.Database.Path
		}
		fmt.Printf("Checking Database write permissions at %s... ", dbPath)
		dbDir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			fmt.Printf("[✗] (Cannot create directory structure: %v)\n", err)
			failed = true
		} else {
			testFile := filepath.Join(dbDir, ".chaosguard_write_test")
			err := os.WriteFile(testFile, []byte("test"), 0644)
			if err != nil {
				fmt.Printf("[✗] (Write test failed: %v)\n", err)
				failed = true
			} else {
				os.Remove(testFile)
				fmt.Println("[✓] Permitted")
			}
		}

		fmt.Println("--------------------------------------------------")
		if failed {
			fmt.Println("Warning: Some checks failed. ChaosGuard might not function correctly.")
		} else {
			fmt.Println("All checks passed! ChaosGuard is fully ready to run on this system.")
		}
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
