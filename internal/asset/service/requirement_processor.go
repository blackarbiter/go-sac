package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// RequirementProcessor 需求文档处理器
type RequirementProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewRequirementProcessor 创建需求文档处理器
func NewRequirementProcessor(repo repository.Repository) *RequirementProcessor {
	return &RequirementProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建需求文档资产
func (p *RequirementProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	req, ok := extension.(*model.RequirementAsset)
	if !ok {
		return nil, fmt.Errorf("invalid requirement asset type")
	}

	// 创建需求文档资产
	if err := p.repo.CreateRequirement(ctx, base, req); err != nil {
		return nil, err
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新需求文档资产
func (p *RequirementProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	req, ok := extension.(*model.RequirementAsset)
	if !ok {
		return fmt.Errorf("invalid requirement asset type")
	}

	// 更新需求文档资产
	return p.repo.UpdateRequirement(ctx, base, req)
}

// Get 获取需求文档资产
func (p *RequirementProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, req, err := p.repo.GetRequirement(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, req, nil
}

// Validate 验证需求文档资产数据
func (p *RequirementProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证需求文档资产数据
	req, ok := extension.(*model.RequirementAsset)
	if !ok {
		return fmt.Errorf("invalid requirement asset type")
	}

	// 验证版本号
	if req.Version == "" {
		return fmt.Errorf("version is required")
	}

	// 验证优先级
	if req.Priority < 1 || req.Priority > 5 {
		return fmt.Errorf("priority must be between 1 and 5")
	}

	return nil
}
