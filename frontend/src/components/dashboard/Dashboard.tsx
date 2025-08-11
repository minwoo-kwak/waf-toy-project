import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
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
    <Box sx={{ flexGrow: 1, bgcolor: 'background.default', minHeight: '100vh' }}>
      {/* App Bar */}
      <AppBar position="static" elevation={1}>
        <Toolbar>
          <SecurityIcon sx={{ mr: 2 }} />
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            WAF SaaS Dashboard
          </Typography>
          
          <Chip
            label={wsConnected ? 'Live' : 'Disconnected'}
            color={wsConnected ? 'success' : 'error'}
            size="small"
            sx={{ mr: 2 }}
          />

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

      {/* Main Content */}
      <Box sx={{ p: 3 }}>
        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          {/* Stats Cards */}
          <StatsCards stats={stats} loading={loading} />

          {/* Charts Row */}
          <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
            {/* Attack Type Chart */}
            <Box sx={{ flex: '1 1 400px', minWidth: '400px' }}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Attack Types Distribution
                  </Typography>
                  <AttackChart stats={stats} />
                </CardContent>
              </Card>
            </Box>

            {/* System Info */}
            <Box sx={{ flex: '1 1 400px', minWidth: '400px' }}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    System Information
                  </Typography>
                  <Box sx={{ mt: 2 }}>
                    <Typography variant="body2" paragraph>
                      <strong>WAF Engine:</strong> ModSecurity 3.x
                    </Typography>
                    <Typography variant="body2" paragraph>
                      <strong>Rule Set:</strong> OWASP CRS 4.x
                    </Typography>
                    <Typography variant="body2" paragraph>
                      <strong>Status:</strong>{' '}
                      <Chip
                        label="Active"
                        color="success"
                        size="small"
                      />
                    </Typography>
                    <Typography variant="body2" paragraph>
                      <strong>Live Connections:</strong>{' '}
                      {wsConnected ? '1' : '0'}
                    </Typography>
                  </Box>
                </CardContent>
              </Card>
            </Box>
          </Box>

          {/* Recent Logs */}
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Security Events
              </Typography>
              <LogsTable logs={recentLogs} loading={loading} />
            </CardContent>
          </Card>
        </Box>
      </Box>
    </Box>
  );
};

export default Dashboard;