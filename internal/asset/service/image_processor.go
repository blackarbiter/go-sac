package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// ImageProcessor 容器镜像处理器
type ImageProcessor struct {
	*BaseProcessor
	repo repository.Repository
}

// NewImageProcessor 创建容器镜像处理器
func NewImageProcessor(repo repository.Repository) *ImageProcessor {
	return &ImageProcessor{
		BaseProcessor: NewBaseProcessor(repo),
		repo:          repo,
	}
}

// Create 创建容器镜像资产
func (p *ImageProcessor) Create(ctx context.Context, base *model.BaseAsset, extension interface{}) (*AssetResponse, error) {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return nil, err
	}

	// 类型断言
	image, ok := extension.(*model.ImageAsset)
	if !ok {
		return nil, fmt.Errorf("invalid image asset type")
	}

	// 创建容器镜像资产
	if err := p.repo.CreateImage(ctx, base, image); err != nil {
		return nil, err
	}

	return &AssetResponse{
		ID:        base.ID,
		Name:      base.Name,
		AssetType: base.AssetType,
		Status:    base.Status,
	}, nil
}

// Update 更新容器镜像资产
func (p *ImageProcessor) Update(ctx context.Context, id uint, base *model.BaseAsset, extension interface{}) error {
	// 验证数据
	if err := p.Validate(base, extension); err != nil {
		return err
	}

	// 类型断言
	image, ok := extension.(*model.ImageAsset)
	if !ok {
		return fmt.Errorf("invalid image asset type")
	}

	// 更新容器镜像资产
	return p.repo.UpdateImage(ctx, base, image)
}

// Get 获取容器镜像资产
func (p *ImageProcessor) Get(ctx context.Context, id uint) (*model.BaseAsset, interface{}, error) {
	base, image, err := p.repo.GetImage(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return base, image, nil
}

// Validate 验证容器镜像资产数据
func (p *ImageProcessor) Validate(base *model.BaseAsset, extension interface{}) error {
	// 验证基础资产数据
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}

	// 验证容器镜像资产数据
	image, ok := extension.(*model.ImageAsset)
	if !ok {
		return fmt.Errorf("invalid image asset type")
	}

	// 验证镜像名称
	if image.ImageName == "" {
		return fmt.Errorf("image name is required")
	}

	// 验证镜像标签
	if image.Tag == "" {
		return fmt.Errorf("image tag is required")
	}

	// 验证镜像仓库
	if image.RegistryURL == "" {
		return fmt.Errorf("image registry is required")
	}

	return nil
}

// CreateImageRequest 创建容器镜像请求
type CreateImageRequest struct {
	Name            string   `json:"name"`
	RegistryURL     string   `json:"registry_url"`
	ImageName       string   `json:"image_name"`
	Tag             string   `json:"tag"`
	Digest          string   `json:"digest"`
	Size            int64    `json:"size"`
	Vulnerabilities []byte   `json:"vulnerabilities"`
	CreatedBy       string   `json:"created_by"`
	UpdatedBy       string   `json:"updated_by"`
	ProjectID       uint     `json:"project_id"`
	OrganizationID  uint     `json:"organization_id"`
	Tags            []string `json:"tags"`
}

// UpdateImageRequest 更新容器镜像请求
type UpdateImageRequest struct {
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	RegistryURL     string   `json:"registry_url"`
	ImageName       string   `json:"image_name"`
	Tag             string   `json:"tag"`
	Digest          string   `json:"digest"`
	Size            int64    `json:"size"`
	Vulnerabilities []byte   `json:"vulnerabilities"`
	UpdatedBy       string   `json:"updated_by"`
	Tags            []string `json:"tags"`
}

// ImageResponse 容器镜像响应
type ImageResponse struct {
	model.BaseAsset
	Extension model.ImageAsset
}
