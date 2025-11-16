package models

import "time"


type PullRequest struct{
	PullRequestId string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status string `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	MergedAt *time.Time `json:"mergedAt,omitempty"`
}


type PullRequestShort struct{
	PullRequestId string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status string `json:"status"`
}
