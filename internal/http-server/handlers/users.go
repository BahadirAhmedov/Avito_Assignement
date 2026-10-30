package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"Backend-trainee-assignment-autumn-2025/internal/lib/logger/sl"
	"Backend-trainee-assignment-autumn-2025/internal/storage"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/responses"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)

func (p *PullRequest) UsersSetIsActive(ctx *gin.Context) {
	const op = "handlers.UsersSetIsActive"

	log := p.log.With(
		slog.String("op", op),
	)

	auth := ctx.GetHeader("Authorization")

	if auth == "" || auth != "Bearer "+adminSecret {
		log.Info("invalid token", slog.Any("token", auth))

		ctx.JSON(http.StatusUnauthorized, responses.Error(UNAUTHORIZED, "invalid token"))
		return
	}

	var request requests.UsersRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	user, err := p.userProvider.SetIsActive(request)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {

			log.Info("user not found", slog.String("user_id", request.UserId))

			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
			return
		}

		log.Error("failed to update user", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to update user"))
		return
	}

	log.Info("user updated", slog.String("user", user.UserId))

	ctx.JSON(http.StatusOK, responses.UserResponse{User: user})
}


func (p *PullRequest) UsersGetReview(ctx *gin.Context) {
	const op = "handlers.UsersGetReview"

	log := p.log.With(
		slog.String("op", op),
	)

	userId := ctx.Query("user_id")
	if userId == "" {
		log.Info("parameter 'user_id' is required")
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "Parameter 'user_id' is required"))
		return
	}

	log.Info("request received", slog.String("user_id", userId))


	pullRequests, err := p.userProvider.ReadReviewUsers(userId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", slog.String("user_id", userId))
			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
			return
		}

		log.Error("failed to read review users", sl.Err(err),
			slog.String("user_id", userId),
		)
	
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to read review users"))
		return
	}

	log.Info("pull requests retrieved successfully",
		slog.String("user_id", userId),
		slog.Int("pull_requests_count", len(pullRequests)),
	)

	responseBody := responses.UsersReadReviewResponse{
		UserId:       userId,
		PullRequests: pullRequests,
	}
	
	ctx.JSON(http.StatusOK, responseBody)

}


func (p *PullRequest) UsersDelete(ctx *gin.Context) {
	const op = "handlers.UsersDelete"

	log := p.log.With(
		slog.String("op", op),
	)

	auth := ctx.GetHeader("Authorization")

	if auth == "" || auth != "Bearer "+adminSecret {
		log.Info("invalid token", slog.Any("token", auth))

		ctx.JSON(http.StatusUnauthorized, responses.Error(UNAUTHORIZED, "invalid token"))
		return
	}

	var request requests.UserDeleteRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	user, err := p.userProvider.DeleteUser(request)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyDeleted) {

			log.Info("user already deleted", slog.String("user_id", request.UserId))

			ctx.JSON(http.StatusConflict, responses.Error(USER_ALREADY_DELETED, "user already deleted"))
			return
		}

		if errors.Is(err, storage.ErrUserNotFound) {

			log.Info("user not found", slog.String("user_id", request.UserId))

			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
			return
		}

		log.Error("failed to delete user", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to delete user"))
		return
	}

	log.Info("user deleted", slog.String("user", user.UserId))

	ctx.JSON(http.StatusOK, responses.UserResponse{User: user})
}


func (p *PullRequest) UsersRestore(ctx *gin.Context) {
	const op = "handlers.UsersRestore"

	log := p.log.With(
		slog.String("op", op),
	)

	auth := ctx.GetHeader("Authorization")

	if auth == "" || auth != "Bearer "+adminSecret {
		log.Info("invalid token", slog.Any("token", auth))

		ctx.JSON(http.StatusUnauthorized, responses.Error(UNAUTHORIZED, "invalid token"))
		return
	}

	var request requests.UserRestoreRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	user, err := p.userProvider.RestoreUser(request)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotDeleted) {

			log.Info("user is not deleted", slog.String("user_id", request.UserId))

			ctx.JSON(http.StatusConflict, responses.Error(USER_NOT_DELETED, "user is not deleted"))
			return
		}

		if errors.Is(err, storage.ErrUserNotFound) {

			log.Info("user not found", slog.String("user_id", request.UserId))

			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
			return
		}

		log.Error("failed to restore user", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to restore user"))
		return
	}

	log.Info("user restored", slog.String("user", user.UserId))

	ctx.JSON(http.StatusOK, responses.UserResponse{User: user})
}
