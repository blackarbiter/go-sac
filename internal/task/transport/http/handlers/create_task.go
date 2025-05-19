package handlers

import (
	"net/http"
	"strconv"

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

// CreateTask 处理创建任务请求
func CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
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
	taskID, err := taskHandler.taskService.CreateTask(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		logger.Logger.Error("failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}

// GetTaskStatus 处理获取任务状态请求
func GetTaskStatus(c *gin.Context) {
	// 定义 URI 参数接收结构体（使用指针）
	var uriParams struct {
		ID uint `uri:"id"`
	}

	// 使用指针绑定参数
	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取解析后的 ID
	taskID := uriParams.ID
	logger.Logger.Info(strconv.Itoa(int(taskID)))

	// 调用服务层获取任务状态
	task, err := taskHandler.taskService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		logger.Logger.Error("failed to get task status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task status"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// 全局任务处理器实例
var taskHandler *TaskHandler
