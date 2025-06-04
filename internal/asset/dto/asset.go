package dto

import (
	"time"
)

// Asset 资产实体
type Asset struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Type      string    `json:"type" gorm:"size:50;not null"`
	Status    string    `json:"status" gorm:"size:20;not null"`
	Value     float64   `json:"value" gorm:"type:decimal(20,2)"`
	Location  string    `json:"location" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAssetRequest 创建资产请求
type CreateAssetRequest struct {
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	Status   string  `json:"status" binding:"required"`
	Value    float64 `json:"value"`
	Location string  `json:"location"`
}

// UpdateAssetRequest 更新资产请求
type UpdateAssetRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	Value    float64 `json:"value"`
	Location string  `json:"location"`
}

// AssetResponse 资产响应
type AssetResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Value     float64   `json:"value"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListAssetsRequest 获取资产列表请求
type ListAssetsRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	Name     string `form:"name"`
	Type     string `form:"type"`
	Status   string `form:"status"`
}

// ListAssetsResponse 获取资产列表响应
type ListAssetsResponse struct {
	Total  int64           `json:"total"`
	Assets []AssetResponse `json:"assets"`
}
