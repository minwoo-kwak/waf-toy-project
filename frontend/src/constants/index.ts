// Application constants for better maintainability

export const SEVERITY_LEVELS = {
  LOW: 'LOW',
  MEDIUM: 'MEDIUM', 
  HIGH: 'HIGH',
  CRITICAL: 'CRITICAL'
} as const;

export const TEST_TYPES = {
  SQL_INJECTION: 'sql_injection',
  XSS: 'xss',
  PATH_TRAVERSAL: 'path_traversal',
  COMMAND_INJECTION: 'command_injection'
} as const;

export const API_ENDPOINTS = {
  AUTH_URL: '/api/v1/public/auth/url',
  AUTH_CALLBACK: '/api/v1/public/auth/callback',
  AUTH_PROFILE: '/api/v1/auth/profile',
  AUTH_LOGOUT: '/api/v1/auth/logout',
  
  WAF_LOGS: '/api/v1/waf/logs',
  WAF_STATS: '/api/v1/waf/stats', 
  WAF_DASHBOARD: '/api/v1/waf/dashboard',
  
  RULES: '/api/v1/rules',
  
  SECURITY_TEST: '/api/v1/security/test',
  SECURITY_TEST_TYPES: '/api/v1/security/test-types',
  SECURITY_QUICK_TEST: '/api/v1/security/quick-test'
} as const;

export const WEBSOCKET_MESSAGE_TYPES = {
  WELCOME: 'welcome',
  NEW_LOG: 'new_log',
  STATS_UPDATE: 'stats_update',
  STATS: 'stats',
  LOGS: 'logs',
  GET_LOGS: 'get_logs',
  GET_STATS: 'get_stats'
} as const;

export const LOCAL_STORAGE_KEYS = {
  TOKEN: 'waf_token',
  USER: 'waf_user'
} as const;

export const DEFAULT_VALUES = {
  LOG_LIMIT: 50,
  MAX_RECONNECT_ATTEMPTS: 5,
  RECONNECT_DELAY: 1000,
  API_TIMEOUT: 10000
} as const;

export type SeverityLevel = typeof SEVERITY_LEVELS[keyof typeof SEVERITY_LEVELS];
export type TestType = typeof TEST_TYPES[keyof typeof TEST_TYPES];