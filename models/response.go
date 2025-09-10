package models

// StandardResponse представляет стандартный ответ API
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorInfo `json:"error"`
}

// ErrorInfo содержит детали ошибки
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationInfo содержит информацию о пагинации
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// PaginatedResponse представляет ответ с пагинацией
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Message    string         `json:"message,omitempty"`
}

// AuthResponse представляет ответ аутентификации
type AuthResponse struct {
	User         UserResponse `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refreshToken"`
}

// Константы кодов ошибок
const (
	ErrAuthRequired       = "AUTH_REQUIRED"
	ErrAuthInvalid        = "AUTH_INVALID"
	ErrValidationError    = "VALIDATION_ERROR"
	ErrNotFound           = "NOT_FOUND"
	ErrForbidden          = "FORBIDDEN"
	ErrConflict           = "CONFLICT"
	ErrRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrInternalError      = "INTERNAL_ERROR"
	ErrUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrInsufficientStock  = "INSUFFICIENT_STOCK"
	ErrProductNotFound    = "PRODUCT_NOT_FOUND"
	ErrCartItemNotFound   = "CART_ITEM_NOT_FOUND"
	ErrOrderNotFound      = "ORDER_NOT_FOUND"
)

// SuccessResponse создает успешный ответ
func SuccessResponse(data interface{}, message ...string) StandardResponse {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}
	if len(message) > 0 {
		response.Message = message[0]
	}
	return response
}

// ErrorResponseWithCode создает ответ с ошибкой
func ErrorResponseWithCode(code, message string, details ...interface{}) ErrorResponse {
	errorInfo := ErrorInfo{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		errorInfo.Details = details[0]
	}
	return ErrorResponse{
		Success: false,
		Error:   errorInfo,
	}
}

// PaginatedSuccessResponse создает успешный ответ с пагинацией
func PaginatedSuccessResponse(data interface{}, pagination PaginationInfo, message ...string) PaginatedResponse {
	response := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}
	if len(message) > 0 {
		response.Message = message[0]
	}
	return response
}
