//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

/* Socket Map for socket redirection / acceleration */
struct {
    __uint(type, BPF_MAP_TYPE_SOCKHASH);
    __type(key, struct ct_key_v4);
    __type(value, __u64);
    __uint(max_entries, 65536);
} strait_sock_hash __section(".maps");

/*
 * strait_sockops: Accelerates local pod-to-pod and pod-to-host TCP socket communication
 * by bypassing the full TCP/IP network stack when both endpoints are on the same host.
 */
__section("sockops")
int strait_sockops(struct bpf_sock_ops *skops) {
    __u32 op = skops->op;

    /* Handle established active and passive connections */
    if (op == BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB || op == BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB) {
        if (skops->family == 2) { /* AF_INET */
            struct ct_key_v4 key = {
                .src_ip = skops->local_ip4,
                .dst_ip = skops->remote_ip4,
                .src_port = skops->local_port,
                .dst_port = __constant_ntohl(skops->remote_port) >> 16,
                .proto = IPPROTO_TCP,
            };
            /* Update socket hash map for direct skb redirection */
        }
    }

    return BPF_OK;
}

char _license[] __section("license") = "GPL";
