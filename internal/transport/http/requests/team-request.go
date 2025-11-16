package requests


import (
	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
)

type TeamRequest struct{
	TeamName string `json:"team_name" binding:"required"`
	Members []models.User `json:"members" binding:"required"`
}
