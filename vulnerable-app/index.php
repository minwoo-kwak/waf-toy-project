<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WAF 테스트용 취약한 웹 애플리케이션</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .test-section { border: 1px solid #ddd; margin: 20px 0; padding: 15px; border-radius: 5px; background-color: #f9f9f9; }
        .attack-btn { background-color: #dc3545; color: white; padding: 15px 30px; border: none; border-radius: 5px; cursor: pointer; margin: 10px; font-size: 16px; font-weight: bold; }
        .attack-btn:hover { background-color: #c82333; }
        .test-all-btn { background-color: #6f42c1; color: white; padding: 20px 40px; border: none; border-radius: 5px; cursor: pointer; margin: 20px 10px; font-size: 18px; font-weight: bold; }
        .payload { background-color: #f8f9fa; padding: 10px; margin: 10px 0; border-radius: 3px; font-family: monospace; border-left: 4px solid #dc3545; }
        .result { background-color: #e9ecef; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .blocked { background-color: #d4edda; color: #155724; border-left: 4px solid #28a745; }
        .passed { background-color: #f8d7da; color: #721c24; border-left: 4px solid #dc3545; }
        .warning { background-color: #fff3cd; border-left-color: #ffc107; color: #856404; padding: 15px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚨 WAF 테스트용 취약한 웹 애플리케이션</h1>
        <div class="warning">
            <strong>⚠️ 경고:</strong> 이 애플리케이션은 의도적으로 취약하게 작성되었습니다. WAF 차단 테스트 목적으로만 사용하세요.
        </div>

        <div style="text-align: center; margin: 30px 0;">
            <button class="test-all-btn" onclick="runAllTests()">🚀 모든 공격 테스트 실행</button>
        </div>

        <div class="test-section">
            <h2>1. 💉 SQL Injection 테스트</h2>
            <div class="payload">페이로드: ' OR '1'='1</div>
            <button class="attack-btn" onclick="testAttack('sql')">SQL Injection 공격 실행</button>
            <div id="sql-result"></div>
        </div>

        <div class="test-section">
            <h2>2. 🎭 XSS 테스트</h2>
            <div class="payload">페이로드: &lt;script&gt;alert('XSS')&lt;/script&gt;</div>
            <button class="attack-btn" onclick="testAttack('xss')">XSS 공격 실행</button>
            <div id="xss-result"></div>
        </div>

        <div class="test-section">
            <h2>3. 💻 Command Injection 테스트</h2>
            <div class="payload">페이로드: 127.0.0.1; ls -la</div>
            <button class="attack-btn" onclick="testAttack('cmd')">Command Injection 공격 실행</button>
            <div id="cmd-result"></div>
        </div>

        <div class="test-section">
            <h2>4. 📁 LFI 테스트</h2>
            <div class="payload">페이로드: ../../../etc/passwd</div>
            <button class="attack-btn" onclick="testAttack('lfi')">LFI 공격 실행</button>
            <div id="lfi-result"></div>
        </div>

        <div class="test-section">
            <h2>5. 📤 파일 업로드 테스트</h2>
            <div class="payload">업로드: malware.php, backdoor.exe</div>
            <button class="attack-btn" onclick="testAttack('upload')">악성 파일 업로드 실행</button>
            <div id="upload-result"></div>
        </div>

        <div class="test-section">
            <h2>📊 결과 해석</h2>
            <div class="blocked" style="padding: 10px; margin: 10px 0; border-radius: 3px; font-weight: bold;">
                ✅ WAF 차단 성공: HTTP 403, 406, 400
            </div>
            <div class="passed" style="padding: 10px; margin: 10px 0; border-radius: 3px; font-weight: bold;">
                ❌ WAF 우회됨: HTTP 200
            </div>
        </div>
    </div>

    <script>
        async function testAttack(attackType) {
            const resultDiv = document.getElementById(attackType + '-result');
            resultDiv.innerHTML = '<div class="result">🔄 테스트 중...</div>';

            try {
                const response = await fetch('?attack=' + attackType);
                const statusCode = response.status;

                if (statusCode === 403 || statusCode === 406 || statusCode === 400) {
                    resultDiv.innerHTML = '<div class="blocked" style="padding: 10px; margin: 10px 0; border-radius: 3px; font-weight: bold;">✅ WAF 차단! (HTTP ' + statusCode + ')</div>';
                } else if (statusCode === 200) {
                    resultDiv.innerHTML = '<div class="passed" style="padding: 10px; margin: 10px 0; border-radius: 3px; font-weight: bold;">❌ WAF 우회! (HTTP 200)</div>';
                } else {
                    resultDiv.innerHTML = '<div class="result">⚠️ 응답: HTTP ' + statusCode + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result">❌ 에러: ' + error.message + '</div>';
            }
        }

        async function runAllTests() {
            if (!confirm('모든 공격 테스트를 실행하시겠습니까?')) return;
            
            const attacks = ['sql', 'xss', 'cmd', 'lfi', 'upload'];
            for (let i = 0; i < attacks.length; i++) {
                await testAttack(attacks[i]);
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
            alert('테스트 완료!');
        }
    </script>
</body>
</html>
