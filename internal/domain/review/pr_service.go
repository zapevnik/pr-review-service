package review

import (
	"context"
	"time"
)

func (s *Service) CreatePR(ctx context.Context, pr PullRequest) (PullRequest, error) {
	s.log.Info("CreatePR called", "author_id", pr.AuthorID, "title", pr.Title)

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		s.log.Error("failed to get author", "error", err, "author_id", pr.AuthorID)
		return PullRequest{}, err
	}

	pr.Status = StatusOpen
	teamName := author.Team

	stats, err := s.prRepo.ListReviewerStats(ctx, teamName)
	if err != nil {
		s.log.Error("failed to list reviewer stats", "error", err, "team", teamName)
		return PullRequest{}, err
	}

	candidates := make([]ReviewerStats, 0, len(stats))
	for _, r := range stats {
		if r.UserID != pr.AuthorID {
			candidates = append(candidates, r)
		}
	}

	if len(candidates) > 1 {
		s.randSrc.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
	}
	if len(candidates) > 2 {
		candidates = candidates[:2]
	}

	pr.ReviewerIDs = make([]string, 0, len(candidates))
	for _, c := range candidates {
		pr.ReviewerIDs = append(pr.ReviewerIDs, c.UserID)
	}

	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = time.Now().UTC()
	}

	created, err := s.prRepo.Create(ctx, pr)
	if err != nil {
		s.log.Error("failed to create PR", "error", err, "pr_id", pr.ID)
		return PullRequest{}, err
	}

	s.log.Info("PR created successfully", "pr_id", created.ID, "reviewers", created.ReviewerIDs)
	return created, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID string, reviewerOldID string) (PullRequest, string, error) {
	s.log.Info("ReassignReviewer called", "pr_id", prID, "old_reviewer", reviewerOldID)

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get PR", "error", err, "pr_id", prID)
		return PullRequest{}, "", err
	}

	if pr.Status == StatusMerged {
		s.log.Warn("cannot reassign reviewer for merged PR", "pr_id", prID)
		return PullRequest{}, "", ErrPRMerged
	}

	found := false
	for _, id := range pr.ReviewerIDs {
		if id == reviewerOldID {
			found = true
			break
		}
	}
	if !found {
		s.log.Warn("old reviewer not assigned to PR", "pr_id", prID, "reviewer_id", reviewerOldID)
		return PullRequest{}, "", ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, reviewerOldID)
	if err != nil {
		s.log.Error("failed to get old reviewer", "error", err, "reviewer_id", reviewerOldID)
		return PullRequest{}, "", err
	}

	teamMembers, err := s.userRepo.ListByTeam(ctx, oldReviewer.Team)
	if err != nil {
		s.log.Error("failed to list team members", "error", err, "team", oldReviewer.Team)
		return PullRequest{}, "", err
	}

	current := map[string]struct{}{}
	for _, id := range pr.ReviewerIDs {
		current[id] = struct{}{}
	}

	candidates := make([]User, 0, len(teamMembers))
	for _, u := range teamMembers {
		if !u.IsActive || u.ID == pr.AuthorID {
			continue
		}
		if _, exists := current[u.ID]; exists {
			continue
		}
		candidates = append(candidates, u)
	}

	if len(candidates) == 0 {
		s.log.Warn("no candidate to reassign", "pr_id", prID)
		return PullRequest{}, "", ErrNoCandidate
	}

	newID := candidates[s.randSrc.Intn(len(candidates))].ID
	for i, id := range pr.ReviewerIDs {
		if id == reviewerOldID {
			pr.ReviewerIDs[i] = newID
			break
		}
	}

	updated, err := s.prRepo.Update(ctx, pr)
	if err != nil {
		s.log.Error("failed to update PR with new reviewer", "error", err, "pr_id", prID)
		return PullRequest{}, "", err
	}

	s.log.Info("reviewer reassigned successfully", "pr_id", prID, "old_reviewer", reviewerOldID, "new_reviewer", newID)
	return updated, newID, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (PullRequest, error) {
	s.log.Info("MergePR called", "pr_id", prID)

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get PR", "error", err, "pr_id", prID)
		return PullRequest{}, err
	}

	if pr.Status == StatusMerged {
		s.log.Info("PR already merged", "pr_id", prID)
		if pr.MergedAt == nil {
			t := time.Now().UTC()
			pr.MergedAt = &t
			return s.prRepo.Update(ctx, pr)
		}
		return pr, nil
	}

	pr.Status = StatusMerged
	t := time.Now().UTC()
	pr.MergedAt = &t

	updated, err := s.prRepo.Update(ctx, pr)
	if err != nil {
		s.log.Error("failed to merge PR", "error", err, "pr_id", prID)
		return PullRequest{}, err
	}

	s.log.Info("PR merged successfully", "pr_id", prID)
	return updated, nil
}
