package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/handlers"
)

func NewRouter(
	team *handlers.TeamHandler,
	user *handlers.UserHandler,
	pr *handlers.PRHandler,
) http.Handler {
	r := chi.NewRouter()
	UseMiddlewares(r)

	registerTeamRoutes(r, team)
	registerUserRoutes(r, user)
	registerPRRoutes(r, pr)

	return r
}

func registerTeamRoutes(r chi.Router, h *handlers.TeamHandler) {
	r.Route("/team", func(r chi.Router) {
		r.Post("/add", h.CreateTeam)
		r.Get("/get", h.GetTeam)
	})
}

func registerUserRoutes(r chi.Router, h *handlers.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.SetUserActive)
		r.Get("/getReview", h.GetAssignedPRs)
	})
}

func registerPRRoutes(r chi.Router, h *handlers.PRHandler) {
	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.CreatePR)
		r.Post("/merge", h.MergePR)
		r.Post("/reassign", h.ReassignPR)
	})
}
