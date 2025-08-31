import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  AppBar,
  Toolbar,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
  Chip,
  Alert,
  Container,
  Paper,
  Slide,
  Fade,
} from '@mui/material';
import {
  Security as SecurityIcon,
  Logout as LogoutIcon,
  Refresh as RefreshIcon,
  AccountCircle as AccountCircleIcon,
} from '@mui/icons-material';
import { useAuth } from '../../contexts/AuthContext';
import { wafAPI } from '../../services/api';
import { WAFStats, WAFLog } from '../../types/waf';
import webSocketService from '../../services/websocket';
import StatsCards from './StatsCards';
import LogsTable from './LogsTable';
import AttackChart from './AttackChart';
import LiveLogMonitor from './LiveLogMonitor';

const Dashboard: React.FC = () => {
  const { authState, logout } = useAuth();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [stats, setStats] = useState<WAFStats | null>(null);
  const [recentLogs, setRecentLogs] = useState<WAFLog[]>([]);
  const [wsConnected, setWsConnected] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboardData();
    setupWebSocket();

    return () => {
      webSocketService.disconnect();
    };
  }, []);

  const setupWebSocket = () => {
    if (authState.token && !webSocketService.isConnected()) {
      webSocketService.connect(authState.token);
    }

    webSocketService.onWelcome((data) => {
      console.log('WebSocket connected:', data);
      setWsConnected(true);
    });

    webSocketService.onStatsUpdate((stats) => {
      setStats(stats);
    });

    webSocketService.onNewLog((log) => {
      setRecentLogs(prev => [log, ...prev.slice(0, 49)]);
    });

    webSocketService.onStats((stats) => {
      setStats(stats);
    });

    webSocketService.onLogs((logs) => {
      setRecentLogs(logs);
    });

    // 초기 데이터 요청
    setTimeout(() => {
      webSocketService.requestStats();
      webSocketService.requestLogs(20);
    }, 1000);
  };

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      const response = await wafAPI.getDashboard();
      setStats(response.stats);
      setRecentLogs(response.recent_logs);
      setError(null);
    } catch (error: any) {
      console.error('Failed to load dashboard data:', error);
      setError('Failed to load dashboard data. Please try refreshing.');
    } finally {
      setLoading(false);
    }
  };

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = async () => {
    handleMenuClose();
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleRefresh = () => {
    loadDashboardData();
    webSocketService.requestStats();
    webSocketService.requestLogs(20);
  };

  return (
    <Box sx={{ flexGrow: 1, bgcolor: '#f8fafc', minHeight: '100vh' }}>
      {/* Modern App Bar */}
      <AppBar position="static" elevation={0} sx={{ 
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        backdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255,255,255,0.1)'
      }}>
        <Toolbar sx={{ minHeight: '72px !important' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mr: 3 }}>
            <SecurityIcon sx={{ mr: 1.5, fontSize: 28 }} />
            <Typography variant="h5" component="div" sx={{ fontWeight: 700, letterSpacing: -0.5 }}>
              WAF Guardian
            </Typography>
          </Box>
          
          <Box sx={{ flexGrow: 1 }} />
          
          <Slide direction="left" in={true} mountOnEnter unmountOnExit>
            <Chip
              label={wsConnected ? '🟢 Live Monitoring' : '🔴 Disconnected'}
              sx={{ 
                mr: 2,
                bgcolor: wsConnected ? 'rgba(76, 175, 80, 0.9)' : 'rgba(244, 67, 54, 0.9)',
                color: 'white',
                fontWeight: 600,
                '& .MuiChip-label': {
                  px: 2
                }
              }}
            />
          </Slide>

          <IconButton color="inherit" onClick={handleRefresh}>
            <RefreshIcon />
          </IconButton>

          <IconButton
            size="large"
            edge="end"
            aria-label="account of current user"
            aria-controls="account-menu"
            aria-haspopup="true"
            onClick={handleMenuOpen}
            color="inherit"
          >
            {authState.user?.picture ? (
              <Avatar
                src={authState.user.picture}
                alt={authState.user.name}
                sx={{ width: 32, height: 32 }}
              />
            ) : (
              <AccountCircleIcon />
            )}
          </IconButton>
          
          <Menu
            id="account-menu"
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={handleMenuClose}
            onClick={handleMenuClose}
            PaperProps={{
              elevation: 0,
              sx: {
                overflow: 'visible',
                filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                mt: 1.5,
                '& .MuiAvatar-root': {
                  width: 32,
                  height: 32,
                  ml: -0.5,
                  mr: 1,
                },
              },
            }}
            transformOrigin={{ horizontal: 'right', vertical: 'top' }}
            anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
          >
            <MenuItem>
              <Typography variant="subtitle2">
                {authState.user?.name}
              </Typography>
            </MenuItem>
            <MenuItem>
              <Typography variant="body2" color="textSecondary">
                {authState.user?.email}
              </Typography>
            </MenuItem>
            <MenuItem onClick={handleLogout}>
              <LogoutIcon sx={{ mr: 1 }} />
              Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>

      {/* Modern Main Content */}
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        <Fade in={true} timeout={800}>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {/* Stats Cards */}
            <StatsCards stats={stats} loading={loading} />

            {/* Enhanced Charts Row */}
            <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
              {/* Enhanced Attack Type Chart */}
              <Box sx={{ flex: '2 1 600px', minWidth: '600px' }}>
                <Paper
                  elevation={0}
                  sx={{ 
                    p: 3,
                    background: 'linear-gradient(145deg, #ffffff 0%, #f8fafc 100%)',
                    border: '1px solid #e2e8f0',
                    borderRadius: 3,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: '0 12px 32px rgba(0,0,0,0.1)'
                    }
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                    <Box sx={{ 
                      p: 1.5, 
                      borderRadius: 2, 
                      background: 'linear-gradient(135deg, #ff6b6b, #ee5a52)',
                      mr: 2
                    }}>
                      <SecurityIcon sx={{ color: 'white', fontSize: 24 }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, color: '#1e293b' }}>
                      Threat Intelligence Dashboard
                    </Typography>
                  </Box>
                  <AttackChart stats={stats} logs={recentLogs} />
                </Paper>
              </Box>

              {/* Enhanced System Info */}
              <Box sx={{ flex: '1 1 400px', minWidth: '400px' }}>
                <Paper
                  elevation={0}
                  sx={{ 
                    p: 3,
                    height: '100%',
                    background: 'linear-gradient(145deg, #ffffff 0%, #f1f5f9 100%)',
                    border: '1px solid #e2e8f0',
                    borderRadius: 3,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: '0 12px 32px rgba(0,0,0,0.1)'
                    }
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                    <Box sx={{ 
                      p: 1.5, 
                      borderRadius: 2, 
                      background: 'linear-gradient(135deg, #4ecdc4, #44a08d)',
                      mr: 2
                    }}>
                      <SecurityIcon sx={{ color: 'white', fontSize: 24 }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, color: '#1e293b' }}>
                      System Status
                    </Typography>
                  </Box>
                  
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: '#64748b' }}>
                        WAF Engine
                      </Typography>
                      <Chip
                        label="ModSecurity 3.x"
                        size="small"
                        sx={{ bgcolor: '#e0f2fe', color: '#0277bd', fontWeight: 600 }}
                      />
                    </Box>
                    
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: '#64748b' }}>
                        Rule Set
                      </Typography>
                      <Chip
                        label="OWASP CRS 4.x"
                        size="small"
                        sx={{ bgcolor: '#f3e5f5', color: '#7b1fa2', fontWeight: 600 }}
                      />
                    </Box>
                    
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: '#64748b' }}>
                        Protection Status
                      </Typography>
                      <Chip
                        label="🛡️ Active"
                        size="small"
                        sx={{ bgcolor: '#e8f5e8', color: '#2e7d32', fontWeight: 600 }}
                      />
                    </Box>
                    
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600, color: '#64748b' }}>
                        Live Monitoring
                      </Typography>
                      <Chip
                        label={wsConnected ? '🟢 Connected' : '🔴 Offline'}
                        size="small"
                        sx={{ 
                          bgcolor: wsConnected ? '#e8f5e8' : '#ffebee', 
                          color: wsConnected ? '#2e7d32' : '#c62828', 
                          fontWeight: 600 
                        }}
                      />
                    </Box>
                  </Box>
                </Paper>
              </Box>
            </Box>

            {/* Live Security Monitor - Full Width */}
            <LiveLogMonitor logs={recentLogs} isConnected={wsConnected} />
          </Box>
        </Fade>
      </Container>
    </Box>
  );
};

export default Dashboard;