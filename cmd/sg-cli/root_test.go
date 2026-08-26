// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"testing"
)

func TestCLICommandsRegistration(t *testing.T) {
	expectedCommands := []string{
		"status",
		"gateway",
		"node",
		"bgp",
		"policy",
		"transit",
		"endpoint",
		"cluster",
		"wireguard",
		"ipsec",
		"config",
		"export",
		"import",
		"install",
		"upgrade",
		"version",
	}

	cmdMap := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		cmdMap[cmd.Name()] = true
	}

	for _, name := range expectedCommands {
		if !cmdMap[name] {
			t.Errorf("missing expected subcommand: %s", name)
		}
	}
}

func TestCLIExecution(t *testing.T) {
	for _, name := range []string{"status", "version", "config", "gateway", "node", "transit", "policy"} {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{name})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("command %q failed: %v", name, err)
		}
	}
}
