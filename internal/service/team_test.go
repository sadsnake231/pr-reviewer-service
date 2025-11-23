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

func TestAddTeam_Success(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	mockTeamRepo.On("Exists", ctx, "backend").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(nil, pgx.ErrNoRows)
	mockUserRepo.On("GetByID", ctx, "u2").Return(nil, pgx.ErrNoRows)
	mockTeamRepo.On("Create", ctx, team).Return(nil)

	result, err := service.AddTeam(ctx, team)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "backend", result.TeamName)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAddTeam_TeamAlreadyExists(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockTeamRepo.On("Exists", ctx, "backend").Return(true, nil)

	result, err := service.AddTeam(ctx, team)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.TeamExists, appErr.Code)
	mockTeamRepo.AssertExpectations(t)
}

func TestAddTeam_NoMembers(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	team := &domain.Team{
		TeamName: "backend",
		Members:  []domain.User{},
	}

	mockTeamRepo.On("Exists", ctx, "backend").Return(false, nil)

	result, err := service.AddTeam(ctx, team)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NoMembers, appErr.Code)
	mockTeamRepo.AssertExpectations(t)
}

func TestAddTeam_UserAlreadyExists(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	existingUser := &domain.User{UserID: "u1", Username: "Alice", TeamName: "frontend"}

	mockTeamRepo.On("Exists", ctx, "backend").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(existingUser, nil)

	result, err := service.AddTeam(ctx, team)

	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NonUniqueUser, appErr.Code)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAddTeam_DatabaseError(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	dbErr := errors.New("database connection failed")

	mockTeamRepo.On("Exists", ctx, "backend").Return(false, dbErr)

	result, err := service.AddTeam(ctx, team)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbErr, err)
	mockTeamRepo.AssertExpectations(t)
}

func TestGetTeam_Success(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	expectedTeam := &domain.Team{
		TeamName: teamName,
		Members: []domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockTeamRepo.On("GetWithMembers", ctx, teamName, mockUserRepo).Return(expectedTeam, nil)

	result, err := service.GetTeam(ctx, teamName)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, teamName, result.TeamName)
	mockTeamRepo.AssertExpectations(t)
}

func TestGetTeam_NotFound(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	teamName := "nonexistent"

	mockTeamRepo.On("GetWithMembers", ctx, teamName, mockUserRepo).Return(nil, pgx.ErrNoRows)

	result, err := service.GetTeam(ctx, teamName)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockTeamRepo.AssertExpectations(t)
}

func TestDeactivateMembers_Success(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	teamName := "backend"

	mockTeamRepo.On("Exists", ctx, teamName).Return(true, nil)
	mockUserRepo.On("DeactivateTeamMembers", ctx, teamName).Return(nil)

	err := service.DeactivateMembers(ctx, teamName)

	assert.NoError(t, err)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestDeactivateMembers_TeamNotFound(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	teamName := "nonexistent"

	mockTeamRepo.On("Exists", ctx, teamName).Return(false, nil)

	err := service.DeactivateMembers(ctx, teamName)

	assert.Error(t, err)

	var appErr *domain.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, domain.NotFound, appErr.Code)
	mockTeamRepo.AssertExpectations(t)
}

func TestDeactivateMembers_DatabaseError(t *testing.T) {
	mockTeamRepo := new(mocks.MockTeamRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	logger := zap.NewNop()
	service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

	ctx := context.Background()
	teamName := "backend"
	dbErr := errors.New("database error")

	mockTeamRepo.On("Exists", ctx, teamName).Return(true, nil)
	mockUserRepo.On("DeactivateTeamMembers", ctx, teamName).Return(dbErr)

	err := service.DeactivateMembers(ctx, teamName)

	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
