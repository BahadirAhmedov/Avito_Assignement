package app

import (
	"log/slog"

	"Backend-trainee-assignment-autumn-2025/internal/http-server/handlers"
	"Backend-trainee-assignment-autumn-2025/internal/storage/postgres"
)

func New(
	log *slog.Logger, 
	host string, 
	port int, 
	user string, 
	password string, 
	dbname string,
) (*handlers.PullRequest) {
	storage, err := postgres.New(host, port, user, password, dbname)
	if err != nil {
		panic(err)
	}

	SubscriptionHandlers := handlers.New(*log, storage, storage, storage)
	return SubscriptionHandlers
}