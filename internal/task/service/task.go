package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/task/repository"
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
)

// TaskDTO 表示任务数据传输对象
type TaskDTO struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	SubType     string `json:"sub_type"`
	UserID      uint   `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	ErrorMsg    string `json:"error_msg,omitempty"`
	RetryCount  int    `json:"retry_count"`
}

// CreateScanTaskRequest 表示创建扫描任务请求
type CreateScanTaskRequest struct {
	AssetID   string                 `json:"asset_id" binding:"required"`
	AssetType string                 `json:"asset_type" binding:"required"`
	ScanType  string                 `json:"scan_type" binding:"required"`
	Options   map[string]interface{} `json:"options"`
	Priority  int                    `json:"priority"`
}

// CreateAssetTaskRequest 表示创建资产任务请求
type CreateAssetTaskRequest struct {
	AssetID   string                 `json:"asset_id" binding:"required"`
	AssetType string                 `json:"asset_type" binding:"required"`
	Operation string                 `json:"operation" binding:"required,oneof=create update delete"`
	Data      map[string]interface{} `json:"data"`
}

// BatchCreateScanTaskRequest 批量创建扫描任务请求
type BatchCreateScanTaskRequest struct {
	Tasks []CreateScanTaskRequest `json:"tasks" binding:"required,min=1"`
}

// BatchCreateAssetTaskRequest 批量创建资产任务请求
type BatchCreateAssetTaskRequest struct {
	Tasks []CreateAssetTaskRequest `json:"tasks" binding:"required,min=1"`
}

// UpdateTaskStatusRequest 更新任务状态请求
type UpdateTaskStatusRequest struct {
	Status   string `json:"status" binding:"required,oneof=pending running completed failed cancelled"`
	ErrorMsg string `json:"error_msg"`
}

// TaskQueryParams 任务查询参数
type TaskQueryParams struct {
	UserID uint   `form:"user_id"`
	Status string `form:"status"`
	Type   string `form:"type"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Total int64     `json:"total"`
	Items []TaskDTO `json:"items"`
}

// TaskService 定义任务服务接口
type TaskService interface {
	CreateScanTask(ctx context.Context, req *CreateScanTaskRequest, userID uint) (string, error)
	CreateAssetTask(ctx context.Context, req *CreateAssetTaskRequest, userID uint) (string, error)
	BatchCreateScanTasks(ctx context.Context, req *BatchCreateScanTaskRequest, userID uint) ([]string, error)
	BatchCreateAssetTasks(ctx context.Context, req *BatchCreateAssetTaskRequest, userID uint) ([]string, error)
	GetTaskStatus(ctx context.Context, id string) (*TaskDTO, error)
	BatchGetTaskStatus(ctx context.Context, ids []string) ([]*TaskDTO, error)
	UpdateTaskStatus(ctx context.Context, id string, req *UpdateTaskStatusRequest) error
	ListTasks(ctx context.Context, params *TaskQueryParams) (*TaskListResponse, error)
	CancelTask(ctx context.Context, id string) error
	BatchCancelTasks(ctx context.Context, ids []string) ([]string, error)
}

// taskService 是TaskService的具体实现
type taskService struct {
	taskRepo      repository.TaskRepository
	taskPublisher *rabbitmq.TaskPublisher
}

// NewTaskService 创建一个新的任务服务实例
func NewTaskService(taskRepo repository.TaskRepository, taskPublisher *rabbitmq.TaskPublisher) TaskService {
	return &taskService{
		taskRepo:      taskRepo,
		taskPublisher: taskPublisher,
	}
}

// convertToDTO 将任务实体转换为DTO
func convertToDTO(task *repository.Task) *TaskDTO {
	dto := &TaskDTO{
		ID:         task.ID,
		Type:       task.Type,
		Status:     task.Status,
		Priority:   task.Priority,
		SubType:    task.SubType,
		UserID:     task.UserID,
		CreatedAt:  task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ErrorMsg:   task.ErrorMsg,
		RetryCount: task.RetryCount,
	}

	if task.StartedAt != nil {
		dto.StartedAt = task.StartedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	if task.CompletedAt != nil {
		dto.CompletedAt = task.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return dto
}

// CreateScanTask 创建扫描任务
func (s *taskService) CreateScanTask(ctx context.Context, req *CreateScanTaskRequest, userID uint) (string, error) {
	// 解析扫描类型
	scanType, err := domain.ParseScanType(req.ScanType)
	if err != nil {
		return "", err
	}

	// 解析资产类型
	assetType, err := domain.ParseAssetType(req.AssetType)
	if err != nil {
		return "", err
	}

	// 创建任务
	task, err := domain.NewScanTask(
		scanType,
		req.AssetID,
		assetType,
		req.Options,
		domain.TaskPriority(req.Priority),
		userID,
	)
	if err != nil {
		return "", err
	}

	// 转换为仓库实体
	repoTask := &repository.Task{
		Type:     string(task.Type),
		Status:   string(task.Status),
		Priority: int(task.Priority),
		SubType:  task.SubType,
		Payload:  task.Payload,
		UserID:   task.UserID,
	}

	// 保存到数据库
	if err := s.taskRepo.Create(ctx, repoTask); err != nil {
		return "", err
	}

	// 发布到消息队列
	if err := s.taskPublisher.PublishScanTask(ctx, task.SubType, int(task.Priority), task.Payload); err != nil {
		// 如果发布失败，更新任务状态为失败
		_ = s.taskRepo.UpdateStatus(ctx, repoTask.ID, string(domain.TaskStatusFailed), "Failed to publish task to message queue")
		return "", err
	}

	return repoTask.ID, nil
}

// CreateAssetTask 创建资产任务
func (s *taskService) CreateAssetTask(ctx context.Context, req *CreateAssetTaskRequest, userID uint) (string, error) {
	// 解析资产类型
	assetType, err := domain.ParseAssetType(req.AssetType)
	if err != nil {
		return "", err
	}

	// 创建任务
	task, err := domain.NewAssetTask(
		assetType,
		req.AssetID,
		req.Operation,
		req.Data,
		userID,
	)
	if err != nil {
		return "", err
	}

	// 转换为仓库实体
	repoTask := &repository.Task{
		Type:     string(task.Type),
		Status:   string(task.Status),
		Priority: int(task.Priority),
		SubType:  task.SubType,
		Payload:  task.Payload,
		UserID:   task.UserID,
	}

	// 保存到数据库
	if err := s.taskRepo.Create(ctx, repoTask); err != nil {
		return "", err
	}

	// 发布到消息队列
	if err := s.taskPublisher.PublishAssetTask(ctx, req.Operation, task.Payload); err != nil {
		// 如果发布失败，更新任务状态为失败
		_ = s.taskRepo.UpdateStatus(ctx, repoTask.ID, string(domain.TaskStatusFailed), "Failed to publish task to message queue")
		return "", err
	}

	return repoTask.ID, nil
}

// BatchCreateScanTasks 批量创建扫描任务
func (s *taskService) BatchCreateScanTasks(ctx context.Context, req *BatchCreateScanTaskRequest, userID uint) ([]string, error) {
	tasks := make([]*repository.Task, 0, len(req.Tasks))
	taskIDs := make([]string, 0, len(req.Tasks))

	// 创建所有任务
	for _, taskReq := range req.Tasks {
		// 解析扫描类型
		scanType, err := domain.ParseScanType(taskReq.ScanType)
		if err != nil {
			return nil, err
		}

		// 解析资产类型
		assetType, err := domain.ParseAssetType(taskReq.AssetType)
		if err != nil {
			return nil, err
		}

		// 创建任务
		task, err := domain.NewScanTask(
			scanType,
			taskReq.AssetID,
			assetType,
			taskReq.Options,
			domain.TaskPriority(taskReq.Priority),
			userID,
		)
		if err != nil {
			return nil, err
		}

		// 转换为仓库实体
		repoTask := &repository.Task{
			Type:     string(task.Type),
			Status:   string(task.Status),
			Priority: int(task.Priority),
			SubType:  task.SubType,
			Payload:  task.Payload,
			UserID:   task.UserID,
		}

		tasks = append(tasks, repoTask)
	}

	// 批量保存到数据库
	if err := s.taskRepo.BatchCreate(ctx, tasks); err != nil {
		return nil, err
	}

	// 发布到消息队列
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)

		// 解析任务类型
		if task.Type == string(domain.TaskTypeScan) {
			if err := s.taskPublisher.PublishScanTask(ctx, task.SubType, task.Priority, task.Payload); err != nil {
				// 如果发布失败，更新任务状态为失败
				_ = s.taskRepo.UpdateStatus(ctx, task.ID, string(domain.TaskStatusFailed), fmt.Sprintf("Failed to publish task to message queue: %v", err))
				continue
			}
		}
	}

	return taskIDs, nil
}

// BatchCreateAssetTasks 批量创建资产任务
func (s *taskService) BatchCreateAssetTasks(ctx context.Context, req *BatchCreateAssetTaskRequest, userID uint) ([]string, error) {
	tasks := make([]*repository.Task, 0, len(req.Tasks))
	taskIDs := make([]string, 0, len(req.Tasks))

	// 创建所有任务
	for _, taskReq := range req.Tasks {
		// 解析资产类型
		assetType, err := domain.ParseAssetType(taskReq.AssetType)
		if err != nil {
			return nil, err
		}

		// 创建任务
		task, err := domain.NewAssetTask(
			assetType,
			taskReq.AssetID,
			taskReq.Operation,
			taskReq.Data,
			userID,
		)
		if err != nil {
			return nil, err
		}

		// 转换为仓库实体
		repoTask := &repository.Task{
			Type:     string(task.Type),
			Status:   string(task.Status),
			Priority: int(task.Priority),
			SubType:  task.SubType,
			Payload:  task.Payload,
			UserID:   task.UserID,
		}

		tasks = append(tasks, repoTask)
	}

	// 批量保存到数据库
	if err := s.taskRepo.BatchCreate(ctx, tasks); err != nil {
		return nil, err
	}

	// 发布到消息队列
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)

		// 解析任务类型
		if task.Type == string(domain.TaskTypeAsset) {
			// 解析操作类型
			var payload domain.AssetTaskPayload
			if err := json.Unmarshal(task.Payload, &payload); err != nil {
				_ = s.taskRepo.UpdateStatus(ctx, task.ID, string(domain.TaskStatusFailed), fmt.Sprintf("Failed to unmarshal payload: %v", err))
				continue
			}

			if err := s.taskPublisher.PublishAssetTask(ctx, payload.Operation, task.Payload); err != nil {
				// 如果发布失败，更新任务状态为失败
				_ = s.taskRepo.UpdateStatus(ctx, task.ID, string(domain.TaskStatusFailed), fmt.Sprintf("Failed to publish task to message queue: %v", err))
				continue
			}
		}
	}

	return taskIDs, nil
}

// GetTaskStatus 获取任务状态
func (s *taskService) GetTaskStatus(ctx context.Context, id string) (*TaskDTO, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertToDTO(task), nil
}

// BatchGetTaskStatus 批量获取任务状态
func (s *taskService) BatchGetTaskStatus(ctx context.Context, ids []string) ([]*TaskDTO, error) {
	if len(ids) == 0 {
		return nil, errors.New("no task ids provided")
	}

	tasks := make([]*TaskDTO, 0, len(ids))
	for _, id := range ids {
		task, err := s.GetTaskStatus(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get status for task %s: %w", id, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskStatus 更新任务状态
func (s *taskService) UpdateTaskStatus(ctx context.Context, id string, req *UpdateTaskStatusRequest) error {
	return s.taskRepo.UpdateStatus(ctx, id, req.Status, req.ErrorMsg)
}

// ListTasks 列出任务
func (s *taskService) ListTasks(ctx context.Context, params *TaskQueryParams) (*TaskListResponse, error) {
	// 计算分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 || params.Size > 100 {
		params.Size = 10
	}
	offset := (params.Page - 1) * params.Size

	var tasks []*repository.Task
	var total int64
	var err error

	// 根据参数查询
	if params.Status != "" {
		tasks, total, err = s.taskRepo.FindByStatus(ctx, params.Status, params.Size, offset)
	} else if params.UserID > 0 {
		tasks, total, err = s.taskRepo.FindByUserID(ctx, params.UserID, params.Size, offset)
	} else {
		return nil, errors.New("invalid query parameters")
	}

	if err != nil {
		return nil, err
	}

	// 转换为DTO
	dtos := make([]TaskDTO, len(tasks))
	for i, task := range tasks {
		dto := convertToDTO(task)
		dtos[i] = *dto
	}

	return &TaskListResponse{
		Total: total,
		Items: dtos,
	}, nil
}

// CancelTask 取消任务
func (s *taskService) CancelTask(ctx context.Context, id string) error {
	// 获取任务信息
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	// 检查任务状态是否为等待扫描
	if task.Status != string(domain.TaskStatusPending) {
		return fmt.Errorf("task is not in pending status, current status: %s", task.Status)
	}

	// 从消息队列中删除任务
	if task.Type == string(domain.TaskTypeScan) {
		// 解析任务类型和优先级
		var payload domain.ScanTaskPayload
		if err := json.Unmarshal(task.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal task payload: %w", err)
		}

		// 从消息队列中删除任务
		if err := s.taskPublisher.DeleteScanTask(ctx, payload.ScanType.String(), task.Priority, task.Payload); err != nil {
			return fmt.Errorf("failed to delete task from message queue: %w", err)
		}
	} else if task.Type == string(domain.TaskTypeAsset) {
		// 解析任务操作类型
		var payload domain.AssetTaskPayload
		if err := json.Unmarshal(task.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal task payload: %w", err)
		}

		// 从消息队列中删除任务
		if err := s.taskPublisher.DeleteAssetTask(ctx, payload.Operation, task.Payload); err != nil {
			return fmt.Errorf("failed to delete task from message queue: %w", err)
		}
	}

	// 更新任务状态为已取消
	if err := s.taskRepo.UpdateStatus(ctx, id, string(domain.TaskStatusCancelled), "Task cancelled by user"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// BatchCancelTasks 批量取消任务
func (s *taskService) BatchCancelTasks(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, errors.New("no task ids provided")
	}

	failedIDs := make([]string, 0)
	for _, id := range ids {
		if err := s.CancelTask(ctx, id); err != nil {
			failedIDs = append(failedIDs, id)
		}
	}

	if len(failedIDs) > 0 {
		return failedIDs, fmt.Errorf("failed to cancel tasks: %v", failedIDs)
	}

	return nil, nil
}
