package mappers

import (
	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/req"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/resp"
)

// ToDTOUser маппит domain.User -> resp.User
func ToDTOUser(u review.User) resp.User {
	return resp.User{
		UserID:   u.ID,
		Username: u.Name,
		TeamName: u.Team,
		IsActive: u.IsActive,
	}
}

// FromSetIsActiveReq маппит req.SetIsActive -> domain.User
func FromSetIsActiveReq(r req.SetIsActive) review.User {
	return review.User{
		ID:       r.UserID,
		IsActive: r.IsActive,
	}
}
