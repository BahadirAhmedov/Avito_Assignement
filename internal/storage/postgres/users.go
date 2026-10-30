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
	var teamName sql.NullString

	err := s.db.QueryRow(
		`UPDATE users SET is_active = $1, updated_at = NOW()
             WHERE user_id = $2 AND deleted_at IS NULL
             RETURNING user_id, username, team_name, is_active`,
		request.IsActive, request.UserId,
	).Scan(&user.UserId, &user.UserName, &teamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.TeamName = teamName.String

	return user, nil
}


func (s *Storage) DeleteUser(request req.UserDeleteRequest) (models.User, error) {
	const op = "storage.postgres.DeleteUser"

	var user models.User
	var teamName sql.NullString

	err := s.db.QueryRow(
		`UPDATE users SET deleted_at = NOW(), updated_at = NOW()
             WHERE user_id = $1 AND deleted_at IS NULL
             RETURNING user_id, username, team_name, is_active`,
		request.UserId,
	).Scan(&user.UserId, &user.UserName, &teamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			deleted, existsErr := s.userIsDeleted(request.UserId)
			if existsErr != nil {
				return models.User{}, fmt.Errorf("%s: %w", op, existsErr)
			}
			if deleted {
				return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyDeleted)
			}
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.TeamName = teamName.String

	return user, nil
}


func (s *Storage) RestoreUser(request req.UserRestoreRequest) (models.User, error) {
	const op = "storage.postgres.RestoreUser"

	var user models.User
	var teamName sql.NullString

	err := s.db.QueryRow(
		`UPDATE users SET deleted_at = NULL, updated_at = NOW()
             WHERE user_id = $1 AND deleted_at IS NOT NULL
             RETURNING user_id, username, team_name, is_active`,
		request.UserId,
	).Scan(&user.UserId, &user.UserName, &teamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			deleted, existsErr := s.userIsDeleted(request.UserId)
			if existsErr != nil {
				return models.User{}, fmt.Errorf("%s: %w", op, existsErr)
			}
			if !deleted {
				return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotDeleted)
			}
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.TeamName = teamName.String

	return user, nil
}


// userIsDeleted различает «пользователя нет» и «пользователь уже в нужном состоянии»:
// UPDATE с фильтром по deleted_at в обоих случаях отдаёт sql.ErrNoRows.
func (s *Storage) userIsDeleted(userId string) (bool, error) {
	const op = "storage.postgres.userIsDeleted"

	var deleted bool

	err := s.db.QueryRow(
		"SELECT deleted_at IS NOT NULL FROM users WHERE user_id = $1", userId,
	).Scan(&deleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return deleted, nil
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


func (s *Storage) ReadUserExport(userId string) (models.UserExport, error) {
	const op = "storage.postgres.ReadUserExport"

	var export models.UserExport
	var teamName sql.NullString

	err := s.db.QueryRow(`
        SELECT user_id, username, team_name, is_active
        FROM users
        WHERE user_id = $1 AND deleted_at IS NULL`, userId,
	).Scan(&export.User.UserId, &export.User.UserName, &teamName, &export.User.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserExport{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.UserExport{}, fmt.Errorf("%s: %w", op, err)
	}

	export.User.TeamName = teamName.String

	export.Teammates, err = s.readTeammates(userId, teamName)
	if err != nil {
		return models.UserExport{}, fmt.Errorf("%s: %w", op, err)
	}

	export.AuthoredPullRequests, err = s.readAuthoredPullRequests(userId)
	if err != nil {
		return models.UserExport{}, fmt.Errorf("%s: %w", op, err)
	}

	export.ReviewPullRequests, err = s.readReviewPullRequests(userId)
	if err != nil {
		return models.UserExport{}, fmt.Errorf("%s: %w", op, err)
	}

	return export, nil
}


// readTeammates отдаёт живых коллег по команде без самого пользователя;
// team_name nullable, у пользователя без команды коллег нет.
func (s *Storage) readTeammates(userId string, teamName sql.NullString) ([]models.User, error) {
	const op = "storage.postgres.readTeammates"

	var teammates []models.User

	if teamName.Valid {
		rows, err := s.db.Query(`
            SELECT user_id, username, team_name, is_active
            FROM users
            WHERE team_name = $1 AND user_id <> $2 AND deleted_at IS NULL
            ORDER BY user_id`, teamName.String, userId)

		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		defer rows.Close()

		for rows.Next() {
			var teammate models.User

			if err = rows.Scan(&teammate.UserId, &teammate.UserName, &teammate.TeamName, &teammate.IsActive); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			teammates = append(teammates, teammate)
		}
	}

	if teammates == nil {
		teammates = []models.User{}
	}

	return teammates, nil
}


func (s *Storage) readAuthoredPullRequests(userId string) ([]models.PullRequestShort, error) {
	const op = "storage.postgres.readAuthoredPullRequests"

	rows, err := s.db.Query(`
        SELECT pull_request_id, pull_request_name, author_id, status
        FROM pull_requests
        WHERE author_id = $1
        ORDER BY created_at DESC`, userId)

	if err != nil {
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


func (s *Storage) readReviewPullRequests(userId string) ([]models.PullRequestShort, error) {
	const op = "storage.postgres.readReviewPullRequests"

	rows, err := s.db.Query(`
        SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
        FROM pull_requests p
        JOIN pr_reviewers r ON r.pull_request_id = p.pull_request_id
        WHERE r.user_id = $1
        ORDER BY p.created_at DESC`, userId)

	if err != nil {
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
