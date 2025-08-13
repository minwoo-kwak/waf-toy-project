# WAF 환경 설정 가이드

WAF 시스템의 URL과 환경을 중앙에서 관리하는 방법입니다.

## 🚀 빠른 환경 전환

### 1. 스크립트로 환경 설정
```bash
# Local development (localhost:3000)
./scripts/set-environment.sh local

# Docker environment (localhost:80)  
./scripts/set-environment.sh docker

# Kubernetes environment (waf-local.dev)
./scripts/set-environment.sh kubernetes
```

### 2. 수동 환경 설정

#### Local Development
```bash
# Frontend 환경변수
cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://localhost:3000
REACT_APP_WS_URL=ws://localhost:3000/api/v1/ws
REACT_APP_DEV_MODE=true
REACT_APP_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
EOF

# 포트포워딩 실행
kubectl port-forward service/ingress-nginx-controller -n ingress-nginx 3000:80
```

#### Docker Environment  
```bash
# Frontend 환경변수
cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://localhost
REACT_APP_WS_URL=ws://localhost/api/v1/ws
REACT_APP_DEV_MODE=false
REACT_APP_OAUTH_REDIRECT_URL=http://localhost/auth/callback
EOF
```

#### Kubernetes Environment
```bash
# Frontend 환경변수
cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://waf-local.dev
REACT_APP_WS_URL=ws://waf-local.dev/api/v1/ws
REACT_APP_DEV_MODE=false
REACT_APP_OAUTH_REDIRECT_URL=http://waf-local.dev/auth/callback
EOF

# hosts 파일에 추가
echo "127.0.0.1 waf-local.dev" >> /etc/hosts  # Linux/Mac
# Windows: C:\Windows\System32\drivers\etc\hosts 파일 편집
```

## 📁 파일 구조

```
├── config/
│   └── environments.yaml     # 모든 환경 설정 정의
├── frontend/
│   ├── .env.example         # 환경변수 템플릿
│   └── .env.local           # 로컬 환경변수 (gitignore)
└── scripts/
    └── set-environment.sh   # 환경 자동 전환 스크립트
```

## 🔧 환경변수 목록

### Frontend (.env.local)
- `REACT_APP_API_URL`: API 베이스 URL
- `REACT_APP_WS_URL`: WebSocket URL  
- `REACT_APP_DEV_MODE`: 개발 모드 여부
- `REACT_APP_OAUTH_REDIRECT_URL`: OAuth 콜백 URL

### Backend (환경변수)
- `TARGET_URL`: 보안 테스트 타겟 URL
- `OAUTH_REDIRECT_URL`: OAuth 리다이렉트 URL

## 📋 체크리스트

환경 전환 후 확인사항:

- [ ] `frontend/.env.local` 파일 생성됨
- [ ] URL이 올바르게 설정됨
- [ ] Docker 이미지 재빌드 필요시 실행
- [ ] 포트포워딩 또는 Ingress 설정 확인
- [ ] OAuth 콜백 URL이 Google Console과 일치
- [ ] 브라우저에서 접속 테스트

## 🐛 트러블슈팅

**문제: API 요청이 실패함**
```bash
# 현재 환경변수 확인
cat frontend/.env.local

# API 연결 테스트
curl http://localhost:3000/api/v1/health
```

**문제: WebSocket 연결 안됨**
```bash
# WebSocket URL 확인
echo $REACT_APP_WS_URL

# nginx WebSocket 프록시 설정 확인
kubectl logs -l app=waf-frontend
```

**문제: OAuth 콜백 실패**
```bash  
# 콜백 URL 확인
echo $REACT_APP_OAUTH_REDIRECT_URL

# Google OAuth 설정과 비교
# https://console.developers.google.com
```