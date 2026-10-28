package requests


type UsersRequest struct{
	UserId string `json:"user_id" binding:"required"`
	IsActive *bool `json:"is_active" binding:"required"`
}

type UserDeleteRequest struct{
	UserId string `json:"user_id" binding:"required"`
}

type UserRestoreRequest struct{
	UserId string `json:"user_id" binding:"required"`
}
