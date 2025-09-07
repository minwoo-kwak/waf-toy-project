#!/usr/bin/env python3
import requests
import json

# Test without authentication to check if service is responding
def test_health():
    try:
        response = requests.get("http://localhost:80/health", timeout=5)
        print(f"Health check: {response.status_code}")
        if response.status_code == 200:
            print(json.dumps(response.json(), indent=2))
        return response.status_code == 200
    except Exception as e:
        print(f"Health check failed: {e}")
        return False

# Test rule creation directly through pod exec
def test_db_direct():
    import subprocess
    try:
        # Get current rules count
        result = subprocess.run([
            "kubectl", "exec", "pod/waf-backend-cddcb55b6-gwws4", "--", 
            "sh", "-c", "sqlite3 /data/waf.db 'SELECT count(*) FROM custom_rules;'"
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode == 0:
            count = result.stdout.strip()
            print(f"Current rules count: {count}")
            
            # Show existing rules
            result2 = subprocess.run([
                "kubectl", "exec", "pod/waf-backend-cddcb55b6-gwws4", "--",
                "sh", "-c", "sqlite3 /data/waf.db 'SELECT id, name, user_id FROM custom_rules;'"
            ], capture_output=True, text=True, timeout=10)
            
            if result2.returncode == 0:
                print("Existing rules:")
                print(result2.stdout)
                return True
        
        print(f"DB check failed: {result.stderr}")
        return False
    except Exception as e:
        print(f"DB direct test failed: {e}")
        return False

if __name__ == "__main__":
    print("Testing WAF Backend...")
    
    print("\n1. Health Check:")
    health_ok = test_health()
    
    print("\n2. Direct DB Check:")
    db_ok = test_db_direct()
    
    print(f"\nResults: Health={health_ok}, DB={db_ok}")
    if health_ok and db_ok:
        print("✅ Backend is running and DB is accessible")
    else:
        print("❌ Backend has issues")