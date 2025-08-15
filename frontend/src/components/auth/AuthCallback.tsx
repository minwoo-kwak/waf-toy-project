import React, { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Box, CircularProgress, Typography, Alert } from '@mui/material';
import { useAuth } from '../../contexts/AuthContext';
import { LOCAL_STORAGE_KEYS } from '../../constants';

const AuthCallback: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setAuth } = useAuth();
  const [error, setError] = React.useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        
        if (!code) {
          throw new Error('Authorization code not found');
        }

        console.log('Processing OAuth callback with code:', code?.substring(0, 10) + '...');
        console.log('API Base URL:', process.env.REACT_APP_API_BASE_URL);
        console.log('Full callback URL:', `${process.env.REACT_APP_API_BASE_URL}/api/v1/public/auth/callback`);
        
        // Call backend with the authorization code (use relative path for ingress)
        const response = await fetch(`/api/v1/public/auth/callback`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            code: code,
            state: state || '',
          }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: 'Authentication failed' }));
          throw new Error(errorData.error || 'Authentication failed');
        }

        const data = await response.json();
        
        if (data.token && data.user) {
          // Store token and user info with the correct keys
          localStorage.setItem(LOCAL_STORAGE_KEYS.TOKEN, data.token);
          localStorage.setItem(LOCAL_STORAGE_KEYS.USER, JSON.stringify(data.user));
          
          // Update auth context directly
          setAuth(data.user, data.token);
          
          console.log('Authentication successful, redirecting to dashboard');
          navigate('/dashboard');
        } else {
          throw new Error('Invalid response from server');
        }
      } catch (error) {
        console.error('Auth callback error:', error);
        setError(error instanceof Error ? error.message : 'Authentication failed');
        
        // Redirect to login page after 3 seconds
        setTimeout(() => {
          navigate('/login');
        }, 3000);
      }
    };

    handleCallback();
  }, [searchParams, navigate, setAuth]);

  if (error) {
    return (
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '100vh',
          gap: 2,
        }}
      >
        <Alert severity="error" sx={{ mb: 2 }}>
          <Typography variant="h6">Authentication Error</Typography>
          <Typography>{error}</Typography>
        </Alert>
        <Typography color="text.secondary">
          Redirecting to login page...
        </Typography>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        gap: 2,
      }}
    >
      <CircularProgress size={60} />
      <Typography variant="h6" color="primary">
        Completing Authentication...
      </Typography>
      <Typography color="text.secondary">
        Please wait while we verify your credentials
      </Typography>
    </Box>
  );
};

export default AuthCallback;