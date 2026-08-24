package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	trainingdomain "atcoder_shojin/backend/internal/domain/training"
	traininghandler "atcoder_shojin/backend/internal/handler/training"
	"atcoder_shojin/backend/internal/infrastructure/atcoderproblems"
	trainingrepo "atcoder_shojin/backend/internal/infrastructure/training"
	trainingusecase "atcoder_shojin/backend/internal/usecase/training"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := openDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = runMigrations(context.Background(), db); err != nil {
		log.Fatal(err)
	}
	repository := trainingrepo.NewRepository(db)
	client := atcoderproblems.NewClient(config.AtCoderProblemsURL, nil)
	usecase := trainingusecase.New(repository, client, trainingdomain.DefaultConfig())
	handler := traininghandler.New(usecase)
	server := echo.New()
	server.HideBanner = config.Environment == "production"
	server.Use(middleware.Logger(), middleware.Recover(), middleware.BodyLimit("1M"), middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{config.FrontendOrigin}, AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions}, AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept}}))
	server.GET("/healthz", healthHandler(db))
	api := server.Group("/apis")
	api.POST("/training/sessions", handler.Start)
	api.GET("/training/sessions/active", handler.Active)
	api.GET("/training/sessions", handler.History)
	api.GET("/training/sessions/:id", handler.Get)
	api.POST("/training/sessions/:id/sync", handler.Sync)
	api.POST("/training/sessions/:id/abort", handler.Abort)
	registerStaticFrontend(server, config.StaticDir)
	appContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errorsChannel := make(chan error, 1)
	go func() { errorsChannel <- server.Start("0.0.0.0:" + config.Port) }()
	select {
	case serverError := <-errorsChannel:
		if serverError != nil && !errors.Is(serverError, http.ErrServerClosed) {
			log.Fatal(serverError)
		}
	case <-appContext.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = server.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}
}
