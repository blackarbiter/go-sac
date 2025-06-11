package http

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"net/http"
	"strconv"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/service"
	"github.com/gin-gonic/gin"
)

// Handler 处理资产相关的HTTP请求
type Handler struct {
	binder  *AssetBinder
	factory service.AssetProcessorFactory
}

// NewHandler 创建资产处理器实例
func NewHandler(binder *AssetBinder, factory service.AssetProcessorFactory) *Handler {
	return &Handler{
		binder:  binder,
		factory: factory,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	assets := r.Group("/api/v1/assets")
	{
		// 创建资产
		assets.POST("/:type", h.CreateAsset)

		// 更新资产
		assets.PUT("/:type/:id", h.UpdateAsset)

		// 获取资产
		assets.GET("/:type/:id", h.GetAsset)

		// 删除资产
		assets.DELETE("/:type/:id", h.DeleteAsset)

		// 列出资产
		assets.GET("/:type", h.ListAssets)
	}
}

// CreateAsset 创建资产
func (h *Handler) CreateAsset(c *gin.Context) {
	assetType := c.Param("type")

	// 1. 绑定请求
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	req, err := h.binder.Bind(assetType, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 获取处理器
	parseAssetType, _ := domain.ParseAssetType(assetType)
	processor, err := h.factory.GetProcessor(parseAssetType.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported asset type"})
		return
	}

	// 3. 转换请求为基础资产和扩展资产
	baseReq, ok := req.(dto.BaseRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	baseAsset := baseReq.ToBaseAsset(assetType)

	// 4. 执行创建
	response, err := processor.Create(c.Request.Context(), baseAsset, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateAsset 更新资产
func (h *Handler) UpdateAsset(c *gin.Context) {
	assetType := c.Param("type")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset ID"})
		return
	}

	// 1. 绑定请求
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	req, err := h.binder.Bind(assetType, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 获取处理器
	parseAssetType, _ := domain.ParseAssetType(assetType)
	processor, err := h.factory.GetProcessor(parseAssetType.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported asset type"})
		return
	}

	// 3. 转换请求为基础资产和扩展资产
	baseReq, ok := req.(dto.BaseRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	baseAsset := baseReq.ToBaseAsset(assetType)

	// 4. 执行更新
	if err := processor.Update(c.Request.Context(), uint(id), baseAsset, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAsset 获取资产
func (h *Handler) GetAsset(c *gin.Context) {
	assetType := c.Param("type")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset ID"})
		return
	}

	// 1. 获取处理器
	parseAssetType, err := domain.ParseAssetType(assetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset type"})
		return
	}
	processor, err := h.factory.GetProcessor(parseAssetType.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported asset type"})
		return
	}

	// 2. 执行获取
	base, extension, err := processor.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"base":      base,
		"extension": extension,
	})
}

// DeleteAsset 删除资产
func (h *Handler) DeleteAsset(c *gin.Context) {
	assetType := c.Param("type")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset ID"})
		return
	}

	// 1. 获取处理器
	parseAssetType, err := domain.ParseAssetType(assetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset type"})
		return
	}
	processor, err := h.factory.GetProcessor(parseAssetType.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported asset type"})
		return
	}

	// 2. 执行删除
	if err := processor.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAssets 列出资产
func (h *Handler) ListAssets(c *gin.Context) {
	assetType := c.Param("type")

	// 1. 获取处理器
	parseAssetType, err := domain.ParseAssetType(assetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset type"})
		return
	}
	processor, err := h.factory.GetProcessor(parseAssetType.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported asset type"})
		return
	}

	// 2. 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 3. 构建过滤条件
	filters := make(map[string]interface{})
	if projectID := c.Query("project_id"); projectID != "" {
		if id, err := strconv.ParseUint(projectID, 10, 32); err == nil {
			filters["project_id"] = uint(id)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// 4. 执行列表查询
	assets, total, err := processor.List(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": assets,
	})
}
