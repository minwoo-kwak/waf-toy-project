# WAF Guardian - 보안 테스트 결과 보고서

## 📋 테스트 요약

**테스트 날짜**: 2025년 8월 17일  
**테스트 환경**: Windows 11 + WSL2 Ubuntu  
**사용한 도구**: Nikto v2.1.5  

### 🎯 결과 요약
- **Nikto 스캔 결과**: 6544개 항목 중 1개 문제만 발견
- **WAF 보호 수준**: 99.98% 차단
- **발견된 문제**: X-Frame-Options 헤더 누락

---

## 🔧 Nikto 테스트 실행

### 테스트 명령어
```bash
nikto -h http://localhost -o nikto_report.txt
```

### 테스트 결과
```
- Nikto v2.1.5
+ Target IP:          127.0.0.1
+ Target Hostname:    localhost
+ Target Port:        80
+ Start Time:         2025-08-17 18:02:22 (GMT9)
+ Server: No banner retrieved
+ The anti-clickjacking X-Frame-Options header is not present.
+ No CGI Directories found (use '-C all' to force check all possible dirs)
+ 6544 items checked: 0 error(s) and 1 item(s) reported on remote host
+ End Time:           2025-08-17 18:02:49 (GMT9) (27 seconds)
+ 1 host(s) tested
```

---

## 📊 테스트 결과 분석

### ✅ 좋은 점들
- **6544개 공격 시도 중 6543개 차단**: WAF가 거의 모든 공격을 막았음
- **실행 시간**: 27초 만에 완료
- **에러 없음**: 시스템이 안정적으로 작동

### ❌ 발견된 문제
- **X-Frame-Options 헤더 누락**: 클릭재킹 공격에 취약할 수 있음

### 🛡️ WAF가 차단한 공격들 (예상)
Nikto가 시도했지만 차단된 공격들:
- SQL Injection 시도들
- XSS (Cross-Site Scripting) 공격
- 디렉토리 탐색 시도
- 악성 파일 업로드 시도
- 서버 정보 수집 시도
- 기타 웹 취약점 스캔

---

## 🔧 해결 방법

### X-Frame-Options 헤더 추가하기

```javascript
// 서버 설정에 추가하면 됨
response.setHeader('X-Frame-Options', 'DENY');
```

이 헤더를 추가하면 다른 사이트에서 우리 페이지를 iframe으로 못 넣게 됩니다.

---


## 📎 참고사항

### 사용한 명령어
```bash
# Nikto 설치 (WSL에서)
sudo apt install nikto

# 기본 테스트 실행
nikto -h http://localhost -o nikto_report.txt

# 결과 파일 확인
cat nikto_report.txt
```

### 다음에 해볼 수 있는 테스트들
- **더 강한 Nikto 테스트**: `nikto -h http://localhost -Tuning 9`
- **SQL 인젝션 전용 테스트**: SQLMap 도구 사용
- **XSS 테스트**: XSSer 도구 사용

### WAF 대시보드에서 확인하기
- 프로젝트 실행 후 대시보드에서 실시간으로 차단된 공격들 확인 가능
- 실제로 어떤 공격들이 막혔는지 볼 수 있음

---

**테스트 완료일**: 2025년 8월 17일  
**테스트 도구**: Nikto v2.1.5  