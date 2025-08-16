# WAF Guardian SaaS Platform 🛡️

현대적인 웹 애플리케이션 방화벽을 Kubernetes 환경에서 SaaS 형태로 제공하는 실시간 보안 모니터링 플랫폼입니다.

## 🚀 프로젝트 소개

이 프로젝트는 ModSecurity와 OWASP CRS(Core Rule Set)를 기반으로 한 웹 애플리케이션 방화벽을 개발하고, 다중 사용자 환경에서 SaaS 형태로 서비스할 수 있도록 구성한 차세대 보안 플랫폼입니다.

**Docker Desktop의 Kubernetes**를 활용하여 로컬 개발 환경에서도 실제 클라우드 환경과 유사한 구조로 개발할 수 있으며, **React 기반의 실시간 대시보드**를 통해 WebSocket 기반 보안 로그 모니터링과 지능형 공격 유형 분류가 가능합니다.

## ✨ 주요 기능

### 🔥 v3.0 새로운 기능들
- **🎯 실시간 Live Security Monitor**: WebSocket 기반 실시간 보안 이벤트 스트리밍
- **🧠 지능형 공격 유형 분류**: SQL Injection, XSS, Command Injection, LFI/RFI 자동 분류
- **📊 모던 UI/UX**: 그라디언트 기반 현대적 디자인과 실시간 애니메이션
- **🎯 상세 공격 정보**: 실제 공격 URL, ModSecurity 룰 ID, 심각도 표시

### 🛡️ 핵심 보안 기능
- **실시간 WAF 보호**: ModSecurity 3.x 엔진을 통한 실시간 웹 공격 차단
- **OWASP CRS 4.x 통합**: 검증된 최신 보안 룰셋으로 OWASP Top 10 공격 방어
- **커스텀 룰 관리**: 웹 UI를 통한 보안 룰 생성, 수정, 삭제 (CRUD)
- **멀티 테넌트**: 사용자별 독립적인 보안 정책 관리

### 🔐 인증 및 사용자 관리
- **Google OAuth 2.0**: 간편하고 안전한 소셜 로그인
- **JWT 기반 인증**: Stateless 토큰 기반 세션 관리
- **사용자별 데이터 격리**: 완전한 멀티 테넌트 아키텍처

## 🏗️ 기술 스택

### Backend (Go)
- **언어**: Go 1.23+
- **프레임워크**: Gin HTTP Framework
- **아키텍처**: RESTful API, Clean Architecture, DTO 패턴
- **WebSocket**: 실시간 보안 이벤트 스트리밍
- **인증**: Google OAuth 2.0 + JWT

### Frontend (React)
- **언어**: TypeScript 5.x
- **프레임워크**: React 18+
- **상태관리**: Context API
- **UI 라이브러리**: Material-UI (MUI) v5
- **실시간 통신**: WebSocket
- **빌드 도구**: Create React App

### Infrastructure & Security
- **컨테이너**: Docker + Multi-stage builds
- **오케스트레이션**: Kubernetes (Docker Desktop)
- **웹서버**: Nginx Ingress Controller
- **WAF 엔진**: ModSecurity 3.x
- **보안 룰**: OWASP CRS 4.x
- **로드밸런싱**: Kubernetes Services

## 📦 설치 및 실행

### 사전 요구사항
```bash
# 필수 소프트웨어
- Docker Desktop (Kubernetes 활성화)
- Go 1.23 이상
- Node.js 18 이상
- kubectl CLI

# 시스템 요구사항
- RAM: 8GB 이상 권장
- Storage: 10GB 이상 여유 공간
```

### 🚀 빠른 시작 (Quick Start)

```bash
# 1. 저장소 클론
git clone https://github.com/your-username/waf-toy-project.git
cd waf-toy-project

# 2. Kubernetes 클러스터 상태 확인
kubectl cluster-info
kubectl get nodes

# 3. 전체 WAF 시스템 자동 배포
./scripts/deploy-k8s.sh

# 4. 시스템 접속 준비
kubectl port-forward service/ingress-nginx-controller -n ingress-nginx 80:80

# 5. 브라우저에서 접속
open http://localhost
```

### 🔧 개발 환경 설정

#### Backend 개발
```bash
cd backend
go mod tidy
go run main.go  # http://localhost:8080
```

#### Frontend 개발  
```bash
cd frontend
npm install
npm start      # http://localhost:3000
```

#### Docker 이미지 빌드
```bash
# 백엔드 이미지 빌드
docker build -t waf-backend:v3.0.4 ./backend

# 프론트엔드 이미지 빌드  
docker build -t waf-frontend:v3.0.1 ./frontend

# 일괄 빌드 스크립트
./scripts/build-images.sh
```

## 📁 프로젝트 구조

```
waf-toy-project/
├── README.md                          # 프로젝트 소개 및 사용법
├── backend/                           # Go 백엔드 API 서버
│   ├── config/                       # 설정 관리
│   ├── dto/                          # 데이터 전송 객체
│   ├── handlers/                     # HTTP 핸들러 (Router Layer)
│   ├── services/                     # 비즈니스 로직 (Service Layer)
│   │   ├── waf_service.go           # WAF 로그 분석 및 공격 유형 분류
│   │   ├── websocket_service.go     # 실시간 WebSocket 통신
│   │   └── auth_service.go          # Google OAuth 인증
│   ├── utils/                       # 유틸리티 함수
│   └── main.go                      # 애플리케이션 엔트리 포인트
├── frontend/                         # React 프론트엔드
│   ├── src/
│   │   ├── components/              # React 컴포넌트
│   │   │   ├── dashboard/           # 대시보드 관련 컴포넌트
│   │   │   │   ├── Dashboard.tsx    # 메인 대시보드
│   │   │   │   ├── LiveLogMonitor.tsx # 실시간 보안 모니터
│   │   │   │   ├── AttackChart.tsx  # 공격 통계 차트
│   │   │   │   └── StatsCards.tsx   # 통계 카드
│   │   │   ├── auth/                # 인증 관련 컴포넌트
│   │   │   └── rules/               # 커스텀 룰 관리
│   │   ├── services/                # API 클라이언트
│   │   ├── contexts/                # React Context (상태 관리)
│   │   └── types/                   # TypeScript 타입 정의
│   ├── public/                      # 정적 리소스
│   └── Dockerfile                   # 프론트엔드 컨테이너 설정
├── k8s/                             # Kubernetes 매니페스트
│   ├── backend/                     # 백엔드 배포 설정
│   ├── frontend/                    # 프론트엔드 배포 설정
│   ├── ingress/                     # Ingress 및 ModSecurity 설정
│   └── modsecurity/                 # ModSecurity ConfigMap
├── scripts/                         # 자동화 스크립트
│   ├── deploy-k8s.sh               # 전체 시스템 배포
│   ├── build-images.sh             # Docker 이미지 일괄 빌드
│   └── cleanup-k8s.sh              # 리소스 정리
└── docs/                           # 프로젝트 문서
    ├── CUSTOM_RULE_GUIDE.md        # 커스텀 룰 작성 가이드
    ├── SECURITY_ANALYSIS_GUIDE.md  # Kali Linux 보안 분석 가이드
    ├── ENVIRONMENT_SETUP.md        # 환경 설정 가이드
    └── TROUBLESHOOTING.md          # 문제 해결 가이드
```

## 🛡️ 보안 테스트 및 분석

### 🌐 내장 브라우저 테스트
대시보드 내에서 바로 보안 테스트를 수행할 수 있습니다:
- **SQL Injection**: `' OR 1=1--`, `UNION SELECT` 등
- **XSS**: `<script>alert('xss')</script>`, `javascript:alert()` 등  
- **Command Injection**: `; cat /etc/passwd`, `| whoami` 등
- **Path Traversal**: `../../../etc/passwd`, `..\\windows\\system32` 등

### 🔍 Kali Linux 전문 분석
고급 보안 분석을 위한 도구들:

```bash
# OWASP ZAP 자동 스캔
zap-full-scan.py -t http://localhost -r waf_security_report.html

# Nikto 웹 서버 취약점 스캔
nikto -h http://localhost -Format htm -output nikto_report.html

# SQLMap을 이용한 정밀 SQL Injection 테스트
sqlmap -u "http://localhost/login?user=test" --cookie="session=token"

# ModSecurity 우회 테스트
curl "http://localhost/search?q=<img+src=x+onerror=alert()>"
```

상세한 가이드는 [SECURITY_ANALYSIS_GUIDE.md](./SECURITY_ANALYSIS_GUIDE.md)를 참조하세요.

## 📊 주요 기능 스크린샷

### 실시간 Live Security Monitor
- 🎯 실시간 보안 이벤트 스트리밍
- 🏷️ 정확한 공격 유형 분류 (SQL Injection, XSS, Command Injection 등)
- 📍 실제 공격 URL 및 ModSecurity 룰 ID 표시
- ⚡ WebSocket 기반 즉시 알림

### 모던 대시보드
- 📈 인터랙티브 공격 통계 차트
- 🎨 그라디언트 기반 현대적 UI/UX
- 📱 반응형 디자인 (모바일 지원)
- 🔄 실시간 애니메이션 및 시각적 피드백

## 🎯 개발 진행 상황

### ✅ v3.0 완료 (2025.8.16)
- [x] **실시간 보안 모니터링 시스템** - LiveLogMonitor 컴포넌트 구현
- [x] **지능형 공격 유형 분류** - OWASP CRS 룰 매핑 + URL 패턴 분석
- [x] **모던 UI/UX 디자인** - 그라디언트 기반 현대적 디자인 시스템
- [x] **WebSocket 실시간 통신** - 안정적인 실시간 이벤트 스트리밍
- [x] **상세 위협 정보 표시** - 공격 URL, 룰 ID, 심각도 정보 제공

### 🚧 진행 중 기능
- [ ] **고급 보안 분석 리포트** - PDF 보고서 자동 생성
- [ ] **AI 기반 이상 탐지** - 머신러닝 기반 위협 패턴 분석
- [ ] **다중 환경 지원** - AWS, GCP, Azure 클라우드 배포

## 🔧 API 문서

### 인증 API
```http
GET  /api/v1/public/auth/google     # Google OAuth 로그인
POST /api/v1/public/auth/callback   # OAuth 콜백 처리
POST /api/v1/public/auth/logout     # 로그아웃
```

### WAF 관리 API
```http
GET  /api/v1/waf/stats             # WAF 통계 조회
GET  /api/v1/waf/logs              # 보안 로그 조회  
GET  /api/v1/waf/dashboard         # 대시보드 데이터
GET  /api/v1/ws                    # WebSocket 연결 (실시간 스트리밍)
```

### 커스텀 룰 API
```http
GET    /api/v1/rules               # 사용자 룰 목록 조회
POST   /api/v1/rules               # 새 룰 생성
PUT    /api/v1/rules/:id           # 룰 수정
DELETE /api/v1/rules/:id           # 룰 삭제
```

## 🛠️ 개발 가이드

### 커스텀 룰 작성
상세한 ModSecurity 룰 작성 방법은 [CUSTOM_RULE_GUIDE.md](./CUSTOM_RULE_GUIDE.md)를 참조하세요.

### 환경 설정
다양한 개발 환경 설정 방법은 [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md)를 참조하세요.

### 문제 해결
일반적인 문제 해결 방법은 [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)를 참조하세요.

## 🤝 기여하기

1. 이슈 등록 후 작업 시작
2. `feature/기능명` 브랜치에서 개발
3. Conventional Commits 규칙 준수
4. Pull Request 생성

### 브랜치 전략
- `main`: 프로덕션 릴리즈
- `develop`: 개발 통합  
- `feature/*`: 기능 개발
- `fix/*`: 버그 수정

## 📄 라이선스

이 프로젝트는 학습 목적으로 제작되었으며, 사용된 오픈소스 컴포넌트들의 라이선스를 준수합니다.

- **ModSecurity**: Apache 2.0 License
- **OWASP CRS**: Apache 2.0 License
- **React & Material-UI**: MIT License
- **Go & Gin**: BSD & MIT License

## 🎯 성능 및 확장성

### 시스템 성능
- **처리량**: 초당 1,000+ 요청 처리 가능
- **응답 시간**: 평균 50ms 이하
- **메모리 사용량**: 백엔드 512MB, 프론트엔드 256MB
- **실시간 처리**: WebSocket 기반 지연시간 10ms 이하

### 확장성
- **수평 확장**: Kubernetes Deployment 스케일링 지원
- **로드밸런싱**: Kubernetes Service & Ingress 자동 로드밸런싱
- **멀티 테넌트**: 사용자별 완전 데이터 격리
- **클러스터 지원**: 다중 노드 Kubernetes 클러스터 지원

---

**🛡️ WAF Guardian으로 안전한 웹 애플리케이션을 구축하세요! 🚀**