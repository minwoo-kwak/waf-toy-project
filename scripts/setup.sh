#!/bin/bash

# WAF-K8s-Project 환경 구성 스크립트
# ModSecurity + OWASP CRS 기반 WAF 배포

set -e  # 에러 발생 시 스크립트 종료

# 색상 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 필수 명령어 확인
check_prerequisites() {
    log_info "필수 명령어 확인 중..."
    
    commands=("kubectl" "docker")
    for cmd in "${commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "$cmd 명령어를 찾을 수 없습니다. 설치 후 다시 시도하세요."
            exit 1
        fi
    done
    
    log_success "필수 명령어 확인 완료"
}

# Kubernetes 클러스터 연결 확인
check_kubernetes() {
    log_info "Kubernetes 클러스터 연결 확인 중..."
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Kubernetes 클러스터에 연결할 수 없습니다."
        log_warning "Docker Desktop의 Kubernetes 설정을 확인해주세요:"
        log_warning "1. Docker Desktop → Settings → Kubernetes → Enable Kubernetes"
        log_warning "2. Apply & Restart 클릭"
        exit 1
    fi
    
    log_success "Kubernetes 클러스터 연결 확인 완료"
    kubectl cluster-info
}

# 네임스페이스 생성
create_namespace() {
    log_info "네임스페이스 생성 중..."
    
    if kubectl get namespace waf-system &> /dev/null; then
        log_warning "waf-system 네임스페이스가 이미 존재합니다."
    else
        kubectl apply -f deployments/k8s/namespace.yaml
        log_success "waf-system 네임스페이스 생성 완료"
    fi
}

# ConfigMap 적용
apply_configmaps() {
    log_info "WAF 설정 파일 적용 중..."
    
    # OWASP CRS 설정
    kubectl apply -f deployments/k8s/owasp-crs-config.yaml
    log_success "OWASP CRS 설정 적용 완료"
    
    # ModSecurity 설정이 nginx-ingress-modsecurity.yaml에 포함되어 있으므로 함께 적용됨
}

# NGINX Ingress Controller with ModSecurity 배포
deploy_waf() {
    log_info "NGINX Ingress Controller with ModSecurity 배포 중..."
    
    kubectl apply -f deployments/k8s/nginx-ingress-modsecurity.yaml
    log_success "WAF 배포 완료"
    
    log_info "WAF 파드 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=nginx-ingress -n waf-system --timeout=300s
    log_success "WAF 파드 준비 완료"
}

# 테스트 애플리케이션 배포
deploy_test_app() {
    log_info "테스트 웹 애플리케이션 배포 중..."
    
    kubectl apply -f deployments/k8s/test-app.yaml
    log_success "테스트 애플리케이션 배포 완료"
    
    log_info "테스트 애플리케이션 파드 시작 대기 중..."
    kubectl wait --for=condition=ready pod -l app=test-webapp -n waf-system --timeout=300s
    log_success "테스트 애플리케이션 준비 완료"
}

# 서비스 상태 확인
check_services() {
    log_info "서비스 상태 확인 중..."
    
    echo -e "\n${YELLOW}=== 배포된 파드 목록 ===${NC}"
    kubectl get pods -n waf-system
    
    echo -e "\n${YELLOW}=== 서비스 목록 ===${NC}"
    kubectl get svc -n waf-system
    
    echo -e "\n${YELLOW}=== Ingress 목록 ===${NC}"
    kubectl get ingress -n waf-system
}

# 접속 정보 표시
show_access_info() {
    log_info "접속 정보 확인 중..."
    
    # LoadBalancer 타입의 서비스 IP 확인
    EXTERNAL_IP=$(kubectl get svc nginx-ingress -n waf-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [ -z "$EXTERNAL_IP" ] || [ "$EXTERNAL_IP" = "null" ]; then
        # Docker Desktop의 경우 localhost 사용
        EXTERNAL_IP="localhost"
        PORT=$(kubectl get svc nginx-ingress -n waf-system -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "80")
        ACCESS_URL="http://$EXTERNAL_IP:$PORT"
    else
        ACCESS_URL="http://$EXTERNAL_IP"
    fi
    
    echo -e "\n${GREEN}=== 🎉 WAF 시스템 배포 완료! ===${NC}"
    echo -e "${YELLOW}접속 URL:${NC} $ACCESS_URL"
    echo -e "${YELLOW}테스트 페이지:${NC} $ACCESS_URL/"
    echo -e "${YELLOW}헬스체크:${NC} $ACCESS_URL/health"
    echo -e "${YELLOW}WAF 상태:${NC} $ACCESS_URL/waf/status"
    
    echo -e "\n${BLUE}=== 📋 다음 단계 ===${NC}"
    echo "1. 웹 브라우저에서 $ACCESS_URL 접속"
    echo "2. 공격 테스트 실행: bash scripts/test-attacks.sh"
    echo "3. WAF 로그 확인: kubectl logs -f deployment/nginx-ingress-modsecurity -n waf-system"
    echo "4. 대시보드 접속 (향후 구현 예정)"
    
    echo -e "\n${YELLOW}=== 🛠️ 유용한 명령어 ===${NC}"
    echo "• 파드 상태 확인: kubectl get pods -n waf-system"
    echo "• 로그 실시간 확인: kubectl logs -f deployment/nginx-ingress-modsecurity -n waf-system"
    echo "• 설정 재로드: kubectl rollout restart deployment/nginx-ingress-modsecurity -n waf-system"
    echo "• 전체 삭제: kubectl delete namespace waf-system"
}

# 메인 실행 함수
main() {
    echo -e "${GREEN}"
    echo "🛡️ ==============================================="
    echo "   WAF-K8s-Project 자동 배포 스크립트"
    echo "   ModSecurity + OWASP CRS + Kubernetes"
    echo "===============================================${NC}"
    echo
    
    check_prerequisites
    check_kubernetes
    create_namespace
    apply_configmaps
    deploy_waf
    deploy_test_app
    check_services
    show_access_info
    
    echo -e "\n${GREEN}🚀 모든 설정이 완료되었습니다!${NC}"
}

# 스크립트 실행
main "$@"