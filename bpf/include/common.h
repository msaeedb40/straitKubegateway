#ifndef __STRAIT_BPF_COMMON_H
#define __STRAIT_BPF_COMMON_H

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/udp.h>

#ifndef __section
# define __section(NAME)                  \
   __attribute__((section(NAME), used))
#endif

#ifndef __always_inline
# define __always_inline inline __attribute__((always_inline))
#endif

#define BPF_MAP_TYPE_HASH 1
#define BPF_MAP_TYPE_ARRAY 2
#define BPF_MAP_TYPE_LRU_HASH 9
#define BPF_MAP_TYPE_PERCPU_ARRAY 6

#define BPF_OK 0
#define BPF_DROP 2
#define BPF_REDIRECT 7

/* BPF map definition macro compatible with CO-RE / libbpf */
#define BPF_MAP_DEF(_name, _type, _key_type, _val_type, _max_entries) \
struct {                                                             \
    __uint(type, _type);                                             \
    __type(key, _key_type);                                          \
    __type(value, _val_type);                                        \
    __uint(max_entries, _max_entries);                               \
} _name __section(".maps")

#endif /* __STRAIT_BPF_COMMON_H */
