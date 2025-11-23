package service

import (
	"context"
	"testing"

	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockUserService struct {
	mocks.MockUserService
}

func TestCreatePR_Success(t *testing.T) {
	mockPRRepo := new(mocks.MockPullRequestRepository)
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserService := new(mocks.MockUserService)
	logger := zap.NewNop()

	service := NewPullRequestService(mockPRRepo, mockUserService, mockTeamRepo, logger)

	ctx := context.Background()
	prID := "pr-1"
	prName := "Feature A"
	authorID := "u1"
	author := &domain.User{UserID: authorID, TeamName: "backend"}
	reviewers := []string{"u2", "u3"}

	mockPRRepo.On("Exists", ctx, prID).Return(false, nil)
	mockUserService.On("GetByID", ctx, authorID).Return(author, nil)
	mockUserService.On("SelectRandomReviewers", ctx, "backend", authorID, 2).Return(reviewers, nil)

	mockPRRepo.On("Create", ctx, mock.MatchedBy(func(pr *domain.PullRequest) bool {
		return pr.ID == prID && pr.AuthorID == authorID && pr.Status == "OPEN"
	}), reviewers).Return(nil)

	result, err := service.Create(ctx, prID, prName, authorID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, prID, result.ID)
	assert.Equal(t, reviewers, result.AssignedReviewers)

	mockPRRepo.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestMergePR_Success(t *testing.T) {
	mockPRRepo := new(mocks.MockPullRequestRepository)
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserService := new(mocks.MockUserService)
	logger := zap.NewNop()

	service := NewPullRequestService(mockPRRepo, mockUserService, mockTeamRepo, logger)

	ctx := context.Background()
	prID := "pr-1"

	openPR := &domain.PullRequest{ID: prID, Status: "OPEN"}
	mergedPR := &domain.PullRequest{ID: prID, Status: "MERGED"}

	mockPRRepo.On("GetByID", ctx, prID).Return(openPR, nil)
	mockPRRepo.On("Merge", ctx, prID).Return(mergedPR, nil)

	result, err := service.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.Equal(t, "MERGED", result.Status)
	mockPRRepo.AssertExpectations(t)
}

func TestReassign_Success(t *testing.T) {
	mockPRRepo := new(mocks.MockPullRequestRepository)
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserService := new(mocks.MockUserService)
	logger := zap.NewNop()

	service := NewPullRequestService(mockPRRepo, mockUserService, mockTeamRepo, logger)

	ctx := context.Background()
	prID := "pr-1"
	oldReviewerID := "u2"
	newReviewerID := "u4"
	authorID := "u1"

	currentPR := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []string{"u2", "u3"},
	}

	mockPRRepo.On("GetByID", ctx, prID).Return(currentPR, nil).Once()

	mockUserService.On("FindReplacementReviewer", ctx, oldReviewerID, mock.Anything).Return(newReviewerID, nil)

	mockPRRepo.On("ReplaceReviewer", ctx, prID, oldReviewerID, newReviewerID).Return(nil).Once()

	updatedPR := &domain.PullRequest{
		ID:                prID,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []string{newReviewerID, "u3"},
	}
	mockPRRepo.On("GetByID", ctx, prID).Return(updatedPR, nil).Once()

	pr, newID, err := service.Reassign(ctx, prID, oldReviewerID)

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.NotNil(t, newID)
	assert.Equal(t, newReviewerID, *newID)
	assert.Contains(t, pr.AssignedReviewers, newReviewerID)
	assert.NotContains(t, pr.AssignedReviewers, oldReviewerID)

	mockPRRepo.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestGetReviewerPRs_Success(t *testing.T) {
	mockPRRepo := new(mocks.MockPullRequestRepository)
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserService := new(mocks.MockUserService)
	logger := zap.NewNop()

	service := NewPullRequestService(mockPRRepo, mockUserService, mockTeamRepo, logger)

	ctx := context.Background()
	userID := "u1"
	expectedPRs := []domain.PullRequest{
		{ID: "pr1", Status: "OPEN"},
	}

	mockPRRepo.On("GetByReviewer", ctx, userID).Return(expectedPRs, nil)

	result, err := service.GetReviewerPRs(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPRs, result)
	mockPRRepo.AssertExpectations(t)
}
