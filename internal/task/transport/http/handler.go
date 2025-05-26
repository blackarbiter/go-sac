package http

import (
	"net/http"

	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/gin-gonic/gin"
)

// Response 定义HTTP响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler 定义HTTP处理器
type Handler struct {
	taskService service.TaskService
}

// NewHandler 创建新的HTTP处理器
func NewHandler(taskService service.TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

// CancelTask godoc
// @Summary 取消任务
// @Description 取消一个等待扫描的任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /tasks/{id}/cancel [post]
func (h *Handler) CancelTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    http.StatusBadRequest,
			Message: "task id is required",
		})
		return
	}

	if err := h.taskService.CancelTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "task cancelled successfully",
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.POST("/batch", h.BatchCreateTasks)
		tasks.GET("/:id", h.GetTaskStatus)
		tasks.PUT("/:id/status", h.UpdateTaskStatus)
		tasks.GET("", h.ListTasks)
		tasks.POST("/:id/cancel", h.CancelTask)
	}
}
