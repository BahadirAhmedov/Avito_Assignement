package requests


type PullRequestCreateRequest struct {
	PullRequestId string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorId string `json:"author_id" binding:"required"`
}

type PullRequestMergeRequest struct {
	PullRequestId string `json:"pull_request_id" binding:"required"`
}

type PullRequestReassignRequest struct {
	PullRequestId string `json:"pull_request_id" binding:"required"`
	OldUserId string `json:"old_user_id" binding:"required"`
}