import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Typography,
  Box,
  Skeleton,
  Tooltip,
} from '@mui/material';
import {
  Block as BlockIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { WAFLog } from '../../types/waf';
import { formatDistanceToNow } from 'date-fns';
import { extractRuleIdsFromMessage, getMostSevereAttackType, getAttackTypeFromAnomalyScore, extractAnomalyScore } from '../../utils/crsMapping';

interface LogsTableProps {
  logs: WAFLog[];
  loading: boolean;
}

const LogsTable: React.FC<LogsTableProps> = ({ logs, loading }) => {
  const getAttackTypeColor = (attackType: string) => {
    const colors: Record<string, string> = {
      'SQL Injection': '#fee2e2',
      'Cross-Site Scripting (XSS)': '#fef3c7',
      'Local File Inclusion (LFI)': '#f3e8ff',
      'Remote File Inclusion (RFI)': '#ede9fe',
      'Command Injection': '#fecaca',
      'Path Traversal': '#fed7aa',
      'HTTP Protocol Violation': '#dbeafe',
      'HTTP Protocol Anomaly': '#e0f2fe',
      'Security Policy Violation': '#dcfce7',
      'PHP Injection': '#fdba74',
      'Java Injection': '#e5e7eb',
      'Session Fixation': '#a7f3d0',
      'Remote Code Execution': '#fca5a5',
    };
    return colors[attackType] || '#f3f4f6';
  };

  const getAttackTypeTextColor = (attackType: string) => {
    const colors: Record<string, string> = {
      'SQL Injection': '#dc2626',
      'Cross-Site Scripting (XSS)': '#d97706',
      'Local File Inclusion (LFI)': '#7c3aed',
      'Remote File Inclusion (RFI)': '#8b5cf6',
      'Command Injection': '#ef4444',
      'Path Traversal': '#ea580c',
      'HTTP Protocol Violation': '#2563eb',
      'HTTP Protocol Anomaly': '#0284c7',
      'Security Policy Violation': '#16a34a',
      'PHP Injection': '#ea580c',
      'Java Injection': '#374151',
      'Session Fixation': '#059669',
      'Remote Code Execution': '#dc2626',
    };
    return colors[attackType] || '#6b7280';
  };

  const getSeverityBgColor = (severity: string) => {
    switch (severity?.toLowerCase()) {
      case 'critical': return '#fee2e2';
      case 'high': return '#fef3c7';
      case 'medium': return '#fef3c7';
      case 'low': return '#e0f2fe';
      default: return '#f3f4f6';
    }
  };

  const getSeverityTextColor = (severity: string) => {
    switch (severity?.toLowerCase()) {
      case 'critical': return '#dc2626';
      case 'high': return '#dc2626';
      case 'medium': return '#d97706';
      case 'low': return '#0284c7';
      default: return '#6b7280';
    }
  };

  const getSeverityBorderColor = (severity: string) => {
    switch (severity?.toLowerCase()) {
      case 'critical': return '#fecaca';
      case 'high': return '#fecaca';
      case 'medium': return '#fde68a';
      case 'low': return '#bae6fd';
      default: return '#e5e7eb';
    }
  };
  const getSeverityIcon = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return <ErrorIcon sx={{ fontSize: 16, color: 'error.main' }} />;
      case 'high':
        return <ErrorIcon sx={{ fontSize: 16, color: 'error.main' }} />;
      case 'medium':
        return <WarningIcon sx={{ fontSize: 16, color: 'warning.main' }} />;
      case 'low':
        return <InfoIcon sx={{ fontSize: 16, color: 'info.main' }} />;
      default:
        return <InfoIcon sx={{ fontSize: 16, color: 'grey.500' }} />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'error';
      case 'high':
        return 'error';
      case 'medium':
        return 'warning';
      case 'low':
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

  if (loading) {
    return (
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Time</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>IP</TableCell>
              <TableCell>Attack Type</TableCell>
              <TableCell>Severity</TableCell>
              <TableCell>Message</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {[...Array(5)].map((_, index) => (
              <TableRow key={index}>
                <TableCell><Skeleton width={80} /></TableCell>
                <TableCell><Skeleton width={60} /></TableCell>
                <TableCell><Skeleton width={100} /></TableCell>
                <TableCell><Skeleton width={120} /></TableCell>
                <TableCell><Skeleton width={80} /></TableCell>
                <TableCell><Skeleton width={200} /></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    );
  }

  if (!logs.length) {
    return (
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          py: 4,
        }}
      >
        <Typography variant="h6" color="textSecondary">
          No security events found
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Your WAF is protecting your applications
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer component={Paper} variant="outlined" sx={{ maxHeight: 400 }}>
      <Table size="small" stickyHeader>
        <TableHead sx={{ 
          '& .MuiTableCell-head': {
            backgroundColor: '#f8fafc',
            borderBottom: '2px solid #e2e8f0',
            fontSize: '0.875rem'
          }
        }}>
          <TableRow>
            <TableCell sx={{ fontWeight: 'bold' }}>Time</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Status</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>IP Address</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Attack Type</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Severity</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Message</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {logs.map((log) => (
            <TableRow key={log.id} hover>
              <TableCell>
                <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                  {formatTimestamp(log.timestamp)}
                </Typography>
              </TableCell>
              <TableCell>
                <Chip
                  icon={
                    log.blocked ? (
                      <BlockIcon sx={{ fontSize: 16 }} />
                    ) : (
                      <CheckCircleIcon sx={{ fontSize: 16 }} />
                    )
                  }
                  label={log.blocked ? 'Blocked' : 'Allowed'}
                  color={log.blocked ? 'error' : 'success'}
                  size="small"
                  variant="outlined"
                />
              </TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {log.client_ip}
                </Typography>
              </TableCell>
              <TableCell>
                {(() => {
                  // 개선된 공격 유형 분석: 패턴 매칭 + Anomaly Score
                  const message = log.message || '';
                  const ruleIds = extractRuleIdsFromMessage(message);
                  const anomalyScore = extractAnomalyScore(message);
                  
                  let attackType;
                  
                  if (ruleIds.length > 0 && ruleIds[0] !== 949110) {
                    // 구체적인 CRS 룰 ID가 있으면 사용
                    attackType = getMostSevereAttackType(ruleIds);
                  } else if (anomalyScore > 0) {
                    // Anomaly Score 기반 패턴 매칭 분석
                    attackType = getAttackTypeFromAnomalyScore(message, anomalyScore);
                  } else {
                    // 기본값
                    attackType = { category: log.attack_type || 'Unknown', color: '#636e72', icon: '❓' };
                  }
                  
                  return (
                    <Chip
                      label={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          <span>{attackType.icon}</span>
                          {attackType.category}
                        </Box>
                      }
                      size="small"
                      variant="outlined"
                      sx={{
                        bgcolor: attackType.color + '20',
                        borderColor: attackType.color,
                        color: attackType.color,
                        fontWeight: 600
                      }}
                    />
                  );
                })()}
              </TableCell>
              <TableCell>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  {getSeverityIcon(log.severity)}
                  <Chip
                    label={log.severity || 'Unknown'}
                    size="small"
                    color={getSeverityColor(log.severity) as any}
                    variant="outlined"
                  />
                </Box>
              </TableCell>
              <TableCell>
                <Tooltip title={log.message} arrow>
                  <Typography
                    variant="body2"
                    sx={{
                      maxWidth: 200,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {log.message || 'No message'}
                  </Typography>
                </Tooltip>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default LogsTable;