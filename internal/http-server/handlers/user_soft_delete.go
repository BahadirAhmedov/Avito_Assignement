package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"Backend-trainee-assignment-autumn-2025/internal/storage"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/problem"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/responses"
)

func (p *PullRequest) UsersDelete(c *gin.Context) {
	if !adminAuthorized(c) {
		c.JSON(http.StatusUnauthorized, problem.New(http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token"))
		return
	}

	var in requests.UserDeleteRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, problem.New(http.StatusBadRequest, "bad request", "request body is invalid"))
		return
	}

	user, err := p.userProvider.DeleteUser(in)
	switch {
	case errors.Is(err, storage.ErrUserNotFound):
		c.JSON(http.StatusNotFound, problem.New(http.StatusNotFound, "not found", "user does not exist"))
	case errors.Is(err, storage.ErrUserAlreadyDeleted):
		c.JSON(http.StatusConflict, problem.New(http.StatusConflict, "conflict", "user is already marked as deleted"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, problem.New(http.StatusInternalServerError, "internal error", "failed to mark user as deleted"))
	default:
		c.JSON(http.StatusOK, responses.UserDeleteResponse{User: user})
	}
}

func (p *PullRequest) UsersRestore(c *gin.Context) {
	if !adminAuthorized(c) {
		c.JSON(http.StatusUnauthorized, problem.New(http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token"))
		return
	}

	var in requests.UserRestoreRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, problem.New(http.StatusBadRequest, "bad request", "request body is invalid"))
		return
	}

	user, err := p.userProvider.RestoreUser(in)
	switch {
	case errors.Is(err, storage.ErrUserNotFound):
		c.JSON(http.StatusNotFound, problem.New(http.StatusNotFound, "not found", "user does not exist"))
	case errors.Is(err, storage.ErrUserNotDeleted):
		c.JSON(http.StatusConflict, problem.New(http.StatusConflict, "conflict", "user is not marked as deleted"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, problem.New(http.StatusInternalServerError, "internal error", "failed to restore user"))
	default:
		c.JSON(http.StatusOK, responses.UserRestoreResponse{User: user})
	}
}

func adminAuthorized(c *gin.Context) bool {
	header := c.GetHeader("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	return ok && token == adminSecret
}
