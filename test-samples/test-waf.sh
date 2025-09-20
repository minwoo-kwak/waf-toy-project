#!/bin/bash

# WAF 공격 테스트 스크립트
# 사용법: ./test-waf.sh [BASE_URL]

# 기본 URL (프록시를 통해 보호되는 DVWA)
BASE_URL=${1:-"http://dvwatest-n-e-kr-protected.waftest.p-e.kr"}

echo "🛡️ WAF 공격 테스트 시작"
echo "대상 URL: $BASE_URL"
echo "=========================================="

# 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 테스트 결과 체크 함수
check_blocked() {
    local response_code=$1
    local test_name="$2"

    if [[ $response_code -eq 403 || $response_code -eq 406 || $response_code -eq 400 ]]; then
        echo -e "${GREEN}✅ BLOCKED${NC} - $test_name (응답코드: $response_code)"
        return 0
    elif [[ $response_code -eq 200 ]]; then
        echo -e "${RED}❌ ALLOWED${NC} - $test_name (응답코드: $response_code)"
        return 1
    else
        echo -e "${YELLOW}⚠️  UNKNOWN${NC} - $test_name (응답코드: $response_code)"
        return 2
    fi
}

echo -e "${YELLOW}1. SQL Injection 테스트${NC}"
echo "--------------------"

# SQL Injection 테스트
sql_payloads=(
    "' OR '1'='1"
    "' OR 1=1 --"
    "admin'--"
    "' UNION SELECT null, user(), version() --"
    "1' ORDER BY 3--+"
)

for payload in "${sql_payloads[@]}"; do
    encoded_payload=$(printf '%s' "$payload" | jq -sRr @uri 2>/dev/null || python3 -c "import urllib.parse; print(urllib.parse.quote('$payload'))" 2>/dev/null || echo "$payload")
    url="$BASE_URL/vulnerabilities/sqli/?id=$encoded_payload&Submit=Submit"

    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
    check_blocked $response_code "SQL: $payload"
done

echo ""
echo -e "${YELLOW}2. XSS (Cross-Site Scripting) 테스트${NC}"
echo "--------------------"

# XSS 테스트
xss_payloads=(
    "<script>alert(1)</script>"
    "<img src=x onerror=alert('XSS')>"
    "<svg onload=alert(1)>"
    "javascript:alert('XSS')"
    "<iframe src=\"javascript:alert('XSS')\"></iframe>"
)

for payload in "${xss_payloads[@]}"; do
    encoded_payload=$(printf '%s' "$payload" | jq -sRr @uri 2>/dev/null || python3 -c "import urllib.parse; print(urllib.parse.quote('$payload'))" 2>/dev/null || echo "$payload")
    url="$BASE_URL/vulnerabilities/xss_r/?name=$encoded_payload"

    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
    check_blocked $response_code "XSS: $payload"
done

echo ""
echo -e "${YELLOW}3. Command Injection 테스트${NC}"
echo "--------------------"

# Command Injection 테스트
cmd_payloads=(
    "127.0.0.1; ls -la"
    "127.0.0.1 && cat /etc/passwd"
    "127.0.0.1 | whoami"
    "127.0.0.1\`id\`"
    "127.0.0.1\$(uname -a)"
)

for payload in "${cmd_payloads[@]}"; do
    encoded_payload=$(printf '%s' "$payload" | jq -sRr @uri 2>/dev/null || python3 -c "import urllib.parse; print(urllib.parse.quote('$payload'))" 2>/dev/null || echo "$payload")
    url="$BASE_URL/vulnerabilities/exec/?ip=$encoded_payload&Submit=Submit"

    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
    check_blocked $response_code "CMD: $payload"
done

echo ""
echo -e "${YELLOW}4. Local File Inclusion (LFI) 테스트${NC}"
echo "--------------------"

# LFI 테스트
lfi_payloads=(
    "../../../etc/passwd"
    "..\\..\\..\\..\\windows\\system32\\drivers\\etc\\hosts"
    "....//....//....//etc/passwd"
    "/etc/passwd%00"
    "php://filter/read=convert.base64-encode/resource=/etc/passwd"
)

for payload in "${lfi_payloads[@]}"; do
    encoded_payload=$(printf '%s' "$payload" | jq -sRr @uri 2>/dev/null || python3 -c "import urllib.parse; print(urllib.parse.quote('$payload'))" 2>/dev/null || echo "$payload")
    url="$BASE_URL/vulnerabilities/fi/?page=$encoded_payload"

    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
    check_blocked $response_code "LFI: $payload"
done

echo ""
echo -e "${YELLOW}5. 추가 웹 공격 패턴 테스트${NC}"
echo "--------------------"

# 추가 공격 패턴
additional_payloads=(
    "<?php system('whoami'); ?>"
    "../../../proc/self/environ"
    "data:text/html,<script>alert('XSS')</script>"
    "file:///etc/passwd"
    "http://evil.com/malware.exe"
)

for payload in "${additional_payloads[@]}"; do
    encoded_payload=$(printf '%s' "$payload" | jq -sRr @uri 2>/dev/null || python3 -c "import urllib.parse; print(urllib.parse.quote('$payload'))" 2>/dev/null || echo "$payload")
    url="$BASE_URL/?test=$encoded_payload"

    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
    check_blocked $response_code "Additional: $payload"
done

echo ""
echo "=========================================="
echo -e "${YELLOW}📊 테스트 완료${NC}"
echo ""
echo "WAF 로그 확인 명령어:"
echo "kubectl logs -l app=nginx-ingress -n ingress-nginx | grep -i \"denied\\|blocked\\|403\""
echo ""
echo "💡 팁:"
echo "- 403/406/400 응답 = WAF 차단 (정상)"
echo "- 200 응답 = WAF 통과 (문제 가능성)"
echo "- 실시간 로그는 WAF 대시보드에서 확인하세요"

echo ""
echo -e "${GREEN}✅ 모든 테스트가 완료되었습니다!${NC}"