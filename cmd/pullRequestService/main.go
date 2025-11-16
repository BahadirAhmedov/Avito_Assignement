package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"Backend-trainee-assignment-autumn-2025/internal/app"
	"Backend-trainee-assignment-autumn-2025/internal/config"
	"Backend-trainee-assignment-autumn-2025/internal/lib/logger/sl"

)


const (
	envLocal = "local"
	envDev =  "dev"
	envProd = "prod"
)

func main() {
	
	router := gin.Default()
	cfg := config.MustLoad()


	log := setupLogger(cfg.Env)

	// Prod
	// handlers := app.New(log, "postgres", 5432, "postgres", "postgres", "avito")
	// Local
	handlers := app.New(log, cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.Password, cfg.Db.Name)


	router.POST("team/add", handlers.TeamAdd)
	router.GET("team/get", handlers.TeamGet)
	router.POST("/users/setIsActive", handlers.UsersSetIsActive)
	router.GET("/users/getReview", handlers.UsersGetReview)
	router.POST("/pullRequest/create", handlers.PullRequestCreate)
	router.POST("/pullRequest/merge", handlers.PullRequestMerge)
	router.POST("/pullRequest/reassign", handlers.PullRequestReassign)


	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen: %s\\n", sl.Err(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown: %s\\n", sl.Err(err))
		os.Exit(1)	
	}

	<-ctx.Done()
	log.Info("imeout of 5 seconds.")
	log.Info("Server exiting")	

}


func setupLogger(env string) (*slog.Logger){
	var log *slog.Logger
	
	switch env{
		case envLocal:
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
		case envDev:
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)

		case envProd:
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
			)
		}  
	return log
}

