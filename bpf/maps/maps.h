#ifndef __STRAIT_BPF_MAPS_H
#define __STRAIT_BPF_MAPS_H

#include "../include/common.h"

/* LB Algorithm Identifiers */
#define LB_ALGO_ROUND_ROBIN          1
#define LB_ALGO_WEIGHTED_ROUND_ROBIN 2
#define LB_ALGO_MAGLEV_HASH          3
#define LB_ALGO_LEAST_CONNECTIONS    4
#define LB_ALGO_IP_HASH              5
#define LB_ALGO_RANDOM               6
#define LB_ALGO_FAILOVER             7

#define MAGLEV_TABLE_SIZE            128

/* Conntrack States */
#define CT_STATE_NEW                 1
#define CT_STATE_ESTABLISHED         2
#define CT_STATE_REPLY               3
#define CT_STATE_CLOSING             4
#define CT_STATE_CLOSED              5

/* Endpoint Map Key: IPv4 address */
struct endpoint_key_v4 {
    __u32 ip;
};

/* Endpoint Map Value: Pod networking metadata */
struct endpoint_info {
    __u32 ifindex;      /* Host netkit ifindex */
    __u32 identity;     /* BPF identity */
    __u32 segment_id;   /* Segment ID */
    __u8  mac[6];       /* Pod MAC address */
    __u16 pad;
    __u64 rx_bytes;
    __u64 tx_bytes;
};

/* Service Map Key: VIP, Port, Protocol */
struct svc_key_v4 {
    __u32 ip;           /* ClusterIP / VIP */
    __u16 port;         /* Service Port */
    __u8  proto;        /* IPPROTO_TCP or IPPROTO_UDP */
    __u8  pad;
};

/* Service Map Value: Service metadata and LB configuration */
struct svc_info {
    __u32 svc_id;           /* Unique 32-bit Service ID */
    __u32 backend_count;    /* Number of active backends */
    __u16 algo;             /* LB_ALGO_* */
    __u16 flags;            /* e.g. SessionAffinity, DSR, NodePort */
    __u32 affinity_timeout; /* Session affinity timeout in seconds (0 = disabled) */
};

/* Backend Map Key: Backend ID */
struct backend_key {
    __u32 backend_id;
};

/* Backend Map Value: Backend endpoint destination */
struct backend_info {
    __u32 ip;           /* Pod / Host IP */
    __u16 port;         /* Target Port */
    __u8  proto;        /* Protocol */
    __u8  flags;        /* Healthy / Active / DSR */
    __u32 weight;       /* Weight for WRR (default 100) */
    __u64 active_conns; /* Active connection counter for least_conns */
};

/* Maglev Lookup Table Key: Service ID + 0..127 slot index */
struct maglev_lut_key {
    __u32 svc_id;
    __u32 slot_index;   /* 0 .. MAGLEV_TABLE_SIZE-1 */
};

/* Maglev Lookup Table Value: Target Backend ID */
struct maglev_lut_val {
    __u32 backend_id;
};

/* Session Affinity Map Key: Client IP + Service ID */
struct affinity_key {
    __u32 client_ip;
    __u32 svc_id;
};

/* Session Affinity Map Value: Target Backend ID + Timestamp */
struct affinity_val {
    __u32 backend_id;
    __u64 last_seen_ns;
};

/* Conntrack Map Key: Forward 5-tuple */
struct ct_key_v4 {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8  proto;
    __u8  pad[3];
};

/* Conntrack Map Value: Reverse tuple translation & state */
struct ct_entry_v4 {
    __u32 rev_src_ip;   /* Translated source or destination IP */
    __u32 rev_dst_ip;
    __u16 rev_src_port;
    __u16 rev_dst_port;
    __u8  proto;
    __u8  state;        /* CT_STATE_* */
    __u16 flags;        /* SNAT, DNAT, MASQUERADE */
    __u64 last_updated_ns;
    __u64 rx_packets;
    __u64 rx_bytes;
    __u64 tx_packets;
    __u64 tx_bytes;
};

/* SNAT Mapping Key: (Original IP, Port, Proto) */
struct snat_key_v4 {
    __u32 src_ip;
    __u16 src_port;
    __u8  proto;
    __u8  pad;
};

/* SNAT Mapping Value: (Translated Egress IP, Port) */
struct snat_entry_v4 {
    __u32 nat_ip;
    __u16 nat_port;
    __u16 active_conns;
};

/* NAT64 Prefix Key */
struct nat64_key {
    __u32 id;
};

/* NAT64 Prefix Value: 96-bit prefix (default 64:ff9b::/96) */
struct nat64_val {
    __u32 prefix[3];    /* 96 bits */
    __u32 flags;
};

/* Routing Map Key: Destination CIDR prefix */
struct route_key_v4 {
    __u32 prefixlen;
    __u32 prefix;
};

/* Routing Map Value: Next-hop forwarding decision */
struct route_info {
    __u32 nexthop_ip;
    __u32 ifindex;
    __u32 flags;
    __u32 segment_id;
};

/* Policy Map Key: Source ID, Dest ID, Port, Protocol */
struct policy_key {
    __u32 src_identity;
    __u32 dst_identity;
    __u16 dst_port;
    __u8  protocol;
    __u8  direction;    /* 0=Ingress, 1=Egress */
};

/* Policy Map Value: Action and Priority */
struct policy_entry {
    __u32 action;       /* 1=Allow, 2=Deny, 3=Reject */
    __u32 priority;     /* lower = higher priority */
    __u32 rule_no;      /* 1-based index */
    __u64 hit_count;
};

/* Metrics Map Key: Metric ID */
struct metric_key {
    __u32 metric_id;
};

/* Metrics Map Value: Counter */
struct metric_value {
    __u64 count;
    __u64 bytes;
};

#endif /* __STRAIT_BPF_MAPS_H */
