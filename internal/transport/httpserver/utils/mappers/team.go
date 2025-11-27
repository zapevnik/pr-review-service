package mappers

import (
	"github.com/google/uuid"
	"github.com/zapevnik/pr-review-service/internal/domain/review"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/req"
	"github.com/zapevnik/pr-review-service/internal/transport/httpserver/dto/resp"
)

// TeamAddRequestToArgs конвертирует req.TeamAdd -> имя команды + слайс domain.User
func TeamAddRequestToArgs(r req.TeamAdd) (string, []review.User, error) {
	members := make([]review.User, 0, len(r.Members))

	for _, m := range r.Members {
		id := m.UserID
		if id == "" {
			id = uuid.New().String()
		}

		members = append(members, review.User{
			ID:       id,
			Name:     m.Username,
			IsActive: m.IsActive,
		})
	}

	return r.TeamName, members, nil
}

// TeamToResponse конвертирует domain.Team и []domain.User -> resp.Team
func TeamToResponse(team review.Team, members []review.User) resp.Team {
	respMembers := make([]resp.TeamMember, 0, len(members))

	for _, u := range members {
		respMembers = append(respMembers, resp.TeamMember{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}

	return resp.Team{
		TeamName: team.Name,
		Members:  respMembers,
	}
}
