package service

import (
	"context"
	"fmt"

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
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	file, ok := extension.(*model.UploadedFileAsset)
	if !ok {
		return nil, fmt.Errorf("invalid uploaded file asset type")
	}

	// 创建上传文件资产
	if err := p.repo.CreateUploadedFile(ctx, base, file); err != nil {
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
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	file, ok := extension.(*model.UploadedFileAsset)
	if !ok {
		return fmt.Errorf("invalid uploaded file asset type")
	}

	// 更新上传文件资产
	return p.repo.UpdateUploadedFile(ctx, base, file)
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
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证上传文件资产数据
	file, ok := extension.(*model.UploadedFileAsset)
	if !ok {
		return fmt.Errorf("invalid uploaded file asset type")
	}

	// 验证文件路径
	if file.FilePath == "" {
		return fmt.Errorf("file path is required")
	}

	// 验证文件大小
	if file.FileSize <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	// 验证文件类型
	if file.FileType == "" {
		return fmt.Errorf("file type is required")
	}

	// 验证校验和
	if file.Checksum == "" {
		return fmt.Errorf("checksum is required")
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
