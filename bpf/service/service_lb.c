/*
 * straitKubegateway Service Load Balancer eBPF Dataplane Program
 * Provides kernel-native L4 TCP/UDP load balancing, Maglev consistent hashing (128 slots),
 * round robin, session affinity, and NodePort acceleration.
 */

#include "../include/common.h"
#include "../maps/maps.h"

/* Service LB Map Declarations */
BPF_MAP_DEF(services_map, BPF_MAP_TYPE_HASH, struct svc_key_v4, struct svc_info, 16384);
BPF_MAP_DEF(backends_map, BPF_MAP_TYPE_HASH, struct backend_key, struct backend_info, 65536);
BPF_MAP_DEF(maglev_map, BPF_MAP_TYPE_HASH, struct maglev_lut_key, struct maglev_lut_val, 65536);
BPF_MAP_DEF(affinity_map, BPF_MAP_TYPE_LRU_HASH, struct affinity_key, struct affinity_val, 65536);

/* Simple 5-tuple hash function for Maglev slot selection */
static __always_inline __u32 hash_5tuple(__u32 saddr, __u32 daddr, __u16 sport, __u16 dport, __u8 proto)
{
    __u32 hash = saddr ^ daddr;
    hash ^= ((__u32)sport << 16) | dport;
    hash ^= proto;
    /* Murmur-like mixing */
    hash ^= hash >> 16;
    hash *= 0x85ebca6b;
    hash ^= hash >> 13;
    hash *= 0xc2b2ae35;
    hash ^= hash >> 16;
    return hash;
}

/*
 * service_lb_ingress: Evaluates incoming packets on host/bridge/TCX interfaces for Service VIP matching
 */
__section("tcx/service_lb")
int service_lb_ingress(struct __sk_buff *skb)
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

    __u16 dport = 0;
    __u16 sport = 0;
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        dport = tcp->dest;
        sport = tcp->source;
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        dport = udp->dest;
        sport = udp->source;
    } else {
        return BPF_OK;
    }

    /* Check if destination IP:Port is a Kubernetes Service VIP */
    struct svc_key_v4 svc_key = {
        .ip = ip->daddr,
        .port = __constant_ntohs(dport),
        .proto = proto,
        .pad = 0,
    };

    struct svc_info *svc = bpf_map_lookup_elem(&services_map, &svc_key);
    if (!svc || svc->backend_count == 0)
        return BPF_OK;

    __u32 target_backend_id = 0;

    /* Check Session Affinity */
    if (svc->affinity_timeout > 0) {
        struct affinity_key aff_key = {
            .client_ip = ip->saddr,
            .svc_id = svc->svc_id,
        };
        struct affinity_val *aff = bpf_map_lookup_elem(&affinity_map, &aff_key);
        if (aff) {
            target_backend_id = aff->backend_id;
        }
    }

    /* Fallback to LB Algorithm if no session affinity hit */
    if (target_backend_id == 0) {
        __u32 hash = hash_5tuple(ip->saddr, ip->daddr, sport, dport, proto);

        if (svc->algo == LB_ALGO_MAGLEV_HASH) {
            /* Maglev Hash: lookup in 128-slot Maglev LUT */
            struct maglev_lut_key mkey = {
                .svc_id = svc->svc_id,
                .slot_index = hash % MAGLEV_TABLE_SIZE,
            };
            struct maglev_lut_val *mval = bpf_map_lookup_elem(&maglev_map, &mkey);
            if (mval) {
                target_backend_id = mval->backend_id;
            }
        } else {
            /* Default Round-Robin / Hash: modulo backend count */
            __u32 slot = hash % svc->backend_count;
            struct maglev_lut_key mkey = {
                .svc_id = svc->svc_id,
                .slot_index = slot,
            };
            struct maglev_lut_val *mval = bpf_map_lookup_elem(&maglev_map, &mkey);
            if (mval) {
                target_backend_id = mval->backend_id;
            }
        }
    }

    if (target_backend_id == 0)
        return BPF_OK;

    /* Lookup chosen backend */
    struct backend_key bkey = {
        .backend_id = target_backend_id,
    };
    struct backend_info *backend = bpf_map_lookup_elem(&backends_map, &bkey);
    if (!backend)
        return BPF_OK;

    /* Perform DNAT: replace destination IP and Port */
    ip->daddr = backend->ip;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) <= data_end) {
            tcp->dest = __constant_htons(backend->port);
        }
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) <= data_end) {
            udp->dest = __constant_htons(backend->port);
        }
    }

    return BPF_OK;
}

char _license[] __section("license") = "GPL";
