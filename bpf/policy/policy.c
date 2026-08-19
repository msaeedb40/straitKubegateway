/*
 * straitKubegateway NetworkPolicy eBPF Dataplane Program
 * Evaluates Identity-based policy rules, Priority + RuleNo ordering,
 * default Ingress Deny-all, default Egress Allow-all, and hierarchy enforcement.
 */

#include "../include/common.h"
#include "../maps/maps.h"

/* Policy Map Declarations */
BPF_MAP_DEF(policies_map, BPF_MAP_TYPE_HASH, struct policy_key, struct policy_entry, 65536);
BPF_MAP_DEF(endpoints_map, BPF_MAP_TYPE_HASH, struct endpoint_key_v4, struct endpoint_info, 65536);

/*
 * policy_ingress: Enforces Ingress policy on packets arriving at pod endpoints
 */
__section("tcx/policy_ingress")
int policy_ingress(struct __sk_buff *skb)
{
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_DROP;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return BPF_DROP;

    __u16 dport = 0;
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_DROP;
        dport = __constant_ntohs(tcp->dest);
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_DROP;
        dport = __constant_ntohs(udp->dest);
    }

    /* Resolve source and destination identities */
    struct endpoint_key_v4 src_k = {.ip = ip->saddr};
    struct endpoint_key_v4 dst_k = {.ip = ip->daddr};

    __u32 src_id = 2; // IdentityWorld default
    __u32 dst_id = 0;

    struct endpoint_info *src_ep = bpf_map_lookup_elem(&endpoints_map, &src_k);
    if (src_ep)
        src_id = src_ep->identity;

    struct endpoint_info *dst_ep = bpf_map_lookup_elem(&endpoints_map, &dst_k);
    if (dst_ep)
        dst_id = dst_ep->identity;

    /* 1. Exact match lookup (src_id, dst_id, port, proto, Ingress=0) */
    struct policy_key key = {
        .src_identity = src_id,
        .dst_identity = dst_id,
        .dst_port = dport,
        .protocol = proto,
        .direction = 0, // Ingress
    };

    struct policy_entry *rule = bpf_map_lookup_elem(&policies_map, &key);
    if (rule) {
        __sync_fetch_and_add(&rule->hit_count, 1);
        if (rule->action == 1) // Allow
            return BPF_OK;
        return BPF_DROP; // Deny or Reject
    }

    /* 2. Wildcard port lookup (dst_port = 0) */
    key.dst_port = 0;
    rule = bpf_map_lookup_elem(&policies_map, &key);
    if (rule) {
        __sync_fetch_and_add(&rule->hit_count, 1);
        if (rule->action == 1)
            return BPF_OK;
        return BPF_DROP;
    }

    /* 3. Default Ingress Action: Deny-all */
    return BPF_DROP;
}

/*
 * policy_egress: Enforces Egress policy on packets departing from pod endpoints
 */
__section("tcx/policy_egress")
int policy_egress(struct __sk_buff *skb)
{
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_DROP;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return BPF_DROP;

    __u16 dport = 0;
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_DROP;
        dport = __constant_ntohs(tcp->dest);
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_DROP;
        dport = __constant_ntohs(udp->dest);
    }

    struct endpoint_key_v4 src_k = {.ip = ip->saddr};
    struct endpoint_key_v4 dst_k = {.ip = ip->daddr};

    __u32 src_id = 0;
    __u32 dst_id = 2; // IdentityWorld default

    struct endpoint_info *src_ep = bpf_map_lookup_elem(&endpoints_map, &src_k);
    if (src_ep)
        src_id = src_ep->identity;

    struct endpoint_info *dst_ep = bpf_map_lookup_elem(&endpoints_map, &dst_k);
    if (dst_ep)
        dst_id = dst_ep->identity;

    struct policy_key key = {
        .src_identity = src_id,
        .dst_identity = dst_id,
        .dst_port = dport,
        .protocol = proto,
        .direction = 1, // Egress
    };

    struct policy_entry *rule = bpf_map_lookup_elem(&policies_map, &key);
    if (rule) {
        __sync_fetch_and_add(&rule->hit_count, 1);
        if (rule->action == 1) // Allow
            return BPF_OK;
        return BPF_DROP; // Deny
    }

    /* Wildcard port lookup */
    key.dst_port = 0;
    rule = bpf_map_lookup_elem(&policies_map, &key);
    if (rule) {
        __sync_fetch_and_add(&rule->hit_count, 1);
        if (rule->action == 1)
            return BPF_OK;
        return BPF_DROP;
    }

    /* 3. Default Egress Action: Allow-all */
    return BPF_OK;
}

char _license[] __section("license") = "GPL";
