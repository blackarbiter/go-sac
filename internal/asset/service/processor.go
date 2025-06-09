package service

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// AssetResponse 资产操作响应
type AssetResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AssetType string `json:"asset_type"`
	Status    string `json:"status"`
}

// AssetProcessor 资产处理器接口
type AssetProcessor interface {
	// Create 创建资产
	Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error)

	// Update 更新资产
	Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error

	// Get 获取资产
	Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error)

	// Delete 删除资产
	Delete(ctx context.Context, id uint) error

	// List 列出资产
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*AssetResponse, int64, error)

	// Validate 验证资产数据
	Validate(base *model.BaseAsset, extension interface{}) error
}

// AssetProcessorFactory 资产处理器工厂接口
type AssetProcessorFactory interface {
	// GetProcessor 获取指定类型的处理器
	GetProcessor(assetType string) (AssetProcessor, error)

	// RegisterProcessor 注册处理器
	RegisterProcessor(assetType string, processor AssetProcessor)
}
