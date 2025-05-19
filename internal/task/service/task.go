package service

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/task/repository"
)

// TaskDTO 表示任务数据传输对象
type TaskDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// CreateTaskRequest 表示创建任务请求
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// TaskService 定义任务服务接口
type TaskService interface {
	CreateTask(ctx context.Context, req *CreateTaskRequest, userID uint) (uint, error)
	GetTaskStatus(ctx context.Context, id uint) (*TaskDTO, error)
}

// taskService 是TaskService的具体实现
type taskService struct {
	taskRepo repository.TaskRepository
}

// NewTaskService 创建一个新的任务服务实例
func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

// CreateTask 创建新任务
func (s *taskService) CreateTask(ctx context.Context, req *CreateTaskRequest, userID uint) (uint, error) {
	task := &repository.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		UserID:      userID,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return 0, err
	}

	return task.ID, nil
}

// GetTaskStatus 获取任务状态
func (s *taskService) GetTaskStatus(ctx context.Context, id uint) (*TaskDTO, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TaskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
	}, nil
}
