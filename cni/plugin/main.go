// Package main implements the straitKubegateway CNI plugin binary.
package main

import (
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	cniconfig "github.com/straitKubegateway/straitKubegateway/cni/config"
	"github.com/straitKubegateway/straitKubegateway/dataplane"
	sgtypes "github.com/straitKubegateway/straitKubegateway/pkg/types"
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

	// Construct CNI 1.0.0 Result
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
