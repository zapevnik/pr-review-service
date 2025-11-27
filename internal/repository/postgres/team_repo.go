package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/zapevnik/pr-review-service/internal/domain/review"
)

type TeamRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewTeamRepo(db *DB, l *slog.Logger) *TeamRepo {
	return &TeamRepo{db: db.sql, log: l}
}

func (r *TeamRepo) Create(ctx context.Context, team review.Team) (review.Team, error) {
	r.log.Info("creating team", "team_name", team.Name)

	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE team_name=$1)`, team.Name).Scan(&exists)
	if err != nil {
		r.log.Error("failed to check team existence", "error", err, "team_name", team.Name)
		return review.Team{}, err
	}
	if exists {
		r.log.Warn("team already exists", "team_name", team.Name)
		return review.Team{}, review.ErrTeamExists
	}

	_, err = r.db.ExecContext(ctx, `INSERT INTO teams (team_name) VALUES ($1)`, team.Name)
	if err != nil {
		r.log.Error("failed to insert team", "error", err, "team_name", team.Name)
		return review.Team{}, err
	}

	r.log.Info("team created successfully", "team_name", team.Name)
	return team, nil
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (review.Team, error) {
	r.log.Info("fetching team by name", "team_name", name)

	var t review.Team
	err := r.db.QueryRowContext(ctx, `SELECT team_name FROM teams WHERE team_name=$1`, name).Scan(&t.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("team not found", "team_name", name)
			return review.Team{}, review.ErrNotFound
		}
		r.log.Error("failed to fetch team", "error", err, "team_name", name)
		return review.Team{}, err
	}

	return t, nil
}
