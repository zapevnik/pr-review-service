package review

import (
	"context"
)

func (s *Service) GetAssignedForUser(ctx context.Context, userID string) ([]PullRequest, error) {
	return s.prRepo.ListAssignedTo(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, u User) (User, error) {
	return s.userRepo.Update(ctx, u)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (User, error) {
	return s.userRepo.GetByID(ctx, id)
}
