package repository

import (
	"context"
	"errors"
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskEntity 表示任务数据库实体
type TaskEntity struct {
	ID          string    `gorm:"type:varchar(36);primaryKey"`
	Type        string    `gorm:"type:varchar(10);not null;index"` // scan或asset
	Status      string    `gorm:"type:varchar(20);not null;index"`
	Priority    int       `gorm:"type:int;not null;index"`
	SubType     string    `gorm:"type:varchar(50);not null;index"`
	Payload     []byte    `gorm:"type:json;not null"`
	UserID      uint      `gorm:"type:int;not null;index"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	StartedAt   *time.Time
	CompletedAt *time.Time
	ErrorMsg    string `gorm:"type:text"`
	RetryCount  int    `gorm:"type:int;default:0"`
}

// TableName 指定表名
func (TaskEntity) TableName() string {
	return "tasks"
}

// Task 表示任务实体
type Task struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	Priority    int        `json:"priority"`
	SubType     string     `json:"sub_type"`
	Payload     []byte     `json:"payload"`
	UserID      uint       `json:"user_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ErrorMsg    string     `json:"error_msg"`
	RetryCount  int        `json:"retry_count"`
}

// TaskRepository 定义任务仓库接口
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*Task, int64, error)
	FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*Task, int64, error)
	Update(ctx context.Context, task *Task) error
	UpdateStatus(ctx context.Context, id, status string, errorMsg string) error
	BatchCreate(ctx context.Context, tasks []*Task) error
}

// taskRepository 是TaskRepository的具体实现
type taskRepository struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewTaskRepository 创建一个新的任务仓库实例
func NewTaskRepository(db *gorm.DB, cfg *config.Config) TaskRepository {
	return &taskRepository{
		db:  db,
		cfg: cfg,
	}
}

// convertToEntity 将领域模型转换为数据库实体
func convertToEntity(task *Task) *TaskEntity {
	return &TaskEntity{
		ID:          task.ID,
		Type:        task.Type,
		Status:      task.Status,
		Priority:    task.Priority,
		SubType:     task.SubType,
		Payload:     task.Payload,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		StartedAt:   task.StartedAt,
		CompletedAt: task.CompletedAt,
		ErrorMsg:    task.ErrorMsg,
		RetryCount:  task.RetryCount,
	}
}

// convertToDomain 将数据库实体转换为领域模型
func convertToDomain(entity *TaskEntity) *Task {
	return &Task{
		ID:          entity.ID,
		Type:        entity.Type,
		Status:      entity.Status,
		Priority:    entity.Priority,
		SubType:     entity.SubType,
		Payload:     entity.Payload,
		UserID:      entity.UserID,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		StartedAt:   entity.StartedAt,
		CompletedAt: entity.CompletedAt,
		ErrorMsg:    entity.ErrorMsg,
		RetryCount:  entity.RetryCount,
	}
}

// Create 创建新任务
func (r *taskRepository) Create(ctx context.Context, task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	entity := convertToEntity(task)
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return result.Error
	}

	task.ID = entity.ID
	return nil
}

// FindByID 根据ID查找任务
func (r *taskRepository) FindByID(ctx context.Context, id string) (*Task, error) {
	var entity TaskEntity
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, result.Error
	}

	return convertToDomain(&entity), nil
}

// FindByStatus 根据状态查找任务
func (r *taskRepository) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*Task, int64, error) {
	var entities []*TaskEntity
	var count int64

	if err := r.db.WithContext(ctx).Model(&TaskEntity{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]*Task, len(entities))
	for i, entity := range entities {
		tasks[i] = convertToDomain(entity)
	}

	return tasks, count, nil
}

// FindByUserID 根据用户ID查找任务
func (r *taskRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*Task, int64, error) {
	var entities []*TaskEntity
	var count int64

	if err := r.db.WithContext(ctx).Model(&TaskEntity{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]*Task, len(entities))
	for i, entity := range entities {
		tasks[i] = convertToDomain(entity)
	}

	return tasks, count, nil
}

// Update 更新任务
func (r *taskRepository) Update(ctx context.Context, task *Task) error {
	entity := convertToEntity(task)
	result := r.db.WithContext(ctx).Save(entity)
	return result.Error
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, id, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == string(domain.TaskStatusRunning) {
		now := time.Now()
		updates["started_at"] = now
	}

	if status == string(domain.TaskStatusCompleted) || status == string(domain.TaskStatusFailed) {
		now := time.Now()
		updates["completed_at"] = now
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	result := r.db.WithContext(ctx).Model(&TaskEntity{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// BatchCreate 批量创建任务
func (r *taskRepository) BatchCreate(ctx context.Context, tasks []*Task) error {
	if len(tasks) == 0 {
		return nil
	}

	entities := make([]*TaskEntity, len(tasks))
	for i, task := range tasks {
		if task.ID == "" {
			task.ID = uuid.New().String()
		}
		entities[i] = convertToEntity(task)
	}

	result := r.db.WithContext(ctx).Create(entities)
	if result.Error != nil {
		return result.Error
	}

	// 更新ID
	for i, entity := range entities {
		tasks[i].ID = entity.ID
	}

	return nil
}
