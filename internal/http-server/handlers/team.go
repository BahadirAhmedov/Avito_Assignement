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

func (p *PullRequest) TeamAdd(ctx *gin.Context) {
	const op = "handlers.TeamAdd"

	log := p.log.With(
		slog.String("op", op),
	)

	var request requests.TeamRequest

	err := ctx.BindJSON(&request)
	if err != nil {

		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))

		return
	}
	log.Info("request body decoded", slog.Any("request", request))

	err = p.TteamProvide.SaveTeam(request)

	if err != nil {
		if errors.Is(err, storage.ErrTeamNameExists) {
			log.Info("team_name already exists", slog.String("url", request.TeamName))
			ctx.JSON(http.StatusBadRequest, responses.Error(TEAM_EXISTS, "team_name already exists"))
			return
		}

		log.Error("failed to add team", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to add team"))
		return
	}

	log.Info("team added")
	ctx.JSON(http.StatusOK, responses.TeamResponse{Team: request})

}

func (p *PullRequest) TeamGet(ctx *gin.Context) {
	const op = "handlers.TeamGet"

	log := p.log.With(
		slog.String("op", op),
	)

	teamName := ctx.Query("team_name")

	if teamName == "" {
		log.Info("parameter 'team_name' is required")
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "parameter 'team_name' is required"))
		return
	}

	log.Info("request parameter", slog.String("team_name", teamName))

	teamName, members, err := p.TteamProvide.ReadTeam(teamName)

	if err != nil {
		if errors.Is(err, storage.ErrTeamNameNotFound) {
			log.Info("team_name not found", slog.String("team_name", teamName))
			ctx.JSON(http.StatusNotFound, responses.Error(TEAM_NOT_FOUND, "team_name not found"))
			return
		}

		log.Error("failed to get team", sl.Err(err))
		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to add team"))
	}

	log.Info("team retrieved",
		slog.String("team_name", teamName),
		slog.Int("members_count", len(members)),
	)

	ctx.JSON(http.StatusOK, responses.TeamGetResponse{TeamName: teamName, Members: members})
}