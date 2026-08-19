//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

/* BPF maps referenced by TCX */
BPF_MAP_DEF(strait_routes, BPF_MAP_TYPE_LPM_TRIE, struct route_key_v4, struct route_info, 16384);
BPF_MAP_DEF(strait_endpoints, BPF_MAP_TYPE_HASH, struct endpoint_key_v4, struct endpoint_info, 65536);

/*
 * strait_tcx_ingress: TCX hook on host network interfaces.
 * Provides fallback routing and host-level packet processing for non-NetKit traffic.
 */
__section("tcx/ingress")
int strait_tcx_ingress(struct __sk_buff *skb) {
    void *data_end = (void *)(long)skb->data_end;
    void *data = (void *)(long)skb->data;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return BPF_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return BPF_OK;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return BPF_OK;

    /* Process L3 routing for host ingress */
    return BPF_OK;
}

/*
 * strait_tcx_egress: TCX hook on host network egress.
 */
__section("tcx/egress")
int strait_tcx_egress(struct __sk_buff *skb) {
    return BPF_OK;
}

char _license[] __section("license") = "GPL";
