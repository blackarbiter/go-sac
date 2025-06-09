package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// DesignDocumentProcessor 设计文档处理器
type DesignDocumentProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewDesignDocumentProcessor 创建设计文档处理器
func NewDesignDocumentProcessor(repo repository.Repository) *DesignDocumentProcessor {
	return &DesignDocumentProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建设计文档资产
func (p *DesignDocumentProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	doc, ok := extension.(*model.DesignDocumentAsset)
	if !ok {
		return nil, fmt.Errorf("invalid design document asset type")
	}

	// 创建设计文档资产
	if err := p.repo.CreateDesignDocument(ctx, base, doc); err != nil {
		return nil, err
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新设计文档资产
func (p *DesignDocumentProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	doc, ok := extension.(*model.DesignDocumentAsset)
	if !ok {
		return fmt.Errorf("invalid design document asset type")
	}

	// 更新设计文档资产
	return p.repo.UpdateDesignDocument(ctx, base, doc)
}

// Get 获取设计文档资产
func (p *DesignDocumentProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, doc, err := p.repo.GetDesignDocument(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, doc, nil
}

// Validate 验证设计文档资产数据
func (p *DesignDocumentProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证设计文档资产数据
	doc, ok := extension.(*model.DesignDocumentAsset)
	if !ok {
		return fmt.Errorf("invalid design document asset type")
	}

	// 验证设计类型
	if doc.DesignType == "" {
		return fmt.Errorf("design type is required")
	}

	// 验证组件
	if doc.Components == nil {
		return fmt.Errorf("components are required")
	}

	return nil
}

// CreateDesignDocumentRequest 创建设计文档请求
type CreateDesignDocumentRequest struct {
	Name            string          `json:"name"`
	DesignType      string          `json:"design_type"`
	Components      json.RawMessage `json:"components"`
	Diagrams        json.RawMessage `json:"diagrams"`
	Dependencies    json.RawMessage `json:"dependencies"`
	TechnologyStack []string        `json:"technology_stack"`
	CreatedBy       string          `json:"created_by"`
	UpdatedBy       string          `json:"updated_by"`
	ProjectID       uint            `json:"project_id"`
	OrganizationID  uint            `json:"organization_id"`
	Tags            []string        `json:"tags"`
}

// UpdateDesignDocumentRequest 更新设计文档请求
type UpdateDesignDocumentRequest struct {
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	DesignType      string          `json:"design_type"`
	Components      json.RawMessage `json:"components"`
	Diagrams        json.RawMessage `json:"diagrams"`
	Dependencies    json.RawMessage `json:"dependencies"`
	TechnologyStack []string        `json:"technology_stack"`
	UpdatedBy       string          `json:"updated_by"`
	Tags            []string        `json:"tags"`
}

// DesignDocumentResponse 设计文档响应
type DesignDocumentResponse struct {
	model.BaseAsset
	Extension model.DesignDocumentAsset
}
