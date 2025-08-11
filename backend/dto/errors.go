package dto

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Authentication errors
	ErrNoAuthHeader       ErrorCode = "ERR_NO_AUTH_HEADER"
	ErrInvalidAuthFormat  ErrorCode = "ERR_INVALID_AUTH_FORMAT"
	ErrInvalidToken       ErrorCode = "ERR_INVALID_TOKEN"
	ErrAuthFailed         ErrorCode = "ERR_AUTH_FAILED"
	
	// Request errors
	ErrInvalidRequest     ErrorCode = "ERR_INVALID_REQUEST"
	ErrValidationFailed   ErrorCode = "ERR_VALIDATION_FAILED"
	ErrResourceNotFound   ErrorCode = "ERR_RESOURCE_NOT_FOUND"
	
	// Server errors
	ErrInternal           ErrorCode = "ERR_INTERNAL"
	ErrServiceUnavailable ErrorCode = "ERR_SERVICE_UNAVAILABLE"
	ErrDatabaseError      ErrorCode = "ERR_DATABASE_ERROR"
	
	// WAF specific errors
	ErrWAFLogParsing      ErrorCode = "ERR_WAF_LOG_PARSING"
	ErrRuleValidation     ErrorCode = "ERR_RULE_VALIDATION"
	ErrSecurityTestFailed ErrorCode = "ERR_SECURITY_TEST_FAILED"
)

// ErrorResponse represents standardized error response format
type ErrorResponse struct {
	Error   string    `json:"error"`
	Code    ErrorCode `json:"code"`
	Details *string   `json:"details,omitempty"`
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(message string, code ErrorCode, details ...string) *ErrorResponse {
	resp := &ErrorResponse{
		Error: message,
		Code:  code,
	}
	
	if len(details) > 0 && details[0] != "" {
		resp.Details = &details[0]
	}
	
	return resp
}

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Code    ErrorCode         `json:"code"`
	Errors  []ValidationError `json:"errors"`
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Error:  "Validation failed",
		Code:   ErrValidationFailed,
		Errors: errors,
	}
}