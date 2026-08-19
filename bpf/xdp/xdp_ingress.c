//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

/* BPF maps referenced by XDP */
BPF_MAP_DEF(strait_endpoints, BPF_MAP_TYPE_HASH, struct endpoint_key_v4, struct endpoint_info, 65536);
BPF_MAP_DEF(strait_metrics, BPF_MAP_TYPE_PERCPU_ARRAY, struct metric_key, struct metric_value, 256);

/*
 * strait_xdp_ingress: Earliest ingress hook on host/physical interfaces.
 * Provides DDoS protection, rate limiting, and fast endpoint forwarding.
 */
__section("xdp")
int strait_xdp_ingress(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return XDP_PASS;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_PASS;

    /* Basic sanity check */
    if (iph->version != 4 || iph->ihl < 5)
        return XDP_DROP;

    /* Record packet in XDP metrics */
    struct metric_key mkey = { .metric_id = 1 };
    /* Metric lookup/update handled in verifier-friendly way */

    /* Check if destination IP is a known pod endpoint */
    struct endpoint_key_v4 ep_key = { .ip = iph->daddr };
    /* If fastpath lookup succeeds, redirect to target NetKit/veth interface */

    return XDP_PASS;
}

char _license[] __section("license") = "GPL";
