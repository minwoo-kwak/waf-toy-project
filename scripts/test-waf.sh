#!/bin/bash

# WAF Toy Project - Security Testing Script
echo "🔐 Testing WAF Security Features..."

# Set target URL
TARGET_URL="http://waf-local.dev"
API_URL="$TARGET_URL/api/v1/ping"

echo "🎯 Target: $TARGET_URL"

# Check if target is reachable
if ! curl -s --connect-timeout 5 $TARGET_URL > /dev/null; then
    echo "❌ Target $TARGET_URL is not reachable. Make sure the application is deployed."
    echo "💡 Run: ./scripts/deploy-k8s.sh first"
    exit 1
fi

echo "✅ Target is reachable"
echo ""

# Test 1: Normal request (should pass)
echo "🧪 Test 1: Normal API request"
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/test_response $API_URL)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ Normal request: PASSED (HTTP $RESPONSE)"
else
    echo "❌ Normal request: FAILED (HTTP $RESPONSE)"
fi

# Test 2: SQL Injection (should be blocked)
echo ""
echo "🧪 Test 2: SQL Injection attack"
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_URL?id=1' OR '1'='1")
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "406" ]; then
    echo "✅ SQL Injection: BLOCKED (HTTP $RESPONSE)"
else
    echo "❌ SQL Injection: NOT BLOCKED (HTTP $RESPONSE) - WAF may not be working properly"
fi

# Test 3: XSS Attack (should be blocked)
echo ""
echo "🧪 Test 3: XSS attack"
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_URL?search=<script>alert('xss')</script>")
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "406" ]; then
    echo "✅ XSS Attack: BLOCKED (HTTP $RESPONSE)"
else
    echo "❌ XSS Attack: NOT BLOCKED (HTTP $RESPONSE) - WAF may not be working properly"
fi

# Test 4: Path Traversal (should be blocked)
echo ""
echo "🧪 Test 4: Path Traversal attack"
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_URL?file=../../../etc/passwd")
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "406" ]; then
    echo "✅ Path Traversal: BLOCKED (HTTP $RESPONSE)"
else
    echo "❌ Path Traversal: NOT BLOCKED (HTTP $RESPONSE) - WAF may not be working properly"
fi

# Test 5: Command Injection (should be blocked)
echo ""
echo "🧪 Test 5: Command Injection attack"
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_URL?cmd=cat /etc/passwd")
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "406" ]; then
    echo "✅ Command Injection: BLOCKED (HTTP $RESPONSE)"
else
    echo "❌ Command Injection: NOT BLOCKED (HTTP $RESPONSE) - WAF may not be working properly"
fi

# Test 6: Malicious User Agent (should be blocked)
echo ""
echo "🧪 Test 6: Malicious User Agent"
RESPONSE=$(curl -s -w "%{http_code}" -H "User-Agent: sqlmap/1.0" -o /tmp/test_response $API_URL)
if [ "$RESPONSE" = "403" ] || [ "$RESPONSE" = "406" ]; then
    echo "✅ Malicious User Agent: BLOCKED (HTTP $RESPONSE)"
else
    echo "❌ Malicious User Agent: NOT BLOCKED (HTTP $RESPONSE) - WAF may not be working properly"
fi

# Test 7: Rate Limiting (stress test)
echo ""
echo "🧪 Test 7: Rate Limiting (sending 20 rapid requests)"
BLOCKED_COUNT=0
for i in {1..20}; do
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null $API_URL)
    if [ "$RESPONSE" = "429" ] || [ "$RESPONSE" = "503" ]; then
        ((BLOCKED_COUNT++))
    fi
    sleep 0.1
done

if [ "$BLOCKED_COUNT" -gt 0 ]; then
    echo "✅ Rate Limiting: WORKING ($BLOCKED_COUNT requests blocked)"
else
    echo "⚠️  Rate Limiting: NO BLOCKS DETECTED (may need configuration adjustment)"
fi

echo ""
echo "🎉 WAF Security Testing Completed!"
echo ""
echo "📊 View ModSecurity logs:"
echo "   kubectl logs -n ingress-nginx deployment/ingress-nginx-controller | grep modsec"
echo ""
echo "🔍 For detailed analysis, check:"
echo "   - Ingress Controller logs: kubectl logs -n ingress-nginx deployment/ingress-nginx-controller"
echo "   - Application logs: kubectl logs deployment/waf-backend"

# Cleanup temp files
rm -f /tmp/test_response