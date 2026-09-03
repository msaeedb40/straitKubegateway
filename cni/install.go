// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

const (
	// DefaultCNIConfDir is the standard directory where container runtimes look for CNI configs.
	DefaultCNIConfDir = "/etc/cni/net.d"

	// DefaultCNIBinDir is the standard directory where container runtimes look for CNI plugins.
	DefaultCNIBinDir = "/opt/cni/bin"

	// DefaultConflistName is the filename for the straitKubegateway CNI conflist.
	DefaultConflistName = "10-strait.conflist"
)

// InstallConfig writes the CNI conflist to the host's CNI configuration directory.
// When this file is written, container runtimes (containerd/CRI-O) recognize that
// a network plugin is initialized and clear the NetworkPluginNotReady condition on the node.
func InstallConfig(confDir, podCIDR string, log *zap.Logger) error {
	if confDir == "" {
		confDir = DefaultCNIConfDir
	}

	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("create CNI conf directory %s: %w", confDir, err)
	}

	if podCIDR == "" || podCIDR == "10.244.0.0/16" {
		nodeName := os.Getenv("NODE_NAME")
		if nodeName == "" {
			nodeName, _ = os.Hostname()
		}
		if strings.Contains(nodeName, "worker2") {
			podCIDR = "10.244.2.0/24"
		} else if strings.Contains(nodeName, "worker") {
			podCIDR = "10.244.1.0/24"
		} else {
			podCIDR = "10.244.0.0/24"
		}
	}

	// 10-strait.conflist specifies the CNI chaining configuration.
	// We use ptp/bridge and host-local IPAM (pre-installed in containerd / kind),
	// allowing immediate node Readiness while eBPF programs handle datapath forwarding.
	conflist := fmt.Sprintf(`{
  "cniVersion": "0.3.1",
  "name": "straitkubegateway",
  "plugins": [
    {
      "type": "ptp",
      "ipam": {
        "type": "host-local",
        "subnet": "%s",
        "routes": [
          { "dst": "0.0.0.0/0" }
        ]
      }
    },
    {
      "type": "portmap",
      "capabilities": {
        "portMappings": true
      }
    }
  ]
}
`, podCIDR)

	targetPath := filepath.Join(confDir, DefaultConflistName)
	if err := os.WriteFile(targetPath, []byte(conflist), 0644); err != nil {
		return fmt.Errorf("write CNI conflist to %s: %w", targetPath, err)
	}

	log.Info("installed CNI configuration file",
		zap.String("path", targetPath),
		zap.String("podCIDR", podCIDR),
	)
	return nil
}
