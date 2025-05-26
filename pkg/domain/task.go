package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// TaskStatus 定义任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待执行
	TaskStatusRunning   TaskStatus = "running"   // 正在执行
	TaskStatusCompleted TaskStatus = "completed" // 执行完成
	TaskStatusFailed    TaskStatus = "failed"    // 执行失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
)

// TaskType 定义任务类型
type TaskType string

const (
	TaskTypeScan  TaskType = "scan"  // 扫描任务
	TaskTypeAsset TaskType = "asset" // 资产更新任务
)

// TaskPriority 定义任务优先级
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityMedium TaskPriority = 1
	PriorityHigh   TaskPriority = 2
)

func (t TaskPriority) String() string {
	switch t {
	case 2:
		return "high"
	case 1:
		return "medium"
	case 0:
		return "low"
	default:
		return "low"
	}
}

// Task 表示一个任务实体
type Task struct {
	ID          string       `json:"id"`
	Type        TaskType     `json:"type"`         // 任务类型：scan或asset
	Status      TaskStatus   `json:"status"`       // 任务状态
	Priority    TaskPriority `json:"priority"`     // 任务优先级
	SubType     string       `json:"sub_type"`     // 子类型：对应ScanType或AssetType的字符串表示
	Payload     []byte       `json:"payload"`      // 任务载荷，JSON格式
	UserID      uint         `json:"user_id"`      // 创建任务的用户ID
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`   // 更新时间
	StartedAt   *time.Time   `json:"started_at"`   // 开始执行时间
	CompletedAt *time.Time   `json:"completed_at"` // 完成时间
	ErrorMsg    string       `json:"error_msg"`    // 错误信息
	RetryCount  int          `json:"retry_count"`  // 重试次数
}

// ScanTaskPayload 扫描任务的载荷
type ScanTaskPayload struct {
	TaskID    string                 `json:"task_id"`    // 任务ID
	AssetID   string                 `json:"asset_id"`   // 资产ID
	AssetType AssetType              `json:"asset_type"` // 资产类型
	ScanType  ScanType               `json:"scan_type"`  // 扫描类型
	Options   map[string]interface{} `json:"options"`    // 扫描选项
}

// AssetTaskPayload 资产更新任务的载荷
type AssetTaskPayload struct {
	TaskID    string                 `json:"task_id"`    // 任务ID
	AssetID   string                 `json:"asset_id"`   // 资产ID
	AssetType AssetType              `json:"asset_type"` // 资产类型
	Operation string                 `json:"operation"`  // 操作类型：create, update, delete
	Data      map[string]interface{} `json:"data"`       // 资产数据
}

// NewScanTask 创建一个新的扫描任务
func NewScanTask(scanType ScanType, assetID string, assetType AssetType, options map[string]interface{}, priority TaskPriority, userID uint) (*Task, error) {
	taskID := uuid.New().String()
	payload := ScanTaskPayload{
		TaskID:    taskID,
		AssetID:   assetID,
		AssetType: assetType,
		ScanType:  scanType,
		Options:   options,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Task{
		ID:        taskID,
		Type:      TaskTypeScan,
		Status:    TaskStatusPending,
		Priority:  priority,
		SubType:   scanType.String(),
		Payload:   payloadBytes,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewAssetTask 创建一个新的资产更新任务
func NewAssetTask(assetType AssetType, assetID, operation string, data map[string]interface{}, userID uint) (*Task, error) {
	taskID := uuid.New().String()
	payload := AssetTaskPayload{
		TaskID:    taskID,
		AssetID:   assetID,
		AssetType: assetType,
		Operation: operation,
		Data:      data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Task{
		ID:        taskID,
		Type:      TaskTypeAsset,
		Status:    TaskStatusPending,
		Priority:  PriorityMedium, // 资产任务默认中优先级
		SubType:   assetType.String(),
		Payload:   payloadBytes,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetScanPayload 获取扫描任务载荷
func (t *Task) GetScanPayload() (*ScanTaskPayload, error) {
	if t.Type != TaskTypeScan {
		return nil, ErrInvalidTaskType
	}

	var payload ScanTaskPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// GetAssetPayload 获取资产任务载荷
func (t *Task) GetAssetPayload() (*AssetTaskPayload, error) {
	if t.Type != TaskTypeAsset {
		return nil, ErrInvalidTaskType
	}

	var payload AssetTaskPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// ErrInvalidTaskType 错误定义
var (
	ErrInvalidTaskType = NewDomainError("invalid task type")
)

// DomainError 领域错误
type DomainError struct {
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

func NewDomainError(message string) DomainError {
	return DomainError{Message: message}
}
