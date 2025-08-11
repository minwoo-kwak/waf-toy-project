# WAF SaaS Platform - 트러블슈팅 가이드

## OAuth 콜백 라우팅 문제 해결

### 문제 상황
Google OAuth 로그인 시 "Authentication Error" 발생하며 콜백이 실패

### 에러 로그
```
API Base URL: undefined
Full callback URL: undefined/api/v1/public/auth/callback
XHR POST http://localhost:3000/api/v1/public/auth/callback [HTTP/1.1 405 Not Allowed 5ms]
Auth callback error: Error: Authentication failed
```

### 원인 분석
1. **환경변수 문제**: `REACT_APP_API_BASE_URL`이 undefined로 설정됨
2. **라우팅 문제**: API 요청이 프론트엔드 서버(`localhost:3000`)로 직접 전송
3. **프록시 설정 누락**: 프론트엔드 nginx에서 백엔드로의 프록시 설정이 없음

### 해결 과정

#### 1. 문제 식별
- 콜백 요청이 `http://localhost:3000/api/v1/public/auth/callback`으로 전송
- 프론트엔드에는 해당 엔드포인트가 없어 405 Not Allowed 에러 발생
- 백엔드 서비스로 라우팅되지 않음

#### 2. nginx 프록시 설정 추가
**파일**: `frontend/default.conf`

**변경 전**:
```nginx
server {
    listen       80;
    server_name  localhost;

    location / {
        root   /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
}
```

**변경 후**:
```nginx
server {
    listen       80;
    server_name  localhost;

    # Proxy API requests to backend
    location /api/ {
        proxy_pass http://waf-backend-service:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With";
        
        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With";
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
    }

    location / {
        root   /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
}
```

#### 3. 서비스명 오류 수정
**문제**: 처음에 `http://waf-backend:8080`로 설정하여 호스트를 찾을 수 없는 에러 발생
```
nginx: [emerg] host not found in upstream "waf-backend"
```

**해결**: Kubernetes 서비스명 확인 후 올바른 이름으로 수정
```bash
kubectl get services
# 결과: waf-backend-service가 실제 서비스명
```

**수정**: `http://waf-backend-service:8080/api/`로 변경

#### 4. Docker 이미지 재빌드 및 배포
```bash
# 1. nginx 설정이 포함된 새 이미지 빌드
docker build -t waf-frontend:v2.0.8 ./frontend

# 2. Kubernetes 배포 업데이트
kubectl set image deployment/waf-frontend waf-frontend=waf-frontend:v2.0.8

# 3. 롤아웃 상태 확인
kubectl rollout status deployment/waf-frontend
```

### 해결 결과
- `/api/v1/public/auth/callback` 요청이 nginx 프록시를 통해 백엔드로 올바르게 라우팅
- OAuth 콜백 처리가 백엔드에서 정상적으로 수행
- 405 Not Allowed 에러 해결

### 핵심 포인트
1. **프론트엔드와 백엔드 분리**: SPA에서 API 호출 시 프록시 설정 필수
2. **Kubernetes 서비스 디스커버리**: 올바른 서비스명 사용 중요
3. **CORS 설정**: API 프록시에서 CORS 헤더 적절히 설정
4. **nginx 설정**: location 블록 순서와 프록시 헤더 설정 중요

### 예방 방법
1. nginx 설정 시 Kubernetes 서비스명을 정확히 확인
2. 프록시 설정 후 즉시 테스트하여 연결성 확인
3. Docker 이미지 빌드 시 nginx 설정 파일 포함 여부 확인
4. 롤아웃 후 pod 로그를 통해 에러 여부 즉시 확인

### 관련 파일들
- `frontend/default.conf` - nginx 프록시 설정
- `frontend/Dockerfile` - nginx 설정 파일 복사
- `frontend/src/components/auth/AuthCallback.tsx` - OAuth 콜백 처리
- `k8s/waf-frontend-deployment.yaml` - 프론트엔드 배포 설정