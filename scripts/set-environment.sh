#!/bin/bash

# WAF Environment Configuration Script
# Usage: ./scripts/set-environment.sh [local|docker|kubernetes]

ENV_TYPE=${1:-local}

echo "🔧 Setting WAF environment to: $ENV_TYPE"

case $ENV_TYPE in
  "local")
    echo "📍 Local development environment (localhost:3000)"
    
    # Update frontend .env.local
    cat > frontend/.env.local << EOF
# WAF Frontend Local Development Configuration
REACT_APP_API_URL=http://localhost:3000
REACT_APP_WS_URL=ws://localhost:3000/api/v1/ws
REACT_APP_DEV_MODE=true
REACT_APP_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
EOF
    
    # Update backend target URL
    export TARGET_URL="http://localhost:3000"
    
    echo "✅ Local environment configured"
    echo "   Frontend: http://localhost:3000"
    echo "   API: http://localhost:3000/api"
    echo "   WebSocket: ws://localhost:3000/api/v1/ws"
    ;;
    
  "docker")
    echo "🐳 Docker environment (localhost:80)"
    
    cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://localhost
REACT_APP_WS_URL=ws://localhost/api/v1/ws
REACT_APP_DEV_MODE=false
REACT_APP_OAUTH_REDIRECT_URL=http://localhost/auth/callback
EOF
    
    export TARGET_URL="http://localhost"
    
    echo "✅ Docker environment configured"
    echo "   Frontend: http://localhost"
    echo "   API: http://localhost/api"
    ;;
    
  "kubernetes")
    echo "☸️  Kubernetes environment (via port-forward)"
    
    cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://localhost:3000
REACT_APP_WS_URL=ws://localhost:3000/api/v1/ws
REACT_APP_DEV_MODE=false
REACT_APP_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
EOF
    
    export TARGET_URL="http://localhost:3000"
    
    echo "✅ Kubernetes environment configured"
    echo "   Frontend: http://localhost:3000 (via port-forward)"
    echo "   API: http://localhost:3000/api"
    echo "   Run: kubectl port-forward service/ingress-nginx-controller -n ingress-nginx 3000:80"
    ;;
    
  *)
    echo "❌ Invalid environment type: $ENV_TYPE"
    echo "Usage: $0 [local|docker|kubernetes]"
    exit 1
    ;;
esac

echo ""
echo "🚀 Environment set to: $ENV_TYPE"
echo "💡 Run './scripts/build-images.sh' to rebuild with new configuration"