# WAF SaaS Platform - Kali Linux 보안 취약점 분석 가이드

## 🛡️ 개요
이 가이드는 Kali Linux를 사용하여 WAF SaaS 플랫폼의 보안 취약점을 분석하는 방법을 설명합니다.

## 🔧 필요한 도구들

### 1. Kali Linux 기본 도구
```bash
# 시스템 업데이트
sudo apt update && sudo apt upgrade -y

# 필요한 도구 설치 확인
which nmap nikto sqlmap gobuster dirb
```

### 2. 추가 설치가 필요한 도구들
```bash
# OWASP ZAP 설치
sudo apt install zaproxy

# Burp Suite Community (필요시)
sudo apt install burpsuite

# 추가 웹 취약점 도구
sudo apt install wapiti w3af-console
```

## 🎯 테스트 시나리오

### 1. 정보 수집 (Information Gathering)

#### 포트 스캔
```bash
# 기본 포트 스캔
nmap -sV -sC localhost

# 서비스 식별 및 OS 검출
nmap -O -sV --script=default localhost

# 모든 포트 스캔
nmap -p- localhost
```

#### 웹 서비스 분석
```bash
# HTTP 헤더 분석
curl -I http://localhost

# 기술 스택 식별
whatweb http://localhost

# 디렉토리 구조 탐색
dirb http://localhost
gobuster dir -u http://localhost -w /usr/share/wordlists/dirb/common.txt
```

### 2. 웹 애플리케이션 취약점 스캔

#### NIKTO 스캔
```bash
# 기본 취약점 스캔
nikto -h http://localhost

# 상세 스캔 (플러그인 모두 사용)
nikto -h http://localhost -Plugins @@ALL

# 결과를 파일로 저장
nikto -h http://localhost -output nikto_scan_result.html -Format htm
```

#### OWASP ZAP을 이용한 자동 스캔
```bash
# ZAP 헤드리스 모드로 실행
zap.sh -cmd -quickurl http://localhost -quickout zap_report.html
```

#### SQLMap을 이용한 SQL Injection 테스트
```bash
# GET 파라미터 테스트
sqlmap -u "http://localhost/api/v1/waf/logs?limit=50" --cookie="waf_token=YOUR_TOKEN"

# POST 요청 테스트
sqlmap -u "http://localhost/api/v1/rules" --data="name=test&rule_text=test" --cookie="waf_token=YOUR_TOKEN"

# 데이터베이스 정보 추출 시도
sqlmap -u "http://localhost/vulnerable_endpoint" --dbs --cookie="waf_token=YOUR_TOKEN"
```

### 3. ModSecurity WAF 우회 테스트

#### 기본 WAF 우회 페이로드
```bash
# SQL Injection 우회 시도
curl "http://localhost/dashboard?id=1/**/UNION/**/SELECT/**/*/*" -H "User-Agent: Mozilla/5.0"

# XSS 우회 시도
curl "http://localhost/dashboard?search=%3Cimg+src%3Dx+onerror%3Dalert%28%29%3E" -H "User-Agent: Mozilla/5.0"

# Path Traversal 우회 시도
curl "http://localhost/dashboard?file=....//....//etc/passwd" -H "User-Agent: Mozilla/5.0"

# Command Injection 우회 시도
curl "http://localhost/dashboard?cmd=;echo+%22pwned%22" -H "User-Agent: Mozilla/5.0"
```

#### 인코딩 기반 우회 시도
```bash
# URL 인코딩 우회
curl "http://localhost/dashboard?search=%253Cscript%253Ealert%2528%2529%253C%252Fscript%253E"

# 헥스 인코딩 우회
curl "http://localhost/dashboard?cmd=%0x3B%0x69%0x64"

# 더블 인코딩 우회
curl "http://localhost/dashboard?search=%2525%2533%2543script%2525%2533%2545"
```

### 4. 인증/인가 취약점 테스트

#### JWT 토큰 분석
```bash
# JWT 토큰 디코딩 (jwt-cli 필요시 설치)
echo "YOUR_JWT_TOKEN" | base64 -d

# JWT 취약점 테스트
python3 -c "
import jwt
token = 'YOUR_JWT_TOKEN'
try:
    decoded = jwt.decode(token, verify=False)
    print('JWT Payload:', decoded)
except:
    print('Invalid JWT token')
"
```

#### 세션 관리 테스트
```bash
# 세션 고정 공격 테스트
curl -c cookies.txt http://localhost/login
curl -b cookies.txt -d "username=admin&password=admin" http://localhost/api/v1/public/auth/callback

# CSRF 토큰 없이 요청 시도
curl -X POST http://localhost/api/v1/rules -H "Content-Type: application/json" -d '{"name":"malicious","rule_text":"malicious"}'
```

### 5. 비즈니스 로직 취약점

#### 권한 상승 테스트
```bash
# 다른 사용자의 리소스 접근 시도
curl -H "Authorization: Bearer USER1_TOKEN" http://localhost/api/v1/rules/USER2_RULE_ID

# 관리자 기능 접근 시도
curl -H "Authorization: Bearer REGULAR_USER_TOKEN" http://localhost/api/v1/admin/users
```

#### Rate Limiting 테스트
```bash
# 연속 요청으로 Rate Limiting 테스트
for i in {1..100}; do
  curl -w "%{http_code}\n" -o /dev/null -s http://localhost/api/v1/waf/stats &
done
wait
```

## 📊 자동화된 보안 스캔 스크립트

### 종합 보안 스캔 스크립트
```bash
#!/bin/bash
# waf_security_scan.sh

TARGET="http://localhost"
OUTPUT_DIR="security_scan_results"

echo "🔍 WAF SaaS Platform Security Assessment Started"
echo "Target: $TARGET"
echo "Output Directory: $OUTPUT_DIR"

# 결과 디렉토리 생성
mkdir -p $OUTPUT_DIR

# 1. 포트 스캔
echo "[1/6] Port Scanning..."
nmap -sV -sC $TARGET > $OUTPUT_DIR/nmap_scan.txt

# 2. 웹 서비스 스캔
echo "[2/6] Web Service Scanning..."
nikto -h $TARGET -output $OUTPUT_DIR/nikto_scan.html -Format htm

# 3. 디렉토리 브루트포싱
echo "[3/6] Directory Brute Force..."
gobuster dir -u $TARGET -w /usr/share/wordlists/dirb/common.txt -o $OUTPUT_DIR/gobuster_scan.txt

# 4. WAF 우회 테스트
echo "[4/6] WAF Bypass Testing..."
python3 << EOF > $OUTPUT_DIR/waf_bypass_results.txt
import requests
import urllib.parse

payloads = [
    "' OR '1'='1",
    "<script>alert('XSS')</script>",
    "../../../../etc/passwd",
    "; cat /etc/passwd",
    "' UNION SELECT * FROM users--"
]

for payload in payloads:
    encoded_payload = urllib.parse.quote(payload)
    try:
        response = requests.get(f"$TARGET/dashboard?test={encoded_payload}")
        print(f"Payload: {payload}")
        print(f"Status: {response.status_code}")
        print(f"Response Length: {len(response.text)}")
        print("---")
    except Exception as e:
        print(f"Error with payload {payload}: {e}")
EOF

# 5. SSL/TLS 테스트
echo "[5/6] SSL/TLS Testing..."
testssl.sh $TARGET > $OUTPUT_DIR/ssl_scan.txt 2>/dev/null || echo "testssl.sh not available"

# 6. 보고서 생성
echo "[6/6] Generating Report..."
cat << EOF > $OUTPUT_DIR/summary_report.md
# WAF SaaS Platform Security Assessment Report

## 날짜
$(date)

## 대상
$TARGET

## 테스트 수행 항목
1. ✅ 포트 스캔 (nmap_scan.txt)
2. ✅ 웹 취약점 스캔 (nikto_scan.html) 
3. ✅ 디렉토리 탐색 (gobuster_scan.txt)
4. ✅ WAF 우회 테스트 (waf_bypass_results.txt)
5. ✅ SSL/TLS 테스트 (ssl_scan.txt)

## 주요 발견사항
- [수동으로 추가 필요]

## 권장사항
- [수동으로 추가 필요]

## 상세 결과
각 .txt 및 .html 파일에서 상세 결과 확인 가능
EOF

echo "✅ Security Assessment Completed!"
echo "📁 Results saved in: $OUTPUT_DIR/"
echo "📄 Summary report: $OUTPUT_DIR/summary_report.md"
```

### 스크립트 실행 방법
```bash
# 스크립트에 실행 권한 부여
chmod +x waf_security_scan.sh

# 스캔 실행
./waf_security_scan.sh
```

## 🎯 특별 테스트 케이스

### ModSecurity 특화 테스트
```bash
# OWASP CRS 룰셋 우회 시도
curl "http://localhost/dashboard?test=<script>/**/alert()/**/<\/script>"
curl "http://localhost/dashboard?test=java%00script:alert()"
curl "http://localhost/dashboard?test=&#x3C;script&#x3E;alert()&#x3C;/script&#x3E;"
```

### API 보안 테스트
```bash
# API 엔드포인트 열거
curl -X OPTIONS http://localhost/api/v1/ -v
curl -X TRACE http://localhost/api/v1/ -v

# 부적절한 HTTP 메소드 테스트
curl -X DELETE http://localhost/api/v1/rules -H "Authorization: Bearer TOKEN"
curl -X PUT http://localhost/api/v1/users -H "Authorization: Bearer TOKEN"
```

## 📝 보고서 작성 템플릿

분석 완료 후 다음 형식으로 보고서를 작성하세요:

```markdown
# WAF SaaS Platform 보안 취약점 분석 보고서

## Executive Summary
- 전체적인 보안 상태 평가
- 주요 발견사항 요약
- 위험도 평가

## 테스트 환경
- Kali Linux 버전
- 사용된 도구들
- 테스트 수행 일시

## 발견된 취약점
### 1. High Risk
- [취약점 명]
- 설명: [상세 설명]
- 영향도: [영향 분석]
- 재현 방법: [PoC]

### 2. Medium Risk
- [취약점 명]

### 3. Low Risk
- [취약점 명]

## ModSecurity 효과성 분석
- 차단된 공격: X건
- 우회된 공격: Y건
- 전체 차단율: Z%

## 권장사항
1. 즉시 조치 필요
2. 단기 개선사항
3. 장기 보안 전략

## 부록
- 상세 스캔 결과
- 사용된 페이로드 목록
- 참고 자료
```

이 가이드를 따라 Kali Linux에서 종합적인 보안 분석을 수행하실 수 있습니다! 🛡️