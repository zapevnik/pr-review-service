package review

import (
	"context"
	"errors"
)

func (s *Service) ListActiveByTeam(ctx context.Context, teamName string) ([]User, error) {
	s.log.Info("ListActiveByTeam called", "team", teamName)

	users, err := s.userRepo.ListActiveByTeam(ctx, teamName)
	if err != nil {
		s.log.Error("failed to list active users", "team", teamName, "error", err)
		return nil, err
	}

	active := make([]User, 0, len(users))
	for _, u := range users {
		if u.IsActive {
			active = append(active, u)
		}
	}

	s.log.Info("ListActiveByTeam completed", "team", teamName, "count", len(active))
	return active, nil
}

func (s *Service) ListByTeam(ctx context.Context, teamName string) ([]User, error) {
	s.log.Info("ListByTeam called", "team", teamName)

	users, err := s.userRepo.ListByTeam(ctx, teamName)
	if err != nil {
		s.log.Error("failed to list users", "team", teamName, "error", err)
		return nil, err
	}

	s.log.Info("ListByTeam completed", "team", teamName, "count", len(users))
	return users, nil
}

func (s *Service) CreateTeam(ctx context.Context, name string, members []User) (Team, error) {
	s.log.Info("CreateTeam called", "team", name, "members_count", len(members))

	_, err := s.teamRepo.GetByName(ctx, name)
	if err == nil {
		s.log.Warn("team already exists", "team", name)
		return Team{}, ErrTeamExists
	}
	if !errors.Is(err, ErrNotFound) {
		s.log.Error("failed to get team by name", "team", name, "error", err)
		return Team{}, err
	}
	s.log.Info("team does not exist, proceeding", "team", name)

	team := Team{Name: name}

	if _, err := s.teamRepo.Create(ctx, team); err != nil {
		s.log.Error("failed to create team", "team", name, "error", err)
		return Team{}, err
	}

	for _, u := range members {
		u.Team = name
		s.log.Info("checking user existence", "user_id", u.ID, "username", u.Name)

		existing, err := s.userRepo.GetByID(ctx, u.ID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			s.log.Error("failed to get user by ID", "user_id", u.ID, "error", err)
			return Team{}, err
		}

		if err == nil {
			if existing.Team != "" && existing.Team != name {
				return Team{}, ErrUserInAnotherTeam
			}
			if _, err := s.userRepo.Update(ctx, u); err != nil {
				s.log.Error("failed to update user", "user_id", u.ID, "error", err)
				return Team{}, err
			}
			s.log.Info("user updated successfully", "user_id", u.ID)
		} else {
			if _, err := s.userRepo.Create(ctx, u); err != nil {
				s.log.Error("failed to create user", "user_id", u.ID, "error", err)
				return Team{}, err
			}
		}
	}

	return team, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (Team, error) {
	s.log.Info("GetByName called", "team", name)

	team, err := s.teamRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Team{}, err
		}
		s.log.Error("failed to get team by name", "team", name, "error", err)
		return Team{}, err
	}

	s.log.Info("GetByName completed", "team", name)
	return team, nil
}
