package responses

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
)

// Выгрузка в CSV — один плоский файл на все разделы: колонка section говорит,
// из какой части выгрузки строка, лишние для раздела ячейки остаются пустыми.
var userExportCSVHeader = []string{
	"section", "user_id", "username", "team_name", "is_active",
	"pull_request_id", "pull_request_name", "author_id", "status",
}

// UserExportCSV раскладывает ту же структуру, что уходит клиенту в JSON, в CSV-файл.
func UserExportCSV(response UserExportResponse) ([]byte, error) {
	var buf bytes.Buffer

	rows := [][]string{userExportCSVHeader}

	rows = append(rows, userExportUserRow("profile", response.User))

	for _, teammate := range response.Teammates {
		rows = append(rows, userExportUserRow("teammate", teammate))
	}

	for _, pullRequest := range response.AuthoredPullRequests {
		rows = append(rows, userExportPullRequestRow("authored_pull_request", pullRequest))
	}

	for _, pullRequest := range response.ReviewPullRequests {
		rows = append(rows, userExportPullRequestRow("review_pull_request", pullRequest))
	}

	writer := csv.NewWriter(&buf)

	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func userExportUserRow(section string, user models.User) []string {
	return []string{
		section, user.UserId, user.UserName, user.TeamName, strconv.FormatBool(user.IsActive),
		"", "", "", "",
	}
}

func userExportPullRequestRow(section string, pullRequest models.PullRequestShort) []string {
	return []string{
		section, "", "", "", "",
		pullRequest.PullRequestId, pullRequest.PullRequestName, pullRequest.AuthorId, pullRequest.Status,
	}
}
