package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

func (h *Handler) CreatePullRequest(c *gin.Context) {
	var req struct {
		PullRequestID   string `json:"pull_request_id" binding:"required"`
		PullRequestName string `json:"pull_request_name" binding:"required"`
		AuthorID        string `json:"author_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create PR request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("creating pull request",
		zap.String("pr_id", req.PullRequestID),
		zap.String("author_id", req.AuthorID))

	pr, err := h.prService.Create(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case domain.PRExists:
				h.logger.Warn("PR already exists", zap.String("pr_id", req.PullRequestID))
				c.JSON(http.StatusBadRequest, errorResponse(domain.PRExists, appErr.Message))
				return
			case domain.NotFound:
				h.logger.Warn("author not found", zap.String("author_id", req.AuthorID))
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		h.logger.Error("failed to create PR", zap.Error(err), zap.String("pr_id", req.PullRequestID))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("PR created successfully",
		zap.String("pr_id", req.PullRequestID),
		zap.Strings("reviewers", pr.AssignedReviewers))

	c.JSON(http.StatusCreated, gin.H{"pull_request": pr})
}

func (h *Handler) MergePullRequest(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid merge PR request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("merging pull request", zap.String("pr_id", req.PullRequestID))

	pr, err := h.prService.Merge(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			if appErr.Code == domain.NotFound {
				h.logger.Warn("PR not found", zap.String("pr_id", req.PullRequestID))
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		h.logger.Error("failed to merge PR", zap.Error(err), zap.String("pr_id", req.PullRequestID))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("PR merged successfully", zap.String("pr_id", req.PullRequestID))
	c.JSON(http.StatusOK, gin.H{"pull_request": pr})
}

func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
		OldUserID     string `json:"old_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid reassign request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("reassigning reviewer",
		zap.String("pr_id", req.PullRequestID),
		zap.String("old_user_id", req.OldUserID))

	pr, newReviewerID, err := h.prService.Reassign(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case domain.PRMerged:
				h.logger.Warn("cannot reassign on merged PR", zap.String("pr_id", req.PullRequestID))
				c.JSON(http.StatusBadRequest, errorResponse(domain.PRMerged, appErr.Message))
				return
			case domain.NotAssigned:
				h.logger.Warn("reviewer not assigned to PR",
					zap.String("pr_id", req.PullRequestID),
					zap.String("user_id", req.OldUserID))
				c.JSON(http.StatusBadRequest, errorResponse(domain.NotAssigned, appErr.Message))
				return
			case domain.NoCandidate:
				h.logger.Warn("no replacement candidate found", zap.String("pr_id", req.PullRequestID))
				c.JSON(http.StatusBadRequest, errorResponse(domain.NoCandidate, appErr.Message))
				return
			case domain.NotFound:
				h.logger.Warn("PR or user not found", zap.String("pr_id", req.PullRequestID))
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		h.logger.Error("failed to reassign reviewer", zap.Error(err), zap.String("pr_id", req.PullRequestID))
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("reviewer reassigned successfully",
		zap.String("pr_id", req.PullRequestID),
		zap.String("old_reviewer", req.OldUserID),
		zap.String("new_reviewer", *newReviewerID))

	c.JSON(http.StatusOK, gin.H{
		"pull_request":    pr,
		"new_reviewer_id": *newReviewerID,
	})
}
