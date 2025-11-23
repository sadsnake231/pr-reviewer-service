package repository

import (
	"context"

	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
)

type UserRepository interface {
	CreateOrUpdate(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	DeactivateTeamMembers(ctx context.Context, teamName string) error
	GetReviewerStats(ctx context.Context) (map[string]int64, error)
}

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	Exists(ctx context.Context, teamName string) (bool, error)
	GetWithMembers(ctx context.Context, teamName string, userRepo UserRepository) (*domain.Team, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	ReplaceReviewer(ctx context.Context, prID, oldID, newID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	Merge(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
	GetStats(ctx context.Context) (*domain.PRStats, error)
}
