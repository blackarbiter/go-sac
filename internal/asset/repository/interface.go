package repository

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
)

// Repository 定义资产仓储接口
type Repository interface {
	// 基础资产操作
	CreateBase(ctx context.Context, base *model.BaseAsset) error
	UpdateBase(ctx context.Context, base *model.BaseAsset) error
	GetBase(ctx context.Context, id uint) (*model.BaseAsset, error)
	DeleteBase(ctx context.Context, id uint) error
	ListBase(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*model.BaseAsset, int64, error)

	// 需求文档资产操作
	CreateRequirement(ctx context.Context, base *model.BaseAsset, ext *model.RequirementAsset) error
	UpdateRequirement(ctx context.Context, base *model.BaseAsset, ext *model.RequirementAsset) error
	GetRequirement(ctx context.Context, id uint) (*model.BaseAsset, *model.RequirementAsset, error)

	// 设计文档资产操作
	CreateDesignDocument(ctx context.Context, base *model.BaseAsset, ext *model.DesignDocumentAsset) error
	UpdateDesignDocument(ctx context.Context, base *model.BaseAsset, ext *model.DesignDocumentAsset) error
	GetDesignDocument(ctx context.Context, id uint) (*model.BaseAsset, *model.DesignDocumentAsset, error)

	// 代码仓库资产操作
	CreateRepository(ctx context.Context, base *model.BaseAsset, ext *model.RepositoryAsset) error
	UpdateRepository(ctx context.Context, base *model.BaseAsset, ext *model.RepositoryAsset) error
	GetRepository(ctx context.Context, id uint) (*model.BaseAsset, *model.RepositoryAsset, error)

	// 上传文件资产操作
	CreateUploadedFile(ctx context.Context, base *model.BaseAsset, ext *model.UploadedFileAsset) error
	UpdateUploadedFile(ctx context.Context, base *model.BaseAsset, ext *model.UploadedFileAsset) error
	GetUploadedFile(ctx context.Context, id uint) (*model.BaseAsset, *model.UploadedFileAsset, error)

	// 容器镜像资产操作
	CreateImage(ctx context.Context, base *model.BaseAsset, ext *model.ImageAsset) error
	UpdateImage(ctx context.Context, base *model.BaseAsset, ext *model.ImageAsset) error
	GetImage(ctx context.Context, id uint) (*model.BaseAsset, *model.ImageAsset, error)

	// 域名资产操作
	CreateDomain(ctx context.Context, base *model.BaseAsset, ext *model.DomainAsset) error
	UpdateDomain(ctx context.Context, base *model.BaseAsset, ext *model.DomainAsset) error
	GetDomain(ctx context.Context, id uint) (*model.BaseAsset, *model.DomainAsset, error)

	// IP地址资产操作
	CreateIP(ctx context.Context, base *model.BaseAsset, ext *model.IPAsset) error
	UpdateIP(ctx context.Context, base *model.BaseAsset, ext *model.IPAsset) error
	GetIP(ctx context.Context, id uint) (*model.BaseAsset, *model.IPAsset, error)
}
