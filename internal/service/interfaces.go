package service

import (
	"context"

	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
)

type UserService interface {
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	SelectRandomReviewers(ctx context.Context, teamName, excludeUserID string, count int) ([]string, error)
	FindReplacementReviewer(ctx context.Context, currentReviewerID string, excludeUserIDs []string) (string, error)
}

type TeamService interface {
	AddTeam(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	DeactivateMembers(ctx context.Context, teamName string) error
}

type PullRequestService interface {
	Create(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (*domain.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, *string, error)
	GetReviewerPRs(ctx context.Context, userID string) ([]domain.PullRequest, error)
	GetStats(ctx context.Context) (*domain.PRStats, error)
}
