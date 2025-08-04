// eBPF 프로그램: 커널 레벨에서 DDoS 공격 차단
// Facebook, Cloudflare에서 실제 사용하는 방식과 유사

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

// Rate Limiting을 위한 맵 구조
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 10000);
    __type(key, __u32);     // IP 주소
    __type(value, __u64);   // 패킷 카운트 + 타임스탬프
} ip_rate_limit SEC(".maps");

// 차단된 IP 목록
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1000);
    __type(key, __u32);     // IP 주소
    __type(value, __u64);   // 차단 시작 시간
} blocked_ips SEC(".maps");

// 통계 정보
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 10);
    __type(key, __u32);
    __type(value, __u64);
} stats SEC(".maps");

// 통계 키 정의
#define STAT_TOTAL_PACKETS    0
#define STAT_BLOCKED_PACKETS  1
#define STAT_SYN_FLOOD        2
#define STAT_UDP_FLOOD        3
#define STAT_HTTP_FLOOD       4

// Rate limiting 설정 (초당 패킷 수)
#define MAX_PACKETS_PER_SEC   100
#define BLOCK_DURATION_NS     (60 * 1000000000ULL)  // 60초

// 통계 업데이트 헬퍼 함수
static __always_inline void update_stats(__u32 key) {
    __u64 *value = bpf_map_lookup_elem(&stats, &key);
    if (value) {
        __sync_fetch_and_add(value, 1);
    } else {
        __u64 init_val = 1;
        bpf_map_update_elem(&stats, &key, &init_val, BPF_ANY);
    }
}

// IP 주소 rate limiting 체크
static __always_inline int check_rate_limit(__u32 src_ip) {
    __u64 now = bpf_ktime_get_ns();
    __u64 *rate_data = bpf_map_lookup_elem(&ip_rate_limit, &src_ip);
    
    if (!rate_data) {
        // 새로운 IP - 초기 카운트 설정
        __u64 new_data = (now & 0xFFFFFFFF00000000ULL) | 1;
        bpf_map_update_elem(&ip_rate_limit, &src_ip, &new_data, BPF_ANY);
        return XDP_PASS;
    }
    
    __u64 last_time = (*rate_data) >> 32;
    __u32 packet_count = (*rate_data) & 0xFFFFFFFF;
    
    // 1초가 지났으면 카운터 리셋
    if ((now - last_time) > 1000000000ULL) {
        __u64 new_data = (now & 0xFFFFFFFF00000000ULL) | 1;
        bpf_map_update_elem(&ip_rate_limit, &src_ip, &new_data, BPF_ANY);
        return XDP_PASS;
    }
    
    // Rate limit 초과 체크
    if (packet_count >= MAX_PACKETS_PER_SEC) {
        // IP를 차단 목록에 추가
        bpf_map_update_elem(&blocked_ips, &src_ip, &now, BPF_ANY);
        update_stats(STAT_BLOCKED_PACKETS);
        return XDP_DROP;
    }
    
    // 카운터 증가
    __u64 new_data = (last_time << 32) | (packet_count + 1);
    bpf_map_update_elem(&ip_rate_limit, &src_ip, &new_data, BPF_ANY);
    
    return XDP_PASS;
}

// 차단된 IP 체크
static __always_inline int check_blocked_ip(__u32 src_ip) {
    __u64 *block_time = bpf_map_lookup_elem(&blocked_ips, &src_ip);
    if (!block_time) {
        return XDP_PASS;
    }
    
    __u64 now = bpf_ktime_get_ns();
    
    // 차단 시간이 만료되었는지 체크
    if ((now - *block_time) > BLOCK_DURATION_NS) {
        bpf_map_delete_elem(&blocked_ips, &src_ip);
        return XDP_PASS;
    }
    
    update_stats(STAT_BLOCKED_PACKETS);
    return XDP_DROP;
}

// SYN Flood 공격 탐지
static __always_inline int detect_syn_flood(struct tcphdr *tcp, __u32 src_ip) {
    // SYN 패킷이고 ACK가 아닌 경우
    if (tcp->syn && !tcp->ack) {
        // 추가적인 SYN flood 탐지 로직
        // 실제로는 더 정교한 알고리즘 사용
        update_stats(STAT_SYN_FLOOD);
        
        // 간단한 임계값 기반 차단
        static __u32 syn_count = 0;
        syn_count++;
        
        if (syn_count > 50) {  // 임계값
            __u64 now = bpf_ktime_get_ns();
            bpf_map_update_elem(&blocked_ips, &src_ip, &now, BPF_ANY);
            syn_count = 0;  // 카운터 리셋
            return XDP_DROP;
        }
    }
    
    return XDP_PASS;
}

// 메인 XDP 프로그램
SEC("xdp")
int xdp_ddos_protection(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    
    // 이더넷 헤더 파싱
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        return XDP_PASS;
    }
    
    // IP 패킷만 처리
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return XDP_PASS;
    }
    
    // IP 헤더 파싱
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end) {
        return XDP_PASS;
    }
    
    __u32 src_ip = ip->saddr;
    
    // 통계 업데이트
    update_stats(STAT_TOTAL_PACKETS);
    
    // 1. 차단된 IP인지 먼저 체크
    int blocked_result = check_blocked_ip(src_ip);
    if (blocked_result == XDP_DROP) {
        return XDP_DROP;
    }
    
    // 2. Rate limiting 체크
    int rate_result = check_rate_limit(src_ip);
    if (rate_result == XDP_DROP) {
        return XDP_DROP;
    }
    
    // 3. 프로토콜별 공격 탐지
    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
        if ((void *)(tcp + 1) > data_end) {
            return XDP_PASS;
        }
        
        // SYN Flood 탐지
        int syn_result = detect_syn_flood(tcp, src_ip);
        if (syn_result == XDP_DROP) {
            return XDP_DROP;
        }
        
    } else if (ip->protocol == IPPROTO_UDP) {
        // UDP Flood 기본 탐지
        update_stats(STAT_UDP_FLOOD);
        
        // UDP 패킷 크기 체크 (대용량 UDP는 의심)
        if (bpf_ntohs(ip->tot_len) > 1400) {
            __u64 now = bpf_ktime_get_ns();
            bpf_map_update_elem(&blocked_ips, &src_ip, &now, BPF_ANY);
            return XDP_DROP;
        }
    }
    
    // 정상 패킷은 통과
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";