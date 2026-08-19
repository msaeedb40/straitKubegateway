// Package main implements the straitKubegateway CNI plugin binary.
// It delegates to the straitd daemon via UNIX socket IPC for persistent
// IPAM and dataplane state, with a local fallback for standalone environments.
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	cniconfig "github.com/straitKubegateway/straitKubegateway/cni/config"
	"github.com/straitKubegateway/straitKubegateway/dataplane"
	sgtypes "github.com/straitKubegateway/straitKubegateway/pkg/types"
)

const (
	socketDialTimeout = 3 * time.Second
)

// K8sArgs represents standard Kubernetes CNI arguments passed via CNI_ARGS.
type K8sArgs struct {
	types.CommonArgs
	K8S_POD_NAME               types.UnmarshallableString `json:"K8S_POD_NAME,omitempty"`
	K8S_POD_NAMESPACE          types.UnmarshallableString `json:"K8S_POD_NAMESPACE,omitempty"`
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString `json:"K8S_POD_INFRA_CONTAINER_ID,omitempty"`
}

func main() {
	skel.PluginMain(
		cmdAdd,
		cmdCheck,
		cmdDel,
		version.All,
		fmt.Sprintf("straitKubegateway CNI plugin"),
	)
}

// dialDaemon attempts to connect to the straitd CNI daemon socket.
func dialDaemon() (net.Conn, error) {
	return net.DialTimeout("unix", dataplane.DefaultSocketPath, socketDialTimeout)
}

// sendRequest sends a CNIRequest to the daemon and returns the response.
func sendRequest(conn net.Conn, req *dataplane.CNIRequest) (*dataplane.CNIResponse, error) {
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	var resp dataplane.CNIResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("daemon error: %s", resp.Error)
	}
	return &resp, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	netConf, err := cniconfig.LoadNetConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("load cni config: %w", err)
	}

	var k8sArgs K8sArgs
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return fmt.Errorf("load k8s args: %w", err)
	}

	podName := string(k8sArgs.K8S_POD_NAME)
	namespace := string(k8sArgs.K8S_POD_NAMESPACE)
	if namespace == "" {
		namespace = "default"
	}

	cidr := netConf.IPAM.Subnet
	if cidr == "" {
		cidr = "10.244.0.0/16" // Default fallback
	}

	// Try daemon socket first for persistent state
	conn, dialErr := dialDaemon()
	if dialErr == nil {
		defer conn.Close()
		return cmdAddViaDaemon(conn, args, netConf, namespace, podName, cidr)
	}

	// Fallback: local dataplane (standalone / test environments)
	return cmdAddLocal(args, netConf, namespace, podName, cidr)
}

// cmdAddViaDaemon delegates the ADD to the straitd daemon via UNIX socket.
func cmdAddViaDaemon(conn net.Conn, args *skel.CmdArgs, netConf *cniconfig.NetConf, namespace, podName, cidr string) error {
	req := &dataplane.CNIRequest{
		Command:     "ADD",
		ContainerID: args.ContainerID,
		NetnsPath:   args.Netns,
		IfName:      args.IfName,
		Namespace:   namespace,
		PodName:     podName,
		SegmentID:   uint32(netConf.SegmentID),
		Labels:      map[string]string{"k8s.io/namespace": namespace, "k8s.io/pod": podName},
		MTU:         netConf.MTU,
		PodCIDR:     cidr,
	}

	resp, err := sendRequest(conn, req)
	if err != nil {
		return fmt.Errorf("daemon ADD failed: %w", err)
	}

	// Build CNI 1.0.0 Result from daemon response
	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			{
				Name:    args.IfName,
				Sandbox: args.Netns,
			},
			{
				Name: resp.HostVeth,
			},
		},
		IPs: []*current.IPConfig{
			{
				Interface: current.Int(0),
				Address: net.IPNet{
					IP:   net.ParseIP(resp.IP),
					Mask: net.CIDRMask(resp.PrefixLen, 32),
				},
				Gateway: net.ParseIP(resp.Gateway),
			},
		},
		Routes: []*types.Route{
			{
				Dst: net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
				GW:  net.ParseIP(resp.Gateway),
			},
		},
	}

	return types.PrintResult(result, netConf.CNIVersion)
}

// cmdAddLocal is the fallback that runs the dataplane locally (standalone/test mode).
func cmdAddLocal(args *skel.CmdArgs, netConf *cniconfig.NetConf, namespace, podName, cidr string) error {
	dpManager, err := dataplane.NewManager(dataplane.Config{
		PodCIDR: cidr,
	})
	if err != nil {
		return fmt.Errorf("init dataplane manager: %w", err)
	}

	req := dataplane.PodNetworkRequest{
		ContainerID: args.ContainerID,
		NetnsPath:   args.Netns,
		IfName:      args.IfName,
		Namespace:   namespace,
		PodName:     podName,
		SegmentID:   sgtypes.SegmentID(netConf.SegmentID),
		Labels:      map[string]string{"k8s.io/namespace": namespace, "k8s.io/pod": podName},
		MTU:         netConf.MTU,
	}

	res, err := dpManager.AddPodNetwork(req)
	if err != nil {
		return fmt.Errorf("add pod network failed: %w", err)
	}

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			{
				Name:    args.IfName,
				Sandbox: args.Netns,
			},
			{
				Name: res.HostVeth,
			},
		},
		IPs: []*current.IPConfig{
			{
				Interface: current.Int(0),
				Address: net.IPNet{
					IP:   net.ParseIP(res.IP.String()),
					Mask: net.CIDRMask(res.PrefixLen, 32),
				},
				Gateway: net.ParseIP(res.Gateway.String()),
			},
		},
		Routes: []*types.Route{
			{
				Dst: net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
				GW:  net.ParseIP(res.Gateway.String()),
			},
		},
	}

	return types.PrintResult(result, netConf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	// Try daemon socket first
	conn, dialErr := dialDaemon()
	if dialErr == nil {
		defer conn.Close()
		req := &dataplane.CNIRequest{
			Command:     "DEL",
			ContainerID: args.ContainerID,
		}
		_, err := sendRequest(conn, req)
		if err != nil {
			// Graceful: DEL failures are non-fatal
			return nil
		}
		return nil
	}

	// Fallback: local dataplane
	netConf, err := cniconfig.LoadNetConf(args.StdinData)
	if err != nil {
		return nil // Graceful exit on invalid config during delete
	}

	cidr := netConf.IPAM.Subnet
	if cidr == "" {
		cidr = "10.244.0.0/16"
	}

	dpManager, err := dataplane.NewManager(dataplane.Config{
		PodCIDR: cidr,
	})
	if err != nil {
		return nil
	}

	_ = dpManager.DelPodNetwork(args.ContainerID)
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
