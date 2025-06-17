package http

import (
	"net/http"

	"github.com/blackarbiter/go-sac/internal/storage/dto"
	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests
type Handler struct {
	factory service.StorageProcessorFactory
}

// NewHandler creates a new Handler instance
func NewHandler(factory service.StorageProcessorFactory) *Handler {
	return &Handler{factory: factory}
}

// RegisterRoutes registers HTTP routes
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// SAST routes
	sast := r.Group("/api/v1/sast")
	{
		sast.GET("/:task_id", h.handleQuery(domain.ScanTypeStaticCodeAnalysis))
		sast.POST("/batch", h.handleBatchQuery(domain.ScanTypeStaticCodeAnalysis))
	}

	// DAST routes
	dast := r.Group("/api/v1/dast")
	{
		dast.GET("/:task_id", h.handleQuery(domain.ScanTypeDast))
		dast.POST("/batch", h.handleBatchQuery(domain.ScanTypeDast))
	}

	// SCA routes
	sca := r.Group("/api/v1/sca")
	{
		sca.GET("/:task_id", h.handleQuery(domain.ScanTypeSca))
		sca.POST("/batch", h.handleBatchQuery(domain.ScanTypeSca))
	}
}

// handleQuery handles single query request
func (h *Handler) handleQuery(scanType domain.ScanType) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("task_id")
		processor, err := h.factory.GetProcessor(scanType)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
			return
		}

		result, err := processor.Query(c.Request.Context(), taskID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(500, err.Error()))
			return
		}

		c.JSON(http.StatusOK, dto.NewSuccessResponse(result))
	}
}

// handleBatchQuery handles batch query request
func (h *Handler) handleBatchQuery(scanType domain.ScanType) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TaskIDs []string `json:"task_ids" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
			return
		}

		processor, err := h.factory.GetProcessor(scanType)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
			return
		}

		results, err := processor.BatchQuery(c.Request.Context(), req.TaskIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(500, err.Error()))
			return
		}

		c.JSON(http.StatusOK, dto.NewSuccessResponse(results))
	}
}
