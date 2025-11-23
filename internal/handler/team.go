package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

func (h *Handler) AddTeam(c *gin.Context) {
	var req struct {
		TeamName string              `json:"team_name" binding:"required"`
		Members  []domain.TeamMember `json:"members" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid team add request", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "invalid request"))
		return
	}

	h.logger.Info("adding team", zap.String("team_name", req.TeamName), zap.Int("members_count", len(req.Members)))

	members := make([]domain.User, len(req.Members))
	for i, m := range req.Members {
		members[i] = domain.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		}
	}

	team := &domain.Team{
		TeamName: req.TeamName,
		Members:  members,
	}

	result, err := h.teamService.AddTeam(c.Request.Context(), team)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case domain.TeamExists, domain.NoMembers:
				c.JSON(http.StatusBadRequest, errorResponse(appErr.Code, appErr.Message))
				return
			case domain.NonUniqueUser:
				c.JSON(http.StatusConflict, errorResponse(appErr.Code, appErr.Message))
				return
			}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	h.logger.Info("team added successfully", zap.String("team_name", req.TeamName))
	c.JSON(http.StatusCreated, gin.H{"team": result})
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.logger.Warn("team_name parameter missing")
		c.JSON(http.StatusBadRequest, errorResponse(domain.NotFound, "team_name is required"))
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			if appErr.Code == domain.NotFound {
				c.JSON(http.StatusNotFound, errorResponse(domain.NotFound, appErr.Message))
				return
			}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(domain.NotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, team)
}
