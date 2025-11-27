package review

import (
	"context"
)

type PRRepository interface {
	Create(ctx context.Context, pr PullRequest) (PullRequest, error)
	Update(ctx context.Context, pr PullRequest) (PullRequest, error)
	GetByID(ctx context.Context, id string) (PullRequest, error)
	ListAssignedTo(ctx context.Context, userID string) ([]PullRequest, error)
	ListReviewerStats(ctx context.Context, teamName string) ([]ReviewerStats, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (User, error)
	Create(ctx context.Context, u User) (User, error)
	Update(ctx context.Context, u User) (User, error)
	ListByTeam(ctx context.Context, teamName string) ([]User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]User, error)
}

type TeamRepository interface {
	GetByName(ctx context.Context, name string) (Team, error)
	Create(ctx context.Context, t Team) (Team, error)
}
