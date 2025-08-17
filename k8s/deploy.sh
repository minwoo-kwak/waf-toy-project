#!/bin/bash

echo "🚀 Deploying WAF SaaS to Kubernetes..."

# Apply RBAC first
echo "📋 Creating ServiceAccount and RBAC..."
kubectl apply -f rbac.yaml

# Apply ConfigMap and Secrets
echo "🔧 Creating ConfigMaps and Secrets..."
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# Apply Deployments and Services
echo "🏗️ Deploying Backend..."
kubectl apply -f backend-deployment.yaml

echo "🏗️ Deploying Frontend..."
kubectl apply -f frontend-deployment.yaml

echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/waf-saas-backend
kubectl wait --for=condition=available --timeout=300s deployment/waf-saas-frontend

echo "✅ Deployment completed!"
echo ""
echo "📊 Checking status..."
kubectl get pods -l app=waf-saas-backend
kubectl get pods -l app=waf-saas-frontend
kubectl get services

echo ""
echo "🌐 Access URLs:"
echo "Frontend: http://localhost:3000"
echo "Backend API: http://localhost:8080"
echo ""
echo "📝 To view logs:"
echo "kubectl logs -f deployment/waf-saas-backend"
echo "kubectl logs -f deployment/waf-saas-frontend"