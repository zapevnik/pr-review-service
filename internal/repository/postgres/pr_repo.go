package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"log/slog"
	"time"
)

type PRRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewPRRepo(db *DB, l *slog.Logger) *PRRepo {
	return &PRRepo{db: db.sql, log: l}
}

func (r *PRRepo) Create(ctx context.Context, pr review.PullRequest) (review.PullRequest, error) {
	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = time.Now().UTC()
	}

	r.log.Info("creating pull request", "pr_id", pr.ID, "title", pr.Title, "author", pr.AuthorID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.Error("failed to begin transaction", "error", err)
		return review.PullRequest{}, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO pull_requests (pr_id, pr_title, author_id, pr_status, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
		pr.ID, pr.Title, pr.AuthorID, pr.Status, pr.CreatedAt,
	)
	if err != nil {
		_ = tx.Rollback()
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return review.PullRequest{}, review.ErrPRExists
		}
		r.log.Error("failed to insert pull request", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	for _, reviewerID := range pr.ReviewerIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`,
			pr.ID, reviewerID,
		)
		if err != nil {
			_ = tx.Rollback()
			r.log.Error("failed to insert reviewer", "error", err, "pr_id", pr.ID, "reviewer_id", reviewerID)
			return review.PullRequest{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		r.log.Error("failed to commit transaction", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	r.log.Info("pull request created successfully", "pr_id", pr.ID)
	return r.GetByID(ctx, pr.ID)
}

func (r *PRRepo) GetByID(ctx context.Context, id string) (review.PullRequest, error) {
	var pr review.PullRequest
	r.log.Info("fetching pull request by ID", "pr_id", id)

	row := r.db.QueryRowContext(ctx,
		`SELECT pr_id, pr_title, author_id, pr_status, created_at, merged_at
		 FROM pull_requests WHERE pr_id=$1`,
		id,
	)
	err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn("pull request not found", "pr_id", id)
			return pr, review.ErrNotFound
		}
		r.log.Error("failed to scan pull request", "error", err, "pr_id", id)
		return pr, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1`, pr.ID)
	if err != nil {
		r.log.Error("failed to fetch reviewers", "error", err, "pr_id", pr.ID)
		return pr, err
	}
	defer rows.Close()

	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return pr, err
		}
		pr.ReviewerIDs = append(pr.ReviewerIDs, rid)
	}

	return pr, rows.Err()
}

func (r *PRRepo) Update(ctx context.Context, pr review.PullRequest) (review.PullRequest, error) {
	r.log.Info("updating pull request", "pr_id", pr.ID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.Error("failed to begin transaction", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE pull_requests SET pr_title=$1, pr_status=$2, merged_at=$3 WHERE pr_id=$4`,
		pr.Title, pr.Status, pr.MergedAt, pr.ID,
	)
	if err != nil {
		_ = tx.Rollback()
		r.log.Error("failed to update pull request", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM pr_reviewers WHERE pr_id=$1`, pr.ID)
	if err != nil {
		_ = tx.Rollback()
		r.log.Error("failed to delete old reviewers", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	for _, rid := range pr.ReviewerIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`,
			pr.ID, rid,
		)
		if err != nil {
			_ = tx.Rollback()
			r.log.Error("failed to insert reviewer", "error", err, "pr_id", pr.ID, "reviewer_id", rid)
			return review.PullRequest{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		r.log.Error("failed to commit transaction", "error", err, "pr_id", pr.ID)
		return review.PullRequest{}, err
	}

	r.log.Info("pull request updated successfully", "pr_id", pr.ID)
	return r.GetByID(ctx, pr.ID)
}

func (r *PRRepo) ListAssignedTo(ctx context.Context, userID string) ([]review.PullRequest, error) {
	r.log.Info("listing pull requests assigned to user", "user_id", userID)

	rows, err := r.db.QueryContext(ctx,
		`SELECT p.pr_id
		   FROM pull_requests p
		   JOIN pr_reviewers prr ON p.pr_id = prr.pr_id
		  WHERE prr.reviewer_id = $1`,
		userID,
	)
	if err != nil {
		r.log.Error("failed to query assigned PRs", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	var result []review.PullRequest
	for _, id := range ids {
		pr, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, pr)
	}

	return result, nil
}

func (r *PRRepo) ListReviewerStats(ctx context.Context, team string) ([]review.ReviewerStats, error) {
	r.log.Info("listing reviewer stats", "team", team)

	rows, err := r.db.QueryContext(ctx,
		`SELECT u.user_id, u.user_name, u.team_name, COUNT(*)
		   FROM pr_reviewers prr
		   JOIN pull_requests p ON p.pr_id = prr.pr_id
		   JOIN users u ON u.user_id = prr.reviewer_id
		  WHERE p.pr_status = 'OPEN' AND u.team_name = $1
		  GROUP BY u.user_id, u.user_name, u.team_name
		  ORDER BY COUNT(*) ASC`,
		team,
	)
	if err != nil {
		r.log.Error("failed to query reviewer stats", "error", err, "team", team)
		return nil, err
	}
	defer rows.Close()

	var result []review.ReviewerStats
	for rows.Next() {
		var s review.ReviewerStats
		if err := rows.Scan(&s.UserID, &s.Username, &s.TeamName, &s.AssignedOpenPRs); err != nil {
			r.log.Error("failed to scan reviewer stats", "error", err)
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}
