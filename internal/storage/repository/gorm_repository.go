package repository

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/storage/repository/model"
	"gorm.io/gorm"
)

// GormRepository implements the Repository interface using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GormRepository instance
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// AutoMigrate performs database migrations
func (r *GormRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&model.SASTModel{},
		&model.DASTModel{},
		&model.SCAModel{},
	)
}

// SAST operations
func (r *GormRepository) CreateSAST(ctx context.Context, result *model.SASTModel) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *GormRepository) BatchCreateSAST(ctx context.Context, results []*model.SASTModel) error {
	return r.db.WithContext(ctx).CreateInBatches(results, 100).Error
}

func (r *GormRepository) FindSASTByTaskID(ctx context.Context, taskID string) (*model.SASTModel, error) {
	var result model.SASTModel
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormRepository) FindSASTByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.SASTModel, error) {
	var results []*model.SASTModel
	err := r.db.WithContext(ctx).Where("task_id IN ?", taskIDs).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormRepository) UpdateSAST(ctx context.Context, result *model.SASTModel) error {
	return r.db.WithContext(ctx).Save(result).Error
}

// DAST operations
func (r *GormRepository) CreateDAST(ctx context.Context, result *model.DASTModel) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *GormRepository) BatchCreateDAST(ctx context.Context, results []*model.DASTModel) error {
	return r.db.WithContext(ctx).CreateInBatches(results, 100).Error
}

func (r *GormRepository) FindDASTByTaskID(ctx context.Context, taskID string) (*model.DASTModel, error) {
	var result model.DASTModel
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormRepository) FindDASTByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.DASTModel, error) {
	var results []*model.DASTModel
	err := r.db.WithContext(ctx).Where("task_id IN ?", taskIDs).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormRepository) UpdateDAST(ctx context.Context, result *model.DASTModel) error {
	return r.db.WithContext(ctx).Save(result).Error
}

// SCA operations
func (r *GormRepository) CreateSCA(ctx context.Context, result *model.SCAModel) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *GormRepository) BatchCreateSCA(ctx context.Context, results []*model.SCAModel) error {
	return r.db.WithContext(ctx).CreateInBatches(results, 100).Error
}

func (r *GormRepository) FindSCAByTaskID(ctx context.Context, taskID string) (*model.SCAModel, error) {
	var result model.SCAModel
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormRepository) FindSCAByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.SCAModel, error) {
	var results []*model.SCAModel
	err := r.db.WithContext(ctx).Where("task_id IN ?", taskIDs).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormRepository) UpdateSCA(ctx context.Context, result *model.SCAModel) error {
	return r.db.WithContext(ctx).Save(result).Error
}
