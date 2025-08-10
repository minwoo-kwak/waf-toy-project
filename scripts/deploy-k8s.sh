#!/bin/bash

# WAF Toy Project - Kubernetes Deployment Script
echo "🚀 Deploying WAF Toy Project to Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if Docker Desktop Kubernetes is running
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Kubernetes cluster is not running. Please start Docker Desktop Kubernetes."
    exit 1
fi

echo "🔍 Kubernetes cluster info:"
kubectl cluster-info

# Install Ingress Nginx Controller
echo "📥 Installing Ingress Nginx Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Wait for Ingress Controller to be ready
echo "⏳ Waiting for Ingress Controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

# Apply ModSecurity ConfigMaps
echo "🛡️ Applying ModSecurity configuration..."
kubectl apply -f k8s/modsecurity/modsecurity-configmap.yaml

# Deploy backend services
echo "📦 Deploying backend services..."
kubectl apply -f k8s/backend/deployment.yaml
kubectl apply -f k8s/backend/service.yaml

# Deploy frontend services  
echo "🌐 Deploying frontend services..."
kubectl apply -f k8s/frontend/deployment.yaml
kubectl apply -f k8s/frontend/service.yaml

# Deploy ingress
echo "🚪 Deploying ingress..."
kubectl apply -f k8s/ingress/ingress.yaml

# Wait for deployments to be ready
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/waf-backend
kubectl wait --for=condition=available --timeout=300s deployment/waf-frontend

# Show deployment status
echo "📊 Deployment Status:"
kubectl get pods -o wide
kubectl get services
kubectl get ingress

# Get Ingress Controller LoadBalancer IP
INGRESS_IP=$(kubectl get service/ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ -z "$INGRESS_IP" ]; then
    INGRESS_IP=$(kubectl get service/ingress-nginx-controller -n ingress-nginx -o jsonpath='{.spec.clusterIP}')
fi

echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "📋 Access Information:"
echo "   Frontend: http://waf-local.dev"
echo "   Backend API: http://waf-local.dev/api"
echo "   Ingress IP: $INGRESS_IP"
echo ""
echo "🔧 To access the application:"
echo "   1. Add this line to your /etc/hosts (Linux/Mac) or C:\\Windows\\System32\\drivers\\etc\\hosts (Windows):"
echo "      $INGRESS_IP waf-local.dev"
echo "   2. Open browser: http://waf-local.dev"
echo ""
echo "🛡️ ModSecurity is enabled with OWASP CRS"
echo "📝 View logs: kubectl logs -n ingress-nginx deployment/ingress-nginx-controller"