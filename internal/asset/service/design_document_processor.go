package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
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
	var req *model.DesignDocumentAsset
	switch v := extension.(type) {
	case *model.DesignDocumentAsset:
		req = v
	case *dto.CreateDesignDocumentRequest:
		req = v.ToDesignDocumentAsset()
	default:
		return nil, fmt.Errorf("invalid design document asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateDesignDocument(ctx, base, req); err != nil {
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
	var req *model.DesignDocumentAsset
	switch v := extension.(type) {
	case *model.DesignDocumentAsset:
		req = v
	case *dto.CreateDesignDocumentRequest:
		req = v.ToDesignDocumentAsset()
	default:
		return fmt.Errorf("invalid design document asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateDesignDocument(ctx, base, req)
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
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.DesignDocumentAsset
	switch v := extension.(type) {
	case *model.DesignDocumentAsset:
		req = v
	case *dto.CreateDesignDocumentRequest:
		req = v.ToDesignDocumentAsset()
	default:
		return fmt.Errorf("invalid design document asset type")
	}
	if req.DesignType == "" {
		return fmt.Errorf("design type is required")
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
