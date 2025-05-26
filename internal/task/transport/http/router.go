package http

import (
	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/blackarbiter/go-sac/internal/task/transport/http/handlers"
	"github.com/blackarbiter/go-sac/pkg/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 依赖
var (
	taskService service.TaskService
)

// SetTaskService 设置任务服务
func SetTaskService(svc service.TaskService) {
	taskService = svc
	// 初始化处理程序
	handlers.InitHandlers(taskService)
}

func NewRouter() *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.JWTValidation(), // JWT验证
	)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	api := r.Group("/api/v1")
	{
		tasks := api.Group("/tasks")
		{
			// 获取任务处理器
			h := handlers.GetTaskHandler()

			// 任务管理
			tasks.POST("/scan", h.CreateScanTask)               // 创建扫描任务
			tasks.POST("/asset", h.CreateAssetTask)             // 创建资产任务
			tasks.POST("/scan/batch", h.BatchCreateScanTasks)   // 批量创建扫描任务
			tasks.POST("/asset/batch", h.BatchCreateAssetTasks) // 批量创建资产任务
			tasks.GET("/:id", h.GetTaskStatus)                  // 获取任务状态
			tasks.POST("/batch/status", h.BatchGetTaskStatus)   // 批量获取任务状态
			tasks.PUT("/:id/status", h.UpdateTaskStatus)        // 更新任务状态
			tasks.GET("", h.ListTasks)                          // 列出任务
			tasks.POST("/:id/cancel", h.CancelTask)             // 取消任务
			tasks.POST("/batch/cancel", h.BatchCancelTasks)     // 批量取消任务
		}
	}

	return r
}
