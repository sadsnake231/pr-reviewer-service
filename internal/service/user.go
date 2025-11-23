package service

import (
	"context"
	"errors"
	"math/rand"

	"github.com/jackc/pgx/v5"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/repository"
	"go.uber.org/zap"
)

type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{repo: repo, logger: logger}
}

func (s *userService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("user_id " + userID)
		}
		s.logger.Error("failed to get user by id", zap.Error(err))
		return nil, err
	}

	s.logger.Info("setting user active status", zap.String("user_id", userID), zap.Bool("is_active", isActive))
	return s.repo.SetActive(ctx, userID, isActive)
}

func (s *userService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	s.logger.Debug("getting user by id", zap.String("user_id", userID))
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	s.logger.Debug("getting active users by team", zap.String("team_name", teamName))
	return s.repo.GetActiveByTeam(ctx, teamName)
}

func (s *userService) SelectRandomReviewers(ctx context.Context, teamName, excludeUserID string, count int) ([]string, error) {
	s.logger.Info("selecting random reviewers",
		zap.String("team_name", teamName),
		zap.String("exclude_user_id", excludeUserID),
		zap.Int("count", count))

	active, err := s.GetActiveByTeam(ctx, teamName)
	if err != nil {
		s.logger.Error("failed to get active users", zap.Error(err))
		return nil, err
	}

	available := make([]domain.User, 0, len(active))
	for _, u := range active {
		if u.UserID != excludeUserID {
			available = append(available, u)
		}
	}

	if len(available) == 0 {
		s.logger.Warn("no available reviewers found", zap.String("team_name", teamName))
		return []string{}, nil
	}

	if len(available) <= count {
		result := make([]string, len(available))
		for i, u := range available {
			result[i] = u.UserID
		}
		s.logger.Info("selected all available reviewers", zap.Int("count", len(result)))
		return result, nil
	}

	selected := make([]string, 0, count)
	indices := rand.Perm(len(available))
	for i := 0; i < count && i < len(indices); i++ {
		selected = append(selected, available[indices[i]].UserID)
	}

	s.logger.Info("randomly selected reviewers", zap.Strings("reviewers", selected))
	return selected, nil
}

func (s *userService) FindReplacementReviewer(ctx context.Context, currentReviewerID string, excludeUserIDs []string) (string, error) {
	s.logger.Info("finding replacement reviewer",
		zap.String("current_reviewer_id", currentReviewerID),
		zap.Strings("exclude_user_ids", excludeUserIDs))

	reviewer, err := s.GetByID(ctx, currentReviewerID)
	if err != nil {
		s.logger.Error("failed to get current reviewer", zap.Error(err))
		return "", domain.NewNotFoundError("user")
	}

	active, err := s.GetActiveByTeam(ctx, reviewer.TeamName)
	if err != nil {
		s.logger.Error("failed to get active team members", zap.Error(err))
		return "", err
	}

	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	available := make([]domain.User, 0, len(active))
	for _, u := range active {
		if !excludeMap[u.UserID] {
			available = append(available, u)
		}
	}

	if len(available) == 0 {
		s.logger.Warn("no candidate for replacement found", zap.String("team_name", reviewer.TeamName))
		return "", domain.NewNoCandidateError()
	}

	idx := rand.Intn(len(available))
	replacement := available[idx].UserID

	s.logger.Info("found replacement reviewer", zap.String("replacement_id", replacement))
	return replacement, nil
}
