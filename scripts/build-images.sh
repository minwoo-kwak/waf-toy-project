#!/bin/bash

# WAF Toy Project - Docker Images Build Script
echo "🏗️ Building WAF Toy Project Docker Images..."

# Set image tags - Enhanced UI/UX and Security v3.0.0
BACKEND_IMAGE="waf-backend:v3.0.0"
FRONTEND_IMAGE="waf-frontend:v3.0.0"

# Build backend image
echo "📦 Building backend image..."
cd backend
if docker build -t $BACKEND_IMAGE .; then
    echo "✅ Backend image built successfully: $BACKEND_IMAGE"
else
    echo "❌ Failed to build backend image"
    exit 1
fi

# Build frontend image
echo "📦 Building frontend image..."
cd ../frontend
if docker build -t $FRONTEND_IMAGE .; then
    echo "✅ Frontend image built successfully: $FRONTEND_IMAGE"
else
    echo "❌ Failed to build frontend image"
    exit 1
fi

cd ..

# Show built images
echo "🎯 Docker images built:"
docker images | grep -E "(waf-backend|waf-frontend)"

echo "🎉 All images built successfully!"
echo ""
echo "Next steps:"
echo "1. Run: kubectl apply -f scripts/deploy-k8s.sh"
echo "2. Or run: docker-compose up for local testing"