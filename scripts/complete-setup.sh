#!/bin/bash

# 완전 자동화된 WAF 시스템 배포 스크립트
# ELK Stack + Prometheus + Grafana + Redis + WAF Gateway

set -e

# 색상 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 로깅 함수들
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "${PURPLE}[HEADER]${NC} $1"; }

# 배너 출력
print_banner() {
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║  🛡️  차세대 WAF 시스템 완전 자동 배포 스크립트                      ║
║                                                            ║
║  • ModSecurity + OWASP CRS                                 ║
║  • Redis Rate Limiting                                     ║
║  • ELK Stack (Elasticsearch + Logstash + Kibana)          ║
║  • Prometheus + Grafana 모니터링                            ║
║  • 멀티테넌트 관리 시스템                                      ║
║  • 실시간 위협 탐지                                         ║
║                                                            ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# 사전 요구사항 확인
check_prerequisites() {
    log_header "=== 사전 요구사항 확인 ==="
    
    local missing_tools=()
    
    # 필수 도구들 확인
    tools=("kubectl" "helm" "docker")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        else
            version=$(command -v "$tool" --version 2>/dev/null | head -1 || echo "unknown")
            log_info "$tool: ✅ 설치됨 ($version)"
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "다음 도구들이 설치되지 않았습니다: ${missing_tools[*]}"
        log_info "설치 가이드:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                kubectl)
                    echo "  kubectl: https://kubernetes.io/docs/tasks/tools/"
                    ;;
                helm)
                    echo "  helm: https://helm.sh/docs/intro/install/"
                    ;;
                docker)
                    echo "  docker: https://docs.docker.com/get-docker/"
                    ;;
            esac
        done
        exit 1
    fi
    
    log_success "모든 사전 요구사항이 충족되었습니다"
}

# Kubernetes 클러스터 확인
check_kubernetes_cluster() {
    log_header "=== Kubernetes 클러스터 연결 확인 ==="
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Kubernetes 클러스터에 연결할 수 없습니다"
        log_warning "Docker Desktop의 Kubernetes 설정을 확인하세요:"
        log_warning "1. Docker Desktop → Settings → Kubernetes"
        log_warning "2. Enable Kubernetes 체크"
        log_warning "3. Apply & Restart 클릭"
        exit 1
    fi
    
    # 클러스터 정보 출력
    log_info "클러스터 정보:"
    kubectl cluster-info | sed 's/^/  /'
    
    # 노드 상태 확인
    log_info "노드 상태:"
    kubectl get nodes | sed 's/^/  /'
    
    log_success "Kubernetes 클러스터 연결 확인 완료"
}

# Helm 리포지토리 추가
setup_helm_repositories() {
    log_header "=== Helm 리포지토리 설정 ==="
    
    # 필요한 Helm 리포지토리들
    repos=(
        "prometheus-community https://prometheus-community.github.io/helm-charts"
        "grafana https://grafana.github.io/helm-charts"
        "elastic https://helm.elastic.co"
        "ingress-nginx https://kubernetes.github.io/ingress-nginx"
        "bitnami https://charts.bitnami.com/bitnami"
    )
    
    for repo in "${repos[@]}"; do
        name=$(echo $repo | cut -d' ' -f1)
        url=$(echo $repo | cut -d' ' -f2)
        
        log_info "Helm 리포지토리 추가: $name"
        helm repo add "$name" "$url" > /dev/null 2>&1 || true
    done
    
    log_info "Helm 리포지토리 업데이트 중..."
    helm repo update > /dev/null 2>&1
    
    log_success "Helm 리포지토리 설정 완료"
}

# 네임스페이스 생성
create_namespaces() {
    log_header "=== 네임스페이스 생성 ==="
    
    namespaces=("waf-system" "monitoring" "logging")
    
    for ns in "${namespaces[@]}"; do
        if kubectl get namespace "$ns" &> /dev/null; then
            log_warning "네임스페이스 '$ns'가 이미 존재합니다"
        else
            kubectl create namespace "$ns"
            log_success "네임스페이스 '$ns' 생성 완료"
        fi
    done
}

# WAF 시스템 배포
deploy_waf_system() {
    log_header "=== WAF 시스템 배포 ==="
    
    # WAF 네임스페이스 생성
    kubectl apply -f deployments/k8s/namespace.yaml
    
    # Redis 배포
    log_info "Redis 배포 중..."
    kubectl apply -f deployments/k8s/redis.yaml
    kubectl wait --for=condition=ready pod -l app=redis -n waf-system --timeout=300s
    log_success "Redis 배포 완료"
    
    # OWASP CRS 설정 적용
    log_info "OWASP CRS 설정 적용 중..."
    kubectl apply -f deployments/k8s/owasp-crs-config.yaml
    log_success "OWASP CRS 설정 적용 완료"
    
    # ModSecurity + NGINX Ingress 배포
    log_info "ModSecurity + NGINX Ingress 배포 중..."
    kubectl apply -f deployments/k8s/nginx-ingress-modsecurity.yaml
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=nginx-ingress -n waf-system --timeout=600s
    log_success "WAF 시스템 배포 완료"
    
    # 테스트 애플리케이션 배포
    log_info "테스트 애플리케이션 배포 중..."
    kubectl apply -f deployments/k8s/test-app.yaml
    kubectl wait --for=condition=ready pod -l app=test-webapp -n waf-system --timeout=300s
    log_success "테스트 애플리케이션 배포 완료"
}

# 모니터링 스택 배포 (Prometheus + Grafana)
deploy_monitoring_stack() {
    log_header "=== 모니터링 스택 배포 ==="
    
    # Prometheus + Grafana 배포
    log_info "Prometheus + Grafana 배포 중..."
    kubectl apply -f deployments/k8s/prometheus-grafana.yaml
    
    # Prometheus 준비 대기
    log_info "Prometheus 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app=prometheus -n waf-system --timeout=600s
    
    # Grafana 준비 대기
    log_info "Grafana 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app=grafana -n waf-system --timeout=600s
    
    log_success "모니터링 스택 배포 완료"
}

# 로깅 스택 배포 (ELK)
deploy_logging_stack() {
    log_header "=== 로깅 스택 배포 (ELK) ==="
    
    # ELK Stack 배포
    log_info "Elasticsearch 배포 중..."
    kubectl apply -f deployments/k8s/elasticsearch-kibana.yaml
    
    # Elasticsearch 준비 대기 (더 긴 시간 필요)
    log_info "Elasticsearch 시작 대기 중... (최대 10분)"
    kubectl wait --for=condition=ready pod -l app=elasticsearch -n waf-system --timeout=600s
    
    # Kibana 준비 대기
    log_info "Kibana 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app=kibana -n waf-system --timeout=600s
    
    # Logstash 준비 대기
    log_info "Logstash 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app=logstash -n waf-system --timeout=600s
    
    log_success "로깅 스택 배포 완료"
}

# WAF Gateway 애플리케이션 빌드 및 배포
deploy_waf_gateway() {
    log_header "=== WAF Gateway 애플리케이션 배포 ==="
    
    # Docker 이미지 빌드
    log_info "WAF Gateway Docker 이미지 빌드 중..."
    docker build -t waf-gateway:latest -f deployments/docker/Dockerfile .
    
    # Kubernetes에 배포
    log_info "WAF Gateway 배포 중..."
    
    # WAF Gateway Deployment 생성
    cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: waf-gateway
  namespace: waf-system
  labels:
    app: waf-gateway
spec:
  replicas: 2
  selector:
    matchLabels:
      app: waf-gateway
  template:
    metadata:
      labels:
        app: waf-gateway
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: waf-gateway
        image: waf-gateway:latest
        imagePullPolicy: Never  # 로컬 이미지 사용
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: REDIS_URL
          value: "redis://redis-service.waf-system.svc.cluster.local:6379"
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: 256Mi
            cpu: 200m
          limits:
            memory: 512Mi
            cpu: 500m
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: waf-gateway-service
  namespace: waf-system
  labels:
    app: waf-gateway
spec:
  type: ClusterIP
  ports:
    - port: 8080
      targetPort: 8080
      protocol: TCP
      name: http
  selector:
    app: waf-gateway
EOF

    # 배포 완료 대기
    kubectl wait --for=condition=ready pod -l app=waf-gateway -n waf-system --timeout=300s
    
    log_success "WAF Gateway 배포 완료"
}

# 포트 포워딩 설정
setup_port_forwarding() {
    log_header "=== 서비스 접근을 위한 포트 포워딩 설정 ==="
    
    # 백그라운드에서 포트 포워딩 실행
    log_info "포트 포워딩 설정 중..."
    
    # WAF Gateway
    kubectl port-forward svc/waf-gateway-service 8080:8080 -n waf-system > /dev/null 2>&1 &
    WAF_PID=$!
    
    # Grafana
    kubectl port-forward svc/grafana 3000:3000 -n waf-system > /dev/null 2>&1 &
    GRAFANA_PID=$!
    
    # Kibana
    kubectl port-forward svc/kibana 5601:5601 -n waf-system > /dev/null 2>&1 &
    KIBANA_PID=$!
    
    # Prometheus
    kubectl port-forward svc/prometheus 9090:9090 -n waf-system > /dev/null 2>&1 &
    PROMETHEUS_PID=$!
    
    # 포트 포워딩이 시작될 시간 대기
    sleep 5
    
    log_success "포트 포워딩 설정 완료"
    
    # PID 파일에 저장 (나중에 종료용)
    echo "$WAF_PID $GRAFANA_PID $KIBANA_PID $PROMETHEUS_PID" > /tmp/waf-port-forwards.pid
}

# 배포 상태 확인
check_deployment_status() {
    log_header "=== 배포 상태 확인 ==="
    
    echo -e "${YELLOW}📊 전체 Pod 상태:${NC}"
    kubectl get pods -n waf-system -o wide
    
    echo -e "\n${YELLOW}🌐 서비스 상태:${NC}"
    kubectl get svc -n waf-system
    
    echo -e "\n${YELLOW}🔗 Ingress 상태:${NC}"
    kubectl get ingress -n waf-system
    
    echo -e "\n${YELLOW}📈 리소스 사용량:${NC}"
    kubectl top pods -n waf-system 2>/dev/null || echo "  metrics-server가 설치되지 않아 리소스 사용량을 확인할 수 없습니다"
}

# 접속 정보 출력
show_access_information() {
    log_header "=== 🎉 배포 완료! 접속 정보 ==="
    
    cat << EOF

${GREEN}✅ WAF 시스템이 성공적으로 배포되었습니다!${NC}

${YELLOW}📋 서비스 접속 정보:${NC}
┌─────────────────────────────────────────────────────────────┐
│ 🛡️  WAF Gateway:     http://localhost:8080                  │
│     • 헬스체크:        /health                              │
│     • WAF 상태:        /waf/status                          │
│     • 메트릭:          /metrics                             │
│     • 테스트:          /test                                │
│                                                           │
│ 📊 Grafana:          http://localhost:3000                 │
│     • 사용자명/비밀번호: admin / admin123!                    │
│     • WAF 대시보드 자동 구성됨                               │
│                                                           │
│ 🔍 Kibana:           http://localhost:5601                 │
│     • 로그 분석 및 검색                                      │
│     • 실시간 보안 이벤트 모니터링                             │
│                                                           │
│ 📈 Prometheus:       http://localhost:9090                 │
│     • 메트릭 쿼리 및 알림 설정                               │
│     • WAF 성능 지표 확인                                    │
└─────────────────────────────────────────────────────────────┘

${BLUE}🧪 테스트 명령어:${NC}
# 정상 요청 테스트
curl -I "http://localhost:8080/health"

# 공격 시뮬레이션 테스트 실행
bash scripts/test-attacks.sh

# 실시간 로그 확인
kubectl logs -f deployment/waf-gateway -n waf-system

${YELLOW}📚 추가 작업:${NC}
1. /etc/hosts에 다음 라인 추가 (선택사항):
   127.0.0.1 waf-test.local

2. WAF 설정 커스터마이징:
   vim deployments/k8s/owasp-crs-config.yaml

3. 테넌트 생성:
   curl -X POST http://localhost:8080/admin/tenants \\
     -H "Content-Type: application/json" \\
     -d '{"name":"test-tenant","domain":"example.com"}'

${GREEN}🚀 이제 차세대 WAF 시스템을 체험해보세요!${NC}

EOF
}

# 정리 함수 (Ctrl+C 시 호출)
cleanup() {
    log_warning "배포가 중단되었습니다. 정리 중..."
    
    # 포트 포워딩 프로세스 종료
    if [ -f /tmp/waf-port-forwards.pid ]; then
        for pid in $(cat /tmp/waf-port-forwards.pid); do
            kill $pid 2>/dev/null || true
        done
        rm -f /tmp/waf-port-forwards.pid
    fi
    
    exit 1
}

# 신호 핸들러 설정
trap cleanup SIGINT SIGTERM

# 메인 실행 함수
main() {
    print_banner
    
    check_prerequisites
    check_kubernetes_cluster
    setup_helm_repositories
    create_namespaces
    
    deploy_waf_system
    deploy_monitoring_stack
    deploy_logging_stack
    deploy_waf_gateway
    
    setup_port_forwarding
    check_deployment_status
    show_access_information
    
    log_success "모든 배포가 완료되었습니다! 🎉"
    log_info "포트 포워딩은 백그라운드에서 계속 실행됩니다"
    log_info "종료하려면 'ps aux | grep kubectl' 로 프로세스를 찾아 종료하세요"
}

# 스크립트 실행
main "$@"