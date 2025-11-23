package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	userService service.UserService
	teamService service.TeamService
	prService   service.PullRequestService
	logger      *zap.Logger
}

func NewHandler(
	userService service.UserService,
	teamService service.TeamService,
	prService service.PullRequestService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		userService: userService,
		teamService: teamService,
		prService:   prService,
		logger:      logger,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/team/add", h.AddTeam)
	router.GET("/team/get", h.GetTeam)

	router.POST("/users/setIsActive", h.SetUserActive)
	router.GET("/users/getReview", h.GetUserReviews)
	router.POST("/users/deactivateTeam", h.DeactivateTeamMembers)

	router.POST("/pullRequest/create", h.CreatePullRequest)
	router.POST("/pullRequest/merge", h.MergePullRequest)
	router.POST("/pullRequest/reassign", h.ReassignReviewer)

	router.GET("/stats", h.GetStats)
}
