package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"go.uber.org/zap"
)

// TaskStatusUpdaterImpl 实现任务状态更新接口
type TaskStatusUpdaterImpl struct {
	apiBaseURL string
	authToken  string
	client     *http.Client
}

// NewTaskStatusUpdater 创建任务状态更新器
func NewTaskStatusUpdater(apiBaseURL, authToken string) *TaskStatusUpdaterImpl {
	return &TaskStatusUpdaterImpl{
		apiBaseURL: apiBaseURL,
		authToken:  authToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UpdateTaskStatus 更新任务状态
func (u *TaskStatusUpdaterImpl) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/status", u.apiBaseURL, taskID)

	// 构建请求体
	reqBody := map[string]string{
		"status": string(status),
	}

	// 序列化请求体
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+u.authToken)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Logger.Info("Task status updated",
		zap.String("taskID", taskID),
		zap.String("status", string(status)))

	return nil
}
