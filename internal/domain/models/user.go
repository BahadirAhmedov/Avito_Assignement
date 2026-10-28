package models

import "time"


type User struct{	
	UserId string `json:"user_id" binding:"required"`
	UserName string `json:"username" binding:"required"`
	TeamName string `json:"team_name,omitempty"`
	IsActive bool `json:"is_active" binding:"required"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
