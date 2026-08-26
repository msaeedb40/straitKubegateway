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

/* TCP Conntrack States */
#define CT_STATE_SYN_SENT     1
#define CT_STATE_SYN_RECV     2
#define CT_STATE_ESTABLISHED  3
#define CT_STATE_FIN_WAIT     4
#define CT_STATE_CLOSE        5
#define CT_STATE_UDP          6

/* 5-tuple Conntrack Key */
struct ct_tuple {
    __u32 saddr;
    __u32 daddr;
    __u16 sport;
    __u16 dport;
    __u8  proto;
    __u8  pad[3];
};

/* Conntrack Entry */
struct ct_entry {
    __u64 rx_packets;
    __u64 tx_packets;
    __u64 rx_bytes;
    __u64 tx_bytes;
    __u64 last_seen_ns;
    __u32 state;
    __u32 nat_ip;
    __u16 nat_port;
    __u16 flags;
};

/* Bi-directional Conntrack Table Map */
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct ct_tuple);
    __type(value, struct ct_entry);
    __uint(max_entries, 262144);
} ct_map SEC(".maps");

/* Conntrack Ingress Tracking Hook */
SEC("tc/conntrack_ingress")
int conntrack_ingress_track(struct __sk_buff *skb)
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

    __u16 sport = 0, dport = 0;
    __u8 tcp_flags = 0;

    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        sport = tcp->source;
        dport = tcp->dest;
        tcp_flags = ((__u8 *)tcp)[13];
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)ip + (ip->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        sport = udp->source;
        dport = udp->dest;
    } else {
        return BPF_OK;
    }

    struct ct_tuple key = {
        .saddr = ip->saddr,
        .daddr = ip->daddr,
        .sport = sport,
        .dport = dport,
        .proto = ip->protocol,
    };

    struct ct_entry *entry = bpf_map_lookup_elem(&ct_map, &key);
    __u64 now = bpf_ktime_get_ns();

    if (!entry) {
        struct ct_entry new_entry = {
            .rx_packets = 1,
            .rx_bytes = skb->len,
            .last_seen_ns = now,
            .state = (ip->protocol == IPPROTO_TCP) ? CT_STATE_SYN_SENT : CT_STATE_UDP,
        };
        bpf_map_update_elem(&ct_map, &key, &new_entry, BPF_ANY);
    } else {
        __sync_fetch_and_add(&entry->rx_packets, 1);
        __sync_fetch_and_add(&entry->rx_bytes, skb->len);
        entry->last_seen_ns = now;
        if (tcp_flags & 0x02) { // SYN
            entry->state = CT_STATE_SYN_RECV;
        } else if (tcp_flags & 0x10) { // ACK
            entry->state = CT_STATE_ESTABLISHED;
        } else if (tcp_flags & 0x01) { // FIN
            entry->state = CT_STATE_FIN_WAIT;
        } else if (tcp_flags & 0x04) { // RST
            entry->state = CT_STATE_CLOSE;
        }
    }

    return BPF_OK;
}

char _license[] SEC("license") = "Apache-2.0";
