// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package bpf

import (
	"testing"
	"unsafe"
)

func TestBPFMapCreateAttrAlignment(t *testing.T) {
	var attr bpfMapCreateAttr

	// In Linux kernel uapi for union bpf_attr BPF_MAP_CREATE:
	// map_type (4) + key_size (4) + value_size (4) + max_entries (4) + map_flags (4)
	// + inner_map_fd (4) + numa_node (4) = 28 bytes offset to map_name.
	offset := unsafe.Offsetof(attr.mapName)
	if offset != 28 {
		t.Fatalf("expected offset of mapName to be 28, got %d", offset)
	}
}

func TestByteOrderHelpers(t *testing.T) {
	var port uint16 = 8080
	netPort := Htons(port)
	hostPort := Ntohs(netPort)
	if hostPort != port {
		t.Fatalf("expected Ntohs(Htons(%d)) == %d, got %d", port, port, hostPort)
	}

	var ip uint32 = 0x0A000105
	netIP := Htonl(ip)
	hostIP := Ntohl(netIP)
	if hostIP != ip {
		t.Fatalf("expected Ntohl(Htonl(%x)) == %x, got %x", ip, ip, hostIP)
	}
}
