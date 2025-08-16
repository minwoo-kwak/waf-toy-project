import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box, Drawer, List, ListItem, ListItemIcon, ListItemText, Toolbar, AppBar, Typography } from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Security as SecurityIcon,
  Rule as RuleIcon,
  BugReport as BugReportIcon,
} from '@mui/icons-material';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import LoginPage from './components/auth/LoginPage';
import AuthCallback from './components/auth/AuthCallback';
import Dashboard from './components/dashboard/Dashboard';
import SecurityTests from './components/security/SecurityTests';
import RulesManagement from './components/rules/RulesManagement';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#667eea',
      dark: '#764ba2',
      light: '#a8b5ff',
    },
    secondary: {
      main: '#10b981',
      dark: '#059669',
      light: '#34d399',
    },
    error: {
      main: '#dc2626',
      light: '#ef4444',
      dark: '#b91c1c',
    },
    warning: {
      main: '#d97706',
      light: '#f59e0b',
      dark: '#b45309',
    },
    success: {
      main: '#10b981',
      light: '#34d399',
      dark: '#059669',
    },
    background: {
      default: '#f8fafc',
      paper: '#ffffff',
    },
    text: {
      primary: '#1e293b',
      secondary: '#64748b',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h4: {
      fontWeight: 700,
      letterSpacing: '-0.025em',
    },
    h5: {
      fontWeight: 700,
      letterSpacing: '-0.025em',
    },
    h6: {
      fontWeight: 600,
      letterSpacing: '-0.025em',
    },
    button: {
      fontWeight: 600,
      textTransform: 'none',
    },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
          borderRadius: 12,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 600,
          fontSize: '0.8rem',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          padding: '8px 24px',
        },
      },
    },
  },
});

const drawerWidth = 240;

const navigationItems = [
  { text: 'Dashboard', icon: <DashboardIcon />, path: '/dashboard' },
  { text: 'Security Tests', icon: <BugReportIcon />, path: '/security' },
  { text: 'Custom Rules', icon: <RuleIcon />, path: '/rules' },
];

interface MainLayoutProps {
  children: React.ReactNode;
  currentPath: string;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children, currentPath }) => {
  return (
    <Box sx={{ display: 'flex' }}>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          [`& .MuiDrawer-paper`]: {
            width: drawerWidth,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar>
          <SecurityIcon sx={{ mr: 1, color: 'primary.main' }} />
          <Typography variant="h6" noWrap component="div" sx={{ fontWeight: 'bold' }}>
            WAF SaaS
          </Typography>
        </Toolbar>
        <Box sx={{ overflow: 'auto' }}>
          <List>
            {navigationItems.map((item) => (
              <ListItem
                key={item.text}
                component="a"
                href={item.path}
                sx={{
                  cursor: 'pointer',
                  textDecoration: 'none',
                  backgroundColor: currentPath === item.path ? 'primary.main' : 'transparent',
                  color: currentPath === item.path ? 'white' : 'inherit',
                  '&:hover': {
                    backgroundColor: currentPath === item.path ? 'primary.dark' : 'action.hover',
                  },
                  '& .MuiListItemIcon-root': {
                    color: currentPath === item.path ? 'white' : 'inherit',
                  },
                }}
              >
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.text} />
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, width: `calc(100% - ${drawerWidth}px)` }}>
        {children}
      </Box>
    </Box>
  );
};

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { authState } = useAuth();

  if (authState.loading) {
    return <div>Loading...</div>;
  }

  if (!authState.isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

const AppContent: React.FC = () => {
  const { authState } = useAuth();
  const currentPath = window.location.pathname;

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/auth/callback" element={<AuthCallback />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <MainLayout currentPath={currentPath}>
              <Dashboard />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/security"
        element={
          <ProtectedRoute>
            <MainLayout currentPath={currentPath}>
              <SecurityTests />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/rules"
        element={
          <ProtectedRoute>
            <MainLayout currentPath={currentPath}>
              <RulesManagement />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/"
        element={
          authState.isAuthenticated ? (
            <Navigate to="/dashboard" replace />
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
    </Routes>
  );
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
        <Router>
          <AppContent />
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;