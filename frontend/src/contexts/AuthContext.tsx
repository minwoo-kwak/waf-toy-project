import React, { createContext, useContext, useReducer, useEffect } from 'react';
import { AuthState, AuthContextType, User } from '../types/auth';
import { authAPI } from '../services/api';
import webSocketService from '../services/websocket';

interface AuthAction {
  type: 'LOGIN_START' | 'LOGIN_SUCCESS' | 'LOGIN_FAILURE' | 'LOGOUT' | 'RESTORE_SESSION';
  payload?: any;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
  loading: true,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN_START':
      return {
        ...state,
        loading: true,
      };
    case 'LOGIN_SUCCESS':
      return {
        ...state,
        isAuthenticated: true,
        user: action.payload.user,
        token: action.payload.token,
        loading: false,
      };
    case 'LOGIN_FAILURE':
      return {
        ...state,
        isAuthenticated: false,
        user: null,
        token: null,
        loading: false,
      };
    case 'LOGOUT':
      return {
        ...initialState,
        loading: false,
      };
    case 'RESTORE_SESSION':
      return {
        ...state,
        isAuthenticated: true,
        user: action.payload.user,
        token: action.payload.token,
        loading: false,
      };
    default:
      return state;
  }
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [authState, dispatch] = useReducer(authReducer, initialState);

  useEffect(() => {
    // 페이지 로드시 저장된 토큰으로 세션 복원 시도
    const restoreSession = async () => {
      const token = localStorage.getItem('waf_token');
      const userData = localStorage.getItem('waf_user');

      if (token && userData) {
        try {
          const user: User = JSON.parse(userData);
          dispatch({
            type: 'RESTORE_SESSION',
            payload: { token, user },
          });

          // WebSocket 연결
          webSocketService.connect(token);
        } catch (error) {
          console.error('Session restoration failed:', error);
          localStorage.removeItem('waf_token');
          localStorage.removeItem('waf_user');
          dispatch({ type: 'LOGIN_FAILURE' });
        }
      } else {
        dispatch({ type: 'LOGIN_FAILURE' });
      }
    };

    restoreSession();
  }, []);

  const login = async (code: string, state: string): Promise<void> => {
    dispatch({ type: 'LOGIN_START' });

    try {
      const response = await authAPI.handleCallback(code, state);
      
      // 토큰과 사용자 정보 저장
      localStorage.setItem('waf_token', response.token);
      localStorage.setItem('waf_user', JSON.stringify(response.user));

      dispatch({
        type: 'LOGIN_SUCCESS',
        payload: {
          token: response.token,
          user: response.user,
        },
      });

      // WebSocket 연결
      webSocketService.connect(response.token);
    } catch (error) {
      console.error('Login failed:', error);
      dispatch({ type: 'LOGIN_FAILURE' });
      throw error;
    }
  };

  const logout = async (): Promise<void> => {
    try {
      await authAPI.logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      // 로컬 스토리지 정리
      localStorage.removeItem('waf_token');
      localStorage.removeItem('waf_user');
      
      // WebSocket 연결 해제
      webSocketService.disconnect();
      
      dispatch({ type: 'LOGOUT' });
    }
  };

  const getAuthUrl = async (): Promise<string> => {
    const response = await authAPI.getAuthUrl();
    return response.auth_url;
  };

  const value: AuthContextType = {
    authState,
    login,
    logout,
    getAuthUrl,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export default AuthContext;