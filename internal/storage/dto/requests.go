package dto

// StorageRequest 表示存储请求
type StorageRequest struct {
	Path     string            `json:"path" binding:"required"`
	Type     string            `json:"type" binding:"required"`
	Metadata map[string]string `json:"metadata"`
}

// StorageQueryParams 存储查询参数
type StorageQueryParams struct {
	Status string `form:"status"`
	Type   string `form:"type"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

// BatchCreateStorageRequest 批量创建存储请求
type BatchCreateStorageRequest struct {
	Storages []StorageRequest `json:"storages" binding:"required,min=1"`
}
