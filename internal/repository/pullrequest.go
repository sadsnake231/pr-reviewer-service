package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
	"go.uber.org/zap"
)

type prRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPullRequestRepository(pool *pgxpool.Pool, logger *zap.Logger) PullRequestRepository {
	return &prRepository{pool: pool, logger: logger}
}

func (r *prRepository) Create(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error {
	r.logger.Info("creating pull request",
		zap.String("pr_id", pr.ID),
		zap.String("author_id", pr.AuthorID),
		zap.Int("reviewers_count", len(reviewerIDs)))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, query, pr.ID, pr.Name, pr.AuthorID, pr.Status)
	if err != nil {
		r.logger.Error("failed to insert PR", zap.Error(err), zap.String("pr_id", pr.ID))
		return err
	}

	for _, reviewerID := range reviewerIDs {
		reviewerQuery := `
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`
		_, err = tx.Exec(ctx, reviewerQuery, pr.ID, reviewerID)
		if err != nil {
			r.logger.Error("failed to assign reviewer", zap.Error(err), zap.String("reviewer_id", reviewerID))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	r.logger.Info("PR created successfully", zap.String("pr_id", pr.ID))
	return nil
}

func (r *prRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	r.logger.Debug("getting PR by ID", zap.String("pr_id", prID))

	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`
	row := r.pool.QueryRow(ctx, query, prID)

	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("PR not found", zap.String("pr_id", prID))
			return nil, domain.NewNotFoundError("PR")
		}
		r.logger.Error("failed to get PR", zap.Error(err), zap.String("pr_id", prID))
		return nil, err
	}

	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepository) Exists(ctx context.Context, prID string) (bool, error) {
	r.logger.Debug("checking if PR exists", zap.String("pr_id", prID))

	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check PR existence", zap.Error(err), zap.String("pr_id", prID))
		return false, err
	}
	return exists, nil
}

func (r *prRepository) ReplaceReviewer(ctx context.Context, prID, oldID, newID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`, prID, oldID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`, prID, newID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *prRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	r.logger.Debug("getting reviewers for PR", zap.String("pr_id", prID))

	query := `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY assigned_at`
	rows, err := r.pool.Query(ctx, query, prID)
	if err != nil {
		r.logger.Error("failed to get reviewers", zap.Error(err), zap.String("pr_id", prID))
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			r.logger.Error("failed to scan reviewer", zap.Error(err))
			return nil, err
		}
		reviewers = append(reviewers, userID)
	}

	return reviewers, rows.Err()
}

func (r *prRepository) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	r.logger.Info("merging PR", zap.String("pr_id", prID))

	now := time.Now()
	query := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = $1
		WHERE pull_request_id = $2
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	`
	row := r.pool.QueryRow(ctx, query, now, prID)

	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		r.logger.Error("failed to merge PR", zap.Error(err), zap.String("pr_id", prID))
		return nil, err
	}

	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	r.logger.Info("PR merged successfully", zap.String("pr_id", prID))
	return &pr, nil
}

func (r *prRepository) GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	r.logger.Debug("getting PRs for reviewer", zap.String("user_id", userID))

	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get PRs for reviewer", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			r.logger.Error("failed to scan PR", zap.Error(err))
			return nil, err
		}

		reviewers, err := r.GetReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers

		prs = append(prs, pr)
	}

	r.logger.Info("retrieved PRs for reviewer", zap.String("user_id", userID), zap.Int("count", len(prs)))
	return prs, rows.Err()
}

func (r *prRepository) GetStats(ctx context.Context) (*domain.PRStats, error) {
	r.logger.Debug("getting PR statistics")

	stats := &domain.PRStats{
		AuthoredPRsCount: make(map[string]int64),
		ReviewersCount:   make(map[string]int64),
	}

	query := `SELECT COUNT(*) as total FROM pull_requests`
	if err := r.pool.QueryRow(ctx, query).Scan(&stats.TotalPRs); err != nil {
		r.logger.Error("failed to get total PRs", zap.Error(err))
		return nil, err
	}

	openQuery := `SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'`
	if err := r.pool.QueryRow(ctx, openQuery).Scan(&stats.OpenPRs); err != nil {
		r.logger.Error("failed to get open PRs", zap.Error(err))
		return nil, err
	}

	mergedQuery := `SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'`
	if err := r.pool.QueryRow(ctx, mergedQuery).Scan(&stats.MergedPRs); err != nil {
		r.logger.Error("failed to get merged PRs", zap.Error(err))
		return nil, err
	}

	authorQuery := `SELECT author_id, COUNT(*) FROM pull_requests GROUP BY author_id`
	authorRows, err := r.pool.Query(ctx, authorQuery)
	if err != nil {
		r.logger.Error("failed to get author stats", zap.Error(err))
		return nil, err
	}
	defer authorRows.Close()

	for authorRows.Next() {
		var authorID string
		var count int64
		if err := authorRows.Scan(&authorID, &count); err != nil {
			r.logger.Error("failed to scan author stats", zap.Error(err))
			return nil, err
		}
		stats.AuthoredPRsCount[authorID] = count
	}

	reviewerQuery := `SELECT user_id, COUNT(*) FROM pr_reviewers GROUP BY user_id`
	reviewerRows, err := r.pool.Query(ctx, reviewerQuery)
	if err != nil {
		r.logger.Error("failed to get reviewer stats", zap.Error(err))
		return nil, err
	}
	defer reviewerRows.Close()

	for reviewerRows.Next() {
		var userID string
		var count int64
		if err := reviewerRows.Scan(&userID, &count); err != nil {
			r.logger.Error("failed to scan reviewer stats", zap.Error(err))
			return nil, err
		}
		stats.ReviewersCount[userID] = count
	}

	r.logger.Info("statistics retrieved successfully",
		zap.Int64("total_prs", stats.TotalPRs),
		zap.Int64("open_prs", stats.OpenPRs),
		zap.Int64("merged_prs", stats.MergedPRs))

	return stats, nil
}
