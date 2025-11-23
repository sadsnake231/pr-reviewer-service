package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

type teamRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewTeamRepository(pool *pgxpool.Pool, logger *zap.Logger) TeamRepository {
	return &teamRepository{pool: pool, logger: logger}
}

func (r *teamRepository) Create(ctx context.Context, team *domain.Team) error {
	r.logger.Info("creating team", zap.String("team_name", team.TeamName), zap.Int("members_count", len(team.Members)))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO teams (team_name) VALUES ($1)`
	_, err = tx.Exec(ctx, query, team.TeamName)
	if err != nil {
		r.logger.Error("failed to insert team", zap.Error(err), zap.String("team_name", team.TeamName))
		return err
	}

	for _, member := range team.Members {
		memberQuery := `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE
			SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP
		`
		_, err = tx.Exec(ctx, memberQuery, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			r.logger.Error("failed to insert team member", zap.Error(err), zap.String("user_id", member.UserID))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	r.logger.Info("team created successfully", zap.String("team_name", team.TeamName))
	return nil
}

func (r *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	r.logger.Debug("checking if team exists", zap.String("team_name", teamName))

	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check team existence", zap.Error(err), zap.String("team_name", teamName))
		return false, err
	}
	return exists, nil
}

func (r *teamRepository) GetWithMembers(ctx context.Context, teamName string, userRepo UserRepository) (*domain.Team, error) {
	r.logger.Debug("getting team with members", zap.String("team_name", teamName))

	exists, err := r.Exists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		r.logger.Warn("team not found", zap.String("team_name", teamName))
		return nil, domain.NewNotFoundError("team")
	}

	members, err := userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		r.logger.Error("failed to get team members", zap.Error(err), zap.String("team_name", teamName))
		return nil, err
	}

	team := &domain.Team{
		TeamName: teamName,
		Members:  members,
	}

	r.logger.Info("team retrieved successfully", zap.String("team_name", teamName), zap.Int("members_count", len(members)))
	return team, nil
}
