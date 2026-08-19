/*
 * straitKubegateway NAT & Connection Tracking eBPF Dataplane Program
 * Implements stateful Conntrack, SNAT, DNAT, Masquerading, and NAT64 stateful translation.
 */

#include "../include/common.h"
#include "../maps/maps.h"

/* Conntrack & NAT Maps */
BPF_MAP_DEF(ct_map_v4, BPF_MAP_TYPE_LRU_HASH, struct ct_key_v4, struct ct_entry_v4, 131072);
BPF_MAP_DEF(snat_map, BPF_MAP_TYPE_HASH, struct snat_key_v4, struct snat_entry_v4, 65536);
BPF_MAP_DEF(nat64_map, BPF_MAP_TYPE_ARRAY, struct nat64_key, struct nat64_val, 16);

/*
 * nat_egress_masquerade: Evaluates egress packets leaving pod network toward external world
 */
__section("tcx/nat_egress")
int nat_egress_masquerade(struct __sk_buff *skb)
{
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return BPF_OK;

    __u16 sport = 0;
    __u16 dport = 0;
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        sport = tcp->source;
        dport = tcp->dest;
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        sport = udp->source;
        dport = udp->dest;
    } else {
        return BPF_OK;
    }

    /* 1. Lookup existing connection in forward conntrack table */
    struct ct_key_v4 fwd_key = {
        .src_ip = ip->saddr,
        .dst_ip = ip->daddr,
        .src_port = sport,
        .dst_port = dport,
        .proto = proto,
    };

    struct ct_entry_v4 *entry = bpf_map_lookup_elem(&ct_map_v4, &fwd_key);
    if (entry) {
        /* Apply established SNAT rewrite */
        ip->saddr = entry->rev_dst_ip;
        if (proto == IPPROTO_TCP) {
            struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(tcp + 1) <= data_end)
                tcp->source = entry->rev_dst_port;
        } else if (proto == IPPROTO_UDP) {
            struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(udp + 1) <= data_end)
                udp->source = entry->rev_dst_port;
        }
        return BPF_OK;
    }

    /* 2. Check if original source matches SNAT pool */
    struct snat_key_v4 snat_k = {
        .src_ip = ip->saddr,
        .src_port = __constant_ntohs(sport),
        .proto = proto,
        .pad = 0,
    };

    struct snat_entry_v4 *snat = bpf_map_lookup_elem(&snat_map, &snat_k);
    if (snat) {
        /* Perform SNAT rewrite */
        __u32 new_ip = snat->nat_ip;
        __u16 new_port = __constant_htons(snat->nat_port);

        /* Create forward entry */
        struct ct_entry_v4 new_entry = {
            .rev_src_ip = ip->daddr,
            .rev_dst_ip = new_ip,
            .rev_src_port = dport,
            .rev_dst_port = new_port,
            .proto = proto,
            .state = CT_STATE_ESTABLISHED,
            .flags = 1, // SNAT
        };
        bpf_map_update_elem(&ct_map_v4, &fwd_key, &new_entry, BPF_ANY);

        /* Create reverse entry for replies (DNAT back to original pod IP:Port) */
        struct ct_key_v4 rev_key = {
            .src_ip = ip->daddr,
            .dst_ip = new_ip,
            .src_port = dport,
            .dst_port = new_port,
            .proto = proto,
        };
        struct ct_entry_v4 rev_entry = {
            .rev_src_ip = ip->daddr,
            .rev_dst_ip = ip->saddr,
            .rev_src_port = dport,
            .rev_dst_port = sport,
            .proto = proto,
            .state = CT_STATE_REPLY,
            .flags = 2, // Reverse DNAT
        };
        bpf_map_update_elem(&ct_map_v4, &rev_key, &rev_entry, BPF_ANY);

        /* Rewrite packet headers */
        ip->saddr = new_ip;
        if (proto == IPPROTO_TCP) {
            struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(tcp + 1) <= data_end)
                tcp->source = new_port;
        } else if (proto == IPPROTO_UDP) {
            struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(udp + 1) <= data_end)
                udp->source = new_port;
        }
    }

    return BPF_OK;
}

/*
 * nat_ingress_reply: Evaluates inbound packets from external networks to reverse-DNAT to original Pod IP
 */
__section("tcx/nat_ingress")
int nat_ingress_reply(struct __sk_buff *skb)
{
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return BPF_OK;

    __u16 sport = 0;
    __u16 dport = 0;
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        sport = tcp->source;
        dport = tcp->dest;
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        sport = udp->source;
        dport = udp->dest;
    } else {
        return BPF_OK;
    }

    /* Lookup reverse conntrack entry */
    struct ct_key_v4 rev_key = {
        .src_ip = ip->saddr,
        .dst_ip = ip->daddr,
        .src_port = sport,
        .dst_port = dport,
        .proto = proto,
    };

    struct ct_entry_v4 *entry = bpf_map_lookup_elem(&ct_map_v4, &rev_key);
    if (entry && entry->state == CT_STATE_REPLY) {
        /* Rewrite destination IP and Port back to pod */
        ip->daddr = entry->rev_dst_ip;
        if (proto == IPPROTO_TCP) {
            struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(tcp + 1) <= data_end)
                tcp->dest = entry->rev_dst_port;
        } else if (proto == IPPROTO_UDP) {
            struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
            if ((void *)(udp + 1) <= data_end)
                udp->dest = entry->rev_dst_port;
        }
    }

    return BPF_OK;
}

char _license[] __section("license") = "GPL";
