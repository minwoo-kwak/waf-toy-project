import axios from 'axios';
import { LoginResponse, User } from '../types/auth';
import { WAFLog, WAFStats, CustomRule, CustomRuleRequest, SecurityTest, SecurityTestRequest } from '../types/waf';
import { ErrorResponse } from '../types/errors';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('waf_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('waf_token');
      localStorage.removeItem('waf_user');
      window.location.href = '/login';
    }
    
    // Extract standardized error response
    const errorResponse = error.response?.data as ErrorResponse;
    if (errorResponse?.code) {
      error.standardizedError = errorResponse;
    }
    
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  getAuthUrl: async (state?: string): Promise<{ auth_url: string; state: string }> => {
    const response = await api.get('/api/v1/public/auth/url', {
      params: { state },
    });
    return response.data;
  },

  handleCallback: async (code: string, state: string): Promise<LoginResponse> => {
    const response = await api.post('/api/v1/public/auth/callback', {
      code,
      state,
    });
    return response.data;
  },

  getProfile: async (): Promise<{ user: User }> => {
    const response = await api.get('/api/v1/auth/profile');
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post('/api/v1/auth/logout');
  },
};

// WAF API
export const wafAPI = {
  getLogs: async (limit?: number): Promise<{ logs: WAFLog[]; count: number }> => {
    const response = await api.get('/api/v1/waf/logs', {
      params: { limit },
    });
    return response.data;
  },

  getStats: async (): Promise<{ stats: WAFStats; websocket_clients: number }> => {
    const response = await api.get('/api/v1/waf/stats');
    return response.data;
  },

  getDashboard: async (): Promise<{
    user: User;
    stats: WAFStats;
    recent_logs: WAFLog[];
    system_info: any;
  }> => {
    const response = await api.get('/api/v1/waf/dashboard');
    return response.data;
  },
};

// Rules API
export const rulesAPI = {
  createRule: async (rule: CustomRuleRequest): Promise<{ rule: CustomRule }> => {
    const response = await api.post('/api/v1/rules', rule);
    return response.data;
  },

  getRules: async (): Promise<{ rules: CustomRule[]; count: number }> => {
    const response = await api.get('/api/v1/rules');
    return response.data;
  },

  getRule: async (id: string): Promise<{ rule: CustomRule }> => {
    const response = await api.get(`/api/v1/rules/${id}`);
    return response.data;
  },

  updateRule: async (id: string, rule: CustomRuleRequest): Promise<{ rule: CustomRule }> => {
    const response = await api.put(`/api/v1/rules/${id}`, rule);
    return response.data;
  },

  deleteRule: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/rules/${id}`);
  },
};

// Security Testing API
export const securityAPI = {
  runSecurityTest: async (request: SecurityTestRequest): Promise<{
    test: SecurityTest;
    summary: {
      total_tests: number;
      blocked_tests: number;
      passed_tests: number;
      block_rate: number;
      effectiveness: string;
    };
  }> => {
    const response = await api.post('/api/v1/security/test', request);
    return response.data;
  },

  getTestTypes: async (): Promise<{
    test_types: Array<{
      id: string;
      name: string;
      description: string;
      severity: string;
    }>;
    count: number;
  }> => {
    const response = await api.get('/api/v1/security/test-types');
    return response.data;
  },

  getQuickTests: async (): Promise<{
    quick_tests: Array<{
      test_type: string;
      total_tests: number;
      blocked_tests: number;
      effectiveness: string;
      results: any[];
    }>;
    summary: any;
  }> => {
    const response = await api.get('/api/v1/security/quick-test');
    return response.data;
  },
};

export default api;