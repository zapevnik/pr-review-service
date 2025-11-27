package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/req"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils/mappers"
)

type TeamHandler struct {
	svc *review.Service
	log *slog.Logger
}

func NewTeamHandler(svc *review.Service, l *slog.Logger) *TeamHandler {
	return &TeamHandler{svc: svc, log: l}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body req.TeamAdd
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("invalid request body in CreateTeam", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	teamName, members, err := mappers.TeamAddRequestToArgs(body)
	if err != nil {
		h.log.Warn("failed to parse TeamAdd request", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	h.log.Info("CreateTeam called", "team_name", teamName, "members_count", len(members))

	_, err = h.svc.GetByName(ctx, teamName)
	if err == nil {
		h.log.Warn("team already exists", "team_name", teamName)
		utils.WriteError(w, http.StatusConflict, "TEAM_EXISTS", "team_name already exists")
		return
	}
	if !errors.Is(err, review.ErrNotFound) {
		h.log.Error("failed to check team existence", "team_name", teamName, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check team")
		return
	}

	createdTeam, err := h.svc.CreateTeam(ctx, teamName, members)
	if err != nil {
		if errors.Is(err, review.ErrUserInAnotherTeam) {
			h.log.Warn("user already in another team", "team_name", teamName)
			utils.WriteError(w, http.StatusConflict, "USER_IN_ANOTHER_TEAM", "user is already in another team")
			return
		}
		h.log.Error("failed to create team", "team_name", teamName, "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	h.log.Info("team created successfully", "team_name", teamName)
	respTeam := mappers.TeamToResponse(createdTeam, members)
	utils.RespondJSON(w, http.StatusCreated, map[string]any{"team": respTeam})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.log.Warn("missing team_name in GetTeam")
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
		return
	}

	h.log.Info("GetTeam called", "team_name", teamName)
	team, err := h.svc.GetByName(r.Context(), teamName)
	if err != nil {
		h.log.Error("team not found", "team_name", teamName, "error", err)
		utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	members, err := h.svc.ListByTeam(r.Context(), teamName)
	if err != nil {
		h.log.Error("failed to list team members", "team_name", teamName, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.log.Info("team retrieved successfully", "team_name", teamName, "members_count", len(members))
	respTeam := mappers.TeamToResponse(team, members)
	utils.RespondJSON(w, http.StatusOK, map[string]any{"team": respTeam})
}
