package handlers

import (
	"log/slog"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
	"Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)

const (
	TEAM_EXISTS     = "TEAM_EXISTS"
	INVALID_REQUEST = "INVALID_REQUEST"
	TEAM_NOT_FOUND  = "TEAM_NOT_FOUND"
	NOT_FOUND       = "NOT_FOUND"
	UNAUTHORIZED    = "UNAUTHORIZED"
	PR_EXISTS       = "PR_EXISTS"
	PR_MERGED       = "PR_MERGED"
	NOT_ASSIGNED    = "NOT_ASSIGNED"
	INTERNAL_ERROR  = "INTERNAL_ERROR"
	NO_CANDIDATE    = "NO_CANDIDATE"
	USER_DELETED    = "USER_DELETED"
	USER_NOT_DELETED = "USER_NOT_DELETED"
)

const (
	adminSecret = "secret"
)

type PullRequest struct {
	log *slog.Logger

	pullRequestProvider PullRequestProvider
	userProvider        UserProvider
	TteamProvide        TeamProvider
}


type TeamProvider interface {
	SaveTeam(requests.TeamRequest) error
	ReadTeam(string) (string, []models.User, error)
}

type UserProvider interface {
	SetIsActive(requests.UsersRequest) (models.User, error)
	ReadReviewUsers(request string) ([]models.PullRequestShort, error)
	DeleteUser(requests.UserDeleteRequest) (models.User, error)
	RestoreUser(requests.UserRestoreRequest) (models.User, error)
}

type PullRequestProvider interface {
	SavePullRequest(request requests.PullRequestCreateRequest) ([]string, string, error)
	MergePullRequest(pullRequestId requests.PullRequestMergeRequest) (models.PullRequest, error)
	ReassignPullRequest(request requests.PullRequestReassignRequest) (models.PullRequest, string, error)
}

func New(
	log slog.Logger,
	pullRequestProvider PullRequestProvider,
	userProvider UserProvider,
	TteamProvide TeamProvider,
) *PullRequest {
	return &PullRequest{
		log:                 &log,
		pullRequestProvider: pullRequestProvider,
		userProvider:        userProvider,
		TteamProvide:        TteamProvide}
}



