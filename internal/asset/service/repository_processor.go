package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// RepositoryProcessor 代码仓库处理器
type RepositoryProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewRepositoryProcessor 创建代码仓库处理器
func NewRepositoryProcessor(repo repository.Repository) *RepositoryProcessor {
	return &RepositoryProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建代码仓库资产
func (p *RepositoryProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	var req *model.RepositoryAsset
	switch v := extension.(type) {
	case *model.RepositoryAsset:
		req = v
	case *dto.CreateRepositoryRequest:
		req = v.ToRepositoryAsset()
	default:
		return nil, fmt.Errorf("invalid repository asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateRepository(ctx, base, req); err != nil {
		return nil, err
	}
	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新代码仓库资产
func (p *RepositoryProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	var req *model.RepositoryAsset
	switch v := extension.(type) {
	case *model.RepositoryAsset:
		req = v
	case *dto.CreateRepositoryRequest:
		req = v.ToRepositoryAsset()
	default:
		return fmt.Errorf("invalid repository asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateRepository(ctx, base, req)
}

// Get 获取代码仓库资产
func (p *RepositoryProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, repo, err := p.repo.GetRepository(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, repo, nil
}

// Validate 验证代码仓库资产数据
func (p *RepositoryProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.RepositoryAsset
	switch v := extension.(type) {
	case *model.RepositoryAsset:
		req = v
	case *dto.CreateRepositoryRequest:
		req = v.ToRepositoryAsset()
	default:
		return fmt.Errorf("invalid repository asset type")
	}
	if req.RepoURL == "" {
		return fmt.Errorf("repo url is required")
	}
	return nil
}
