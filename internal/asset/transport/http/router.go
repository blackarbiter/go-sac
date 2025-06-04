package http

import (
	"github.com/blackarbiter/go-sac/internal/asset/service"
	"github.com/blackarbiter/go-sac/internal/asset/transport/http/handlers"
	"github.com/blackarbiter/go-sac/pkg/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 依赖
var (
	assetService service.AssetService
)

// SetAssetService 设置资产服务
func SetAssetService(svc service.AssetService) {
	assetService = svc
	// 初始化处理程序
	handlers.InitHandlers(assetService)
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
		assets := api.Group("/assets")
		{
			// 获取资产处理器
			h := handlers.GetAssetHandler()

			// 资产管理
			assets.POST("", h.CreateAsset)       // 创建资产
			assets.GET("/:id", h.GetAsset)       // 获取资产
			assets.GET("", h.ListAssets)         // 列出资产
			assets.PUT("/:id", h.UpdateAsset)    // 更新资产
			assets.DELETE("/:id", h.DeleteAsset) // 删除资产
		}
	}

	return r
}
