package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// BaseProcessor 基础处理器实现
type BaseProcessor struct {
	repo repository.Repository
}

// NewBaseProcessor 创建基础处理器
func NewBaseProcessor(repo repository.Repository) *BaseProcessor {
	return &BaseProcessor{
		repo: repo,
	}
}

// Create 创建基础资产
func (p *BaseProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 创建资产
	if err := p.repo.CreateBase(ctx, base); err != nil {
		return nil, fmt.Errorf("failed to create base asset: %w", err)
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新基础资产
func (p *BaseProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 更新资产
	return p.repo.UpdateBase(ctx, base)
}

// Get 获取基础资产
func (p *BaseProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, err := p.repo.GetBase(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, nil, nil
}

// Delete 删除基础资产
func (p *BaseProcessor) Delete(ctx context.Context, id uint) error {
	return p.repo.DeleteBase(ctx, id)
}

// List 列出基础资产
func (p *BaseProcessor) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*AssetResponse, int64, error) {
	bases, total, err := p.repo.ListBase(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*AssetResponse, len(bases))
	for i, base := range bases {
		responses[i] = &AssetResponse{
			ID:        base.ID,
			Name:      base.Name,
			AssetType: base.AssetType,
			Status:    base.Status,
		}
	}

	return responses, total, nil
}

// Validate 验证基础资产数据
func (p *BaseProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	if base == nil {
		return fmt.Errorf("base asset cannot be nil")
	}

	if base.Name == "" {
		return fmt.Errorf("asset name cannot be empty")
	}

	if base.AssetType == "" {
		return fmt.Errorf("asset type cannot be empty")
	}

	if base.Status == "" {
		return fmt.Errorf("asset status cannot be empty")
	}

	if base.CreatedBy == "" {
		return fmt.Errorf("created by cannot be empty")
	}

	if base.UpdatedBy == "" {
		return fmt.Errorf("updated by cannot be empty")
	}

	if base.OrganizationID == 0 {
		return fmt.Errorf("organization ID cannot be zero")
	}

	return nil
}
