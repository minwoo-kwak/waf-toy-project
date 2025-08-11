# WAF SaaS Kubernetes Deployment

## 환경변수 관리 방법

### 1. ConfigMap (비민감 정보)
- 서버 포트, CORS 설정, API URL 등
- `k8s/configmap.yaml`에 정의

### 2. Secret (민감 정보)
- Google OAuth 클라이언트 ID/Secret
- JWT 시크릿 키
- `k8s/secret.yaml`에 정의

## 배포 방법

### 1. Docker 이미지 빌드
```bash
# Backend 이미지 빌드
cd backend
docker build -t waf-saas-backend:latest .

# Frontend 이미지 빌드  
cd frontend
docker build -t waf-saas-frontend:latest .
```

### 2. Kubernetes 배포
```bash
cd k8s
chmod +x deploy.sh
./deploy.sh
```

### 3. 개별 리소스 적용
```bash
# RBAC 설정
kubectl apply -f rbac.yaml

# ConfigMap과 Secret 생성
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# 애플리케이션 배포
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml
```

## 환경변수 수정

### ConfigMap 수정
```bash
kubectl edit configmap waf-saas-config
kubectl edit configmap waf-saas-frontend-config
```

### Secret 수정
```bash
kubectl edit secret waf-saas-secrets
```

### 변경사항 적용
```bash
# 팟 재시작으로 새 환경변수 적용
kubectl rollout restart deployment/waf-saas-backend
kubectl rollout restart deployment/waf-saas-frontend
```

## 상태 확인

### 팟 상태
```bash
kubectl get pods
kubectl describe pod <pod-name>
```

### 로그 확인
```bash
kubectl logs -f deployment/waf-saas-backend
kubectl logs -f deployment/waf-saas-frontend
```

### 서비스 확인
```bash
kubectl get services
kubectl port-forward service/waf-saas-frontend-service 3000:3000
kubectl port-forward service/waf-saas-backend-service 8080:8080
```

## 환경변수 파일에서 Kubernetes로 마이그레이션

기존 `.env` 파일의 내용이 다음과 같이 Kubernetes 리소스로 변환됩니다:

**`.env` → `ConfigMap`**
- PORT → waf-saas-config.PORT
- CORS_ORIGIN → waf-saas-config.CORS_ORIGIN
- LOG_LEVEL → waf-saas-config.LOG_LEVEL

**`.env` → `Secret`**  
- GOOGLE_CLIENT_ID → waf-saas-secrets.GOOGLE_CLIENT_ID
- GOOGLE_CLIENT_SECRET → waf-saas-secrets.GOOGLE_CLIENT_SECRET
- JWT_SECRET → waf-saas-secrets.JWT_SECRET