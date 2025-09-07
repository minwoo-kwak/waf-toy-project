#!/bin/bash

# SQLite DB 확인 스크립트

POD_NAME=$(kubectl get pods -l app=waf-backend -o name | head -1)

if [ -z "$POD_NAME" ]; then
    echo "❌ Backend pod not found"
    exit 1
fi

echo "📊 WAF Database Status"
echo "======================"

echo "📁 DB 파일 정보:"
kubectl exec $POD_NAME -- sh -c "ls -lh /data/waf.db"

echo ""
echo "📋 테이블 목록:"
kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db '.tables'"

echo ""
echo "📈 테이블별 데이터 개수:"
echo "- custom_rules: $(kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db 'SELECT count(*) FROM custom_rules;'")"
echo "- users: $(kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db 'SELECT count(*) FROM users;'")"

echo ""
echo "🔍 커스텀 룰 목록:"
kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db 'SELECT id, name, enabled, severity FROM custom_rules;'" | while IFS='|' read -r id name enabled severity; do
    status=$([ "$enabled" = "1" ] && echo "🟢" || echo "🔴")
    echo "  $status $name ($id) - $severity"
done

echo ""
echo "👥 사용자 목록:"
kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db 'SELECT id, name, email FROM users;'" | while IFS='|' read -r id name email; do
    echo "  👤 $name ($email)"
done