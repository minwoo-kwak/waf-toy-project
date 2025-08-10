# WAF SaaS Platform 🛡️

현대적인 웹 애플리케이션 방화벽을 쿠버네티스 환경에서 SaaS 형태로 제공하는 플랫폼입니다.

## 프로젝트 소개

이 프로젝트는 ModSecurity와 OWASP CRS(Core Rule Set)를 기반으로 한 웹 애플리케이션 방화벽을 개발하고, 다중 사용자 환경에서 SaaS 형태로 서비스할 수 있도록 구성한 시스템입니다. 

Docker Desktop의 Kubernetes를 활용하여 로컬 개발 환경에서 실제 클라우드 환경과 유사한 구조로 개발할 수 있으며, React 기반의 대시보드를 통해 실시간 보안 로그 모니터링과 커스텀 룰 관리가 가능합니다.

## 주요 기능

- **실시간 WAF 보호**: ModSecurity 엔진을 통한 실시간 웹 공격 차단
- **OWASP CRS 통합**: 검증된 보안 룰셋으로 OWASP Top 10 공격 방어
- **소셜 로그인**: Google OAuth를 통한 간편 인증
- **대시보드**: 직관적인 웹 인터페이스로 보안 현황 모니터링
- **커스텀 룰 관리**: 웹 UI를 통한 보안 룰 생성, 수정, 삭제
- **멀티 테넌트**: 사용자별 독립적인 보안 정책 관리

## 기술 스택

### Backend
- **언어**: Go 1.21+
- **프레임워크**: Gin/Echo
- **아키텍처**: RESTful API, DTO 패턴

### Frontend  
- **언어**: TypeScript
- **프레임워크**: React 18+
- **상태관리**: Context API
- **UI 라이브러리**: Material-UI / Ant Design

### Infrastructure
- **컨테이너**: Docker
- **오케스트레이션**: Kubernetes (Docker Desktop)
- **웹서버**: Nginx Ingress Controller
- **WAF 엔진**: ModSecurity 3.x
- **보안 룰**: OWASP CRS 4.x

## 개발 진행 상황

### ✅ 1주차 (2025.8.4 - 2025.8.10): 개발환경 구성 및 기본 WAF 구현 **[완료]**
- [x] Git 저장소 초기 설정 및 브랜치 전략 수립
- [x] Docker Desktop Kubernetes 환경 구성
- [x] 프로젝트 구조 설계 및 폴더 생성
- [x] **Go 백엔드 API 서버 구축** (Gin 프레임워크, Docker 컨테이너화)
- [x] **React 프론트엔드 애플리케이션 개발** (TypeScript, Docker 멀티스테이지 빌드)
- [x] **ModSecurity + OWASP CRS 3.3.4 통합** (Kubernetes Ingress Controller)
- [x] **Kubernetes 매니페스트 작성** (Deployment, Service, Ingress, ConfigMap)
- [x] **자동 배포 스크립트** 구현 (`scripts/deploy-k8s.sh`)
- [x] **WAF 보안 테스트** 완료 (SQL Injection, XSS 공격 차단 검증)
- [x] 개발 환경 설정 파일 및 .gitignore 작성
- [x] 상세한 프로젝트 문서화

**🎯 1주차 주요 성과:**
```bash
# 정상 요청 테스트
curl "http://localhost/api/v1/ping" -H "Host: waf-local.dev"
# → {"message":"WAF API is running"} (200 OK)

# SQL Injection 차단 테스트  
curl "http://localhost/api/v1/ping?id=1%27%20OR%20%271%27=%271" -H "Host: waf-local.dev"
# → 403 Forbidden (ModSecurity 차단)

# XSS 공격 차단 테스트
curl "http://localhost/api/v1/ping?search=%3Cscript%3Ealert('xss')%3C/script%3E" -H "Host: waf-local.dev"  
# → 403 Forbidden (ModSecurity 차단)
```

### 🔄 2주차 (2025.8.11 - 2025.8.17): SaaS 기능 구현
- [ ] Google OAuth 로그인 연동
- [ ] WAF 로그 수집 및 시각화 대시보드
- [ ] Custom Rule CRUD 기능 구현
- [ ] 사용자별 보안 정책 관리
- [ ] 실시간 보안 이벤트 모니터링
- [ ] 멀티 테넌트 아키텍처 기반 구축

### 🎯 향후 계획
- **3주차**: 커스텀 룰 최적화 및 성능 튜닝
- **4주차**: 멀티 테넌트 구조 완성
- **5주차**: 보안 테스트 자동화 및 로그 시각화
- **6주차**: 통합 테스트 및 배포

## 설치 및 실행

### 사전 요구사항
- Docker Desktop (Kubernetes 활성화)
- Go 1.21 이상
- Node.js 18 이상
- kubectl CLI

### 로컬 개발 환경 설정

```bash
# 저장소 클론
git clone https://github.com/your-username/waf-toy-project.git
cd waf-toy-project

# Kubernetes 클러스터 상태 확인
kubectl cluster-info

# Docker 이미지 빌드
docker build -t waf-backend:v1.0.1 ./backend
docker build -t waf-frontend:v1.0.1 ./frontend

# 전체 WAF 시스템 배포
./scripts/deploy-k8s.sh

# WAF 기능 테스트
curl "http://localhost/api/v1/ping" -H "Host: waf-local.dev"
curl "http://localhost/api/v1/ping?id=1%27%20OR%20%271%27=%271" -H "Host: waf-local.dev"  # SQL Injection 테스트
```

### 개별 컴포넌트 개발 환경

```bash
# 백엔드 개발 환경 (Go)
cd backend
go mod tidy
go run main.go  # http://localhost:8080

# 프론트엔드 개발 환경 (React)
cd frontend  
npm install
npm start      # http://localhost:3000

# ModSecurity 로그 모니터링
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller -f
```

## 프로젝트 구조

```
waf-toy-project/
├── README.md
├── .gitignore
├── backend/                 # Go 백엔드
│   ├── dto/                # 데이터 전송 객체
│   ├── handlers/           # HTTP 핸들러
│   ├── services/           # 비즈니스 로직
│   ├── config/            # 설정 관리
│   └── main.go            # 엔트리 포인트
├── frontend/               # React 프론트엔드
│   ├── src/
│   │   ├── components/    # React 컴포넌트
│   │   ├── services/      # API 클라이언트
│   │   ├── types/         # TypeScript 타입
│   │   └── utils/         # 유틸리티 함수
│   ├── public/
│   └── package.json
├── k8s/                   # Kubernetes 매니페스트
│   ├── ingress/
│   ├── backend/
│   ├── frontend/
│   └── modsecurity/
├── security-analysis/     # 보안 분석 결과
│   ├── reports/
│   └── test-scenarios/
└── docs/                  # 프로젝트 문서
    ├── architecture.md
    └── api-specification.md
```

## 보안 테스트

본 프로젝트에서는 다양한 보안 테스트 도구를 활용하여 WAF의 효과성을 검증합니다:

- **OWASP ZAP**: 자동화된 웹 애플리케이션 보안 스캔
- **Burp Suite**: 수동 보안 테스트 및 트래픽 분석  
- **Nikto**: 웹서버 취약점 스캔
- **SQLMap**: SQL 인젝션 공격 시뮬레이션

## 참고 자료

### 공식 문서
- [OWASP ModSecurity Core Rule Set](https://github.com/coreruleset/coreruleset) - 공식 CRS 저장소
- [ModSecurity CRS Docker Images](https://github.com/coreruleset/modsecurity-crs-docker) - 컨테이너 이미지
- [Kubernetes Ingress-Nginx ModSecurity](https://kubernetes.github.io/ingress-nginx/user-guide/third-party-addons/modsecurity/) - Kubernetes 연동 가이드
- [OWASP CRS 공식 사이트](https://coreruleset.org/) - 프로젝트 홈페이지

### 기술 문서
- [ModSecurity Reference Manual](https://github.com/SpiderLabs/ModSecurity/wiki/Reference-Manual)
- [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [Go Web Development](https://golang.org/doc/)
- [React Documentation](https://react.dev/)

## 기여 가이드

1. 이슈 등록 후 작업 시작
2. feature 브랜치에서 개발
3. 커밋 메시지는 conventional commits 규칙 준수
4. PR 생성 시 리뷰어 지정

## 브랜치 전략

- `main`: 프로덕션 릴리즈
- `develop`: 개발 통합
- `feature/*`: 기능 개발
- `fix/*`: 버그 수정

## 라이선스

이 프로젝트는 학습 목적으로 제작되었으며, 사용된 오픈소스 컴포넌트들의 라이선스를 준수합니다.

- ModSecurity: Apache 2.0 License
- OWASP CRS: Apache 2.0 License

---

*더 많은 정보가 필요하시면 [GitHub Issues](../../issues)에 문의해 주세요.*