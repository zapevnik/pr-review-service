package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/req"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/resp"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils/mappers"
)

type UserHandler struct {
	svc *review.Service
	log *slog.Logger
}

func NewUserHandler(svc *review.Service, l *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: l}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req req.SetIsActive
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body in SetUserActive", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.UserID == "" {
		h.log.Warn("missing user_id in SetUserActive")
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	h.log.Info("SetUserActive called", "user_id", req.UserID, "is_active", req.IsActive)

	user, err := h.svc.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, review.ErrNotFound) {
			h.log.Warn("user not found", "user_id", req.UserID)
			utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		} else {
			h.log.Error("failed to get user", "user_id", req.UserID, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	user.IsActive = req.IsActive
	updated, err := h.svc.UpdateUser(ctx, user)
	if err != nil {
		h.log.Error("failed to update user active status", "user_id", req.UserID, "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	h.log.Info("user active status updated", "user_id", updated.ID, "is_active", updated.IsActive)
	utils.RespondJSON(w, http.StatusOK, map[string]any{"user": mappers.ToDTOUser(updated)})
}

func (h *UserHandler) GetAssignedPRs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.log.Warn("missing user_id in GetAssignedPRs")
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	h.log.Info("GetAssignedPRs called", "user_id", userID)
	prs, err := h.svc.GetAssignedForUser(ctx, userID)
	if err != nil {
		h.log.Error("failed to get assigned PRs", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	out := make([]resp.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		out = append(out, mappers.ToDTOPRShort(pr))
	}

	h.log.Info("assigned PRs retrieved", "user_id", userID, "count", len(out))
	utils.RespondJSON(w, http.StatusOK, map[string]any{
		"user_id":       userID,
		"pull_requests": out,
	})
}
