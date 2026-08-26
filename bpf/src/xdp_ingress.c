// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>

#include "../headers/bpf_helpers.h"
#include "../headers/bpf_endian.h"

/* Blocked CIDRs LPM Trie for DDoS / Early Filtering */
struct lpm_key {
    __u32 prefixlen;
    __u32 addr;
};

struct {
    __uint(type, BPF_MAP_TYPE_LPM_TRIE);
    __type(key, struct lpm_key);
    __type(value, __u32);
    __uint(max_entries, 16384);
    __uint(map_flags, BPF_F_NO_PREALLOC);
} blocklist_map SEC(".maps");

/* Early Ingress XDP Hook */
SEC("xdp/ingress")
int xdp_ingress_filter(struct xdp_md *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;

    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;

    /* Check blocklist */
    struct lpm_key key = {
        .prefixlen = 32,
        .addr = ip->saddr,
    };

    __u32 *action = bpf_map_lookup_elem(&blocklist_map, &key);
    if (action && *action == 1) {
        return XDP_DROP;
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "Apache-2.0";
