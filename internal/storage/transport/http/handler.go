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
	r.GET("/api/v1/sast/:task_id", h.handleSASTQuery)
	r.POST("/api/v1/sast/batch", h.handleSASTBatchQuery)

	// DAST routes
	r.GET("/api/v1/dast/:task_id", h.handleDASTQuery)
	r.POST("/api/v1/dast/batch", h.handleDASTBatchQuery)

	// SCA routes
	r.GET("/api/v1/sca/:task_id", h.handleSCAQuery)
	r.POST("/api/v1/sca/batch", h.handleSCABatchQuery)
}

// handleSASTQuery handles SAST query request
func (h *Handler) handleSASTQuery(c *gin.Context) {
	taskID := c.Param("task_id")
	processor, err := h.factory.GetProcessor(domain.ScanTypeStaticCodeAnalysis)
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

// handleSASTBatchQuery handles SAST batch query request
func (h *Handler) handleSASTBatchQuery(c *gin.Context) {
	var req dto.SASTBatchQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
		return
	}

	processor, err := h.factory.GetProcessor(domain.ScanTypeStaticCodeAnalysis)
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

// handleDASTQuery handles DAST query request
func (h *Handler) handleDASTQuery(c *gin.Context) {
	taskID := c.Param("task_id")
	processor, err := h.factory.GetProcessor(domain.ScanTypeDast)
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

// handleDASTBatchQuery handles DAST batch query request
func (h *Handler) handleDASTBatchQuery(c *gin.Context) {
	var req dto.DASTBatchQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
		return
	}

	processor, err := h.factory.GetProcessor(domain.ScanTypeDast)
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

// handleSCAQuery handles SCA query request
func (h *Handler) handleSCAQuery(c *gin.Context) {
	taskID := c.Param("task_id")
	processor, err := h.factory.GetProcessor(domain.ScanTypeSca)
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

// handleSCABatchQuery handles SCA batch query request
func (h *Handler) handleSCABatchQuery(c *gin.Context) {
	var req dto.SCABatchQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(400, err.Error()))
		return
	}

	processor, err := h.factory.GetProcessor(domain.ScanTypeSca)
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
