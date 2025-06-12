package handlers

import (
	"net/http"
	"strconv"

	"github.com/blackarbiter/go-sac/internal/storage/dto"
	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/gin-gonic/gin"
)

// StorageHandler 存储服务处理函数
type StorageHandler struct {
	storageService service.StorageService
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(storageService service.StorageService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
	}
}

// CreateStorage 创建存储
func (h *StorageHandler) CreateStorage(c *gin.Context) {
	var req dto.StorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	storage, err := h.storageService.CreateStorage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, storage)
}

// GetStorage 获取存储信息
func (h *StorageHandler) GetStorage(c *gin.Context) {
	id := c.Param("id")
	storage, err := h.storageService.GetStorage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

// UpdateStorage 更新存储信息
func (h *StorageHandler) UpdateStorage(c *gin.Context) {
	id := c.Param("id")
	var req dto.StorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	storage, err := h.storageService.UpdateStorage(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

// DeleteStorage 删除存储
func (h *StorageHandler) DeleteStorage(c *gin.Context) {
	id := c.Param("id")
	if err := h.storageService.DeleteStorage(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListStorages 列出存储
func (h *StorageHandler) ListStorages(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	storages, err := h.storageService.ListStorages(c.Request.Context(), &dto.StorageQueryParams{
		Page: offset,
		Size: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, storages)
}

// UploadFile 上传文件
func (h *StorageHandler) UploadFile(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	if err := h.storageService.UploadFile(c.Request.Context(), id, file); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DownloadFile 下载文件
func (h *StorageHandler) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	file, err := h.storageService.DownloadFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.File(file)
}
