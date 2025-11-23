package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

func (h *Handler) GetStats(c *gin.Context) {
	h.logger.Debug("getting statistics")

	stats, err := h.prService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("statistics retrieved successfully",
		zap.Int64("total_prs", stats.TotalPRs),
		zap.Int64("open_prs", stats.OpenPRs),
		zap.Int64("merged_prs", stats.MergedPRs))

	c.JSON(http.StatusOK, stats)
}
