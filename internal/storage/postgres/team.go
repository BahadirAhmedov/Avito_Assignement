package postgres


import (

	"database/sql"
	"fmt"
	"errors"
	"github.com/lib/pq"

	"Backend-trainee-assignment-autumn-2025/internal/domain/models"
	"Backend-trainee-assignment-autumn-2025/internal/storage"
	req "Backend-trainee-assignment-autumn-2025/internal/transport/http/requests"
)


func (s *Storage) SaveTeam(request req.TeamRequest) error {
	const op = "storage.postgres.SaveTeam"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
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

	_, err = tx.Exec(`INSERT INTO teams (team_name) VALUES ($1)`, request.TeamName)
	if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("%w", storage.ErrTeamNameExists)
			}
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, member := range request.Members {

		_, err = tx.Exec(`
            INSERT INTO users (user_id, username, is_active, team_name)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id) DO UPDATE
            SET username = EXCLUDED.username,
                is_active = EXCLUDED.is_active,
                team_name = EXCLUDED.team_name, 
                deleted_at = NULL,
                updated_at = NOW()
        `, member.UserId, member.UserName, member.IsActive, request.TeamName)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) ReadTeam(teamName string) (string, []models.User, error) {
	const op = "storage.postgres.ReadTeam"


	row := s.db.QueryRow("SELECT team_name FROM teams WHERE team_name = $1", teamName)

	var team string

	err := row.Scan(&team)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, fmt.Errorf("%s: %w", op, storage.ErrTeamNameNotFound)
		}
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.Query(`
	SELECT user_id, username, is_active
	FROM users
	WHERE team_name = $1 AND deleted_at IS NULL`, teamName)
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var member models.User
		if err := rows.Scan(&member.UserId, &member.UserName, &member.IsActive); err != nil {
			return "", nil, fmt.Errorf("%s: %w", op, err)
		}
		members = append(members, member)
	}

	if members == nil {
		members = []models.User{}
	}

	return team, members, nil
}