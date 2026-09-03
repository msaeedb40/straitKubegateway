// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestInstallConfig(t *testing.T) {
	tempDir := t.TempDir()
	log := zap.NewNop()

	err := InstallConfig(tempDir, "10.244.0.0/16", log)
	if err != nil {
		t.Fatalf("InstallConfig failed: %v", err)
	}

	targetPath := filepath.Join(tempDir, DefaultConflistName)
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read written conflist: %v", err)
	}

	content := string(data)
	if !filepath.IsAbs(targetPath) {
		t.Errorf("expected absolute target path, got %s", targetPath)
	}
	if len(content) == 0 {
		t.Errorf("conflist content should not be empty")
	}
}
