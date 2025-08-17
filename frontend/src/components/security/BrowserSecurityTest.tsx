import React, { useState } from 'react';
import {
  Card,
  CardContent,
  Typography,
  Button,
  Alert,
  Box,
  Chip,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  Security,
  OpenInNew,
  BugReport,
  Code,
  Folder,
  Terminal,
  CheckCircle,
  ErrorOutline,
} from '@mui/icons-material';

interface AttackVector {
  id: string;
  name: string;
  description: string;
  icon: React.ReactElement;
  urls: string[];
}

const BrowserSecurityTest: React.FC = () => {
  const [testRunning, setTestRunning] = useState(false);
  const [openedTabs, setOpenedTabs] = useState<number>(0);

  const attackVectors: AttackVector[] = [
    {
      id: 'sql_injection',
      name: 'SQL Injection',
      description: 'SQL 삽입 공격 테스트',
      icon: <BugReport color="error" />,
      urls: [
        "/dashboard?id=' OR '1'='1",
        "/dashboard?user=admin' UNION SELECT * FROM users--",
        "/dashboard?id=1; DROP TABLE users--",
        "/dashboard?search=' AND 1=1--",
      ]
    },
    {
      id: 'xss',
      name: 'Cross-Site Scripting (XSS)',
      description: 'XSS 공격 테스트',
      icon: <Code color="warning" />,
      urls: [
        "/dashboard?search=<script>alert('XSS')</script>",
        "/dashboard?name=<img src=x onerror=alert('XSS')>",
        "/dashboard?comment=<svg onload=alert('XSS')>",
        "/dashboard?data=javascript:alert('XSS')",
      ]
    },
    {
      id: 'path_traversal',
      name: 'Path Traversal',
      description: '경로 순회 공격 테스트',
      icon: <Folder color="info" />,
      urls: [
        "/dashboard?file=../../../etc/passwd",
        "/dashboard?path=..\\..\\..\\boot.ini",
        "/dashboard?include=../../../../etc/shadow",
        "/dashboard?load=..%2F..%2F..%2Fetc%2Fpasswd",
      ]
    },
    {
      id: 'command_injection',
      name: 'Command Injection',
      description: '명령어 삽입 공격 테스트',
      icon: <Terminal color="secondary" />,
      urls: [
        "/dashboard?cmd=; cat /etc/passwd",
        "/dashboard?exec=| whoami",
        "/dashboard?run=`id`",
        "/dashboard?system=; ls -la",
      ]
    }
  ];

  const handleRunAllTests = async () => {
    // 팝업 차단 경고 표시
    const userConfirmed = window.confirm(
      '🚨 팝업 차단 해제 필요\n\n이 테스트는 여러 탭을 열어 보안 테스트를 수행합니다.\n브라우저에서 팝업 차단을 해제해주세요.\n\n계속하시겠습니까?'
    );
    
    if (!userConfirmed) {
      return;
    }
    
    setTestRunning(true);
    setOpenedTabs(0);
    
    try {
      let totalUrls = 0;
      
      // 모든 공격 벡터의 URL 수 계산
      attackVectors.forEach(vector => {
        totalUrls += vector.urls.length;
      });
      
      // 첫 번째 URL만 즉시 열고, 나머지는 사용자가 수동으로 열도록 안내
      let urlsToOpen: string[] = [];
      
      attackVectors.forEach(vector => {
        vector.urls.forEach(url => {
          const testTargetUrl = process.env.REACT_APP_TEST_TARGET_URL || 'http://localhost';
          const fullUrl = testTargetUrl + url;
          urlsToOpen.push(fullUrl);
        });
      });
      
      // 첫 번째 URL을 열어서 팝업 차단 여부 확인
      const firstTab = window.open(urlsToOpen[0], '_blank', 'noopener,noreferrer');
      
      if (!firstTab) {
        alert('❌ 팝업이 차단되었습니다!\n\n브라우저 주소창 옆의 팝업 차단 아이콘을 클릭하여 팝업을 허용해주세요.\n그 후 다시 테스트를 실행하세요.');
        return;
      }
      
      setOpenedTabs(1);
      
      // 나머지 URL들을 순차적으로 열기 (각각 사용자 액션으로 처리)
      for (let i = 1; i < Math.min(urlsToOpen.length, 8); i++) { // 최대 8개만
        await new Promise(resolve => setTimeout(resolve, 800));
        const newTab = window.open(urlsToOpen[i], '_blank', 'noopener,noreferrer');
        if (newTab) {
          setOpenedTabs(prev => prev + 1);
        }
      }
      
      // 결과 확인 안내
      setTimeout(() => {
        alert(`🚀 ${Math.min(urlsToOpen.length, 8)}개의 공격 테스트가 새 탭에서 실행되었습니다!\\n\\n📋 결과 확인 방법:\\n• 403 Forbidden = ModSecurity 차단 성공 ✅\\n• 200 OK = 차단 실패 ❌\\n\\n각 탭을 확인하여 ModSecurity 차단 여부를 확인하세요.`);
      }, 2000);
      
    } catch (error) {
      console.error('Security test failed:', error);
      alert('보안 테스트 실행 중 오류가 발생했습니다.');
    } finally {
      setTestRunning(false);
    }
  };

  const handleRunSingleVector = async (vector: AttackVector) => {
    setTestRunning(true);
    
    try {
      for (const url of vector.urls) {
        const testTargetUrl = process.env.REACT_APP_TEST_TARGET_URL || 'http://localhost';
        const fullUrl = testTargetUrl + url;
        window.open(fullUrl, '_blank', 'noopener,noreferrer');
        await new Promise(resolve => setTimeout(resolve, 300));
      }
      
      setTimeout(() => {
        alert(`🎯 ${vector.name} 테스트가 ${vector.urls.length}개 탭에서 실행되었습니다!\\n\\n각 탭에서 403 Forbidden이 나오면 ModSecurity 차단 성공입니다.`);
      }, 1000);
      
    } catch (error) {
      console.error(`${vector.name} test failed:`, error);
      alert(`${vector.name} 테스트 실행 중 오류가 발생했습니다.`);
    } finally {
      setTestRunning(false);
    }
  };

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Security sx={{ mr: 1, color: 'primary.main' }} />
          <Typography variant="h6" component="h2">
            브라우저 보안 테스트
          </Typography>
        </Box>
        
        <Alert severity="info" sx={{ mb: 3 }}>
          <Typography variant="body2">
            <strong>🚀 새 탭 기반 ModSecurity 테스트</strong><br />
            각 공격 패턴을 새 탭에서 실행하여 ModSecurity 차단 여부를 확인합니다.<br />
            <strong>결과:</strong> 403 Forbidden = 차단 성공 ✅ | 200 OK = 차단 실패 ❌
          </Typography>
        </Alert>

        <Box sx={{ mb: 3 }}>
          <Button
            variant="contained"
            color="primary"
            size="large"
            onClick={handleRunAllTests}
            disabled={testRunning}
            startIcon={<OpenInNew />}
            sx={{ mr: 2 }}
          >
            {testRunning ? '테스트 실행 중...' : '🚀 모든 공격 테스트 실행'}
          </Button>
          
          {openedTabs > 0 && (
            <Chip 
              label={`${openedTabs}개 탭 열림`} 
              color="success" 
              icon={<CheckCircle />}
            />
          )}
        </Box>

        <Divider sx={{ my: 3 }} />

        <Typography variant="h6" gutterBottom>
          개별 공격 벡터 테스트
        </Typography>

        <List>
          {attackVectors.map((vector, index) => (
            <React.Fragment key={vector.id}>
              <ListItem
                sx={{
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  mb: 1,
                }}
              >
                <ListItemIcon>
                  {vector.icon}
                </ListItemIcon>
                <ListItemText
                  primary={vector.name}
                  secondary={
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        {vector.description}
                      </Typography>
                      <Typography variant="caption" color="text.disabled">
                        {vector.urls.length}개 테스트 패턴
                      </Typography>
                    </Box>
                  }
                />
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => handleRunSingleVector(vector)}
                  disabled={testRunning}
                  startIcon={<OpenInNew />}
                >
                  테스트
                </Button>
              </ListItem>
            </React.Fragment>
          ))}
        </List>

        <Alert severity="warning" sx={{ mt: 3 }}>
          <Typography variant="body2">
            <strong>⚠️ 주의사항</strong><br />
            • 이 테스트는 교육 및 보안 검증 목적으로만 사용하세요<br />
            • 실제 운영 환경에서는 사용하지 마세요<br />
            • 팝업 차단이 활성화되어 있으면 새 탭이 열리지 않을 수 있습니다
          </Typography>
        </Alert>
      </CardContent>
    </Card>
  );
};

export default BrowserSecurityTest;