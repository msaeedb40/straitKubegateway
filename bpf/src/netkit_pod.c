// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>

#include "../headers/bpf_helpers.h"
#include "../headers/bpf_endian.h"

/* Endpoint Map: IP -> Identity & Interface */
struct endpoint_info {
    __u32 identity;
    __u32 ifindex;
    __u32 segment_id;
    __u8  mac[6];
    __u8  pad[2];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32); /* IPv4 address */
    __type(value, struct endpoint_info);
    __uint(max_entries, 65536);
} endpoint_map SEC(".maps");

/* Metrics Counter Map */
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 16);
} metrics_map SEC(".maps");

/* NetKit Host Ingress Program */
SEC("netkit/ingress")
int netkit_host_ingress(struct __sk_buff *skb)
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

    /* Lookup destination endpoint */
    struct endpoint_info *ep = bpf_map_lookup_elem(&endpoint_map, &ip->daddr);
    if (ep) {
        /* Set security mark to identity */
        skb->mark = ep->identity;
        return bpf_redirect(ep->ifindex, 0);
    }

    return BPF_OK;
}

/* NetKit Container Egress Program */
SEC("netkit/egress")
int netkit_container_egress(struct __sk_buff *skb)
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

    /* Lookup source identity from map */
    struct endpoint_info *src_ep = bpf_map_lookup_elem(&endpoint_map, &ip->saddr);
    if (src_ep) {
        skb->mark = src_ep->identity;
    }

    return BPF_OK;
}

char _license[] SEC("license") = "Apache-2.0";
