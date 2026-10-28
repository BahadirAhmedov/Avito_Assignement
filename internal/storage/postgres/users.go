package postgres

import (
	"database/sql"
	"fmt"
	"errors"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
	"Backend-trainee-assignment-autumn-2025/internal/storage"
	req "Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)


func (s *Storage) SetIsActive(request req.UsersRequest) (models.User, error) {
	const op = "storage.postgres.SetIsActive"

	var user models.User

	err := s.db.QueryRow(
		`UPDATE users SET is_active = $1, updated_at = NOW()
             WHERE user_id = $2 AND deleted_at IS NULL
             RETURNING user_id, username, team_name, is_active`,
		request.IsActive, request.UserId,
	).Scan(&user.UserId, &user.UserName, &user.TeamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}


func (s *Storage) ReadReviewUsers(request string)([]models.PullRequestShort, error)  {
	const op = "storage.postgres.ReadReviewUsers"

	
	var exists bool
    err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1 AND deleted_at IS NULL)", request).Scan(&exists)
    if err != nil { 
		return nil, fmt.Errorf("%s: %w", op, err) 
	}
    if !exists {
        return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
    }

	rows, err := s.db.Query(`
        SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
        FROM pull_requests p
        JOIN pr_reviewers r ON r.pull_request_id = p.pull_request_id
        WHERE r.user_id = $1
        ORDER BY p.created_at DESC
    `, request)

    if err != nil { 
		fmt.Println(err)
		return nil, fmt.Errorf("%s: %w", op, err) 
	}
    defer rows.Close()

    var pullRequests []models.PullRequestShort
    for rows.Next() {
        var pullRequest models.PullRequestShort
       
		if err = rows.Scan(&pullRequest.PullRequestId, &pullRequest.PullRequestName, &pullRequest.AuthorId, &pullRequest.Status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		pullRequests = append(pullRequests, pullRequest)
    }


	if pullRequests == nil {
		pullRequests = []models.PullRequestShort{}
	}

	return pullRequests, nil
}

func (s *Storage) DeleteUser(request req.UserDeleteRequest) (models.User, error) {
	var deletedAt sql.NullTime
	err := s.db.QueryRow("SELECT deleted_at FROM users WHERE user_id = $1", request.UserId).Scan(&deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if deletedAt.Valid {
		return models.User{}, storage.ErrUserAlreadyDeleted
	}

	var user models.User
	var teamName sql.NullString
	var deleted sql.NullTime
	err = s.db.QueryRow(
		`UPDATE users SET deleted_at = NOW(), updated_at = NOW()
		 WHERE user_id = $1 AND deleted_at IS NULL
		 RETURNING user_id, username, team_name, is_active, deleted_at`,
		request.UserId,
	).Scan(&user.UserId, &user.UserName, &teamName, &user.IsActive, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	if teamName.Valid {
		user.TeamName = teamName.String
	}
	if deleted.Valid {
		user.DeletedAt = &deleted.Time
	}
	return user, nil
}

func (s *Storage) RestoreUser(request req.UserRestoreRequest) (models.User, error) {
	var deletedAt sql.NullTime
	err := s.db.QueryRow("SELECT deleted_at FROM users WHERE user_id = $1", request.UserId).Scan(&deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if !deletedAt.Valid {
		return models.User{}, storage.ErrUserNotDeleted
	}

	var user models.User
	var teamName sql.NullString
	err = s.db.QueryRow(
		`UPDATE users SET deleted_at = NULL, updated_at = NOW()
		 WHERE user_id = $1 AND deleted_at IS NOT NULL
		 RETURNING user_id, username, team_name, is_active`,
		request.UserId,
	).Scan(&user.UserId, &user.UserName, &teamName, &user.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	if teamName.Valid {
		user.TeamName = teamName.String
	}
	return user, nil
}
