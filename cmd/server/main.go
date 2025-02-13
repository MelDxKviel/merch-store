package main

import (
	"log"
	"net/http"
	"time"

	"merch-store/internal/config"
	"merch-store/internal/handlers"
	"merch-store/internal/middleware"
	"merch-store/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := repository.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo, cfg.JWTSecret)

	router := gin.Default()

	router.POST("/api/auth", handler.Auth)

	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
	{
		apiGroup.GET("/info", handler.GetInfo)
		apiGroup.POST("/sendCoin", handler.SendCoin)
		apiGroup.GET("/buy/:item", handler.BuyItem)
	}

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server is running on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

