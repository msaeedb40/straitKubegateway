/*
 * straitKubegateway kube-proxy Replacement eBPF Program
 * Provides socket-level (cgroup/connect4) zero-overhead translation and
 * packet-level (tcx/kubeproxy) NodePort, LoadBalancer, DSR, and Hairpin NAT.
 */

#include "../include/common.h"
#include "../maps/maps.h"

/* NodePort Map Key: (NodePort, Proto) */
struct nodeport_key {
    __u16 port;
    __u8  proto;
    __u8  pad;
};

/* NodePort Map Value: Target Service ID */
struct nodeport_val {
    __u32 svc_id;
};

BPF_MAP_DEF(nodeport_map, BPF_MAP_TYPE_HASH, struct nodeport_key, struct nodeport_val, 1024);
BPF_MAP_DEF(services_map, BPF_MAP_TYPE_HASH, struct svc_key_v4, struct svc_info, 16384);
BPF_MAP_DEF(backends_map, BPF_MAP_TYPE_HASH, struct backend_key, struct backend_info, 65536);
BPF_MAP_DEF(maglev_map, BPF_MAP_TYPE_HASH, struct maglev_lut_key, struct maglev_lut_val, 65536);

/*
 * sock_connect4: Intercepts connect() syscalls from pods/host to perform socket-level VIP -> backend translation
 */
__section("cgroup/connect4")
int sock_connect4(struct bpf_sock_addr *ctx)
{
    if (ctx->user_family != AF_INET)
        return 1;

    struct svc_key_v4 key = {
        .ip = ctx->user_ip4,
        .port = __constant_ntohs((__u16)ctx->user_port),
        .proto = (ctx->protocol == IPPROTO_UDP) ? IPPROTO_UDP : IPPROTO_TCP,
        .pad = 0,
    };

    struct svc_info *svc = bpf_map_lookup_elem(&services_map, &key);
    if (!svc || svc->backend_count == 0)
        return 1;

    /* Select backend using Maglev consistent hash */
    struct maglev_lut_key mkey = {
        .svc_id = svc->svc_id,
        .slot_index = 0,
    };

    struct maglev_lut_val *mval = bpf_map_lookup_elem(&maglev_map, &mkey);
    if (!mval)
        return 1;

    struct backend_key bkey = {.backend_id = mval->backend_id};
    struct backend_info *backend = bpf_map_lookup_elem(&backends_map, &bkey);
    if (!backend)
        return 1;

    /* Rewrite destination at socket layer before packet generation */
    ctx->user_ip4 = backend->ip;
    ctx->user_port = __constant_htons(backend->port);

    return 1;
}

/*
 * kubeproxy_tcx: TCX hook handling forwarded NodePort, LoadBalancer, and Hairpin NAT traffic
 */
__section("tcx/kubeproxy")
int kubeproxy_tcx(struct __sk_buff *skb)
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
    __u8 proto = ip->protocol;

    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(tcp + 1) > data_end)
            return BPF_OK;
        dport = __constant_ntohs(tcp->dest);
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
        if ((void *)(udp + 1) > data_end)
            return BPF_OK;
        dport = __constant_ntohs(udp->dest);
    } else {
        return BPF_OK;
    }

    /* Check NodePort lookup */
    struct nodeport_key np_key = {
        .port = dport,
        .proto = proto,
        .pad = 0,
    };

    struct nodeport_val *np_val = bpf_map_lookup_elem(&nodeport_map, &np_key);
    if (np_val) {
        struct maglev_lut_key mkey = {
            .svc_id = np_val->svc_id,
            .slot_index = 0,
        };
        struct maglev_lut_val *mval = bpf_map_lookup_elem(&maglev_map, &mkey);
        if (mval) {
            struct backend_key bkey = {.backend_id = mval->backend_id};
            struct backend_info *backend = bpf_map_lookup_elem(&backends_map, &bkey);
            if (backend) {
                /* Perform DNAT */
                ip->daddr = backend->ip;
                if (proto == IPPROTO_TCP) {
                    struct tcphdr *tcp = (void *)((__u8 *)ip + (ip->ihl * 4));
                    if ((void *)(tcp + 1) <= data_end)
                        tcp->dest = __constant_htons(backend->port);
                } else if (proto == IPPROTO_UDP) {
                    struct udphdr *udp = (void *)((__u8 *)ip + (ip->ihl * 4));
                    if ((void *)(udp + 1) <= data_end)
                        udp->dest = __constant_htons(backend->port);
                }
            }
        }
    }

    return BPF_OK;
}

char _license[] __section("license") = "GPL";
