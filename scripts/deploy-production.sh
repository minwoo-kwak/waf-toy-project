#!/bin/bash

# WAF Toy Project - Production Deployment Script for Naver Cloud
echo "🚀 Deploying WAF Toy Project to Naver Cloud NKS..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Set kubeconfig
export KUBECONFIG="C:\workspace\kubeconfig"
echo "📝 Using kubeconfig: $KUBECONFIG"

# Check if Naver Cloud cluster is running
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Naver Cloud NKS cluster is not accessible. Please check kubeconfig and network."
    exit 1
fi

echo "🔍 Naver Cloud NKS cluster info:"
kubectl cluster-info

# Check if nginx ingress is ready
echo "🔍 Checking Nginx Ingress Controller..."
kubectl get pods -n ingress-nginx | grep ingress-nginx-controller

# Apply ModSecurity ConfigMaps
echo "🛡️ Applying ModSecurity configuration..."
kubectl apply -f k8s/modsecurity/modsecurity-configmap.yaml
kubectl apply -f k8s/modsecurity/custom-rules-configmap.yaml

# Apply production configuration
echo "⚙️ Applying production configuration..."
kubectl apply -f k8s/production/configmap-production.yaml

# Apply PVC (if not exists)
echo "💾 Applying persistent volume claim..."
kubectl apply -f k8s/backend/pvc.yaml
kubectl apply -f k8s/backend/rbac.yaml

# Deploy backend services
echo "📦 Deploying backend services..."
kubectl apply -f k8s/production/backend-deployment.yaml
kubectl apply -f k8s/backend/service.yaml

# Deploy frontend services  
echo "🌐 Deploying frontend services..."
kubectl apply -f k8s/production/frontend-deployment.yaml
kubectl apply -f k8s/frontend/service.yaml

# Deploy ingress
echo "🚪 Deploying production ingress..."
kubectl apply -f k8s/production/ingress-production.yaml

# Wait for deployments to be ready
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/waf-backend
kubectl wait --for=condition=available --timeout=300s deployment/waf-frontend

# Show deployment status
echo "📊 Deployment Status:"
kubectl get pods -o wide
kubectl get services
kubectl get ingress

echo ""
echo "🎉 Production deployment completed successfully!"
echo ""
echo "📋 Access Information:"
echo "   Frontend: http://waftest.p-e.kr:31264"
echo "   Backend API: http://waftest.p-e.kr:31264/api"
echo "   Domain: waftest.p-e.kr → 175.45.204.150:31264"
echo ""
echo "🔧 Next Steps:"
echo "   1. Build and push Docker images with production environment variables"
echo "   2. Update Google OAuth client credentials in Secret"
echo "   3. Test the application: http://waftest.p-e.kr:31264"
echo ""
echo "🛡️ ModSecurity WAF is enabled with OWASP CRS"
echo "📝 View logs: kubectl logs -n ingress-nginx deployment/ingress-nginx-controller"