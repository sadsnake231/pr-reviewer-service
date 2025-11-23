package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

func (h *Handler) SetUserActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive *bool  `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid set active request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("setting user active status", zap.String("user_id", req.UserID), zap.Bool("is_active", *req.IsActive))

	user, err := h.userService.SetActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			if appErr.Code == domain.NotFound {
				h.logger.Warn("user not found", zap.String("user_id", req.UserID))
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		h.logger.Error("failed to set user active", zap.Error(err), zap.String("user_id", req.UserID))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("user active status updated", zap.String("user_id", req.UserID), zap.Bool("is_active", user.IsActive))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) GetUserReviews(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		h.logger.Warn("user_id parameter missing")
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "user_id is required"))
		return
	}

	h.logger.Debug("getting reviews for user", zap.String("user_id", userID))

	prs, err := h.prService.GetReviewerPRs(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user reviews", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("retrieved user reviews", zap.String("user_id", userID), zap.Int("count", len(prs)))
	c.JSON(http.StatusOK, gin.H{"pull_requests": prs})
}

func (h *Handler) DeactivateTeamMembers(c *gin.Context) {
	var req struct {
		TeamName string `json:"team_name" binding:"required"`
	}

	start := time.Now()

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid deactivate team request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("deactivating team members", zap.String("team_name", req.TeamName))

	err := h.teamService.DeactivateMembers(c.Request.Context(), req.TeamName)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			if appErr.Code == domain.NotFound {
				h.logger.Warn("team not found for deactivation", zap.String("team_name", req.TeamName))
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		h.logger.Error("failed to deactivate team members", zap.Error(err), zap.String("team_name", req.TeamName))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("team members deactivated successfully", zap.String("team_name", req.TeamName))
	c.JSON(http.StatusOK, gin.H{"message": "team members deactivated", "execution time": time.Since(start).Milliseconds()})
}
