package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/api-server/internal/config"
	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/handlers"
	"github.com/solvyd/solvyd/api-server/internal/metrics"
	"github.com/solvyd/solvyd/api-server/internal/scheduler"
	"github.com/solvyd/solvyd/api-server/internal/worker"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting Ritmo API Server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Initialize database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connection established")

	// Initialize metrics
	metricsCollector := metrics.NewCollector()

	// Initialize worker manager
	workerMgr := worker.NewManager(db, metricsCollector)
	go workerMgr.Start(context.Background())

	// Initialize scheduler
	sched := scheduler.NewScheduler(db, workerMgr, metricsCollector)
	go sched.Start(context.Background())

	// Initialize HTTP router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	router.HandleFunc("/ready", handlers.ReadinessCheck(db)).Methods("GET")

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Jobs endpoints
	jobHandler := handlers.NewJobHandler(db)
	apiV1.HandleFunc("/jobs", jobHandler.ListJobs).Methods("GET")
	apiV1.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	apiV1.HandleFunc("/jobs/{id}", jobHandler.GetJob).Methods("GET")
	apiV1.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	apiV1.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")
	apiV1.HandleFunc("/jobs/{id}/trigger", jobHandler.TriggerJob).Methods("POST")

	// Builds endpoints
	buildHandler := handlers.NewBuildHandler(db)
	apiV1.HandleFunc("/builds", buildHandler.ListBuilds).Methods("GET")
	apiV1.HandleFunc("/builds/{id}", buildHandler.GetBuild).Methods("GET")
	apiV1.HandleFunc("/builds/{id}/cancel", buildHandler.CancelBuild).Methods("POST")
	apiV1.HandleFunc("/builds/{id}/logs", buildHandler.GetBuildLogs).Methods("GET")
	apiV1.HandleFunc("/builds/{id}/artifacts", buildHandler.ListArtifacts).Methods("GET")

	// Workers endpoints
	workerHandler := handlers.NewWorkerHandler(db, workerMgr)
	apiV1.HandleFunc("/workers", workerHandler.ListWorkers).Methods("GET")
	apiV1.HandleFunc("/workers/{id}", workerHandler.GetWorker).Methods("GET")
	apiV1.HandleFunc("/workers/{id}", workerHandler.UpdateWorker).Methods("PUT")
	apiV1.HandleFunc("/workers/{id}/drain", workerHandler.DrainWorker).Methods("POST")

	// Deployments endpoints
	deploymentHandler := handlers.NewDeploymentHandler(db)
	apiV1.HandleFunc("/deployments", deploymentHandler.ListDeployments).Methods("GET")
	apiV1.HandleFunc("/deployments", deploymentHandler.CreateDeployment).Methods("POST")
	apiV1.HandleFunc("/deployments/{id}", deploymentHandler.GetDeployment).Methods("GET")
	apiV1.HandleFunc("/deployments/{id}/rollback", deploymentHandler.RollbackDeployment).Methods("POST")

	// Plugins endpoints
	pluginHandler := handlers.NewPluginHandler(db)
	apiV1.HandleFunc("/plugins", pluginHandler.ListPlugins).Methods("GET")
	apiV1.HandleFunc("/plugins/{id}", pluginHandler.GetPlugin).Methods("GET")
	apiV1.HandleFunc("/plugins", pluginHandler.InstallPlugin).Methods("POST")

	// Metrics endpoint (Prometheus)
	router.Handle("/metrics", metrics.Handler())

	// Webhooks endpoint
	webhookHandler := handlers.NewWebhookHandler(db, sched)
	router.HandleFunc("/webhooks/{source}/{jobId}", webhookHandler.HandleWebhook).Methods("POST")

	// WebSocket for real-time updates
	wsHandler := handlers.NewWebSocketHandler()
	router.HandleFunc("/ws", wsHandler.HandleConnection)

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(router)

	// HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Port).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
