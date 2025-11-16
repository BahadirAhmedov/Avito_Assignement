package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
	"Backend-trainee-assignment-autumn-2025/internal/lib/logger/sl"
	"Backend-trainee-assignment-autumn-2025/internal/storage"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/responses"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)


func (p *PullRequest) PullRequestCreate(ctx *gin.Context) {
	const op = "handlers.PullRequestCreate"

	log := p.log.With(
		slog.String("op", op),
	)

	var request requests.PullRequestCreateRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	log.Info("request received",
		slog.String("pull_request_id", request.PullRequestId),
		slog.String("pull_request_name", request.PullRequestName),
		slog.String("author_id", request.AuthorId),
	)

	reviewers, status, err := p.pullRequestProvider.SavePullRequest(request)
	if err != nil {
		if errors.Is(err, storage.ErrPullRequestExists) {
			log.Info("pull request already exists", slog.String("pull_request_id", request.PullRequestId))
			ctx.JSON(http.StatusConflict, responses.Error(PR_EXISTS, "PR already exists"))
			return
		} else if errors.Is(err, storage.ErrAuthorNotFound) {
			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "author not found"))
			return
		} else {
			log.Error("failed to save pull request", sl.Err(err),
				slog.String("pull_request_id", request.PullRequestId),
				slog.String("author_id", request.AuthorId),
			)
			ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to save pull request"))
		}
	}

	log.Info("pull request created",
		slog.String("pull_request_id", request.PullRequestId),
		slog.String("pull_request_name", request.PullRequestName),
		slog.String("author_id", request.AuthorId),
		slog.Int("reviewers_count", len(reviewers)),
		slog.String("status", string(status)),
	)

	responseBody := responses.PullRequestCreateResponse{
		Pr: models.PullRequest{
			PullRequestId:     request.PullRequestId,
			PullRequestName:   request.PullRequestName,
			AuthorId:          request.AuthorId,
			Status:            status,
			AssignedReviewers: reviewers,
		},
	}

	ctx.JSON(http.StatusOK, responseBody)
}

func (p *PullRequest) PullRequestMerge(ctx *gin.Context) {
	const op = "handlers.PullRequestMerge"

	log := p.log.With(
		slog.String("op", op),
	)

	var request requests.PullRequestMergeRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	log.Info("request received",
		slog.String("pull_request_id", request.PullRequestId),
	)

	pullRequest, err := p.pullRequestProvider.MergePullRequest(request)
	if err != nil {
		if errors.Is(err, storage.ErrPullRequestNotFound) {
			log.Info("pull request not found",
				slog.String("pull_request_id", request.PullRequestId),
			)
			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "pull request not found"))
			return
		}

		log.Error("failed to merge pull request", sl.Err(err),
			slog.String("pull_request_id", request.PullRequestId),
		)

		ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to merge pull request"))
		return
	}

	log.Info("pull request merged successfully",
		slog.String("pull_request_id", pullRequest.PullRequestId),
		slog.String("status", string(pullRequest.Status)),
		slog.Int("reviewers_count", len(pullRequest.AssignedReviewers)),
	)

	responseBody := responses.PullRequestMergeResponse{
		Pr: pullRequest,
	}

	ctx.JSON(http.StatusOK, responseBody)

}

func (p *PullRequest) PullRequestReassign(ctx *gin.Context) {
	const op = "handlers.PullRequestReassign"

	log := p.log.With(
		slog.String("op", op),
	)

	var request requests.PullRequestReassignRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
		return
	}

	log.Info("request received",
		slog.String("pull_request_id", request.PullRequestId),
		slog.String("old_user_id", request.OldUserId),
	)

	pullRequest, replacedBy, err := p.pullRequestProvider.ReassignPullRequest(request)
	if err != nil {
		if errors.Is(err, storage.ErrPullRequestNotFound) {
			log.Info("pull request not found",
				slog.String("pull_request_id", request.PullRequestId),
			)
			ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "pull request not found"))
			return
		} else if errors.Is(err, storage.ErrCannotReassingnOnMergedPr) {
			log.Info("cannot reassign on merged PR",
				slog.String("pull_request_id", request.PullRequestId),
			)
			ctx.JSON(http.StatusConflict, responses.Error(PR_MERGED, "cannot reassign on merged PR"))
			return
		} else if errors.Is(err, storage.ErrReviewerIsNotAssignedToThisPr) {
			log.Info("reviewer is not assigned to this PR",
				slog.String("pull_request_id", request.PullRequestId),
			)
			ctx.JSON(http.StatusConflict, responses.Error(NOT_ASSIGNED, "reviewer is not assigned to this PR"))
			return
		} else if errors.Is(err, storage.ErrActiveReplacementCandidateInTeam) {
			log.Info("no active replacement candidate in team")
			ctx.JSON(http.StatusConflict, responses.Error(NO_CANDIDATE, "no active replacement candidate in team"))
			return
		} else {
			log.Error("failed to reassign pull request", sl.Err(err),
				slog.String("pull_request_id", request.PullRequestId),
			)
			ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "unable reassign pull request"))
			return
		}
	}

	log.Info("pull request reassigned successfully",
		slog.String("pull_request_id", pullRequest.PullRequestId),
		slog.Int("reviewers_count", len(pullRequest.AssignedReviewers)),
	)

	responseBody := responses.PullRequestReassignResponse{
		Pr:         pullRequest,
		ReplacedBy: replacedBy,
	}

	ctx.JSON(http.StatusOK, responseBody)

}
