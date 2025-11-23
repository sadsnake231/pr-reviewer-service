package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sadsnake231/pr-reviewer-service/internal/config"
	"github.com/sadsnake231/pr-reviewer-service/internal/database"
	"github.com/sadsnake231/pr-reviewer-service/internal/handler"
	"github.com/sadsnake231/pr-reviewer-service/internal/repository"
	"github.com/sadsnake231/pr-reviewer-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting PR reviewer service")

	cfg := config.LoadConfig()
	logger.Info("configuration loaded",
		zap.String("host", cfg.ServiceHost),
		zap.String("port", cfg.ServicePort))

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("database connection established")

	if err := database.MigrateUp(ctx, pool); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("database migrations applied successfully")

	userRepo := repository.NewUserRepository(pool, logger)
	teamRepo := repository.NewTeamRepository(pool, logger)
	prRepo := repository.NewPullRequestRepository(pool, logger)

	userService := service.NewUserService(userRepo, logger)
	teamService := service.NewTeamService(teamRepo, userRepo, logger)
	prService := service.NewPullRequestService(prRepo, userService, teamRepo, logger)

	h := handler.NewHandler(userService, teamService, prService, logger)

	router := gin.Default()
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort),
		Handler: router,
	}

	go func() {
		logger.Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}
