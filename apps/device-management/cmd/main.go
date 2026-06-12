package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"device-management/internal/db"
	"device-management/internal/handler"
	"device-management/internal/repository"
	"device-management/internal/service"

	"github.com/gin-gonic/gin"
)

type Config struct {
	DatabaseURL    string
	WebhookURL     string
	Port           string
	WebhookTimeout time.Duration
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/smarthome"),
		WebhookURL:     getEnv("TELEMETRY_WEBHOOK_URL", "http://telemetry-service:8080/webhooks/device"),
		Port:           getEnv("PORT", ":8080"),
		WebhookTimeout: 10 * time.Second,
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[DeviceMgmt] ")

	cfg := LoadConfig()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer database.Close()

	log.Println("Connected to database successfully")

	deviceRepo := repository.NewDeviceRepository(database)
	deviceService := service.NewDeviceService(deviceRepo, cfg.WebhookURL)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiRoutes := router.Group("/api/v1")
	deviceHandler := handler.NewDeviceHandler(deviceService)
	deviceHandler.RegisterRoutes(apiRoutes)

	webhookRoutes := router.Group("/webhooks")
	deviceHandler.RegisterWebhookRoutes(webhookRoutes)

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Device Management Service starting on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
