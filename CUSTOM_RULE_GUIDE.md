# Custom Rule 추가 가이드 📝

## 🎯 Custom Rule 추가 방법

### 1. Custom Rules 페이지 접속
1. 로그인 후 좌측 메뉴에서 **"Custom Rules"** 클릭
2. 페이지 상단의 **"+ Create New Rule"** 버튼 클릭

### 2. 필수 입력 필드

#### **Rule Name** (규칙 이름)
```
예시: Block SQL Injection
```

#### **Description** (설명)
```
예시: SQL 인젝션 공격을 차단하는 커스텀 규칙
```

#### **Rule Text** (ModSecurity 규칙 문법)
```
예시: SecRule ARGS "@detectSQLi" "id:1001,phase:2,block,msg:'SQL Injection Attack Detected',logdata:'Matched Data: %{MATCHED_VAR} found within %{MATCHED_VAR_NAME}'"
```

#### **Severity** (심각도)
- **LOW**: 낮은 위험도
- **MEDIUM**: 보통 위험도  
- **HIGH**: 높은 위험도
- **CRITICAL**: 매우 높은 위험도

#### **Enabled** (활성화 여부)
- ✅ 체크: 즉시 적용
- ❌ 해제: 저장만 하고 비활성화

## 🛡️ 실용적인 Custom Rule 예제들

### 1. SQL Injection 차단 규칙
```
Name: Advanced SQL Injection Protection
Description: 고급 SQL 인젝션 공격 패턴 차단
Rule Text: SecRule ARGS "@detectSQLi" "id:2001,phase:2,block,msg:'Advanced SQL Injection Attack',logdata:'%{MATCHED_VAR}'"
Severity: HIGH
```

### 2. XSS 공격 차단 규칙
```
Name: XSS Attack Prevention
Description: Cross-Site Scripting 공격 차단
Rule Text: SecRule ARGS "@detectXSS" "id:2002,phase:2,block,msg:'XSS Attack Detected',logdata:'%{MATCHED_VAR}'"
Severity: HIGH
```

### 3. 특정 IP 차단 규칙
```
Name: Block Malicious IP
Description: 악성 IP 주소 차단
Rule Text: SecRule REMOTE_ADDR "@ipMatch 192.168.1.100" "id:2003,phase:1,deny,msg:'Blocked IP Address'"
Severity: CRITICAL
```

### 4. 파일 업로드 제한 규칙
```
Name: File Upload Restriction
Description: 위험한 파일 확장자 업로드 차단
Rule Text: SecRule FILES_NAMES "@rx (?i)\.(php|jsp|asp|exe|bat)$" "id:2004,phase:2,block,msg:'Dangerous file upload blocked'"
Severity: MEDIUM
```

### 5. User-Agent 필터링 규칙
```
Name: Block Bot Traffic
Description: 악성 봇 트래픽 차단
Rule Text: SecRule REQUEST_HEADERS:User-Agent "@rx (?i)(sqlmap|nmap|nikto|havij)" "id:2005,phase:1,deny,msg:'Malicious bot detected'"
Severity: MEDIUM
```

### 6. Rate Limiting 규칙
```
Name: Rate Limiting Protection
Description: 과도한 요청 빈도 제한
Rule Text: SecRule IP:bf_counter "@gt 10" "id:2006,phase:1,deny,msg:'Rate limit exceeded',expirevar:IP.bf_counter=60"
Severity: LOW
```

## 🔧 ModSecurity 규칙 문법 참고

### 기본 구조
```
SecRule [VARIABLES] [OPERATOR] [ACTIONS]
```

### 주요 Variables
- `ARGS`: 모든 GET/POST 파라미터
- `REQUEST_URI`: 요청 URI
- `REQUEST_HEADERS`: HTTP 헤더
- `REQUEST_BODY`: POST 요청 본문
- `REMOTE_ADDR`: 클라이언트 IP
- `FILES_NAMES`: 업로드 파일명

### 주요 Operators
- `@detectSQLi`: SQL 인젝션 탐지
- `@detectXSS`: XSS 공격 탐지
- `@rx`: 정규표현식 매칭
- `@contains`: 문자열 포함 여부
- `@ipMatch`: IP 주소 매칭
- `@gt`: 숫자 비교 (greater than)

### 주요 Actions
- `id:XXXX`: 규칙 고유 ID (필수)
- `phase:X`: 실행 단계 (1-5)
- `block`: 요청 차단
- `deny`: 요청 거부
- `pass`: 통과 (로깅만)
- `msg:'message'`: 로그 메시지
- `logdata:'data'`: 추가 로그 데이터

## 🚨 주의사항

### 1. 규칙 ID 관리
- **1000-1999**: 시스템 예약
- **2000-2999**: 사용자 커스텀 규칙
- **3000+**: 고급 사용자 규칙

### 2. Phase 단계
- **Phase 1**: 요청 헤더 검사
- **Phase 2**: 요청 본문 검사  
- **Phase 3**: 응답 헤더 검사
- **Phase 4**: 응답 본문 검사
- **Phase 5**: 로깅 단계

### 3. 테스트 방법
1. 규칙 추가 후 **Enabled = false**로 설정
2. 로그에서 매칭 여부 확인
3. 문제없으면 **Enabled = true**로 활성화

## 🧪 규칙 테스트 방법

### 브라우저에서 직접 테스트
```bash
# SQL 인젝션 테스트
http://localhost/dashboard?test=' OR '1'='1

# XSS 테스트  
http://localhost/dashboard?search=<script>alert('test')</script>

# 결과 확인
# - 403 Forbidden = 규칙 적용됨 ✅
# - 200 OK = 규칙 적용 안됨 ❌
```

### curl 명령어로 테스트
```bash
# 규칙 테스트
curl "http://localhost/dashboard?test=your_test_payload"

# 응답 코드 확인
# - 403 = 차단됨
# - 200 = 통과됨
```

## 🎯 실전 시나리오별 규칙

### 1. 로그인 보호
```
SecRule REQUEST_URI "@contains /login" "chain,id:2101,phase:2,block,msg:'Login brute force protection'"
SecRule &ARGS_POST:password "@gt 3" "t:none"
```

### 2. 관리자 페이지 보호
```
SecRule REQUEST_URI "@beginsWith /admin" "id:2102,phase:1,block,msg:'Admin area access denied'"
```

### 3. API 엔드포인트 보호  
```
SecRule REQUEST_URI "@beginsWith /api" "chain,id:2103,phase:2,block,msg:'API abuse detected'"
SecRule REQUEST_HEADERS:Content-Type "!@contains application/json"
```

이제 Custom Rule을 쉽게 추가하실 수 있습니다! 🚀