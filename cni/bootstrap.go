// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// EnsureBootstrapNAT configures kernel forwarding, outbound MASQUERADE,
// and redirects the in-cluster Kubernetes Service VIP (10.96.0.1:443) directly to the
// control plane API server when kube-proxy is disabled (kubeProxyMode: none).
// This allows non-hostNetwork pods like CoreDNS and local-path-provisioner to reach
// the Kubernetes API without depending on an external Service proxy.
func EnsureBootstrapNAT(apiServerHost string, apiServerPort int, podCIDR string, log *zap.Logger) error {
	// 1. Enable IPv4 forwarding
	_ = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0644)

	// 2. Accept forwarded traffic across interfaces
	_ = exec.Command("iptables", "-P", "FORWARD", "ACCEPT").Run()

	// 3. Masquerade outbound traffic from pod CIDR
	if podCIDR == "" {
		podCIDR = "10.244.0.0/16"
	}
	_ = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", podCIDR, "!", "-o", "ptp+", "-j", "MASQUERADE").Run()

	// 4. Resolve real API server endpoint if given in-cluster service IP 10.96.0.1
	realHost, realPort := discoverAPIServer(apiServerHost, apiServerPort)
	if realHost == "" || realHost == "10.96.0.1" {
		log.Warn("could not determine upstream API server address for bootstrap DNAT",
			zap.String("host", realHost),
			zap.Int("port", realPort),
		)
		return nil
	}

	dest := fmt.Sprintf("%s:%d", realHost, realPort)
	log.Info("configuring bootstrap API server DNAT rule", zap.String("destination", dest))

	// PREROUTING for pods
	_ = exec.Command("iptables", "-t", "nat", "-C", "PREROUTING", "-p", "tcp", "-d", "10.96.0.1", "--dport", "443", "-j", "DNAT", "--to-destination", dest).Run()
	_ = exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "-d", "10.96.0.1", "--dport", "443", "-j", "DNAT", "--to-destination", dest).Run()

	// OUTPUT for host
	_ = exec.Command("iptables", "-t", "nat", "-C", "OUTPUT", "-p", "tcp", "-d", "10.96.0.1", "--dport", "443", "-j", "DNAT", "--to-destination", dest).Run()
	_ = exec.Command("iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "-d", "10.96.0.1", "--dport", "443", "-j", "DNAT", "--to-destination", dest).Run()

	return nil
}

func discoverAPIServer(host string, port int) (string, int) {
	if host != "" && host != "10.96.0.1" {
		return resolveHost(host), port
	}

	// Try reading kubeconfig files for the upstream API server URL
	kubeConfigFiles := []string{
		"/etc/kubernetes/kubelet.conf",
		"/etc/kubernetes/admin.conf",
		"/host/etc/kubernetes/kubelet.conf",
		"/var/lib/kubelet/kubeconfig",
	}

	for _, kf := range kubeConfigFiles {
		if f, err := os.Open(kf); err == nil {
			parsedHost, parsedPort, found := parseAPIServerFromKubeconfig(f)
			_ = f.Close()
			if found {
				return resolveHost(parsedHost), parsedPort
			}
		}
	}

	// Dynamically derive control plane candidate names
	var candidates []string

	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		if strings.Contains(nodeName, "-control-plane") {
			candidates = append(candidates, nodeName)
		} else if idx := strings.Index(nodeName, "-worker"); idx != -1 {
			candidates = append(candidates, nodeName[:idx]+"-control-plane")
		} else if idx := strings.LastIndex(nodeName, "-"); idx != -1 {
			candidates = append(candidates, nodeName[:idx]+"-control-plane")
		}
	}

	if clusterName := os.Getenv("CLUSTER_NAME"); clusterName != "" {
		candidates = append(candidates, clusterName+"-control-plane")
	}

	// Common well-known Kind control plane hostnames
	candidates = append(candidates,
		"ci-kind-test-control-plane",
		"straitkubegateway-control-plane",
		"kind-control-plane",
		"cluster0-control-plane",
		"kubernetes.default.svc",
	)

	for _, candidate := range candidates {
		if resolved := resolveHost(candidate); resolved != candidate && resolved != "10.96.0.1" {
			return resolved, 6443
		}
	}

	return host, port
}

func parseAPIServerFromKubeconfig(f *os.File) (string, int, bool) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "server:") {
			rawURL := strings.TrimSpace(strings.TrimPrefix(line, "server:"))
			if u, parseErr := url.Parse(rawURL); parseErr == nil {
				h, p, splitErr := net.SplitHostPort(u.Host)
				if splitErr == nil {
					var parsedPort int
					fmt.Sscanf(p, "%d", &parsedPort)
					return h, parsedPort, true
				}
				return u.Hostname(), 6443, true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", 0, false
	}
	return "", 0, false
}

func resolveHost(host string) string {
	if ips, err := net.LookupHost(host); err == nil && len(ips) > 0 {
		// Prefer IPv4 addresses for iptables NAT compatibility
		for _, ip := range ips {
			parsed := net.ParseIP(ip)
			if parsed != nil && parsed.To4() != nil {
				return ip
			}
		}
		return ips[0]
	}
	return host
}
