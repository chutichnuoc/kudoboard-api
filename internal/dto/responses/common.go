package responses

// APIResponse is a standard response format for all API endpoints
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      *APIError   `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Pagination represents pagination information in the API response
type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse creates a new success response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessResponseWithPagination creates a new success response with pagination
func SuccessResponseWithPagination(data interface{}, pagination *Pagination) APIResponse {
	return APIResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}
}

// ErrorResponse creates a new error response
func ErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}
