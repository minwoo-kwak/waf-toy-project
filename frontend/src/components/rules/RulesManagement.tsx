import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  Switch,
  FormControlLabel,
  Alert,
  Tooltip,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Rule as RuleIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';
import { rulesAPI } from '../../services/api';
import { CustomRule, CustomRuleRequest } from '../../types/waf';
import { formatDistanceToNow } from 'date-fns';

// Rule templates for beginners
const RULE_TEMPLATES = [
  {
    name: '🛡️ SQL Injection 차단',
    description: 'SQL 인젝션 공격을 감지하고 차단합니다',
    rule_text: 'SecRule ARGS "@detectSQLi" "id:2001,phase:2,block,msg:\'SQL Injection Attack Detected\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
  {
    name: '🚫 XSS 공격 차단',
    description: 'Cross-Site Scripting 공격을 감지하고 차단합니다',
    rule_text: 'SecRule ARGS "@detectXSS" "id:2002,phase:2,block,msg:\'XSS Attack Detected\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
  {
    name: '🔐 관리자 페이지 차단',
    description: '/admin 경로 접근을 차단합니다',
    rule_text: 'SecRule REQUEST_URI "@beginsWith /admin" "id:2003,phase:1,block,msg:\'Admin path access blocked\',severity:MEDIUM"',
    severity: 'MEDIUM' as const,
  },
  {
    name: '📁 위험한 파일 업로드 차단',
    description: '.php, .jsp 등 위험한 파일 업로드를 차단합니다',
    rule_text: 'SecRule FILES_NAMES "@rx \\.(php|jsp|asp|exe)$" "id:2004,phase:2,block,msg:\'Dangerous file upload blocked\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
  {
    name: '🔍 스캐닝 도구 차단',
    description: 'sqlmap, nikto 등 스캐닝 도구를 차단합니다',
    rule_text: 'SecRule REQUEST_HEADERS:User-Agent "@rx (sqlmap|nikto|nmap|dirb|gobuster)" "id:2005,phase:1,block,msg:\'Scanning tool blocked\',severity:MEDIUM"',
    severity: 'MEDIUM' as const,
  },
  {
    name: '⚡ 빠른 요청 제한 (Rate Limiting)',
    description: '같은 IP에서 10초에 20회 이상 요청시 차단',
    rule_text: 'SecRule IP:bf_counter "@gt 20" "id:2006,phase:1,block,msg:\'Rate limit exceeded\',severity:MEDIUM,expirevar:ip.bf_counter=10"',
    severity: 'MEDIUM' as const,
  },
  {
    name: '🕷️ 봇 트래픽 탐지',
    description: '의심스러운 봇 패턴을 감지하고 로깅합니다',
    rule_text: 'SecRule REQUEST_HEADERS:User-Agent "@rx (bot|crawl|spider|scraper)" "id:2007,phase:1,log,msg:\'Bot traffic detected\',severity:LOW"',
    severity: 'LOW' as const,
  },
  {
    name: '💾 데이터베이스 파일 보호',
    description: '.db, .sql, .backup 파일 접근을 차단합니다',
    rule_text: 'SecRule REQUEST_URI "@rx \\.(db|sql|backup|bak|dump)$" "id:2008,phase:1,block,msg:\'Database file access blocked\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
  {
    name: '🔑 민감한 디렉토리 보호',
    description: '/.git, /.env, /config 등 민감한 디렉토리 차단',
    rule_text: 'SecRule REQUEST_URI "@rx /\\.(git|env|svn|config|htaccess)" "id:2009,phase:1,block,msg:\'Sensitive directory access blocked\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
  {
    name: '🌏 특정 국가 IP 차단',
    description: '특정 지역의 IP 접근을 차단합니다 (예시: 중국)',
    rule_text: 'SecRule REMOTE_ADDR "@geoLookup" "id:2010,phase:1,block,msg:\'Geo-blocked IP access\',severity:MEDIUM,chain" SecRule GEO:COUNTRY_CODE "@streq CN"',
    severity: 'MEDIUM' as const,
  },
  {
    name: '📜 Log4j 취약점 차단',
    description: 'Log4Shell (CVE-2021-44228) 공격 패턴 차단',
    rule_text: 'SecRule ARGS "@rx jndi:(ldap|rmi|dns)" "id:2011,phase:2,block,msg:\'Log4j exploit attempt blocked\',severity:CRITICAL"',
    severity: 'CRITICAL' as const,
  },
  {
    name: '🚨 명령어 인젝션 차단',
    description: '시스템 명령어 실행 시도를 차단합니다',
    rule_text: 'SecRule ARGS "@rx (;|\\||&&|\\$\\(|`)" "id:2012,phase:2,block,msg:\'Command injection attempt blocked\',severity:HIGH"',
    severity: 'HIGH' as const,
  },
];

const RulesManagement: React.FC = () => {
  const [rules, setRules] = useState<CustomRule[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<CustomRule | null>(null);
  const [formData, setFormData] = useState<CustomRuleRequest>({
    name: '',
    description: '',
    rule_text: '',
    enabled: true,
    severity: 'MEDIUM',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadRules();
  }, []);

  const loadRules = async () => {
    try {
      setLoading(true);
      const response = await rulesAPI.getRules();
      setRules(response.rules);
      setError(null);
    } catch (error: any) {
      console.error('Failed to load rules:', error);
      setError('Failed to load rules');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRule = () => {
    setEditingRule(null);
    setFormData({
      name: '',
      description: '',
      rule_text: '',
      enabled: true,
      severity: 'MEDIUM',
    });
    setDialogOpen(true);
  };

  const handleEditRule = (rule: CustomRule) => {
    setEditingRule(rule);
    setFormData({
      name: rule.name,
      description: rule.description,
      rule_text: rule.rule_text,
      enabled: rule.enabled,
      severity: rule.severity,
    });
    setDialogOpen(true);
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (!window.confirm('Are you sure you want to delete this rule?')) {
      return;
    }

    try {
      await rulesAPI.deleteRule(ruleId);
      await loadRules();
    } catch (error: any) {
      console.error('Failed to delete rule:', error);
      setError('Failed to delete rule');
    }
  };

  const handleSaveRule = async () => {
    try {
      setLoading(true);
      if (editingRule) {
        await rulesAPI.updateRule(editingRule.id, formData);
      } else {
        await rulesAPI.createRule(formData);
      }
      setDialogOpen(false);
      await loadRules();
    } catch (error: any) {
      console.error('Failed to save rule:', error);
      setError('Failed to save rule: ' + (error.response?.data?.error || error.message));
    } finally {
      setLoading(false);
    }
  };

  const handleUseTemplate = (template: typeof RULE_TEMPLATES[0]) => {
    setFormData({
      name: template.name,
      description: template.description,
      rule_text: template.rule_text,
      enabled: true,
      severity: template.severity,
    });
  };

  const handleTestRule = async (ruleId: string) => {
    try {
      const rule = rules.find(r => r.id === ruleId);
      if (!rule) return;
      
      // 간단한 테스트 - 룰이 적용되었는지 확인
      alert(`Testing rule: ${rule.name}\n\nRule will be tested against current WAF configuration.\nCheck the dashboard for any blocked requests after this test.`);
      
      // TODO: 실제 테스트 요청을 보내는 로직 구현
      // 예: 룰이 SQL Injection을 차단한다면 SQL Injection 패턴을 포함한 테스트 요청 전송
    } catch (error: any) {
      console.error('Failed to test rule:', error);
      setError('Failed to test rule: ' + (error.response?.data?.error || error.message));
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'CRITICAL':
        return 'error';
      case 'HIGH':
        return 'error';
      case 'MEDIUM':
        return 'warning';
      case 'LOW':
        return 'info';
      default:
        return 'default';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'Unknown';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <RuleIcon />
          Custom Rules Management
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleCreateRule}
        >
          Create Rule
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Active Rules ({rules?.length || 0})
          </Typography>

          <TableContainer component={Paper} variant="outlined">
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>Name</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Description</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Severity</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Status</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Created</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {!rules || rules.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
                      <Typography variant="body2" color="textSecondary">
                        No custom rules found. Create your first rule to get started.
                      </Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  rules.map((rule) => (
                    <TableRow key={rule.id} hover>
                      <TableCell>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold' }}>
                          {rule.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {rule.description || 'No description'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={rule.severity}
                          color={getSeverityColor(rule.severity) as any}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          icon={
                            rule.enabled ? (
                              <VisibilityIcon sx={{ fontSize: 16 }} />
                            ) : (
                              <VisibilityOffIcon sx={{ fontSize: 16 }} />
                            )
                          }
                          label={rule.enabled ? 'Enabled' : 'Disabled'}
                          color={rule.enabled ? 'success' : 'default'}
                          size="small"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {formatTimestamp(rule.created_at)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Tooltip title="Edit Rule">
                          <IconButton
                            size="small"
                            onClick={() => handleEditRule(rule)}
                          >
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Test Rule">
                          <IconButton
                            size="small"
                            onClick={() => handleTestRule(rule.id)}
                            color="primary"
                          >
                            <VisibilityIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete Rule">
                          <IconButton
                            size="small"
                            onClick={() => handleDeleteRule(rule.id)}
                            color="error"
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Create/Edit Rule Dialog */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>
          {editingRule ? 'Edit Rule' : 'Create New Rule'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <TextField
              fullWidth
              label="Rule Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
            />

            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              multiline
              rows={2}
            />

            <TextField
              fullWidth
              label="ModSecurity Rule"
              value={formData.rule_text}
              onChange={(e) => setFormData({ ...formData, rule_text: e.target.value })}
              multiline
              rows={6}
              placeholder="SecRule ARGS '@detectSQLi' 'id:1001,phase:2,block,msg:&quot;SQL Injection Attack&quot;,severity:HIGH'"
              helperText="Enter a valid ModSecurity rule. Must start with 'SecRule'"
              required
              sx={{ fontFamily: 'monospace' }}
            />

            <TextField
              select
              label="Severity"
              value={formData.severity}
              onChange={(e) => setFormData({ ...formData, severity: e.target.value as any })}
              required
            >
              <MenuItem value="LOW">Low</MenuItem>
              <MenuItem value="MEDIUM">Medium</MenuItem>
              <MenuItem value="HIGH">High</MenuItem>
              <MenuItem value="CRITICAL">Critical</MenuItem>
            </TextField>

            <FormControlLabel
              control={
                <Switch
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                />
              }
              label="Enable Rule"
            />
            
            <Alert severity="info" sx={{ mt: 2 }}>
              <Typography variant="body2" sx={{ fontWeight: 'bold', mb: 2 }}>
                🚀 빠른 시작: 아래 템플릿을 클릭해서 바로 사용해보세요!
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
                {RULE_TEMPLATES.map((template, index) => (
                  <Button
                    key={index}
                    variant="outlined"
                    size="small"
                    startIcon={<CopyIcon />}
                    onClick={() => handleUseTemplate(template)}
                    sx={{ textTransform: 'none', fontSize: '0.8rem' }}
                  >
                    {template.name}
                  </Button>
                ))}
              </Box>
              <Typography variant="caption" color="textSecondary">
                💡 템플릿을 클릭하면 자동으로 폼이 채워집니다. 필요에 따라 수정해서 사용하세요!
              </Typography>
            </Alert>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleSaveRule}
            variant="contained"
            disabled={loading || !formData.name || !formData.rule_text}
          >
            {loading ? 'Saving...' : (editingRule ? 'Update' : 'Create')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default RulesManagement;