package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetActive_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	userID := "u1"
	existingUser := &domain.User{UserID: userID, Username: "Alice", IsActive: false}
	updatedUser := &domain.User{UserID: userID, Username: "Alice", IsActive: true}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("SetActive", ctx, userID, true).Return(updatedUser, nil)

	result, err := service.SetActive(ctx, userID, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestSetActive_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	userID := "nonexistent"

	mockRepo.On("GetByID", ctx, userID).Return(nil, pgx.ErrNoRows)

	result, err := service.SetActive(ctx, userID, true)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NotFound, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestSetActive_DatabaseError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	userID := "u1"
	dbErr := errors.New("database connection failed")

	mockRepo.On("GetByID", ctx, userID).Return(nil, dbErr)

	result, err := service.SetActive(ctx, userID, true)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbErr, err)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	userID := "u1"
	expectedUser := &domain.User{UserID: userID, Username: "Alice", IsActive: true}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	result, err := service.GetByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	mockRepo.AssertExpectations(t)
}

func TestGetActiveByTeam_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	expectedUsers := []domain.User{
		{UserID: "u1", Username: "Alice", TeamName: teamName, IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: teamName, IsActive: true},
	}

	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(expectedUsers, nil)

	result, err := service.GetActiveByTeam(ctx, teamName)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestSelectRandomReviewers_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	excludeUserID := "u1"
	activeUsers := []domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Charlie", IsActive: true},
	}

	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, err := service.SelectRandomReviewers(ctx, teamName, excludeUserID, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.LessOrEqual(t, len(result), 2)

	for _, id := range result {
		assert.NotEqual(t, excludeUserID, id)
	}
	mockRepo.AssertExpectations(t)
}

func TestSelectRandomReviewers_NoAvailableUsers(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	excludeUserID := "u1"
	activeUsers := []domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
	}

	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, err := service.SelectRandomReviewers(ctx, teamName, excludeUserID, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestSelectRandomReviewers_LessUsersThanRequested(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	excludeUserID := "u1"
	activeUsers := []domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Charlie", IsActive: true},
	}

	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, err := service.SelectRandomReviewers(ctx, teamName, excludeUserID, 5)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	mockRepo.AssertExpectations(t)
}

func TestFindReplacementReviewer_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	currentReviewerID := "u1"
	teamName := "backend"
	excludeUserIDs := []string{"u1", "u2"}

	currentReviewer := &domain.User{UserID: currentReviewerID, TeamName: teamName}
	activeUsers := []domain.User{
		{UserID: "u1", Username: "Alice", TeamName: teamName, IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: teamName, IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamName: teamName, IsActive: true},
		{UserID: "u4", Username: "Dave", TeamName: teamName, IsActive: true},
	}

	mockRepo.On("GetByID", ctx, currentReviewerID).Return(currentReviewer, nil)
	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, err := service.FindReplacementReviewer(ctx, currentReviewerID, excludeUserIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.NotContains(t, excludeUserIDs, result)
	mockRepo.AssertExpectations(t)
}

func TestFindReplacementReviewer_CurrentReviewerNotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	currentReviewerID := "nonexistent"

	mockRepo.On("GetByID", ctx, currentReviewerID).Return(nil, pgx.ErrNoRows)

	result, err := service.FindReplacementReviewer(ctx, currentReviewerID, []string{})

	assert.Error(t, err)
	assert.Empty(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NotFound, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestFindReplacementReviewer_NoCandidates(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewUserService(mockRepo, logger)

	ctx := context.Background()
	currentReviewerID := "u1"
	teamName := "backend"
	excludeUserIDs := []string{"u1", "u2", "u3"}

	currentReviewer := &domain.User{UserID: currentReviewerID, TeamName: teamName}
	activeUsers := []domain.User{
		{UserID: "u1", Username: "Alice", TeamName: teamName, IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: teamName, IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamName: teamName, IsActive: true},
	}

	mockRepo.On("GetByID", ctx, currentReviewerID).Return(currentReviewer, nil)
	mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, err := service.FindReplacementReviewer(ctx, currentReviewerID, excludeUserIDs)

	assert.Error(t, err)
	assert.Empty(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NoCandidate, appErr.Code)
	mockRepo.AssertExpectations(t)
}
