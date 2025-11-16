package responses


import(
	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
)

type PullRequestCreateResponse struct {
	Pr models.PullRequest `json:"pr"`
}

type PullRequestMergeResponse struct {
	Pr models.PullRequest `json:"pr"`
}

type PullRequestReassignResponse struct{
	Pr models.PullRequest `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}
