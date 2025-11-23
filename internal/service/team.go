package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"github.com/sadsnake231/pr-reviewer-service/internal/repository"
	"go.uber.org/zap"
)

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository, logger *zap.Logger) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *teamService) AddTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	s.logger.Info("adding team", zap.String("team_name", team.TeamName), zap.Int("members_count", len(team.Members)))

	exists, err := s.teamRepo.Exists(ctx, team.TeamName)
	if err != nil {
		s.logger.Error("failed to check team existence", zap.Error(err))
		return nil, err
	}
	if exists {
		s.logger.Warn("team already exists", zap.String("team_name", team.TeamName))
		return nil, domain.NewTeamExistsError()
	}

	if len(team.Members) == 0 {
		s.logger.Warn("no members in team", zap.String("team_name", team.TeamName))
		return nil, domain.NewNoMembersError()
	}

	for _, member := range team.Members {
		user, err := s.userRepo.GetByID(ctx, member.UserID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}
		if user != nil {
			return nil, domain.NewNonUniqueUserError()
		}
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		s.logger.Error("failed to create team", zap.Error(err), zap.String("team_name", team.TeamName))
		return nil, err
	}

	s.logger.Info("team added successfully", zap.String("team_name", team.TeamName))
	return team, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	s.logger.Info("getting team", zap.String("team_name", teamName))

	team, err := s.teamRepo.GetWithMembers(ctx, teamName, s.userRepo)
	if err != nil {
		s.logger.Error("failed to get team", zap.Error(err), zap.String("team_name", teamName))
		return nil, err
	}

	s.logger.Info("team retrieved successfully", zap.String("team_name", teamName))
	return team, nil
}

func (s *teamService) DeactivateMembers(ctx context.Context, teamName string) error {
	s.logger.Info("deactivating team members", zap.String("team_name", teamName))

	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return err
	}
	if !exists {
		s.logger.Warn("team not found for deactivation", zap.String("team_name", teamName))
		return domain.NewNotFoundError("team")
	}

	if err := s.userRepo.DeactivateTeamMembers(ctx, teamName); err != nil {
		s.logger.Error("failed to deactivate team members", zap.Error(err), zap.String("team_name", teamName))
		return err
	}

	s.logger.Info("team members deactivated successfully", zap.String("team_name", teamName))
	return nil
}
