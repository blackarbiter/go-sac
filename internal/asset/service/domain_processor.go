package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// DomainProcessor 域名处理器
type DomainProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewDomainProcessor 创建域名处理器
func NewDomainProcessor(repo repository.Repository) *DomainProcessor {
	return &DomainProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建域名资产
func (p *DomainProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	var req *model.DomainAsset
	switch v := extension.(type) {
	case *model.DomainAsset:
		req = v
	case *dto.CreateDomainRequest:
		req = dto.ToModelDomainAsset(v)
	default:
		return nil, fmt.Errorf("invalid domain asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateDomain(ctx, base, req); err != nil {
		return nil, err
	}
	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新域名资产
func (p *DomainProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	var req *model.DomainAsset
	switch v := extension.(type) {
	case *model.DomainAsset:
		req = v
	case *dto.CreateDomainRequest:
		req = dto.ToModelDomainAsset(v)
	default:
		return fmt.Errorf("invalid domain asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateDomain(ctx, base, req)
}

// Get 获取域名资产
func (p *DomainProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, domain, err := p.repo.GetDomain(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, domain, nil
}

// Validate 验证域名资产数据
func (p *DomainProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.DomainAsset
	switch v := extension.(type) {
	case *model.DomainAsset:
		req = v
	case *dto.CreateDomainRequest:
		req = dto.ToModelDomainAsset(v)
	default:
		return fmt.Errorf("invalid domain asset type")
	}
	if req.DomainName == "" {
		return fmt.Errorf("domain name is required")
	}
	return nil
}
