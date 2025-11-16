package postgres

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func New(host string, port int, user string, password string, dbname string) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return &Storage{}, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return &Storage{}, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("Successfully connected!")

	return &Storage{db: db}, nil
}










