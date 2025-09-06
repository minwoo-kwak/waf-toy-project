#!/bin/bash

# SQLite DB 관리 스크립트

POD_NAME=$(kubectl get pods -l app=waf-backend -o name | head -1)

if [ -z "$POD_NAME" ]; then
    echo "❌ Backend pod not found"
    exit 1
fi

function show_help() {
    echo "📖 DB 관리 명령어:"
    echo "  check    - DB 상태 확인"
    echo "  rules    - 커스텀 룰 목록"
    echo "  users    - 사용자 목록"
    echo "  sql      - 직접 SQL 실행"
    echo "  backup   - DB 백업"
    echo "  restore  - DB 복원"
    echo "  clear    - 모든 데이터 삭제"
}

function check_db() {
    echo "📊 Database Status:"
    kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db '.schema'"
}

function show_rules() {
    echo "🔍 Custom Rules:"
    kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db -header -column 'SELECT id, name, enabled, severity, created_at FROM custom_rules;'"
}

function show_users() {
    echo "👥 Users:"
    kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db -header -column 'SELECT id, name, email, created_at FROM users;'"
}

function run_sql() {
    echo "💻 Enter SQL query:"
    read -r query
    kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db '$query'"
}

function backup_db() {
    timestamp=$(date +%Y%m%d_%H%M%S)
    echo "💾 Backing up database..."
    kubectl exec $POD_NAME -- sh -c "cp /data/waf.db /data/waf_backup_$timestamp.db"
    echo "✅ Backup created: waf_backup_$timestamp.db"
}

function clear_data() {
    echo "⚠️  This will delete ALL data. Continue? (y/N)"
    read -r confirm
    if [[ $confirm =~ ^[Yy]$ ]]; then
        kubectl exec $POD_NAME -- sh -c "sqlite3 /data/waf.db 'DELETE FROM custom_rules; DELETE FROM users;'"
        echo "🗑️  All data cleared"
    else
        echo "❌ Cancelled"
    fi
}

case "$1" in
    check)   check_db ;;
    rules)   show_rules ;;
    users)   show_users ;;
    sql)     run_sql ;;
    backup)  backup_db ;;
    clear)   clear_data ;;
    *)       show_help ;;
esac