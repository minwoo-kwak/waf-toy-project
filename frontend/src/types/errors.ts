export type ErrorCode = 
  // Authentication errors
  | 'ERR_NO_AUTH_HEADER'
  | 'ERR_INVALID_AUTH_FORMAT'
  | 'ERR_INVALID_TOKEN'
  | 'ERR_AUTH_FAILED'
  // Request errors
  | 'ERR_INVALID_REQUEST'
  | 'ERR_VALIDATION_FAILED'
  | 'ERR_RESOURCE_NOT_FOUND'
  // Server errors
  | 'ERR_INTERNAL'
  | 'ERR_SERVICE_UNAVAILABLE'
  | 'ERR_DATABASE_ERROR'
  // WAF specific errors
  | 'ERR_WAF_LOG_PARSING'
  | 'ERR_RULE_VALIDATION'
  | 'ERR_SECURITY_TEST_FAILED';

export interface ErrorResponse {
  error: string;
  code: ErrorCode;
  details?: string;
}

export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationErrorResponse {
  error: string;
  code: ErrorCode;
  errors: ValidationError[];
}