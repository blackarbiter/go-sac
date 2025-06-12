package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 错误常量
var (
	ErrStorageNotFound = errors.New("storage not found")
)

// StorageEntity 表示存储数据库实体
type StorageEntity struct {
	ID        string            `gorm:"type:varchar(36);primaryKey"`
	Path      string            `gorm:"type:varchar(255);not null"`
	Type      string            `gorm:"type:varchar(50);not null;index"`
	Status    string            `gorm:"type:varchar(50);not null;default:'pending';index"`
	Metadata  map[string]string `gorm:"type:json"`
	CreatedAt time.Time         `gorm:"not null"`
	UpdatedAt time.Time         `gorm:"not null"`
}

// TableName 指定表名
func (StorageEntity) TableName() string {
	return "storages"
}

// Storage 表示存储领域模型
type Storage struct {
	ID        string            `json:"id"`
	Path      string            `json:"path"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// StorageRepository 定义存储仓库接口
type StorageRepository interface {
	Create(ctx context.Context, storage *Storage) error
	FindByID(ctx context.Context, id string) (*Storage, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*Storage, int64, error)
	FindByType(ctx context.Context, storageType string, limit, offset int) ([]*Storage, int64, error)
	Update(ctx context.Context, storage *Storage) error
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Storage, int64, error)
	BatchCreate(ctx context.Context, storages []*Storage) error
}

// storageRepository 是StorageRepository的具体实现
type storageRepository struct {
	db *gorm.DB
}

// convertToEntity 将领域模型转换为数据库实体
func convertToEntity(storage *Storage) *StorageEntity {
	return &StorageEntity{
		ID:        storage.ID,
		Path:      storage.Path,
		Type:      storage.Type,
		Status:    storage.Status,
		Metadata:  storage.Metadata,
		CreatedAt: storage.CreatedAt,
		UpdatedAt: storage.UpdatedAt,
	}
}

// convertToDomain 将数据库实体转换为领域模型
func convertToDomain(entity *StorageEntity) *Storage {
	return &Storage{
		ID:        entity.ID,
		Path:      entity.Path,
		Type:      entity.Type,
		Status:    entity.Status,
		Metadata:  entity.Metadata,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

// Create 创建新存储记录
func (r *storageRepository) Create(ctx context.Context, storage *Storage) error {
	if storage.ID == "" {
		storage.ID = uuid.New().String()
	}

	entity := convertToEntity(storage)
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return result.Error
	}

	storage.ID = entity.ID
	return nil
}

// FindByID 根据ID查找存储记录
func (r *storageRepository) FindByID(ctx context.Context, id string) (*Storage, error) {
	var entity StorageEntity
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrStorageNotFound
		}
		return nil, result.Error
	}

	return convertToDomain(&entity), nil
}

// FindByStatus 根据状态查找存储记录
func (r *storageRepository) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*Storage, int64, error) {
	var entities []*StorageEntity
	var count int64

	if err := r.db.WithContext(ctx).Model(&StorageEntity{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	storages := make([]*Storage, len(entities))
	for i, entity := range entities {
		storages[i] = convertToDomain(entity)
	}

	return storages, count, nil
}

// FindByType 根据类型查找存储记录
func (r *storageRepository) FindByType(ctx context.Context, storageType string, limit, offset int) ([]*Storage, int64, error) {
	var entities []*StorageEntity
	var count int64

	if err := r.db.WithContext(ctx).Model(&StorageEntity{}).Where("type = ?", storageType).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("type = ?", storageType).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	storages := make([]*Storage, len(entities))
	for i, entity := range entities {
		storages[i] = convertToDomain(entity)
	}

	return storages, count, nil
}

// Update 更新存储记录
func (r *storageRepository) Update(ctx context.Context, storage *Storage) error {
	entity := convertToEntity(storage)
	result := r.db.WithContext(ctx).Save(entity)
	return result.Error
}

// UpdateStatus 更新存储状态
func (r *storageRepository) UpdateStatus(ctx context.Context, id, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&StorageEntity{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// Delete 删除存储记录
func (r *storageRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&StorageEntity{}, "id = ?", id)
	return result.Error
}

// List 获取存储记录列表
func (r *storageRepository) List(ctx context.Context, limit, offset int) ([]*Storage, int64, error) {
	var entities []*StorageEntity
	var count int64

	if err := r.db.WithContext(ctx).Model(&StorageEntity{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	storages := make([]*Storage, len(entities))
	for i, entity := range entities {
		storages[i] = convertToDomain(entity)
	}

	return storages, count, nil
}

// BatchCreate 批量创建存储记录
func (r *storageRepository) BatchCreate(ctx context.Context, storages []*Storage) error {
	if len(storages) == 0 {
		return nil
	}

	entities := make([]*StorageEntity, len(storages))
	for i, storage := range storages {
		if storage.ID == "" {
			storage.ID = uuid.New().String()
		}
		entities[i] = convertToEntity(storage)
	}

	result := r.db.WithContext(ctx).Create(entities)
	if result.Error != nil {
		return result.Error
	}

	// 更新ID
	for i, entity := range entities {
		storages[i].ID = entity.ID
	}

	return nil
}
