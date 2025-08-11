import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Box,
  Button,
  Card,
  CardContent,
  Typography,
  Alert,
  CircularProgress,
  Container,
} from '@mui/material';
import { Google as GoogleIcon, Security as SecurityIcon } from '@mui/icons-material';
import { useAuth } from '../../contexts/AuthContext';

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { authState, login, getAuthUrl } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // OAuth 콜백 처리
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const error = searchParams.get('error');

    if (error) {
      setError(`OAuth Error: ${error}`);
      return;
    }

    if (code && state) {
      handleOAuthCallback(code, state);
    }
  }, [searchParams]);

  useEffect(() => {
    // 이미 로그인된 경우 대시보드로 리디렉션
    if (authState.isAuthenticated) {
      navigate('/dashboard');
    }
  }, [authState.isAuthenticated, navigate]);

  const handleOAuthCallback = async (code: string, state: string) => {
    setLoading(true);
    setError(null);

    try {
      await login(code, state);
      navigate('/dashboard');
    } catch (error: any) {
      console.error('OAuth callback failed:', error);
      setError(
        error.response?.data?.error || 
        error.message || 
        'Authentication failed. Please try again.'
      );
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleLogin = async () => {
    setLoading(true);
    setError(null);

    try {
      const authUrl = await getAuthUrl();
      window.location.href = authUrl;
    } catch (error: any) {
      console.error('Failed to get auth URL:', error);
      setError('Failed to initiate login. Please try again.');
      setLoading(false);
    }
  };

  if (authState.loading) {
    return (
      <Container maxWidth="sm">
        <Box
          sx={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          bgcolor: 'background.default',
        }}
      >
        <Card
          sx={{
            width: '100%',
            maxWidth: 400,
            p: 4,
            boxShadow: 3,
          }}
        >
          <CardContent>
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                mb: 3,
              }}
            >
              <SecurityIcon
                sx={{
                  fontSize: 64,
                  color: 'primary.main',
                  mb: 2,
                }}
              />
              <Typography variant="h4" component="h1" gutterBottom align="center">
                WAF SaaS Platform
              </Typography>
              <Typography variant="body1" color="textSecondary" align="center">
                Advanced Web Application Firewall Management Dashboard
              </Typography>
            </Box>

            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}

            <Button
              fullWidth
              variant="contained"
              size="large"
              startIcon={loading ? <CircularProgress size={20} /> : <GoogleIcon />}
              onClick={handleGoogleLogin}
              disabled={loading}
              sx={{
                py: 1.5,
                textTransform: 'none',
                fontSize: '1.1rem',
                bgcolor: '#4285f4',
                '&:hover': {
                  bgcolor: '#357ae8',
                },
              }}
            >
              {loading ? 'Signing in...' : 'Sign in with Google'}
            </Button>

            <Box sx={{ mt: 4, textAlign: 'center' }}>
              <Typography variant="body2" color="textSecondary">
                🛡️ Secure • Real-time • Advanced Protection
              </Typography>
            </Box>

            <Box sx={{ mt: 3 }}>
              <Typography variant="h6" gutterBottom>
                Features:
              </Typography>
              <ul style={{ margin: 0, paddingLeft: '20px' }}>
                <li>
                  <Typography variant="body2" color="textSecondary">
                    Real-time WAF log monitoring
                  </Typography>
                </li>
                <li>
                  <Typography variant="body2" color="textSecondary">
                    Custom security rule management
                  </Typography>
                </li>
                <li>
                  <Typography variant="body2" color="textSecondary">
                    Security testing suite
                  </Typography>
                </li>
                <li>
                  <Typography variant="body2" color="textSecondary">
                    Live attack analytics
                  </Typography>
                </li>
              </ul>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default LoginPage;