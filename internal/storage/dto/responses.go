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

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
