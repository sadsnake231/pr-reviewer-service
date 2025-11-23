package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

type userRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *zap.Logger) UserRepository {
	return &userRepository{pool: pool, logger: logger}
}

func (r *userRepository) CreateOrUpdate(ctx context.Context, user *domain.User) error {
	r.logger.Debug("creating or updating user", zap.String("user_id", user.UserID))

	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.pool.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		r.logger.Error("failed to create/update user", zap.Error(err), zap.String("user_id", user.UserID))
		return err
	}

	r.logger.Info("user created/updated successfully", zap.String("user_id", user.UserID))
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	r.logger.Debug("getting user by id", zap.String("user_id", userID))

	query := `SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1`
	row := r.pool.QueryRow(ctx, query, userID)

	var user domain.User
	err := row.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		r.logger.Error("failed to get user", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	r.logger.Debug("getting users by team", zap.String("team_name", teamName))

	query := `SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 ORDER BY user_id`
	rows, err := r.pool.Query(ctx, query, teamName)
	if err != nil {
		r.logger.Error("failed to get users by team", zap.Error(err), zap.String("team_name", teamName))
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			r.logger.Error("failed to scan user row", zap.Error(err))
			return nil, err
		}
		users = append(users, user)
	}

	r.logger.Info("retrieved users by team", zap.String("team_name", teamName), zap.Int("count", len(users)))
	return users, rows.Err()
}

func (r *userRepository) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	r.logger.Debug("getting active users by team", zap.String("team_name", teamName))

	query := `SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 AND is_active = true ORDER BY user_id`
	rows, err := r.pool.Query(ctx, query, teamName)
	if err != nil {
		r.logger.Error("failed to get active users", zap.Error(err), zap.String("team_name", teamName))
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			r.logger.Error("failed to scan active user row", zap.Error(err))
			return nil, err
		}
		users = append(users, user)
	}

	r.logger.Info("retrieved active users", zap.String("team_name", teamName), zap.Int("count", len(users)))
	return users, rows.Err()
}

func (r *userRepository) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	r.logger.Info("setting user active status", zap.String("user_id", userID), zap.Bool("is_active", isActive))

	query := `UPDATE users SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2 RETURNING user_id, username, team_name, is_active`
	row := r.pool.QueryRow(ctx, query, isActive, userID)

	var user domain.User
	if err := row.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
		r.logger.Error("failed to set user active", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) DeactivateTeamMembers(ctx context.Context, teamName string) error {
	r.logger.Info("deactivating team members", zap.String("team_name", teamName))

	query := `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE team_name = $1 AND is_active = true`
	result, err := r.pool.Exec(ctx, query, teamName)
	if err != nil {
		r.logger.Error("failed to deactivate team members", zap.Error(err), zap.String("team_name", teamName))
		return err
	}

	r.logger.Info("team members deactivated", zap.String("team_name", teamName), zap.Int64("count", result.RowsAffected()))
	return nil
}

func (r *userRepository) GetReviewerStats(ctx context.Context) (map[string]int64, error) {
	r.logger.Debug("getting reviewer stats")

	query := `
		SELECT u.user_id, COUNT(pr.pull_request_id) as review_count
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
		GROUP BY u.user_id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to get reviewer stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var userID string
		var count int64
		if err := rows.Scan(&userID, &count); err != nil {
			r.logger.Error("failed to scan reviewer stats row", zap.Error(err))
			return nil, err
		}
		stats[userID] = count
	}

	r.logger.Info("retrieved reviewer stats", zap.Int("users_count", len(stats)))
	return stats, rows.Err()
}
