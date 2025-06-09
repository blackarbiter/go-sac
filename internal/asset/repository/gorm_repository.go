package repository

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/asset/repository/model"
	"gorm.io/gorm"
)

// GormRepository GORM仓储实现
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository 创建GORM仓储实例
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// 基础资产操作实现
func (r *GormRepository) CreateBase(ctx context.Context, base *model.BaseAsset) error {
	return r.db.WithContext(ctx).Create(base).Error
}

func (r *GormRepository) UpdateBase(ctx context.Context, base *model.BaseAsset) error {
	return r.db.WithContext(ctx).Save(base).Error
}

func (r *GormRepository) GetBase(ctx context.Context, id uint) (*model.BaseAsset, error) {
	var base model.BaseAsset
	err := r.db.WithContext(ctx).First(&base, id).Error
	if err != nil {
		return nil, err
	}
	return &base, nil
}

func (r *GormRepository) DeleteBase(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.BaseAsset{}, id).Error
}

func (r *GormRepository) ListBase(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*model.BaseAsset, int64, error) {
	var bases []*model.BaseAsset
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BaseAsset{})

	// 应用过滤条件
	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&bases).Error; err != nil {
		return nil, 0, err
	}

	return bases, total, nil
}

// 需求文档资产操作实现
func (r *GormRepository) CreateRequirement(ctx context.Context, base *model.BaseAsset, ext *model.RequirementAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateRequirement(ctx context.Context, base *model.BaseAsset, ext *model.RequirementAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetRequirement(ctx context.Context, id uint) (*model.BaseAsset, *model.RequirementAsset, error) {
	var base model.BaseAsset
	var ext model.RequirementAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// 设计文档资产操作实现
func (r *GormRepository) CreateDesignDocument(ctx context.Context, base *model.BaseAsset, ext *model.DesignDocumentAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateDesignDocument(ctx context.Context, base *model.BaseAsset, ext *model.DesignDocumentAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetDesignDocument(ctx context.Context, id uint) (*model.BaseAsset, *model.DesignDocumentAsset, error) {
	var base model.BaseAsset
	var ext model.DesignDocumentAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// 代码仓库资产操作实现
func (r *GormRepository) CreateRepository(ctx context.Context, base *model.BaseAsset, ext *model.RepositoryAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateRepository(ctx context.Context, base *model.BaseAsset, ext *model.RepositoryAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetRepository(ctx context.Context, id uint) (*model.BaseAsset, *model.RepositoryAsset, error) {
	var base model.BaseAsset
	var ext model.RepositoryAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// 上传文件资产操作实现
func (r *GormRepository) CreateUploadedFile(ctx context.Context, base *model.BaseAsset, ext *model.UploadedFileAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateUploadedFile(ctx context.Context, base *model.BaseAsset, ext *model.UploadedFileAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetUploadedFile(ctx context.Context, id uint) (*model.BaseAsset, *model.UploadedFileAsset, error) {
	var base model.BaseAsset
	var ext model.UploadedFileAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// 容器镜像资产操作实现
func (r *GormRepository) CreateImage(ctx context.Context, base *model.BaseAsset, ext *model.ImageAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateImage(ctx context.Context, base *model.BaseAsset, ext *model.ImageAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetImage(ctx context.Context, id uint) (*model.BaseAsset, *model.ImageAsset, error) {
	var base model.BaseAsset
	var ext model.ImageAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// 域名资产操作实现
func (r *GormRepository) CreateDomain(ctx context.Context, base *model.BaseAsset, ext *model.DomainAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateDomain(ctx context.Context, base *model.BaseAsset, ext *model.DomainAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetDomain(ctx context.Context, id uint) (*model.BaseAsset, *model.DomainAsset, error) {
	var base model.BaseAsset
	var ext model.DomainAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}

// IP地址资产操作实现
func (r *GormRepository) CreateIP(ctx context.Context, base *model.BaseAsset, ext *model.IPAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(base).Error; err != nil {
			return err
		}
		ext.ID = base.ID
		return tx.Create(ext).Error
	})
}

func (r *GormRepository) UpdateIP(ctx context.Context, base *model.BaseAsset, ext *model.IPAsset) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(base).Error; err != nil {
			return err
		}
		return tx.Save(ext).Error
	})
}

func (r *GormRepository) GetIP(ctx context.Context, id uint) (*model.BaseAsset, *model.IPAsset, error) {
	var base model.BaseAsset
	var ext model.IPAsset

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&base, id).Error; err != nil {
			return err
		}
		return tx.First(&ext, id).Error
	})

	if err != nil {
		return nil, nil, err
	}

	return &base, &ext, nil
}
