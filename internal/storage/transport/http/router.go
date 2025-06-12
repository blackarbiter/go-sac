package http

import (
	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/internal/storage/transport/http/handlers"
	"github.com/blackarbiter/go-sac/pkg/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 依赖
var (
	storageService service.StorageService
)

// SetStorageService 设置存储服务
func SetStorageService(svc service.StorageService) {
	storageService = svc
	// 初始化处理程序
	handlers.InitHandlers(storageService)
}

// NewRouter 创建新的路由
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
		storage := api.Group("/storage")
		{
			// 获取存储处理器
			h := handlers.GetStorageHandler()

			// 存储管理
			storage.POST("", h.CreateStorage)            // 创建存储
			storage.GET("/:id", h.GetStorage)            // 获取存储信息
			storage.PUT("/:id", h.UpdateStorage)         // 更新存储信息
			storage.DELETE("/:id", h.DeleteStorage)      // 删除存储
			storage.GET("", h.ListStorages)              // 列出存储
			storage.POST("/:id/upload", h.UploadFile)    // 上传文件
			storage.GET("/:id/download", h.DownloadFile) // 下载文件
		}
	}

	return r
}
