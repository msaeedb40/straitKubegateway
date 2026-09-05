// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAPIServerFromKubeconfig(t *testing.T) {
	tempDir := t.TempDir()
	confPath := filepath.Join(tempDir, "kubelet.conf")

	content := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: abc
    server: https://172.18.0.3:6443
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    user: default-auth
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users:
- name: default-auth
  user:
    client-certificate-data: def
    client-key-data: ghi
`
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test kubelet.conf: %v", err)
	}

	f, err := os.Open(confPath)
	if err != nil {
		t.Fatalf("failed to open test kubelet.conf: %v", err)
	}
	defer f.Close()

	host, port, found := parseAPIServerFromKubeconfig(f)
	if !found {
		t.Fatalf("expected to find API server in kubeconfig")
	}
	if host != "172.18.0.3" {
		t.Errorf("got host %q, want 172.18.0.3", host)
	}
	if port != 6443 {
		t.Errorf("got port %d, want 6443", port)
	}
}

func TestDiscoverAPIServerDirectHost(t *testing.T) {
	host, port := discoverAPIServer("1.2.3.4", 8443)
	if host != "1.2.3.4" || port != 8443 {
		t.Errorf("got %s:%d, want 1.2.3.4:8443", host, port)
	}
}

func TestResolveHostIPv4Preference(t *testing.T) {
	// localhost should resolve to 127.0.0.1
	ip := resolveHost("127.0.0.1")
	if ip != "127.0.0.1" {
		t.Errorf("got %q, want 127.0.0.1", ip)
	}
}
