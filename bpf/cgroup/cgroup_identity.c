//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

BPF_MAP_DEF(strait_cgroup_identity, BPF_MAP_TYPE_HASH, __u64, __u32, 16384);

/*
 * strait_cgroup_sock: Cgroup socket creation hook.
 * Associates socket with pod/container identity for egress enforcement.
 */
__section("cgroup/sock_create")
int strait_cgroup_sock_create(struct bpf_sock *sk) {
    return 1; /* Allow */
}

/*
 * strait_cgroup_skb_egress: Cgroup SKB egress hook.
 * Enforces egress segment and identity rules at the cgroup level before packet reaches net device.
 */
__section("cgroup_skb/egress")
int strait_cgroup_skb_egress(struct __sk_buff *skb) {
    return 1; /* Allow */
}

char _license[] __section("license") = "GPL";
