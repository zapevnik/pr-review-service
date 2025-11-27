package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/req"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/resp"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/utils/mappers"
)

type PRHandler struct {
	svc *review.Service
	log *slog.Logger
}

func NewPRHandler(svc *review.Service, log *slog.Logger) *PRHandler {
	return &PRHandler{svc: svc, log: log}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var body req.CreatePR
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("invalid JSON in CreatePR", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
		return
	}
	if body.PullRequestID == "" || body.PullRequestName == "" || body.AuthorID == "" {
		h.log.Warn("missing required fields in CreatePR", "body", body)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id, pull_request_name and author_id are required")
		return
	}

	h.log.Info("CreatePR called", "pr_id", body.PullRequestID, "author_id", body.AuthorID, "pr_title", body.PullRequestName)

	pr := mappers.FromCreatePRReq(body.PullRequestID, body.AuthorID, body.PullRequestName)
	created, err := h.svc.CreatePR(r.Context(), pr)
	if err != nil {
		h.log.Error("failed to create PR", "pr_id", body.PullRequestID, "error", err)
		if utils.HandleDomainError(w, err) {
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	h.log.Info("PR created successfully", "pr_id", created.ID)
	utils.RespondJSON(w, http.StatusCreated, resp.CreatePR{PR: mappers.ToDTOPR(created)})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var body req.MergePR
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("invalid JSON in MergePR", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
		return
	}
	if body.PullRequestID == "" {
		h.log.Warn("missing pull_request_id in MergePR")
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id is required")
		return
	}

	h.log.Info("MergePR called", "pr_id", body.PullRequestID)
	merged, err := h.svc.MergePR(r.Context(), body.PullRequestID)
	if err != nil {
		h.log.Error("failed to merge PR", "pr_id", body.PullRequestID, "error", err)
		if utils.HandleDomainError(w, err) {
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	h.log.Info("PR merged successfully", "pr_id", merged.ID)
	utils.RespondJSON(w, http.StatusOK, resp.MergePR{PR: mappers.ToDTOPR(merged)})
}

func (h *PRHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var body req.ReassignReviewer
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("invalid JSON in ReassignPR", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
		return
	}
	if body.PullRequestID == "" || body.OldUserID == "" {
		h.log.Warn("missing required fields in ReassignPR", "body", body)
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id and old_user_id are required")
		return
	}

	h.log.Info("ReassignPR called", "pr_id", body.PullRequestID, "old_reviewer_id", body.OldUserID)
	pr, replacedBy, err := h.svc.ReassignReviewer(r.Context(), body.PullRequestID, body.OldUserID)
	if err != nil {
		h.log.Error("failed to reassign PR", "pr_id", body.PullRequestID, "error", err)
		if utils.HandleDomainError(w, err) {
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	h.log.Info("reviewer reassigned successfully", "pr_id", pr.ID, "new_reviewer_id", replacedBy)
	utils.RespondJSON(w, http.StatusOK, resp.ReassignReviewer{
		PR:         mappers.ToDTOPR(pr),
		ReplacedBy: replacedBy,
	})
}
