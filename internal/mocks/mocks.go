package mocks

import (
	"context"

	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateOrUpdate(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) DeactivateTeamMembers(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func (m *MockUserRepository) GetReviewerStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

func (m *MockTeamRepository) GetWithMembers(ctx context.Context, teamName string, userRepo repository.UserRepository) (*domain.Team, error) {
	args := m.Called(ctx, teamName, userRepo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) Create(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error {
	args := m.Called(ctx, pr, reviewerIDs)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPullRequestRepository) ReplaceReviewer(ctx context.Context, prID, oldID, newID string) error {
	args := m.Called(ctx, prID, oldID, newID)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPullRequestRepository) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetStats(ctx context.Context) (*domain.PRStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PRStats), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserService) SelectRandomReviewers(ctx context.Context, teamName, excludeUserID string, count int) ([]string, error) {
	args := m.Called(ctx, teamName, excludeUserID, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserService) FindReplacementReviewer(ctx context.Context, currentReviewerID string, excludeUserIDs []string) (string, error) {
	args := m.Called(ctx, currentReviewerID, excludeUserIDs)
	return args.String(0), args.Error(1)
}
