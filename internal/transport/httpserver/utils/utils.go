package utils

import (
	"encoding/json"

	"github.com/zapevnik/pr-review-service/internal/domain/review"

	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func HandleDomainError(w http.ResponseWriter, err error) bool {
	switch err {
	case review.ErrTeamExists:
		WriteError(w, http.StatusBadRequest, "TEAM_EXISTS", "team already exists")
	case review.ErrPRExists:
		WriteError(w, http.StatusConflict, "PR_EXISTS", "pull request already exists")
	case review.ErrPRMerged:
		WriteError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
	case review.ErrNotAssigned:
		WriteError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
	case review.ErrNoCandidate:
		WriteError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
	case review.ErrNotFound:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case review.ErrUserInAnotherTeam:
		WriteError(w, http.StatusBadRequest, "USER_IN_ANOTHER_TEAM", "user already belongs to another team")
	default:
		return false
	}
	return true
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}
