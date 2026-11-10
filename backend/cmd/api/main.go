package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"whisper/backend/internal/config"
	"whisper/backend/internal/database"
	"whisper/backend/internal/handler"
	"whisper/backend/internal/realtime"
	"whisper/backend/internal/repository"
	"whisper/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStartup()
	db, err := database.Connect(startupCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(startupCtx, db); err != nil {
		log.Fatal(err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	adminRepository := repository.NewAdminRepository(db)
	conversationService := service.NewConversationService(adminRepository, repository.NewConversationRepository(db))
	messageService := service.NewMessageService(repository.NewMessageRepository(db))
	authService := service.NewAuthService(adminRepository)
	api := handler.NewAPI(
		authService,
		conversationService,
		messageService,
		realtime.NewHandler(authService, conversationService, messageService, cfg.FrontendOrigin),
		cfg,
	)
	router.Use(api.CORS())
	api.RegisterRoutes(router)
	router.GET("/api/v1/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "service": "whisper-api", "database": "disconnected"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "whisper-api", "database": "connected"})
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Whisper API listening on :%s", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-stopCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		}
		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped with error: %v", err)
		}
	}
}
