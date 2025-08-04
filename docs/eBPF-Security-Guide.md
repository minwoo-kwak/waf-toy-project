# 🔥 eBPF 기반 커널 레벨 보안 완전 가이드

## 📖 eBPF란 무엇인가?

### **개념**
- **Berkeley Packet Filter의 확장판**
- **커널 공간에서 실행되는 가상머신**
- **JIT 컴파일로 네이티브 코드 성능**
- **안전성 보장**: 커널 크래시 없이 실행

### **왜 혁명적인가?**
```
기존 방식:     유저스페이스 ←→ 커널 (느림, 복사 오버헤드)
eBPF 방식:     커널에서 직접 처리 (빠름, 제로 오버헤드)
```

## 🏭 **실무 활용 사례**

### **Facebook (Meta)**
```c
// DDoS 공격을 나노초 단위로 차단
// 초당 수억 패킷 처리 가능
if (is_ddos_attack(packet)) {
    return XDP_DROP;  // 커널에서 즉시 드롭
}
```

### **Cloudflare**
- 전 세계 엣지 서버에서 eBPF로 트래픽 필터링
- 기존 대비 **10배 빠른 처리 속도**

### **Google (Kubernetes)**
- Cilium CNI의 핵심 기술
- 서비스 메시 데이터플레인

### **Netflix**
- 네트워크 성능 최적화
- 실시간 트래픽 분석

## 🛠️ **eBPF 프로그램 구조 분석**

### **1. 맵(Map) 구조**
```c
// LRU 해시맵: 자동으로 오래된 엔트리 제거
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 10000);    // 최대 10,000개 IP 추적
    __type(key, __u32);           // IP 주소 (키)
    __type(value, __u64);         // 패킷 카운트 + 타임스탬프 (값)
} ip_rate_limit SEC(".maps");
```

**핵심 개념:**
- **유저스페이스-커널 공유**: 맵을 통해 데이터 교환
- **다양한 맵 타입**: HASH, ARRAY, LRU, RING_BUF 등
- **원자성 보장**: 동시 접근 시 데이터 무결성

### **2. 패킷 파싱**
```c
// 이더넷 → IP → TCP/UDP 순차적 파싱
struct ethhdr *eth = data;
struct iphdr *ip = (void *)(eth + 1);
struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);

// 경계 검사 필수 (커널 크래시 방지)
if ((void *)(tcp + 1) > data_end) {
    return XDP_PASS;
}
```

### **3. Rate Limiting 알고리즘**
```c
// 시간 기반 슬라이딩 윈도우 방식
__u64 packed_data = (timestamp << 32) | packet_count;

// 1초가 지났으면 카운터 리셋
if ((now - last_time) > 1000000000ULL) {
    // 새로운 윈도우 시작
}
```

## 🚀 **실무 성능 최적화 기법**

### **1. 빠른 경로 우선 처리**
```c
// 자주 발생하는 조건을 먼저 체크
if (likely(normal_traffic)) {
    return XDP_PASS;
}
if (unlikely(attack_detected)) {
    return XDP_DROP;
}
```

### **2. 메모리 효율적 데이터 구조**
```c
// 64비트에 여러 정보 압축
// 상위 32비트: 타임스탬프
// 하위 32비트: 패킷 카운트
__u64 packed = (time << 32) | count;
```

### **3. JIT 컴파일 최적화**
- **인라이닝**: 작은 함수는 인라인으로 처리
- **상수 폴딩**: 컴파일 시 상수 계산
- **레지스터 최적화**: CPU 레지스터 효율적 사용

## 🧪 **공격 탐지 알고리즘**

### **1. SYN Flood 탐지**
```c
static __always_inline int detect_syn_flood(struct tcphdr *tcp, __u32 src_ip) {
    if (tcp->syn && !tcp->ack) {
        // SYN:ACK 비율 분석
        // 임계값 초과 시 차단
    }
}
```

**실무 기법:**
- **적응형 임계값**: 트래픽 패턴에 따라 동적 조정
- **화이트리스트**: 신뢰할 수 있는 IP는 예외 처리
- **시간 기반 해제**: 일정 시간 후 자동 차단 해제

### **2. UDP Flood 탐지**
```c
// 대용량 UDP 패킷 의심
if (bpf_ntohs(ip->tot_len) > 1400) {
    return XDP_DROP;
}
```

### **3. Volumetric 공격 탐지**
```c
// 초당 패킷 수 기반 차단
#define MAX_PACKETS_PER_SEC 100
if (packet_count >= MAX_PACKETS_PER_SEC) {
    block_ip(src_ip);
}
```

## 📊 **모니터링 및 관측성**

### **통계 수집**
```c
// 다양한 공격 유형별 통계
enum stats_key {
    STAT_TOTAL_PACKETS,
    STAT_BLOCKED_PACKETS,
    STAT_SYN_FLOOD,
    STAT_UDP_FLOOD,
    STAT_HTTP_FLOOD
};
```

### **유저스페이스 연동**
```go
// Go 코드에서 eBPF 맵 읽기
func getEBPFStats() map[string]uint64 {
    stats := make(map[string]uint64)
    // eBPF 맵에서 통계 데이터 읽어오기
    return stats
}
```

## 🔧 **실무 배포 방법**

### **1. 컴파일**
```bash
# LLVM/Clang으로 eBPF 바이트코드 생성
clang -O2 -target bpf -c ddos_protection.c -o ddos_protection.o
```

### **2. 로드 및 어태치**
```bash
# XDP 프로그램을 네트워크 인터페이스에 어태치
ip link set dev eth0 xdp obj ddos_protection.o sec xdp
```

### **3. Kubernetes 연동**
```yaml
# DaemonSet으로 모든 노드에 배포
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ebpf-security
spec:
  template:
    spec:
      hostNetwork: true
      privileged: true  # eBPF 로드를 위해 필요
```

## 📈 **성능 벤치마크**

### **처리 성능**
- **기존 iptables**: ~1M pps (초당 백만 패킷)
- **eBPF/XDP**: ~10M pps (초당 천만 패킷)
- **성능 향상**: **10배 이상**

### **지연시간**
- **유저스페이스**: ~50μs
- **커널스페이스**: ~5μs  
- **지연시간 감소**: **10배 이상**

### **메모리 사용량**
- **기존 방식**: 패킷 복사로 메모리 낭비
- **eBPF**: 제자리 처리로 메모리 효율적

## 🎯 **학습 로드맵**

### **1단계: 기초 이해**
- C 언어 기본기
- 리눅스 네트워킹 스택
- 패킷 구조 (Ethernet, IP, TCP/UDP)

### **2단계: eBPF 기본**
- BPF 맵 타입 이해
- 패킷 파싱 기법
- 경계 검사 (Verifier 통과)

### **3단계: 고급 기능**
- 성능 최적화 기법
- 복잡한 공격 탐지 알고리즘
- 유저스페이스 연동

### **4단계: 실무 적용**
- Kubernetes 환경 배포
- 모니터링 및 관측성
- 운영 자동화

## 🛡️ **보안 고려사항**

### **eBPF 프로그램 보안**
- **Verifier**: 커널 크래시 방지
- **권한 관리**: CAP_SYS_ADMIN 필요
- **샌드박싱**: 제한된 커널 함수만 호출 가능

### **운영 보안**
- **화이트리스트**: 중요 서비스 IP 보호
- **점진적 배포**: 카나리 배포로 안전성 확보
- **롤백 계획**: 문제 발생 시 즉시 비활성화

## 📚 **추천 학습 자료**

### **공식 문서**
- [Linux Kernel eBPF Documentation](https://www.kernel.org/doc/html/latest/bpf/index.html)
- [Cilium eBPF Documentation](https://docs.cilium.io/en/latest/bpf/)

### **실습 환경**
- [eBPF Tutorial](https://github.com/xdp-project/xdp-tutorial)
- [BCC Examples](https://github.com/iovisor/bcc/tree/master/examples)

### **도서**
- "Learning eBPF" by Liz Rice
- "Systems Performance" by Brendan Gregg

## 💼 **커리어 관점**

### **시장 가치**
- **희소성**: eBPF 전문가 매우 부족
- **급여**: 일반 개발자 대비 30-50% 프리미엄
- **미래성**: 클라우드 네이티브의 핵심 기술

### **관련 직무**
- **Platform Engineer**: Kubernetes + eBPF
- **Security Engineer**: 커널 레벨 보안
- **Performance Engineer**: 시스템 최적화
- **Cloud Architect**: 클라우드 네이티브 설계

이 가이드를 바탕으로 eBPF를 마스터하면, 시장에서 매우 경쟁력 있는 개발자가 될 수 있습니다! 🚀