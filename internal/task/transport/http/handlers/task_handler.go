package handlers

import (
	"net/http"

	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TaskHandler 处理任务相关请求
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler 创建任务处理程序
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateScanTask 处理创建扫描任务请求
func (h *TaskHandler) CreateScanTask(c *gin.Context) {
	var req service.CreateScanTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户ID（来自JWT中间件）
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
		return
	}

	// 调用服务层创建任务
	taskID, err := h.taskService.CreateScanTask(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		logger.Logger.Error("failed to create scan task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create scan task: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}

// CreateAssetTask 处理创建资产任务请求
func (h *TaskHandler) CreateAssetTask(c *gin.Context) {
	var req service.CreateAssetTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户ID（来自JWT中间件）
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
		return
	}

	// 调用服务层创建任务
	taskID, err := h.taskService.CreateAssetTask(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		logger.Logger.Error("failed to create asset task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset task: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}

// BatchCreateScanTasks 处理批量创建扫描任务请求
func (h *TaskHandler) BatchCreateScanTasks(c *gin.Context) {
	var req service.BatchCreateScanTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户ID（来自JWT中间件）
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
		return
	}

	// 调用服务层批量创建任务
	taskIDs, err := h.taskService.BatchCreateScanTasks(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		logger.Logger.Error("failed to batch create scan tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch create scan tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_ids": taskIDs})
}

// BatchCreateAssetTasks 处理批量创建资产任务请求
func (h *TaskHandler) BatchCreateAssetTasks(c *gin.Context) {
	var req service.BatchCreateAssetTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户ID（来自JWT中间件）
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
		return
	}

	// 调用服务层批量创建任务
	taskIDs, err := h.taskService.BatchCreateAssetTasks(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		logger.Logger.Error("failed to batch create asset tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch create asset tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_ids": taskIDs})
}

// GetTaskStatus 处理获取任务状态请求
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	// 调用服务层获取任务状态
	task, err := h.taskService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		logger.Logger.Error("failed to get task status", zap.Error(err), zap.String("task_id", taskID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task status"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskStatus 处理更新任务状态请求
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	var req service.UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层更新任务状态
	if err := h.taskService.UpdateTaskStatus(c.Request.Context(), taskID, &req); err != nil {
		logger.Logger.Error("failed to update task status", zap.Error(err), zap.String("task_id", taskID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ListTasks 处理列出任务请求
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var params service.TaskQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层列出任务
	result, err := h.taskService.ListTasks(c.Request.Context(), &params)
	if err != nil {
		logger.Logger.Error("failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelTask 处理取消任务请求
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	// 调用服务层取消任务
	if err := h.taskService.CancelTask(c.Request.Context(), taskID); err != nil {
		logger.Logger.Error("failed to cancel task", zap.Error(err), zap.String("task_id", taskID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel task: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task cancelled successfully"})
}

// BatchGetTaskStatus 处理批量获取任务状态请求
func (h *TaskHandler) BatchGetTaskStatus(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层批量获取任务状态
	tasks, err := h.taskService.BatchGetTaskStatus(c.Request.Context(), req.TaskIDs)
	if err != nil {
		logger.Logger.Error("failed to batch get task status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch get task status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// BatchCancelTasks 处理批量取消任务请求
func (h *TaskHandler) BatchCancelTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层批量取消任务
	failedIDs, err := h.taskService.BatchCancelTasks(c.Request.Context(), req.TaskIDs)
	if err != nil {
		logger.Logger.Error("failed to batch cancel tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "failed to batch cancel tasks: " + err.Error(),
			"failed_ids": failedIDs,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tasks cancelled successfully"})
}
