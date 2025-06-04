package handlers

import (
	"net/http"
	"strconv"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/service"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssetHandler 处理资产相关请求
type AssetHandler struct {
	assetService service.AssetService
}

// NewAssetHandler 创建资产处理程序
func NewAssetHandler(assetService service.AssetService) *AssetHandler {
	return &AssetHandler{
		assetService: assetService,
	}
}

// CreateAsset 处理创建资产请求
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req dto.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户ID（来自JWT中间件）
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
		return
	}

	// 调用服务层创建资产
	asset, err := h.assetService.CreateAsset(c.Request.Context(), &req)
	if err != nil {
		logger.Logger.Error("failed to create asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, asset)
}

// GetAsset 处理获取资产请求
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset id"})
		return
	}

	// 调用服务层获取资产
	asset, err := h.assetService.GetAsset(c.Request.Context(), uint(id))
	if err != nil {
		logger.Logger.Error("failed to get asset", zap.Error(err), zap.Uint("asset_id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get asset"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// ListAssets 处理列出资产请求
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var req dto.ListAssetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层列出资产
	assets, err := h.assetService.ListAssets(c.Request.Context(), &req)
	if err != nil {
		logger.Logger.Error("failed to list assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assets"})
		return
	}

	c.JSON(http.StatusOK, assets)
}

// UpdateAsset 处理更新资产请求
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset id"})
		return
	}

	var req dto.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层更新资产
	asset, err := h.assetService.UpdateAsset(c.Request.Context(), uint(id), &req)
	if err != nil {
		logger.Logger.Error("failed to update asset", zap.Error(err), zap.Uint("asset_id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// DeleteAsset 处理删除资产请求
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset id"})
		return
	}

	// 调用服务层删除资产
	err = h.assetService.DeleteAsset(c.Request.Context(), uint(id))
	if err != nil {
		logger.Logger.Error("failed to delete asset", zap.Error(err), zap.Uint("asset_id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete asset"})
		return
	}

	c.Status(http.StatusNoContent)
}
