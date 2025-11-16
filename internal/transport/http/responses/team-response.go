package responses

import(
	req "Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
)

type TeamResponse struct{
	Team req.TeamRequest `json:"team"`
}

type TeamGetResponse struct {
    TeamName string        `json:"team_name" binding:"required"`
    Members  []models.User `json:"members" binding:"required"`
}