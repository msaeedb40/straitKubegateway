// Package ipam provides IP address management for pod IP allocation.
package ipam

import (
	"fmt"
	"net/netip"
	"sync"
)

// Allocator manages allocation and release of IP addresses within a CIDR prefix.
type Allocator struct {
	mu        sync.Mutex
	prefix    netip.Prefix
	gateway   netip.Addr
	allocated map[netip.Addr]string // IP -> ContainerID
	lastAlloc netip.Addr
}

// NewAllocator creates an IP allocator for the given subnet prefix.
func NewAllocator(cidr string) (*Allocator, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	prefix = prefix.Masked()
	baseAddr := prefix.Addr()
	gateway := baseAddr.Next() // 1st usable address is typically gateway

	return &Allocator{
		prefix:    prefix,
		gateway:   gateway,
		allocated: make(map[netip.Addr]string),
		lastAlloc: gateway,
	}, nil
}

// Gateway returns the gateway IP for this pool.
func (a *Allocator) Gateway() netip.Addr {
	return a.gateway
}

// Prefix returns the allocated subnet prefix.
func (a *Allocator) Prefix() netip.Prefix {
	return a.prefix
}

// Allocate assigns the next available IP address to the given container ID.
func (a *Allocator) Allocate(containerID string) (netip.Addr, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cur := a.lastAlloc.Next()
	maxAttempts := int(1 << (a.prefix.Addr().BitLen() - a.prefix.Bits()))
	attempts := 0

	for attempts < maxAttempts {
		if !a.prefix.Contains(cur) {
			// Wrap around to start of prefix (skip network & gateway)
			cur = a.gateway.Next()
		}

		if _, exists := a.allocated[cur]; !exists {
			a.allocated[cur] = containerID
			a.lastAlloc = cur
			return cur, nil
		}

		cur = cur.Next()
		attempts++
	}

	return netip.Addr{}, fmt.Errorf("IP pool %s is exhausted", a.prefix)
}

// Release frees an allocated IP address associated with a container ID.
func (a *Allocator) Release(containerID string) (netip.Addr, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for ip, id := range a.allocated {
		if id == containerID {
			delete(a.allocated, ip)
			return ip, nil
		}
	}

	return netip.Addr{}, fmt.Errorf("no allocated IP found for container %q", containerID)
}

// ReleaseIP frees a specific IP address.
func (a *Allocator) ReleaseIP(ip netip.Addr) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.allocated[ip]; !exists {
		return fmt.Errorf("IP %s is not allocated", ip)
	}
	delete(a.allocated, ip)
	return nil
}

// AllocatedCount returns the number of currently allocated IP addresses.
func (a *Allocator) AllocatedCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.allocated)
}
