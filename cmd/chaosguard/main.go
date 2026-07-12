package main

import (
	"fmt"
	"os"

	"chaosguard/cmd/chaosguard/cmd"
)

// @title ChaosGuard API
// @version 1.0
// @description ChaosGuard Automated Chaos Engineering Platform API
// @host localhost:8080
// @BasePath /
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
