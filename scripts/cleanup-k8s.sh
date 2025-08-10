#!/bin/bash

# WAF Toy Project - Kubernetes Cleanup Script
echo "🧹 Cleaning up WAF Toy Project from Kubernetes..."

# Remove application resources
echo "🗑️ Removing application resources..."
kubectl delete -f k8s/ingress/ingress.yaml --ignore-not-found=true
kubectl delete -f k8s/frontend/ --ignore-not-found=true
kubectl delete -f k8s/backend/ --ignore-not-found=true
kubectl delete -f k8s/modsecurity/modsecurity-configmap.yaml --ignore-not-found=true

# Remove custom ConfigMaps
echo "🗑️ Removing custom ConfigMaps..."
kubectl delete configmap nginx-configuration -n ingress-nginx --ignore-not-found=true
kubectl delete configmap custom-headers -n ingress-nginx --ignore-not-found=true

# Optional: Remove Ingress Nginx Controller (uncomment if needed)
# echo "🗑️ Removing Ingress Nginx Controller..."
# kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml --ignore-not-found=true

echo "🎉 Cleanup completed!"
echo ""
echo "📊 Remaining resources:"
kubectl get pods
kubectl get services
kubectl get ingress