// Package service provides service load balancing abstractions, Maglev consistent
// hashing (128-slot lookup table), Round Robin, Weighted Round Robin, Least Connections,
// and session affinity management for straitKubegateway.
package service

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/netip"
	"sync/atomic"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Algorithm represents the load balancing algorithm.
type Algorithm uint16

const (
	AlgorithmRoundRobin         Algorithm = 1
	AlgorithmWeightedRoundRobin Algorithm = 2
	AlgorithmMaglevHash         Algorithm = 3
	AlgorithmLeastConnections   Algorithm = 4
	AlgorithmIPHash             Algorithm = 5
	AlgorithmRandom             Algorithm = 6
	AlgorithmFailover           Algorithm = 7
)

const (
	// MaglevTableSize is the lookup table size M = 128 as specified in architecture.
	MaglevTableSize = 128
)

// Service represents a Kubernetes Service definition with its load balancing configuration.
type Service struct {
	ID              uint32
	Namespace       string
	Name            string
	VIP             netip.Addr
	Port            uint16
	Protocol        types.Protocol
	Algorithm       Algorithm
	SessionAffinity bool
	AffinityTimeout uint32 // seconds
	Backends        []*Backend
	rrIndex         uint32
}

// Backend represents a target endpoint (Pod IP:Port) serving a Service.
type Backend struct {
	ID          uint32
	IP          netip.Addr
	Port        uint16
	Weight      uint32
	Healthy     bool
	ActiveConns uint64
}

// MaglevTable represents a 128-entry lookup table for a Service.
type MaglevTable struct {
	ServiceID uint32
	Lookup    [MaglevTableSize]uint32 // slot index -> Backend ID
}

// GenerateMaglevLUT computes the 128-slot Maglev lookup table for the given backends.
// It uses permutation generation based on backend hashes as described in the Google Maglev paper.
func GenerateMaglevLUT(serviceID uint32, backends []*Backend) MaglevTable {
	lut := MaglevTable{
		ServiceID: serviceID,
	}

	activeBackends := make([]*Backend, 0, len(backends))
	for _, b := range backends {
		if b.Healthy {
			activeBackends = append(activeBackends, b)
		}
	}

	n := len(activeBackends)
	if n == 0 {
		return lut
	}

	// 1. Generate permutation arrays for each backend
	// offset = hash1(backend) % M
	// skip = (hash2(backend) % (M/2))*2 + 1 (guarantees odd number, coprime to 128)
	permutation := make([][]int, n)
	for i, b := range activeBackends {
		h1 := hashOffset(b.IP.String(), b.Port)
		h2 := hashSkip(b.IP.String(), b.Port)

		offset := int(h1 % MaglevTableSize)
		skip := int((h2%(MaglevTableSize/2))*2 + 1)

		permutation[i] = make([]int, MaglevTableSize)
		for j := 0; j < MaglevTableSize; j++ {
			permutation[i][j] = (offset + j*skip) % MaglevTableSize
		}
	}

	// 2. Populate lookup table
	next := make([]int, n)
	entry := make([]int, MaglevTableSize)
	for i := range entry {
		entry[i] = -1
	}

	filled := 0
	for {
		for i := 0; i < n; i++ {
			c := permutation[i][next[i]%MaglevTableSize]
			for entry[c] >= 0 {
				next[i]++
				c = permutation[i][next[i]%MaglevTableSize]
			}
			entry[c] = i
			next[i]++
			filled++
			if filled == MaglevTableSize {
				for slot := 0; slot < MaglevTableSize; slot++ {
					lut.Lookup[slot] = activeBackends[entry[slot]].ID
				}
				return lut
			}
		}
	}
}

// SelectBackend selects a backend for a given packet using the Service's configured algorithm.
func (s *Service) SelectBackend(srcIP netip.Addr, srcPort uint16) (*Backend, error) {
	healthy := make([]*Backend, 0, len(s.Backends))
	for _, b := range s.Backends {
		if b.Healthy {
			healthy = append(healthy, b)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy backends available for service %s/%s", s.Namespace, s.Name)
	}

	switch s.Algorithm {
	case AlgorithmRoundRobin:
		idx := atomic.AddUint32(&s.rrIndex, 1) % uint32(len(healthy))
		return healthy[idx], nil

	case AlgorithmWeightedRoundRobin:
		var totalWeight uint32
		for _, b := range healthy {
			totalWeight += b.Weight
		}
		if totalWeight == 0 {
			return healthy[0], nil
		}
		r := rand.Uint32() % totalWeight
		var acc uint32
		for _, b := range healthy {
			acc += b.Weight
			if r < acc {
				return b, nil
			}
		}
		return healthy[0], nil

	case AlgorithmLeastConnections:
		var best *Backend
		var minConns uint64 = ^uint64(0)
		for _, b := range healthy {
			conns := atomic.LoadUint64(&b.ActiveConns)
			if conns < minConns {
				minConns = conns
				best = b
			}
		}
		return best, nil

	case AlgorithmIPHash:
		h := fnv.New32a()
		_, _ = h.Write([]byte(srcIP.String()))
		idx := h.Sum32() % uint32(len(healthy))
		return healthy[idx], nil

	case AlgorithmRandom:
		idx := rand.Intn(len(healthy))
		return healthy[idx], nil

	case AlgorithmFailover:
		return healthy[0], nil

	case AlgorithmMaglevHash:
		fallthrough
	default:
		lut := GenerateMaglevLUT(s.ID, healthy)
		h := fnv.New32a()
		_, _ = h.Write([]byte(fmt.Sprintf("%s:%d", srcIP, srcPort)))
		slot := h.Sum32() % MaglevTableSize
		targetID := lut.Lookup[slot]
		for _, b := range healthy {
			if b.ID == targetID {
				return b, nil
			}
		}
		return healthy[0], nil
	}
}

func hashOffset(ip string, port uint16) uint32 {
	h := md5.Sum([]byte(fmt.Sprintf("offset-%s-%d", ip, port)))
	return binary.LittleEndian.Uint32(h[:4])
}

func hashSkip(ip string, port uint16) uint32 {
	h := md5.Sum([]byte(fmt.Sprintf("skip-%s-%d", ip, port)))
	return binary.LittleEndian.Uint32(h[:4])
}
