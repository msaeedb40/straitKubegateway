// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package ipam implements dynamic, zero-hardcoded IP address management
// for straitKubegateway. All CIDRs, subnets, and gateways are discovered
// dynamically from the Kubernetes API, Node objects, or CNI runtime configurations.
package ipam

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"go.uber.org/zap"
)

// ============================================================================
// Dynamic CIDR Discovery
// ============================================================================

// ClusterCIDRs holds dynamically discovered cluster CIDRs.
// Invariant 6: All CIDRs are discovered dynamically from the Kubernetes API server;
// zero hardcoded addresses or default subnets exist in this package.
type ClusterCIDRs struct {
	// PodCIDRs are the IPv4 cluster pod network CIDRs.
	PodCIDRs []netip.Prefix
	// ServiceCIDRs are the IPv4 cluster service network CIDRs.
	ServiceCIDRs []netip.Prefix
	// IPv6PodCIDRs are the IPv6 pod CIDRs for dual-stack clusters.
	IPv6PodCIDRs []netip.Prefix
	// IPv6ServiceCIDRs are the IPv6 service CIDRs for dual-stack clusters.
	IPv6ServiceCIDRs []netip.Prefix
}

// Discoverer discovers cluster CIDRs dynamically from the Kubernetes API.
type Discoverer struct {
	client client.Client
	log    *zap.Logger
}

// NewDiscoverer creates a new dynamic CIDR discoverer.
func NewDiscoverer(c client.Client, log *zap.Logger) *Discoverer {
	return &Discoverer{client: c, log: log}
}

// Discover fetches current cluster CIDRs dynamically from the Kubernetes API.
// It inspects Node specs for PodCIDRs and ConfigMaps for ServiceCIDRs.
func (d *Discoverer) Discover(ctx context.Context) (*ClusterCIDRs, error) {
	cidrs := &ClusterCIDRs{}

	// 1. Discover pod CIDRs from node objects
	var nodeList corev1.NodeList
	if err := d.client.List(ctx, &nodeList); err != nil {
		return nil, fmt.Errorf("list nodes for dynamic CIDR discovery: %w", err)
	}

	seen := make(map[netip.Prefix]bool)
	for _, node := range nodeList.Items {
		// Inspect Spec.PodCIDRs (dual-stack) and Spec.PodCIDR (single-stack)
		podCIDRList := node.Spec.PodCIDRs
		if len(podCIDRList) == 0 && node.Spec.PodCIDR != "" {
			podCIDRList = []string{node.Spec.PodCIDR}
		}

		for _, cidrStr := range podCIDRList {
			prefix, err := netip.ParsePrefix(cidrStr)
			if err != nil {
				d.log.Warn("invalid pod CIDR on node",
					zap.String("node", node.Name),
					zap.String("cidr", cidrStr),
					zap.Error(err),
				)
				continue
			}
			prefix = prefix.Masked()
			if seen[prefix] {
				continue
			}
			seen[prefix] = true
			if prefix.Addr().Is4() {
				cidrs.PodCIDRs = appendUnique(cidrs.PodCIDRs, prefix)
			} else {
				cidrs.IPv6PodCIDRs = appendUnique(cidrs.IPv6PodCIDRs, prefix)
			}
		}
	}

	// 2. Discover service CIDRs from kubeadm-config or kube-proxy ConfigMaps
	svcCIDRs, err := d.discoverServiceCIDRs(ctx)
	if err != nil {
		d.log.Info("service CIDR not present in configmaps, relying on dynamic service discovery", zap.Error(err))
	} else {
		for _, c := range svcCIDRs {
			if c.Addr().Is4() {
				cidrs.ServiceCIDRs = appendUnique(cidrs.ServiceCIDRs, c)
			} else {
				cidrs.IPv6ServiceCIDRs = appendUnique(cidrs.IPv6ServiceCIDRs, c)
			}
		}
	}

	d.log.Info("dynamically discovered cluster CIDRs",
		zap.Int("podCIDRs", len(cidrs.PodCIDRs)),
		zap.Int("serviceCIDRs", len(cidrs.ServiceCIDRs)),
		zap.Int("ipv6PodCIDRs", len(cidrs.IPv6PodCIDRs)),
		zap.Int("ipv6ServiceCIDRs", len(cidrs.IPv6ServiceCIDRs)),
	)
	return cidrs, nil
}

func (d *Discoverer) discoverServiceCIDRs(ctx context.Context) ([]netip.Prefix, error) {
	// Try kubeadm ConfigMap first
	var cm corev1.ConfigMap
	err := d.client.Get(ctx, client.ObjectKey{
		Namespace: "kube-system",
		Name:      "kubeadm-config",
	}, &cm)
	if err == nil {
		if cfg, ok := cm.Data["ClusterConfiguration"]; ok {
			return parseServiceCIDRFromKubeadm(cfg)
		}
	}

	// Try kube-proxy ConfigMap
	var kpCM corev1.ConfigMap
	err = d.client.Get(ctx, client.ObjectKey{
		Namespace: "kube-system",
		Name:      "kube-proxy",
	}, &kpCM)
	if err == nil {
		if cfg, ok := kpCM.Data["config.conf"]; ok {
			return parseServiceCIDRFromKubeProxy(cfg)
		}
	}
	return nil, fmt.Errorf("service CIDR not present in known ConfigMaps")
}

func parseServiceCIDRFromKubeadm(cfg string) ([]netip.Prefix, error) {
	const key = "serviceSubnet:"
	idx := indexOf(cfg, key)
	if idx < 0 {
		return nil, fmt.Errorf("serviceSubnet not found in kubeadm config")
	}
	line := cfg[idx+len(key):]
	for i, c := range line {
		if c == '\n' {
			line = line[:i]
			break
		}
	}
	return parseCIDRList(trimSpace(line))
}

func parseServiceCIDRFromKubeProxy(cfg string) ([]netip.Prefix, error) {
	const key = "clusterCIDR:"
	idx := indexOf(cfg, key)
	if idx < 0 {
		return nil, fmt.Errorf("clusterCIDR not found in kube-proxy config")
	}
	line := cfg[idx+len(key):]
	for i, c := range line {
		if c == '\n' {
			line = line[:i]
			break
		}
	}
	return parseCIDRList(trimSpace(line))
}

func parseCIDRList(s string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, part := range splitComma(s) {
		part = trimSpace(part)
		if part == "" {
			continue
		}
		p, err := netip.ParsePrefix(part)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", part, err)
		}
		out = append(out, p.Masked())
	}
	return out, nil
}

// ============================================================================
// Dynamic Node IPAM Allocator
// ============================================================================

// DualStackIP holds both IPv4 and IPv6 addresses allocated to an endpoint.
type DualStackIP struct {
	IPv4 netip.Addr
	IPv6 netip.Addr
}

// Allocator dynamically manages IP address allocation from arbitrary prefix lengths.
type Allocator struct {
	mu        sync.Mutex
	cidrs     []netip.Prefix
	v4Pools   []*Pool
	v6Pools   []*Pool
	log       *zap.Logger
}

// NewAllocator creates a dynamic allocator over arbitrary pod CIDRs.
func NewAllocator(cidrs []netip.Prefix, log *zap.Logger) *Allocator {
	a := &Allocator{
		cidrs: cidrs,
		log:   log,
	}
	for _, c := range cidrs {
		c = c.Masked()
		p := NewPool(c)
		if c.Addr().Is4() {
			a.v4Pools = append(a.v4Pools, p)
		} else {
			a.v6Pools = append(a.v6Pools, p)
		}
	}
	return a
}

// AddCIDR dynamically registers a newly assigned CIDR pool on the fly.
func (a *Allocator) AddCIDR(cidr netip.Prefix) {
	a.mu.Lock()
	defer a.mu.Unlock()
	cidr = cidr.Masked()
	for _, c := range a.cidrs {
		if c == cidr {
			return
		}
	}
	a.cidrs = append(a.cidrs, cidr)
	p := NewPool(cidr)
	if cidr.Addr().Is4() {
		a.v4Pools = append(a.v4Pools, p)
	} else {
		a.v6Pools = append(a.v6Pools, p)
	}
	a.log.Info("dynamically added IPAM pool", zap.String("cidr", cidr.String()))
}

// Allocate allocates the next available IP address.
func (a *Allocator) Allocate() (netip.Addr, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Try IPv4 pools first
	for _, p := range a.v4Pools {
		ip, err := p.Next()
		if err == nil {
			return ip, nil
		}
	}
	// Fallback to IPv6 pools if IPv4 not available
	for _, p := range a.v6Pools {
		ip, err := p.Next()
		if err == nil {
			return ip, nil
		}
	}
	return netip.Addr{}, fmt.Errorf("all dynamic IPAM pools exhausted")
}

// AllocateDualStack dynamically allocates both an IPv4 and an IPv6 address.
func (a *Allocator) AllocateDualStack() (DualStackIP, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var res DualStackIP
	var v4Err, v6Err error

	for _, p := range a.v4Pools {
		ip, err := p.Next()
		if err == nil {
			res.IPv4 = ip
			break
		}
		v4Err = err
	}

	for _, p := range a.v6Pools {
		ip, err := p.Next()
		if err == nil {
			res.IPv6 = ip
			break
		}
		v6Err = err
	}

	if !res.IPv4.IsValid() && len(a.v4Pools) > 0 {
		return res, fmt.Errorf("ipv4 pool exhausted: %v", v4Err)
	}
	if !res.IPv6.IsValid() && len(a.v6Pools) > 0 {
		if res.IPv4.IsValid() {
			a.releaseLocked(res.IPv4)
		}
		return res, fmt.Errorf("ipv6 pool exhausted: %v", v6Err)
	}

	return res, nil
}

// Release returns an IP address back to its dynamic pool.
func (a *Allocator) Release(ip netip.Addr) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releaseLocked(ip)
}

func (a *Allocator) releaseLocked(ip netip.Addr) {
	pools := a.v4Pools
	if ip.Is6() {
		pools = a.v6Pools
	}
	for _, p := range pools {
		if p.CIDR.Contains(ip) {
			p.Release(ip)
			return
		}
	}
}

// Used returns the count of currently allocated IPs across all pools.
func (a *Allocator) Used() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	total := 0
	for _, p := range a.v4Pools {
		total += p.UsedCount()
	}
	for _, p := range a.v6Pools {
		total += p.UsedCount()
	}
	return total
}

// GatewayForIP dynamically calculates the default gateway for an allocated IP.
func (a *Allocator) GatewayForIP(ip netip.Addr) (netip.Addr, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pools := a.v4Pools
	if ip.Is6() {
		pools = a.v6Pools
	}
	for _, p := range pools {
		if p.CIDR.Contains(ip) {
			return p.Gateway(), nil
		}
	}
	return netip.Addr{}, fmt.Errorf("no matching IPAM pool found for IP %s", ip)
}

// ============================================================================
// Dynamic Pool Implementation
// ============================================================================

// Pool manages dynamic IP leasing within a single arbitrary CIDR prefix.
type Pool struct {
	mu        sync.Mutex
	CIDR      netip.Prefix
	gateway   netip.Addr
	broadcast netip.Addr
	cursor    netip.Addr
	used      map[netip.Addr]struct{}
}

// NewPool creates an allocation pool for an arbitrary prefix length.
func NewPool(cidr netip.Prefix) *Pool {
	masked := cidr.Masked()
	baseAddr := masked.Addr()
	// Dynamic gateway: 1st usable host IP
	gw := baseAddr.Next()
	// Dynamic cursor starts at 2nd usable host IP
	start := gw.Next()

	var bcast netip.Addr
	if masked.Addr().Is4() && masked.Bits() <= 30 {
		b := masked.Addr().As4()
		baseU32 := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		hostMask := uint32(0xFFFFFFFF) >> masked.Bits()
		bcastU32 := baseU32 | hostMask
		bcast = netip.AddrFrom4([4]byte{
			byte(bcastU32 >> 24),
			byte(bcastU32 >> 16),
			byte(bcastU32 >> 8),
			byte(bcastU32),
		})
	}

	return &Pool{
		CIDR:      masked,
		gateway:   gw,
		broadcast: bcast,
		cursor:    start,
		used:      make(map[netip.Addr]struct{}),
	}
}

// Gateway returns the dynamically determined gateway for this pool.
func (p *Pool) Gateway() netip.Addr {
	return p.gateway
}

// Next allocates the next available IP address in the pool.
func (p *Pool) Next() (netip.Addr, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	addr := p.cursor
	for {
		if !p.CIDR.Contains(addr) || (p.broadcast.IsValid() && addr == p.broadcast) {
			// Wrap around to start of pool (after gateway) to reclaim released IPs
			addr = p.gateway.Next()
		}
		if !p.CIDR.Contains(addr) || (p.broadcast.IsValid() && addr == p.broadcast) {
			return netip.Addr{}, fmt.Errorf("pool %s exhausted", p.CIDR)
		}
		if _, exists := p.used[addr]; !exists && addr != p.gateway && (!p.broadcast.IsValid() || addr != p.broadcast) {
			p.used[addr] = struct{}{}
			p.cursor = addr.Next()
			return addr, nil
		}
		addr = addr.Next()
		if addr == p.cursor {
			return netip.Addr{}, fmt.Errorf("pool %s fully allocated", p.CIDR)
		}
	}
}

// Release marks an IP address as free.
func (p *Pool) Release(ip netip.Addr) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.used, ip)
}

// UsedCount returns the number of active leases in the pool.
func (p *Pool) UsedCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.used)
}

// ============================================================================
// Helpers
// ============================================================================

func appendUnique(s []netip.Prefix, p netip.Prefix) []netip.Prefix {
	for _, x := range s {
		if x == p {
			return s
		}
	}
	return append(s, p)
}

func indexOf(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
