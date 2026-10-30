package models


// UserExport — выгрузка данных одного пользователя из всех таблиц, где он упоминается:
// users (собственный профиль и состав его команды), pull_requests (авторство),
// pr_reviewers (назначения на ревью).
type UserExport struct {
	User                 User               `json:"user"`
	Teammates            []User             `json:"teammates"`
	AuthoredPullRequests []PullRequestShort `json:"authored_pull_requests"`
	ReviewPullRequests   []PullRequestShort `json:"review_pull_requests"`
}
