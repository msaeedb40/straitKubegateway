// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>

#include "../headers/bpf_helpers.h"
#include "../headers/bpf_endian.h"

/* Policy Key: (SrcIdentity, DstIdentity, Port, Proto) */
struct policy_key {
    __u32 src_identity;
    __u32 dst_identity;
    __u16 dport;
    __u8  proto;
    __u8  direction; /* 0 = Ingress, 1 = Egress */
};

/* Policy Verdict: 1 = Allow, 0 = Deny, 2 = Reject */
struct policy_val {
    __u8  action;
    __u8  pad[3];
    __u32 rule_id;
};

/* Policy Map */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct policy_key);
    __type(value, struct policy_val);
    __uint(max_entries, 65536);
} policy_map SEC(".maps");

/* Policy Enforcement TC Ingress Hook */
SEC("tc/policy_ingress")
int policy_ingress_eval(struct __sk_buff *skb)
{
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_OK;

    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return BPF_OK;

    /* Extract source identity from skb mark (set by NetKit ingress) */
    __u32 src_identity = skb->mark;

    struct policy_key key = {
        .src_identity = src_identity,
        .dst_identity = 0,
        .dport = 0,
        .proto = ip->protocol,
        .direction = 0, // Ingress
    };

    struct policy_val *val = bpf_map_lookup_elem(&policy_map, &key);
    if (val) {
        if (val->action == 0) {
            /* Deny */
            return BPF_DROP;
        }
    }

    return BPF_OK;
}

char _license[] SEC("license") = "Apache-2.0";
