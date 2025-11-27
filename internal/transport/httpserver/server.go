package httpserver

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/zapevnik/pr-review-service/internal/app/config"
	"github.com/zapevnik/pr-review-service/internal/domain/review"
)

type Server struct {
	log *slog.Logger
	svc review.Service
	cfg config.Server
}

func New(log *slog.Logger, service review.Service, cfg config.Server) *Server {
	return &Server{log: log, svc: service, cfg: cfg}
}
func (s *Server) Run(ctx context.Context, r http.Handler) error {

	srv := &http.Server{
		Addr:              s.cfg.Address,
		Handler:           r,
		ReadTimeout:       s.cfg.ReadTimeout.Duration,
		WriteTimeout:      s.cfg.WriteTimeout.Duration,
		IdleTimeout:       s.cfg.IdleTimeout.Duration,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          slog.NewLogLogger(s.log.Handler(), slog.LevelError),
	}

	errCh := make(chan error, 1)

	go func() {
		s.log.Info("HTTP server starting", "addr", s.cfg.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down http server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.log.Error("http server shutdown error", "err", err)
			return err
		}

		s.log.Info("http server stopped gracefully")
		return nil

	case err := <-errCh:
		s.log.Error("http server error", "err", err)
		return err
	}
}
