//go:build ignore
#include "../include/common.h"
#include "../maps/maps.h"

/* Perf event array / ring buffer for sending flow events to userspace */
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024); /* 256KB ringbuf */
} strait_flow_events __section(".maps");

/*
 * strait_trace_drop: Tracepoint hook on kfree_skb to capture dropped packets
 * for network observability without impacting packet forwarding fast-path.
 */
__section("tracepoint/skb/kfree_skb")
int strait_trace_kfree_skb(void *ctx) {
    /* Capture drop reason and packet metadata for observability pipeline */
    return 0;
}

char _license[] __section("license") = "GPL";
