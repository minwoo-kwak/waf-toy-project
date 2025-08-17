import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  Paper,
  Typography,
  Chip,
  IconButton,
  Tooltip,
  Switch,
  FormControlLabel,
  Slide,
  Alert,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Avatar,
  Badge,
  Divider,
} from '@mui/material';
import {
  PlayArrow as PlayIcon,
  Pause as PauseIcon,
  Clear as ClearIcon,
  FilterList as FilterIcon,
  Security as SecurityIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
} from '@mui/icons-material';
import { WAFLog } from '../../types/waf';
import { formatDistanceToNow } from 'date-fns';

interface LiveLogMonitorProps {
  logs: WAFLog[];
  isConnected: boolean;
}

const LiveLogMonitor: React.FC<LiveLogMonitorProps> = ({ logs, isConnected }) => {
  const [isMonitoring, setIsMonitoring] = useState(true);
  const [showOnlyBlocked, setShowOnlyBlocked] = useState(false);
  const [displayLogs, setDisplayLogs] = useState<WAFLog[]>([]);
  const [newLogCount, setNewLogCount] = useState(0);
  const logEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isMonitoring) {
      const filteredLogs = showOnlyBlocked 
        ? logs.filter(log => log.blocked)
        : logs;
      
      const newLogs = filteredLogs.slice(-20); // Show only last 20 logs
      setDisplayLogs(newLogs);
      
      // Count new logs since last update
      if (newLogs.length > displayLogs.length) {
        setNewLogCount(prev => prev + (newLogs.length - displayLogs.length));
      }
    }
  }, [logs, isMonitoring, showOnlyBlocked]);

  useEffect(() => {
    if (isMonitoring && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [displayLogs, isMonitoring]);

  const handleClearLogs = () => {
    setDisplayLogs([]);
    setNewLogCount(0);
  };

  const handleToggleMonitoring = () => {
    setIsMonitoring(!isMonitoring);
    if (!isMonitoring) {
      setNewLogCount(0);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity?.toLowerCase()) {
      case 'critical': return '#dc2626';
      case 'high': return '#ea580c';
      case 'medium': return '#d97706';
      case 'low': return '#059669';
      default: return '#64748b';
    }
  };

  const getAttackTypeColor = (attackType: string) => {
    const colors: Record<string, string> = {
      'SQL Injection': '#dc2626',
      'Cross-Site Scripting (XSS)': '#d97706',
      'Command Injection': '#dc2626',
      'Path Traversal': '#ea580c',
      'Local File Inclusion (LFI)': '#7c3aed',
      'Remote File Inclusion (RFI)': '#8b5cf6',
      'PHP Injection': '#ea580c',
      'Java Injection': '#374151',
      'Session Fixation': '#059669',
      'HTTP Protocol Violation': '#2563eb',
      'HTTP Protocol Anomaly': '#0284c7',
      'Security Policy Violation': '#16a34a',
    };
    return colors[attackType] || '#64748b';
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleTimeString();
    } catch {
      return 'Unknown';
    }
  };

  return (
    <Paper
      elevation={0}
      sx={{
        height: '600px',
        display: 'flex',
        flexDirection: 'column',
        background: 'linear-gradient(145deg, #1a1a2e 0%, #16213e 100%)',
        border: '1px solid #2d3748',
        borderRadius: 3,
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      {/* Header */}
      <Box
        sx={{
          p: 2,
          background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
          borderBottom: '1px solid #334155',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Box sx={{
            p: 1,
            borderRadius: 2,
            background: 'linear-gradient(135deg, #10b981, #059669)',
          }}>
            <SecurityIcon sx={{ color: 'white', fontSize: 20 }} />
          </Box>
          <Typography variant="h6" sx={{ color: 'white', fontWeight: 700 }}>
            Live Security Monitor
          </Typography>
          <Badge badgeContent={newLogCount} color="error" max={99}>
            <Chip
              label={isConnected ? '🟢 Connected' : '🔴 Offline'}
              size="small"
              sx={{
                bgcolor: isConnected ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)',
                color: isConnected ? '#10b981' : '#ef4444',
                border: `1px solid ${isConnected ? '#10b981' : '#ef4444'}`,
                fontWeight: 600,
              }}
            />
          </Badge>
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <FormControlLabel
            control={
              <Switch
                checked={showOnlyBlocked}
                onChange={(e) => setShowOnlyBlocked(e.target.checked)}
                size="small"
                sx={{
                  '& .MuiSwitch-thumb': { bgcolor: '#10b981' },
                  '& .MuiSwitch-track': { bgcolor: '#374151' },
                }}
              />
            }
            label="Threats Only"
            sx={{ color: '#94a3b8', fontSize: '0.875rem' }}
          />
          
          <Tooltip title={isMonitoring ? 'Pause monitoring' : 'Resume monitoring'}>
            <IconButton
              onClick={handleToggleMonitoring}
              sx={{ color: isMonitoring ? '#ef4444' : '#10b981' }}
            >
              {isMonitoring ? <PauseIcon /> : <PlayIcon />}
            </IconButton>
          </Tooltip>

          <Tooltip title="Clear logs">
            <IconButton
              onClick={handleClearLogs}
              sx={{ color: '#64748b' }}
            >
              <ClearIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Monitoring Status */}
      {!isMonitoring && (
        <Alert
          severity="warning"
          sx={{
            m: 2,
            bgcolor: 'rgba(245, 158, 11, 0.1)',
            color: '#f59e0b',
            border: '1px solid rgba(245, 158, 11, 0.3)',
          }}
        >
          Live monitoring is paused. Click play to resume.
        </Alert>
      )}

      {/* Logs Container */}
      <Box
        ref={containerRef}
        sx={{
          flex: 1,
          overflow: 'auto',
          bgcolor: '#0f172a',
          position: 'relative',
          '&::-webkit-scrollbar': {
            width: '8px',
          },
          '&::-webkit-scrollbar-track': {
            bgcolor: '#1e293b',
          },
          '&::-webkit-scrollbar-thumb': {
            bgcolor: '#475569',
            borderRadius: '4px',
          },
        }}
      >
        {displayLogs.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              color: '#64748b',
            }}
          >
            <SecurityIcon sx={{ fontSize: 48, mb: 2, opacity: 0.5 }} />
            <Typography variant="h6" sx={{ mb: 1 }}>
              No security events
            </Typography>
            <Typography variant="body2">
              Your applications are protected
            </Typography>
          </Box>
        ) : (
          <List sx={{ p: 0 }}>
            {displayLogs.map((log, index) => (
              <Slide
                key={log.id}
                direction="up"
                in={true}
                timeout={300}
                style={{ transitionDelay: `${index * 50}ms` }}
              >
                <ListItem
                  sx={{
                    py: 1.5,
                    px: 2,
                    borderBottom: '1px solid #1e293b',
                    '&:hover': {
                      bgcolor: 'rgba(51, 65, 85, 0.3)',
                    },
                    animation: index === displayLogs.length - 1 ? 'pulse 1s' : 'none',
                    '@keyframes pulse': {
                      '0%': { bgcolor: 'rgba(16, 185, 129, 0.2)' },
                      '100%': { bgcolor: 'transparent' },
                    },
                  }}
                >
                  <ListItemAvatar>
                    <Avatar
                      sx={{
                        bgcolor: log.blocked ? '#dc2626' : '#10b981',
                        width: 36,
                        height: 36,
                      }}
                    >
                      {log.blocked ? (
                        <ErrorIcon sx={{ fontSize: 20 }} />
                      ) : (
                        <CheckCircleIcon sx={{ fontSize: 20 }} />
                      )}
                    </Avatar>
                  </ListItemAvatar>

                  <ListItemText
                    primary={
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                        <Typography
                          variant="body2"
                          sx={{ color: 'white', fontWeight: 600, fontFamily: 'monospace' }}
                        >
                          {log.client_ip}
                        </Typography>
                        <Chip
                          label={log.method}
                          size="small"
                          sx={{
                            bgcolor: 'rgba(59, 130, 246, 0.2)',
                            color: '#3b82f6',
                            fontSize: '0.7rem',
                            height: '20px',
                          }}
                        />
                        <Typography
                          variant="caption"
                          sx={{ color: '#64748b', fontFamily: 'monospace' }}
                        >
                          {formatTimestamp(log.timestamp)}
                        </Typography>
                      </Box>
                    }
                    secondary={
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                          <Chip
                            label={log.attack_type || 'Normal'}
                            size="small"
                            sx={{
                              bgcolor: `${getAttackTypeColor(log.attack_type)}20`,
                              color: getAttackTypeColor(log.attack_type),
                              fontSize: '0.7rem',
                              height: '22px',
                              fontWeight: 600,
                            }}
                          />
                          {log.severity && (
                            <Chip
                              label={log.severity}
                              size="small"
                              sx={{
                                bgcolor: `${getSeverityColor(log.severity)}20`,
                                color: getSeverityColor(log.severity),
                                fontSize: '0.7rem',
                                height: '22px',
                                fontWeight: 600,
                              }}
                            />
                          )}
                          {log.rule_id && (
                            <Chip
                              label={`Rule: ${log.rule_id}`}
                              size="small"
                              sx={{
                                bgcolor: 'rgba(99, 102, 241, 0.1)',
                                color: '#6366f1',
                                fontSize: '0.65rem',
                                height: '20px',
                                fontWeight: 500,
                              }}
                            />
                          )}
                        </Box>
                        
                        {/* URL 정보 표시 */}
                        {log.url && log.url !== '/api/v1/ping' && (
                          <Typography
                            variant="caption"
                            sx={{
                              color: '#ef4444',
                              fontSize: '0.75rem',
                              fontFamily: 'monospace',
                              backgroundColor: 'rgba(239, 68, 68, 0.1)',
                              padding: '2px 6px',
                              borderRadius: '4px',
                              fontWeight: 600,
                            }}
                          >
                            🎯 {log.url}
                          </Typography>
                        )}
                        
                        <Typography
                          variant="caption"
                          sx={{
                            color: '#94a3b8',
                            fontSize: '0.75rem',
                            display: '-webkit-box',
                            WebkitBoxOrient: 'vertical',
                            WebkitLineClamp: 2,
                            overflow: 'hidden',
                          }}
                        >
                          {log.message || 'No message'}
                        </Typography>
                      </Box>
                    }
                  />
                </ListItem>
              </Slide>
            ))}
            <div ref={logEndRef} />
          </List>
        )}
      </Box>
    </Paper>
  );
};

export default LiveLogMonitor;