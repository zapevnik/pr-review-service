package mappers

import (
	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/resp"
)

// ToDTOPR маппит domain.PullRequest -> resp.PullRequest
func ToDTOPR(pr review.PullRequest) resp.PullRequest {
	reviewers := make([]string, 0, len(pr.ReviewerIDs))
	reviewers = append(reviewers, pr.ReviewerIDs...)

	return resp.PullRequest{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// ToDTOPRShort маппит domain.PullRequest -> resp.PullRequestShort
func ToDTOPRShort(pr review.PullRequest) resp.PullRequestShort {
	return resp.PullRequestShort{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Title,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

// FromCreatePRReq маппит req.CreatePR -> review.PullRequest
func FromCreatePRReq(id string, authorID string, title string) review.PullRequest {
	return review.PullRequest{
		ID:       id,
		Title:    title,
		AuthorID: authorID,
	}
}
