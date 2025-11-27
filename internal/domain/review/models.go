package review

import (
	"errors"
	"time"
)

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID          string
	Title       string
	AuthorID    string
	Status      PRStatus
	CreatedAt   time.Time
	MergedAt    *time.Time
	ReviewerIDs []string
}

type Team struct {
	Name string
}

type User struct {
	ID       string
	Team     string
	Name     string
	IsActive bool
}

type ReviewerStats struct {
	UserID          string
	Username        string
	TeamName        string
	AssignedOpenPRs int
}

var (
	ErrTeamExists        = errors.New("TEAM_EXISTS")
	ErrPRExists          = errors.New("PR_EXISTS")
	ErrPRMerged          = errors.New("PR_MERGED")
	ErrNotAssigned       = errors.New("NOT_ASSIGNED")
	ErrNoCandidate       = errors.New("NO_CANDIDATE")
	ErrNotFound          = errors.New("NOT_FOUND")
	ErrUserInAnotherTeam = errors.New("USER_IN_ANOTHER_TEAM")
)
