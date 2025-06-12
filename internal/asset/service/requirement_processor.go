package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
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
	var req *model.RequirementAsset
	switch v := extension.(type) {
	case *model.RequirementAsset:
		req = v
	case *dto.CreateRequirementRequest:
		req = v.ToRequirementAsset()
	default:
		return nil, fmt.Errorf("invalid requirement asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
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
	var req *model.RequirementAsset
	switch v := extension.(type) {
	case *model.RequirementAsset:
		req = v
	case *dto.CreateRequirementRequest:
		req = v.ToRequirementAsset()
	default:
		return fmt.Errorf("invalid requirement asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
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
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.RequirementAsset
	switch v := extension.(type) {
	case *model.RequirementAsset:
		req = v
	case *dto.CreateRequirementRequest:
		req = v.ToRequirementAsset()
	default:
		return fmt.Errorf("invalid requirement asset type")
	}
	if req.BusinessValue == "" {
		return fmt.Errorf("business value is required")
	}
	return nil
}
