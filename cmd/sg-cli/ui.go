// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	uiNamespace string
	uiPort      int
	uiOpen      bool
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Port-forward and open the StraitGateway management web dashboard in your browser",
	Long:  "Port-forward and open the StraitGateway management web dashboard in your browser.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("==> StraitGateway Web UI Dashboard")
		fmt.Printf("    Namespace:  %s\n", uiNamespace)
		fmt.Printf("    Local Port: %d\n", uiPort)
		fmt.Printf("    URL:        http://localhost:%d\n\n", uiPort)

		// Determine candidate services
		candidateServices := []struct {
			name string
			port int
		}{
			{"svc/straitkubegateway-ui", 80},
			{"svc/straitgateway-ui", 8080},
			{"svc/straitgateway-ui", 80},
		}

		targetSvc := candidateServices[0].name
		targetPort := candidateServices[0].port

		// Check which service actually exists in the cluster
		for _, candidate := range candidateServices {
			chk := exec.Command("kubectl", "get", candidate.name, "-n", uiNamespace)
			if err := chk.Run(); err == nil {
				targetSvc = candidate.name
				targetPort = candidate.port
				break
			}
		}

		portMapping := fmt.Sprintf("%d:%d", uiPort, targetPort)
		portForwardCmd := exec.Command("kubectl", "port-forward", "-n", uiNamespace, targetSvc, portMapping)
		portForwardCmd.Stdout = os.Stdout
		portForwardCmd.Stderr = os.Stderr

		if err := portForwardCmd.Start(); err != nil {
			return fmt.Errorf("failed to start port-forward: %w", err)
		}

		fmt.Printf("Forwarding from 127.0.0.1:%d -> %d. Press Ctrl+C to stop.\n", uiPort, targetPort)

		if uiOpen {
			go func() {
				time.Sleep(500 * time.Millisecond)
				_ = openBrowser(fmt.Sprintf("http://localhost:%d", uiPort))
			}()
		}

		// Handle interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			select {
			case <-sigChan:
				if portForwardCmd.Process != nil {
					_ = portForwardCmd.Process.Signal(os.Kill)
				}
				cancel()
			case <-ctx.Done():
			}
		}()

		return portForwardCmd.Wait()
	},
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func init() {
	uiCmd.Flags().StringVarP(&uiNamespace, "namespace", "n", "kube-system", "Kubernetes namespace where StraitGateway is installed")
	uiCmd.Flags().IntVarP(&uiPort, "port", "p", 8080, "Local port to map to the UI service")
	uiCmd.Flags().BoolVarP(&uiOpen, "open", "b", false, "Automatically open the UI in default browser")
}
