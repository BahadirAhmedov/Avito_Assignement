package responses

import "Backend-trainee-assignment-autumn-2025/internal/domain/models"


type UserResponse struct {
	User models.User `json:"user"`
}

type UsersReadReviewResponse struct{
	UserId string`json:"user_id"`
	PullRequests []models.PullRequestShort `json:"pull_requests"`
}