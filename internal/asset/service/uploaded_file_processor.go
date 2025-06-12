package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// UploadedFileProcessor 上传文件处理器
type UploadedFileProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewUploadedFileProcessor 创建上传文件处理器
func NewUploadedFileProcessor(repo repository.Repository) *UploadedFileProcessor {
	return &UploadedFileProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建上传文件资产
func (p *UploadedFileProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	var req *model.UploadedFileAsset
	switch v := extension.(type) {
	case *model.UploadedFileAsset:
		req = v
	case *dto.CreateUploadedFileRequest:
		req = v.ToUploadedFileAsset()
	default:
		return nil, fmt.Errorf("invalid uploaded file asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateUploadedFile(ctx, base, req); err != nil {
		return nil, err
	}
	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新上传文件资产
func (p *UploadedFileProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	var req *model.UploadedFileAsset
	switch v := extension.(type) {
	case *model.UploadedFileAsset:
		req = v
	case *dto.CreateUploadedFileRequest:
		req = v.ToUploadedFileAsset()
	default:
		return fmt.Errorf("invalid uploaded file asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateUploadedFile(ctx, base, req)
}

// Get 获取上传文件资产
func (p *UploadedFileProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, file, err := p.repo.GetUploadedFile(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, file, nil
}

// Validate 验证上传文件资产数据
func (p *UploadedFileProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.UploadedFileAsset
	switch v := extension.(type) {
	case *model.UploadedFileAsset:
		req = v
	case *dto.CreateUploadedFileRequest:
		req = v.ToUploadedFileAsset()
	default:
		return fmt.Errorf("invalid uploaded file asset type")
	}
	if req.FilePath == "" {
		return fmt.Errorf("file path is required")
	}
	return nil
}

// CreateUploadedFileRequest 创建上传文件请求
type CreateUploadedFileRequest struct {
	Name           string   `json:"name"`
	FilePath       string   `json:"file_path"`
	FileSize       int64    `json:"file_size"`
	FileType       string   `json:"file_type"`
	Checksum       string   `json:"checksum"`
	PreviewURL     string   `json:"preview_url"`
	CreatedBy      string   `json:"created_by"`
	UpdatedBy      string   `json:"updated_by"`
	ProjectID      uint     `json:"project_id"`
	OrganizationID uint     `json:"organization_id"`
	Tags           []string `json:"tags"`
}

// UpdateUploadedFileRequest 更新上传文件请求
type UpdateUploadedFileRequest struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	FilePath   string   `json:"file_path"`
	FileSize   int64    `json:"file_size"`
	FileType   string   `json:"file_type"`
	Checksum   string   `json:"checksum"`
	PreviewURL string   `json:"preview_url"`
	UpdatedBy  string   `json:"updated_by"`
	Tags       []string `json:"tags"`
}

// UploadedFileResponse 上传文件响应
type UploadedFileResponse struct {
	model.BaseAsset
	Extension model.UploadedFileAsset
}
