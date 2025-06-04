package repository

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"gorm.io/gorm"
)

// AssetRepository 定义资产仓储接口
type AssetRepository interface {
	// Create 创建资产
	Create(ctx context.Context, asset *dto.Asset) error
	// Get 获取资产
	Get(ctx context.Context, id uint) (*dto.Asset, error)
	// List 获取资产列表
	List(ctx context.Context, req *dto.ListAssetsRequest) ([]*dto.Asset, int64, error)
	// Update 更新资产
	Update(ctx context.Context, asset *dto.Asset) error
	// Delete 删除资产
	Delete(ctx context.Context, id uint) error
}

// assetRepository 实现 AssetRepository 接口
type assetRepository struct {
	db *gorm.DB
}

// NewAssetRepository 创建资产仓储实例
func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{
		db: db,
	}
}

// Create 创建资产
func (r *assetRepository) Create(ctx context.Context, asset *dto.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// Get 获取资产
func (r *assetRepository) Get(ctx context.Context, id uint) (*dto.Asset, error) {
	var asset dto.Asset
	err := r.db.WithContext(ctx).First(&asset, id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// List 获取资产列表
func (r *assetRepository) List(ctx context.Context, req *dto.ListAssetsRequest) ([]*dto.Asset, int64, error) {
	var assets []*dto.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&dto.Asset{})

	// TODO: 添加查询条件

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// Update 更新资产
func (r *assetRepository) Update(ctx context.Context, asset *dto.Asset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

// Delete 删除资产
func (r *assetRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&dto.Asset{}, id).Error
}
