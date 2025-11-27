package app

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/zapevnik/pr-review-service/internal/app/config"
	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/repository/postgres"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/handlers"
)

type App struct {
	log *slog.Logger
	cfg *config.Config
}

func New(logg *slog.Logger, cfg *config.Config) *App {
	return &App{
		log: logg,
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting application")

	db, err := postgres.New(a.log, ctx, postgres.Config{
		DSN:             a.cfg.Database.DSN(),
		MigrationsDir:   "./migrations",
		MaxOpenConns:    a.cfg.Database.Pool.MaxOpenConns,
		MaxIdleConns:    a.cfg.Database.Pool.MaxIdleConns,
		ConnMaxLifetime: a.cfg.Database.Pool.ConnMaxLifetime.Duration,
	})
	if err != nil {
		a.log.Error("failed to init db", "error", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			a.log.Info("db close error", "error", err)
		}
	}()

	prRepo := postgres.NewPRRepo(db, a.log)
	userRepo := postgres.NewUserRepo(db, a.log)
	teamRepo := postgres.NewTeamRepo(db, a.log)

	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	svc := review.NewService(prRepo, userRepo, teamRepo, randSrc, a.log)

	a.log.Info("domain service initialized successfully")

	prHandler := handlers.NewPRHandler(svc, a.log)
	teamHandler := handlers.NewTeamHandler(svc, a.log)
	userHandler := handlers.NewUserHandler(svc, a.log)

	router := httpserver.NewRouter(
		teamHandler,
		userHandler,
		prHandler,
	)

	server := httpserver.New(
		a.log,
		*svc,
		a.cfg.Server,
	)

	a.log.Info("http server initialized successfully")

	return server.Run(ctx, router)
}
