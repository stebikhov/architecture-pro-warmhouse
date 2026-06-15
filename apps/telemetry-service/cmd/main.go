package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"telemetry-service/internal/db"
	"telemetry-service/internal/handler"
	"telemetry-service/internal/repository"
	"telemetry-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Config struct {
	DatabaseURL    string
	DeviceMgmtURL  string
	TemperatureURL string
	Port           string
	ThresholdMin   float64
	ThresholdMax   float64
	WebhookTimeout time.Duration
}

func LoadConfig() *Config {
	thresholdMin, _ := strconv.ParseFloat(getEnv("THRESHOLD_MIN", "0.0"), 64)
	thresholdMax, _ := strconv.ParseFloat(getEnv("THRESHOLD_MAX", "40.0"), 64)

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/smarthome"),
		DeviceMgmtURL:  getEnv("DEVICE_MGMT_URL", "http://device-management:8080"),
		TemperatureURL: getEnv("TEMPERATURE_API_URL", "http://temperature-api:8081"),
		Port:           getEnv("PORT", ":8080"),
		ThresholdMin:   thresholdMin,
		ThresholdMax:   thresholdMax,
		WebhookTimeout: 10 * time.Second,
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[Telemetry] ")

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

	telemetryRepo := repository.NewTelemetryRepository(database)
	telemetryService := service.NewTelemetryService(telemetryRepo, cfg.DeviceMgmtURL, cfg.TemperatureURL, cfg.ThresholdMin, cfg.ThresholdMax)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiRoutes := router.Group("/api/v1")
	telemetryHandler := handler.NewTelemetryHandler(telemetryService)
	telemetryHandler.RegisterRoutes(apiRoutes)

	webhookRoutes := router.Group("/webhooks")
	{
		webhookRoutes.POST("/device", telemetryHandler.HandleDeviceWebhook)
	}

	wsRoutes := router.Group("/ws")
	{
		wsRoutes.GET("/telemetry", telemetryHandler.HandleWebSocket)
	}

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Telemetry Service starting on %s\n", srv.Addr)
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
