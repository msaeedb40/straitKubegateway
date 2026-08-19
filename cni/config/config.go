// Package config provides CNI network configuration parsing and validation
// compliant with the CNI specification 1.1+.
package config

import (
	"encoding/json"
	"fmt"
	"net/netip"
)

// NetConf represents the straitKubegateway CNI network configuration.
type NetConf struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`

	// PrevResult contains the previous plugin's result in chained invocations.
	PrevResult map[string]interface{} `json:"prevResult,omitempty"`

	// SocketPath is the path to the straitd UNIX domain socket for RPC.
	SocketPath string `json:"socketPath,omitempty"`

	// MTU is the maximum transmission unit for the pod interface (0 = auto-detect).
	MTU int `json:"mtu,omitempty"`

	// SegmentID is the default network segment for pods on this node.
	SegmentID uint32 `json:"segmentId,omitempty"`

	// IPAM specifies IPAM configuration for standalone mode.
	IPAM IPAMConfig `json:"ipam,omitempty"`
}

// IPAMConfig defines IPAM settings embedded in the CNI configuration.
type IPAMConfig struct {
	Type     string   `json:"type"`
	Subnet   string   `json:"subnet,omitempty"`
	Subnet6  string   `json:"subnet6,omitempty"`
	Routes   []Route  `json:"routes,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// Route represents a routing table entry in CNI results.
type Route struct {
	Dst string `json:"dst"`
	GW  string `json:"gw,omitempty"`
}

// DefaultSocketPath is the standard UNIX socket path for straitd communication.
const DefaultSocketPath = "/run/straitkubegateway/straitd.sock"

// LoadNetConf parses a raw CNI JSON configuration byte slice.
func LoadNetConf(bytes []byte) (*NetConf, error) {
	conf := &NetConf{
		SocketPath: DefaultSocketPath,
	}

	if err := json.Unmarshal(bytes, conf); err != nil {
		return nil, fmt.Errorf("failed to parse CNI network configuration: %w", err)
	}

	if conf.Name == "" {
		return nil, fmt.Errorf("CNI configuration missing required 'name' field")
	}
	if conf.Type == "" {
		return nil, fmt.Errorf("CNI configuration missing required 'type' field")
	}

	if conf.IPAM.Subnet != "" {
		if _, err := netip.ParsePrefix(conf.IPAM.Subnet); err != nil {
			return nil, fmt.Errorf("invalid IPv4 subnet in IPAM config %q: %w", conf.IPAM.Subnet, err)
		}
	}
	if conf.IPAM.Subnet6 != "" {
		if _, err := netip.ParsePrefix(conf.IPAM.Subnet6); err != nil {
			return nil, fmt.Errorf("invalid IPv6 subnet in IPAM config %q: %w", conf.IPAM.Subnet6, err)
		}
	}

	return conf, nil
}
