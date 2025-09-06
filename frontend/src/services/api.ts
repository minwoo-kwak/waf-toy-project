import axios from 'axios';
import { LoginResponse, User } from '../types/auth';
import { WAFLog, WAFStats, CustomRule, CustomRuleRequest, SecurityTest, SecurityTestRequest } from '../types/waf';
import { ErrorResponse } from '../types/errors';
import { API_ENDPOINTS, LOCAL_STORAGE_KEYS, DEFAULT_VALUES } from '../constants';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || '';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: DEFAULT_VALUES.API_TIMEOUT,
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem(LOCAL_STORAGE_KEYS.TOKEN);
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
      localStorage.removeItem(LOCAL_STORAGE_KEYS.TOKEN);
      localStorage.removeItem(LOCAL_STORAGE_KEYS.USER);
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
    const response = await api.get(API_ENDPOINTS.AUTH_URL, {
      params: { state },
    });
    return response.data;
  },

  handleCallback: async (code: string, state: string): Promise<LoginResponse> => {
    const response = await api.post(API_ENDPOINTS.AUTH_CALLBACK, {
      code,
      state,
    });
    return response.data;
  },

  getProfile: async (): Promise<{ user: User }> => {
    const response = await api.get(API_ENDPOINTS.AUTH_PROFILE);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post(API_ENDPOINTS.AUTH_LOGOUT);
  },
};

// WAF API
export const wafAPI = {
  getLogs: async (limit?: number): Promise<{ logs: WAFLog[]; count: number }> => {
    const response = await api.get(API_ENDPOINTS.WAF_LOGS, {
      params: { limit },
    });
    return response.data;
  },

  getStats: async (): Promise<{ stats: WAFStats; websocket_clients: number }> => {
    const response = await api.get(API_ENDPOINTS.WAF_STATS);
    return response.data;
  },

  getDashboard: async (): Promise<{
    user: User;
    stats: WAFStats;
    recent_logs: WAFLog[];
    system_info: any;
  }> => {
    const response = await api.get(API_ENDPOINTS.WAF_DASHBOARD);
    return response.data;
  },
};

// Rules API
export const rulesAPI = {
  createRule: async (rule: CustomRuleRequest): Promise<{ rule: CustomRule }> => {
    const response = await api.post(API_ENDPOINTS.RULES, rule);
    return response.data;
  },

  getRules: async (): Promise<{ rules: CustomRule[]; count: number }> => {
    const response = await api.get(API_ENDPOINTS.RULES);
    return response.data;
  },

  getRule: async (id: string): Promise<{ rule: CustomRule }> => {
    const response = await api.get(`${API_ENDPOINTS.RULES}/${id}`);
    return response.data;
  },

  updateRule: async (id: string, rule: CustomRuleRequest): Promise<{ rule: CustomRule }> => {
    const response = await api.put(`${API_ENDPOINTS.RULES}/${id}`, rule);
    return response.data;
  },

  deleteRule: async (id: string): Promise<void> => {
    await api.delete(`${API_ENDPOINTS.RULES}/${id}`);
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
    const response = await api.post(API_ENDPOINTS.SECURITY_TEST, request);
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
    const response = await api.get(API_ENDPOINTS.SECURITY_TEST_TYPES);
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
    const response = await api.get(API_ENDPOINTS.SECURITY_QUICK_TEST);
    return response.data;
  },
};

export default api;