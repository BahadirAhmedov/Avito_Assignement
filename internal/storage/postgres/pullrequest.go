package postgres

import (
	"database/sql"
	"fmt"
	"errors"
	"math/rand"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
	"Backend-trainee-assignment-autumn-2025/internal/storage"
	req "Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)

func (s *Storage) SavePullRequest(request req.PullRequestCreateRequest) ([]string, string, error) {
	const op = "storage.postgres.SavePullRequest"

	tx, err := s.db.Begin()
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id=$1)", request.PullRequestId).Scan(&exists)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return nil, "", fmt.Errorf("%s: %w", op, storage.ErrPullRequestExists)
	}

	var teamName string
	err = tx.QueryRow("SELECT team_name FROM users WHERE user_id=$1", request.AuthorId).Scan(&teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("%s: %w", op, storage.ErrAuthorNotFound)
		}
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := tx.Query("SELECT user_id FROM users WHERE team_name=$1 AND user_id<>$2 AND is_active=true ORDER BY RANDOM() LIMIT 2", teamName, request.AuthorId)
	var reviewers []string
	for rows.Next() {
		var uid string

		if err = rows.Scan(&uid); err != nil {
			return nil, "", fmt.Errorf("%s: %w", op, err)
		}

		reviewers = append(reviewers, uid)
	}

	status := "OPEN"
	_, err = tx.Exec("INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES ($1,$2,$3,$4)", request.PullRequestId, request.PullRequestName, request.AuthorId, status)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	for _, r := range reviewers {
		_, err = tx.Exec("INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1,$2)", request.PullRequestId, r)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return reviewers, status, nil
}

func (s *Storage) MergePullRequest(request req.PullRequestMergeRequest) (models.PullRequest, error) {
	const op = "storage.postgres.MergePullRequest"

	tx, err := s.db.Begin()
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var pullRequestId string
	err = tx.QueryRow("SELECT pull_request_id  FROM pull_requests WHERE pull_request_id=$1", request.PullRequestId).Scan(&pullRequestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.PullRequest{},  fmt.Errorf("%s: %w", op, storage.ErrPullRequestNotFound)

		}
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}

	var pullRequest models.PullRequest

	status := "MERGED"

	err = tx.QueryRow(
		`UPDATE pull_requests
	SET status = $1,
    	merged_at = COALESCE(merged_at, NOW())	
	WHERE pull_request_id = $2
	RETURNING pull_request_id, pull_request_name, author_id, status, merged_at;
`, status, request.PullRequestId,
	).Scan(&pullRequest.PullRequestId, &pullRequest.PullRequestName, &pullRequest.AuthorId, &pullRequest.Status, &pullRequest.MergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.PullRequest{}, fmt.Errorf("%s: %w", op, storage.ErrAuthorNotFound)
		}
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := tx.Query("SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1", request.PullRequestId)

	for rows.Next() {
		var uid string
		rows.Scan(&uid)
		if err = rows.Scan(&uid); err != nil {
			return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
		}

		pullRequest.AssignedReviewers = append(pullRequest.AssignedReviewers, uid)
	}

	return pullRequest, nil
}

func (s *Storage) ReassignPullRequest(request req.PullRequestReassignRequest) (models.PullRequest, string, error) {
	const op = "storage.postgres.ReassignPullRequest"

	tx, err := s.db.Begin()
	if err != nil {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			fmt.Println(p)
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()


	var pullRequestName, authorId, status string

	err = tx.QueryRow("SELECT pull_request_name, author_id, status FROM pull_requests WHERE pull_request_id=$1", request.PullRequestId).Scan(&pullRequestName, &authorId, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, storage.ErrPullRequestNotFound)
		}
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	if status == "MERGED" {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, storage.ErrCannotReassingnOnMergedPr)
	}



	var pullRequestId string
	err = tx.QueryRow("SELECT pull_request_id FROM pr_reviewers WHERE pull_request_id=$1 AND user_id =$2", request.PullRequestId, request.OldUserId).Scan(&pullRequestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, storage.ErrReviewerIsNotAssignedToThisPr)
		}
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	var teamName string
	err = tx.QueryRow("SELECT team_name FROM users WHERE user_id =$1", request.OldUserId).Scan(&teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}

	rows, err := tx.Query("SELECT user_id FROM users WHERE team_name=$1 AND user_id <> $2  AND user_id <> $3 AND is_active=true ", teamName, authorId, request.OldUserId)
	if err != nil {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)	
	}


	var reviewers []string
	for rows.Next() {
		var uid string
		if err = rows.Scan(&uid); err != nil {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
		}

		reviewers = append(reviewers, uid)
	}


	if err = rows.Err(); err != nil {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	var reviewersFiltered []string
	for _, r := range reviewers {

		var exist bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE user_id=$1)", r).Scan(&exist)
		if err != nil {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
		}

		if !exist {
			reviewersFiltered = append(reviewersFiltered, r)
		}
	}

	if len(reviewersFiltered) == 0 {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, storage.ErrActiveReplacementCandidateInTeam)
	}

	n := rand.Intn(len(reviewersFiltered))

	var replacedBy string
	err = tx.QueryRow(`
        UPDATE pr_reviewers
        SET user_id = $1
        WHERE user_id = $2 RETURNING user_id;
    `, reviewersFiltered[n], request.OldUserId).Scan(&replacedBy)
	if err != nil {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	rows, err = tx.Query("SELECT user_id FROM pr_reviewers WHERE pull_request_id=$1", pullRequestId)
	if err != nil {
		return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
	}


	var assignedReviewers []string

	for rows.Next() {
		var uid string
		if err = rows.Scan(&uid); err != nil {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
		}
	
		assignedReviewers = append(assignedReviewers, uid)
	
		if err = rows.Err(); err != nil {
			return models.PullRequest{}, "", fmt.Errorf("%s: %w", op, err)
		}
	}
	
	fmt.Println("reviewersAll: ", assignedReviewers)


	pullRequest := models.PullRequest{
		PullRequestId: pullRequestId,
		PullRequestName: pullRequestName,
		AuthorId: authorId,
		Status: status,
		AssignedReviewers: assignedReviewers,
	}
	
	return pullRequest, replacedBy, nil
}





