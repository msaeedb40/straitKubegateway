/* SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause) */
#ifndef __BPF_HELPERS__
#define __BPF_HELPERS__

#include <linux/types.h>
#include <linux/bpf.h>

#define SEC(NAME) __attribute__((section(NAME), used))

/* Define BPF map attributes */
#define __uint(name, val) int (*name)[val]
#define __type(name, val) typeof(val) *name
#define __array(name, val) typeof(val) *name[]

/* Common BPF helper definitions */
static void *(*bpf_map_lookup_elem)(void *map, const void *key) = (void *) 1;
static long (*bpf_map_update_elem)(void *map, const void *key, const void *value, __u64 flags) = (void *) 2;
static long (*bpf_map_delete_elem)(void *map, const void *key) = (void *) 3;
static __u64 (*bpf_ktime_get_ns)(void) = (void *) 5;
static long (*bpf_trace_printk)(const char *fmt, __u32 fmt_size, ...) = (void *) 6;
static __u32 (*bpf_get_prandom_u32)(void) = (void *) 7;
static __u32 (*bpf_get_smp_processor_id)(void) = (void *) 8;
static long (*bpf_skb_store_bytes)(struct __sk_buff *skb, __u32 offset, const void *from, __u32 len, __u64 flags) = (void *) 9;
static long (*bpf_l3_csum_replace)(struct __sk_buff *skb, __u32 offset, __u32 from, __u32 to, __u64 flags) = (void *) 10;
static long (*bpf_l4_csum_replace)(struct __sk_buff *skb, __u32 offset, __u32 from, __u32 to, __u64 flags) = (void *) 11;
static long (*bpf_redirect)(__u32 ifindex, __u64 flags) = (void *) 23;
static long (*bpf_fib_lookup)(void *ctx, struct bpf_fib_lookup *params, int plen, __u32 flags) = (void *) 69;

#endif /* __BPF_HELPERS__ */
