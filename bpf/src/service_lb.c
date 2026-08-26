// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>

#include "../headers/bpf_helpers.h"
#include "../headers/bpf_endian.h"

#define MAGLEV_TABLE_SIZE 127

/* Service Key: VIP + Port + Protocol */
struct service_key {
    __u32 vip;
    __u16 port;
    __u8  proto;
    __u8  pad;
};

/* Service Value */
struct service_val {
    __u32 count;
    __u32 flags; /* DSR, SessionAffinity */
    __u32 maglev_table[MAGLEV_TABLE_SIZE];
};

/* Backend Key / Value */
struct backend_val {
    __u32 ip;
    __u16 port;
    __u8  proto;
    __u8  pad;
    __u8  mac[6];
    __u8  pad2[2];
};

/* Maps */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct service_key);
    __type(value, struct service_val);
    __uint(max_entries, 8192);
} service_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, struct backend_val);
    __uint(max_entries, 65536);
} backend_map SEC(".maps");

/* FNV-1a Hash for 5-tuple */
static __always_inline __u32 hash_flow(__u32 saddr, __u32 daddr, __u16 sport, __u16 dport, __u8 proto)
{
    __u32 hash = 2166136261U;
    hash = (hash ^ (saddr & 0xFF)) * 16777619U;
    hash = (hash ^ ((saddr >> 8) & 0xFF)) * 16777619U;
    hash = (hash ^ ((saddr >> 16) & 0xFF)) * 16777619U;
    hash = (hash ^ ((saddr >> 24) & 0xFF)) * 16777619U;
    hash = (hash ^ (daddr & 0xFF)) * 16777619U;
    hash = (hash ^ ((daddr >> 8) & 0xFF)) * 16777619U;
    hash = (hash ^ ((daddr >> 16) & 0xFF)) * 16777619U;
    hash = (hash ^ ((daddr >> 24) & 0xFF)) * 16777619U;
    hash = (hash ^ sport) * 16777619U;
    hash = (hash ^ dport) * 16777619U;
    hash = (hash ^ proto) * 16777619U;
    return hash;
}

/* Service Load Balancing TC Hook */
SEC("tc/service_lb")
int service_lb_ingress(struct __sk_buff *skb)
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

    if (ip->ihl < 5)
        return BPF_OK;

    __u16 dport = 0;
    __u16 sport = 0;

    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        dport = tcp->dest;
        sport = tcp->source;
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)ip + (ip->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        dport = udp->dest;
        sport = udp->source;
    } else {
        return BPF_OK;
    }

    struct service_key key = {
        .vip = ip->daddr,
        .port = dport,
        .proto = ip->protocol,
    };

    struct service_val *svc = bpf_map_lookup_elem(&service_map, &key);
    if (!svc || svc->count == 0)
        return BPF_OK;

    /* Compute Maglev backend index */
    __u32 flow_hash = hash_flow(ip->saddr, ip->daddr, sport, dport, ip->protocol);
    __u32 idx = svc->maglev_table[flow_hash % MAGLEV_TABLE_SIZE];

    struct backend_val *be = bpf_map_lookup_elem(&backend_map, &idx);
    if (!be)
        return BPF_OK;

    /* DNAT to backend IP/port */
    __u32 old_daddr = ip->daddr;
    __u32 new_daddr = be->ip;

    bpf_l3_csum_replace(skb, sizeof(struct ethhdr) + offsetof(struct iphdr, check),
                        old_daddr, new_daddr, sizeof(new_daddr));

    if (ip->protocol == IPPROTO_TCP) {
        bpf_l4_csum_replace(skb, sizeof(struct ethhdr) + (ip->ihl * 4) + offsetof(struct tcphdr, check),
                            old_daddr, new_daddr, BPF_F_PSEUDO_HDR | sizeof(new_daddr));
    } else if (ip->protocol == IPPROTO_UDP) {
        bpf_l4_csum_replace(skb, sizeof(struct ethhdr) + (ip->ihl * 4) + offsetof(struct udphdr, check),
                            old_daddr, new_daddr, BPF_F_PSEUDO_HDR | sizeof(new_daddr));
    }

    ip->daddr = new_daddr;

    return BPF_OK;
}

/* Socket connect4 Load Balancing Hook (kube-proxy replacement at socket creation) */
SEC("cgroup/connect4")
int sock4_connect_lb(struct bpf_sock_addr *ctx)
{
    struct service_key key = {
        .vip = ctx->user_ip4,
        .port = (__u16)(ctx->user_port >> 16),
        .proto = ctx->protocol,
    };

    struct service_val *svc = bpf_map_lookup_elem(&service_map, &key);
    if (!svc || svc->count == 0)
        return BPF_OK;

    __u32 hash = hash_flow(0, ctx->user_ip4, 0, key.port, ctx->protocol);
    __u32 idx = svc->maglev_table[hash % MAGLEV_TABLE_SIZE];

    struct backend_val *be = bpf_map_lookup_elem(&backend_map, &idx);
    if (be) {
        ctx->user_ip4 = be->ip;
        ctx->user_port = (__u32)be->port << 16;
    }

    return BPF_OK;
}

char _license[] SEC("license") = "Apache-2.0";
