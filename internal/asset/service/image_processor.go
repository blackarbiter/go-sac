package service

import (
	"context"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
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
	var req *model.ImageAsset
	switch v := extension.(type) {
	case *model.ImageAsset:
		req = v
	case *dto.CreateImageRequest:
		req = v.ToImageAsset()
	default:
		return nil, fmt.Errorf("invalid image asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return nil, err
	}
	if err := p.repo.CreateImage(ctx, base, req); err != nil {
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
	var req *model.ImageAsset
	switch v := extension.(type) {
	case *model.ImageAsset:
		req = v
	case *dto.CreateImageRequest:
		req = v.ToImageAsset()
	default:
		return fmt.Errorf("invalid image asset type")
	}
	if err := p.Validate(base, req); err != nil {
		return err
	}
	return p.repo.UpdateImage(ctx, base, req)
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
	if err := p.BaseProcessor.Validate(base, nil); err != nil {
		return err
	}
	var req *model.ImageAsset
	switch v := extension.(type) {
	case *model.ImageAsset:
		req = v
	case *dto.CreateImageRequest:
		req = v.ToImageAsset()
	default:
		return fmt.Errorf("invalid image asset type")
	}
	if req.ImageName == "" {
		return fmt.Errorf("image name is required")
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
