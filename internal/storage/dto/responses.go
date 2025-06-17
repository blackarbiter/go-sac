package dto

// StorageResponse 表示存储响应
type StorageResponse struct {
	ID        string            `json:"id"`
	Path      string            `json:"path"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// StorageListResponse 表示存储列表响应
type StorageListResponse struct {
	Total int64             `json:"total"`
	Items []StorageResponse `json:"items"`
}

// BaseResponse represents the common response structure
type BaseResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) *BaseResponse {
	return &BaseResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}
