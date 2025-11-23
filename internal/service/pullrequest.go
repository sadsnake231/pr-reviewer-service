package service

import (
	"context"

	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/repository"
	"go.uber.org/zap"
)

type prService struct {
	prRepo      repository.PullRequestRepository
	userService UserService
	teamRepo    repository.TeamRepository
	logger      *zap.Logger
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userService UserService,
	teamRepo repository.TeamRepository,
	logger *zap.Logger,
) PullRequestService {
	return &prService{
		prRepo:      prRepo,
		userService: userService,
		teamRepo:    teamRepo,
		logger:      logger,
	}
}

func (s *prService) Create(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	s.logger.Info("creating pull request",
		zap.String("pr_id", prID),
		zap.String("pr_name", prName),
		zap.String("author_id", authorID))

	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		s.logger.Error("failed to check PR existence", zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("PR already exists", zap.String("pr_id", prID))
		return nil, domain.NewPRExistsError()
	}

	author, err := s.userService.GetByID(ctx, authorID)
	if err != nil {
		s.logger.Error("author not found", zap.Error(err), zap.String("author_id", authorID))
		return nil, domain.NewNotFoundError("user")
	}

	reviewers, err := s.userService.SelectRandomReviewers(ctx, author.TeamName, authorID, 2)
	if err != nil {
		s.logger.Error("failed to select reviewers", zap.Error(err))
		return nil, err
	}

	pr := &domain.PullRequest{
		ID:       prID,
		Name:     prName,
		AuthorID: authorID,
		Status:   "OPEN",
	}

	if err := s.prRepo.Create(ctx, pr, reviewers); err != nil {
		s.logger.Error("failed to create PR in repository", zap.Error(err))
		return nil, err
	}

	pr.AssignedReviewers = reviewers

	s.logger.Info("PR created successfully",
		zap.String("pr_id", prID),
		zap.Strings("reviewers", reviewers))

	return pr, nil
}

func (s *prService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	s.logger.Info("merging pull request", zap.String("pr_id", prID))

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("failed to get PR for merge", zap.Error(err), zap.String("pr_id", prID))
		return nil, err
	}

	if pr.Status == "MERGED" {
		s.logger.Info("PR already merged, idempotent operation", zap.String("pr_id", prID))
		return pr, nil
	}

	mergedPR, err := s.prRepo.Merge(ctx, prID)
	if err != nil {
		s.logger.Error("failed to merge PR", zap.Error(err), zap.String("pr_id", prID))
		return nil, err
	}

	s.logger.Info("PR merged successfully", zap.String("pr_id", prID))
	return mergedPR, nil
}

func (s *prService) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, *string, error) {
	s.logger.Info("reassigning reviewer",
		zap.String("pr_id", prID),
		zap.String("old_reviewer_id", oldReviewerID))

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("failed to get PR for reassignment", zap.Error(err), zap.String("pr_id", prID))
		return nil, nil, err
	}

	if pr.Status == "MERGED" {
		s.logger.Warn("cannot reassign on merged PR", zap.String("pr_id", prID))
		return nil, nil, domain.NewPRMergedError()
	}

	found := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		s.logger.Warn("reviewer not assigned to PR",
			zap.String("pr_id", prID),
			zap.String("reviewer_id", oldReviewerID))
		return nil, nil, domain.NewNotAssignedError()
	}

	ExcludeUserIds := append(pr.AssignedReviewers, pr.AuthorID)
	newReviewerID, err := s.userService.FindReplacementReviewer(ctx, oldReviewerID, ExcludeUserIds)
	if err != nil {
		s.logger.Error("failed to find replacement reviewer", zap.Error(err))
		return nil, nil, err
	}

	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return nil, nil, err
	}

	updatedPR, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("failed to get updated PR", zap.Error(err))
		return nil, nil, err
	}

	s.logger.Info("reviewer reassigned successfully",
		zap.String("pr_id", prID),
		zap.String("old_reviewer", oldReviewerID),
		zap.String("new_reviewer", newReviewerID))

	return updatedPR, &newReviewerID, nil
}

func (s *prService) GetReviewerPRs(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	s.logger.Debug("getting PRs for reviewer", zap.String("user_id", userID))

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get PRs for reviewer", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	s.logger.Info("retrieved PRs for reviewer", zap.String("user_id", userID), zap.Int("count", len(prs)))
	return prs, nil
}

func (s *prService) GetStats(ctx context.Context) (*domain.PRStats, error) {
	s.logger.Debug("getting PR statistics")

	stats, err := s.prRepo.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get statistics", zap.Error(err))
		return nil, err
	}

	s.logger.Info("statistics retrieved successfully",
		zap.Int64("total_prs", stats.TotalPRs),
		zap.Int64("open_prs", stats.OpenPRs),
		zap.Int64("merged_prs", stats.MergedPRs))

	return stats, nil
}
