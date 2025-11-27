package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/zapevnik/pr-review-service/internal/domain/review"
)

type UserRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewUserRepo(db *DB, l *slog.Logger) *UserRepo {
	return &UserRepo{db: db.sql, log: l}
}

func (r *UserRepo) Create(ctx context.Context, u review.User) (review.User, error) {
	r.log.Info("creating user", "user_id", u.ID, "username", u.Name, "team", u.Team)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (user_id, user_name, is_active, team_name) VALUES ($1, $2, $3, $4)`,
		u.ID, u.Name, u.IsActive, u.Team,
	)
	if err != nil {
		r.log.Error("failed to insert user", "error", err, "user_id", u.ID)
		return review.User{}, err
	}
	r.log.Info("user created successfully", "user_id", u.ID)
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u review.User) (review.User, error) {
	r.log.Info("updating user", "user_id", u.ID, "team", u.Team)

	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET user_name=$1, is_active=$2, team_name=$3 WHERE user_id=$4`,
		u.Name, u.IsActive, u.Team, u.ID,
	)
	if err != nil {
		r.log.Error("failed to update user", "error", err, "user_id", u.ID)
		return review.User{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		r.log.Warn("user not found for update", "user_id", u.ID)
		return review.User{}, review.ErrNotFound
	}

	return r.GetByID(ctx, u.ID)
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (review.User, error) {
	r.log.Info("fetching user by ID", "user_id", userID)

	query := `SELECT user_id, user_name, is_active, team_name FROM users WHERE user_id=$1`
	var u review.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.Name, &u.IsActive, &u.Team)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn("user not found", "user_id", userID)
			return review.User{}, review.ErrNotFound
		}
		r.log.Error("failed to scan user", "error", err, "user_id", userID)
		return review.User{}, err
	}
	return u, nil
}

func (r *UserRepo) ListActiveByTeam(ctx context.Context, teamName string) ([]review.User, error) {
	r.log.Info("listing active users by team", "team", teamName)

	query := `SELECT user_id, user_name, is_active, team_name FROM users WHERE team_name=$1 AND is_active=true`
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		r.log.Error("failed to query active users", "error", err, "team", teamName)
		return nil, err
	}
	defer rows.Close()

	var users []review.User
	for rows.Next() {
		var u review.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsActive, &u.Team); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) ListByTeam(ctx context.Context, teamName string) ([]review.User, error) {
	r.log.Info("listing all users by team", "team", teamName)

	query := `SELECT user_id, user_name, is_active, team_name FROM users WHERE team_name=$1`
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		r.log.Error("failed to query users by team", "error", err, "team", teamName)
		return nil, err
	}
	defer rows.Close()

	var users []review.User
	for rows.Next() {
		var u review.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsActive, &u.Team); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
