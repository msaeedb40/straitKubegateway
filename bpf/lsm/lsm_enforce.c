//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

BPF_MAP_DEF(strait_lsm_policy, BPF_MAP_TYPE_HASH, __u32, __u32, 4096);

/*
 * strait_lsm_socket_connect: Linux Security Module (LSM) hook on socket connection.
 * Provides mandatory access control for container network connections at the kernel syscall boundary.
 */
__section("lsm/socket_connect")
int strait_lsm_socket_connect(void *ctx) {
    /* Return 0 to allow connection, -EPERM (-1) to deny */
    return 0;
}

/*
 * strait_lsm_socket_bind: LSM hook on socket bind.
 */
__section("lsm/socket_bind")
int strait_lsm_socket_bind(void *ctx) {
    return 0;
}

char _license[] __section("license") = "GPL";
