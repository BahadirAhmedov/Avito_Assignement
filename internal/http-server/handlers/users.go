package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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


const (
	bearerPrefix     = "Bearer "
	exportFormatJSON = "json"
	exportFormatCSV  = "csv"
)

func (p *PullRequest) UsersExport(ctx *gin.Context) {
	const op = "handlers.UsersExport"

	log := p.log.With(
		slog.String("op", op),
	)

	// Ручка не административная: токен — идентификатор самого пользователя,
	// выгрузить можно только собственные данные.
	auth := ctx.GetHeader("Authorization")
	tokenUserId := strings.TrimPrefix(auth, bearerPrefix)

	if !strings.HasPrefix(auth, bearerPrefix) || tokenUserId == "" {
		log.Info("invalid token", slog.Any("token", auth))

		ctx.JSON(http.StatusUnauthorized, responses.Error(UNAUTHORIZED, "invalid token"))
		return
	}

	userId := ctx.Query("user_id")
	if userId == "" {
		log.Info("parameter 'user_id' is required")
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "Parameter 'user_id' is required"))
		return
	}

	if tokenUserId != userId {
		log.Info("foreign data export attempt",
			slog.String("user_id", userId),
			slog.String("token_user_id", tokenUserId),
		)

		ctx.JSON(http.StatusForbidden, responses.Error(FORBIDDEN, "only own data can be exported"))
		return
	}

	format := ctx.DefaultQuery("format", exportFormatJSON)
	if format != exportFormatJSON && format != exportFormatCSV {
		log.Info("unsupported export format", slog.String("format", format))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "Parameter 'format' must be 'json' or 'csv'"))
		return
	}

	export, err := p.userProvider.ReadUserExport(userId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {

			log.Info("user not found", slog.String("user_id", userId))

			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
			return
		}

		log.Error("failed to export user data", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to export user data"))
		return
	}

	responseBody := responses.UserExportResponse{
		User:                 export.User,
		Teammates:            export.Teammates,
		AuthoredPullRequests: export.AuthoredPullRequests,
		ReviewPullRequests:   export.ReviewPullRequests,
	}

	if format == exportFormatCSV {
		body, err := responses.UserExportCSV(responseBody)
		if err != nil {
			log.Error("failed to encode export to csv", sl.Err(err))
			ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to encode export to csv"))
			return
		}

		log.Info("user data exported", slog.String("user_id", userId), slog.String("format", format))

		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="user_%s_export.csv"`, userId))
		ctx.Data(http.StatusOK, "text/csv; charset=utf-8", body)
		return
	}

	log.Info("user data exported", slog.String("user_id", userId), slog.String("format", format))

	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="user_%s_export.json"`, userId))
	ctx.JSON(http.StatusOK, responseBody)
}
