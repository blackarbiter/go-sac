package service

import (
	"context"
	"fmt"

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
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	repo, ok := extension.(*model.RepositoryAsset)
	if !ok {
		return nil, fmt.Errorf("invalid repository asset type")
	}

	// 创建代码仓库资产
	if err := p.repo.CreateRepository(ctx, base, repo); err != nil {
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
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	repo, ok := extension.(*model.RepositoryAsset)
	if !ok {
		return fmt.Errorf("invalid repository asset type")
	}

	// 更新代码仓库资产
	return p.repo.UpdateRepository(ctx, base, repo)
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
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证代码仓库资产数据
	repo, ok := extension.(*model.RepositoryAsset)
	if !ok {
		return fmt.Errorf("invalid repository asset type")
	}

	// 验证仓库URL
	if repo.RepoURL == "" {
		return fmt.Errorf("repository URL is required")
	}

	// 验证编程语言
	if repo.Language == "" {
		return fmt.Errorf("programming language is required")
	}

	return nil
}
