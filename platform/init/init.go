// Package initutil provides init system and container runtime detection.
package initutil

import (
	"os"
)

// InitSystem represents the detected host init system.
type InitSystem string

const (
	InitSystemSystemd   InitSystem = "systemd"
	InitSystemOpenRC    InitSystem = "openrc"
	InitSystemContainer InitSystem = "container"
	InitSystemUnknown   InitSystem = "unknown"
)

// DetectInitSystem determines whether the process runs under systemd, container, or openrc.
func DetectInitSystem() InitSystem {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return InitSystemSystemd
	}
	if _, err := os.Stat("/run/openrc"); err == nil {
		return InitSystemOpenRC
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return InitSystemContainer
	}
	if _, err := os.Stat("/run/containerd"); err == nil {
		return InitSystemContainer
	}
	return InitSystemUnknown
}
