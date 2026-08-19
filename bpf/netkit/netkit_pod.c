/*
 * straitKubegateway NetKit eBPF Dataplane Program
 * Attached to NetKit primary/peer interfaces for fast pod-to-host and pod-to-pod forwarding.
 */

#include "../include/common.h"
#include "../maps/maps.h"

/* Map declarations */
BPF_MAP_DEF(endpoints_map, BPF_MAP_TYPE_HASH, struct endpoint_key_v4, struct endpoint_info, 65536);
BPF_MAP_DEF(routes_map, BPF_MAP_TYPE_HASH, struct route_key_v4, struct route_info, 65536);
BPF_MAP_DEF(policies_map, BPF_MAP_TYPE_HASH, struct policy_key, struct policy_entry, 65536);
BPF_MAP_DEF(metrics_map, BPF_MAP_TYPE_PERCPU_ARRAY, struct metric_key, struct metric_value, 256);

/*
 * netkit_pod_egress: Executed when a packet leaves a container netns via NetKit
 */
__section("netkit/peer")
int netkit_pod_egress(struct __sk_buff *skb)
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

    /* Look up destination in endpoint map for local pod-to-pod fast path */
    struct endpoint_key_v4 ep_key = {
        .ip = ip->daddr,
    };

    struct endpoint_info *ep = bpf_map_lookup_elem(&endpoints_map, &ep_key);
    if (ep) {
        /* Local endpoint destination: direct redirect */
        return bpf_redirect(ep->ifindex, 0);
    }

    /* Fallback to host network stack / routing */
    return BPF_OK;
}

/*
 * netkit_host_ingress: Executed on the host primary interface receiving packets for containers
 */
__section("netkit/primary")
int netkit_host_ingress(struct __sk_buff *skb)
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

    struct endpoint_key_v4 ep_key = {
        .ip = ip->daddr,
    };

    struct endpoint_info *ep = bpf_map_lookup_elem(&endpoints_map, &ep_key);
    if (ep) {
        return bpf_redirect(ep->ifindex, 0);
    }

    return BPF_OK;
}

char _license[] __section("license") = "GPL";
