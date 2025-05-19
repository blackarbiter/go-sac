package http

import (
	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/blackarbiter/go-sac/internal/task/transport/http/handlers"

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
		JWTValidation(), // JWT验证
	)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	api := r.Group("/api/v1")
	{
		tasks := api.Group("/tasks")
		{
			tasks.POST("", handlers.CreateTask)
			tasks.GET("/:id", handlers.GetTaskStatus)
		}
	}

	return r
}
