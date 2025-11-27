package review

import (
	"log/slog"
	"math/rand"
	"time"
)

type Service struct {
	prRepo   PRRepository
	userRepo UserRepository
	teamRepo TeamRepository
	randSrc  *rand.Rand
	log      *slog.Logger
}

func NewService(prRepo PRRepository, userRepo UserRepository, teamRepo TeamRepository, randSrc *rand.Rand, l *slog.Logger) *Service {
	if randSrc == nil {
		randSrc = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Service{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		randSrc:  randSrc,
		log:      l,
	}
}
