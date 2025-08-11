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
import Dashboard from './components/dashboard/Dashboard';
import SecurityTests from './components/security/SecurityTests';
import RulesManagement from './components/rules/RulesManagement';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
    background: {
      default: '#f5f5f5',
    },
  },
  typography: {
    h4: {
      fontWeight: 600,
    },
    h6: {
      fontWeight: 600,
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