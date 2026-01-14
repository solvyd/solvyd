package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"github.com/solvyd/solvyd/worker-agent/internal/agent"
	"github.com/solvyd/solvyd/worker-agent/internal/config"
	"github.com/solvyd/solvyd/worker-agent/internal/executor"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Command-line flags
	var (
		apiServer     = flag.String("api-server", getEnv("SOLVYD_API_URL", "http://localhost:8080"), "API server address")
		workerName    = flag.String("name", getEnv("SOLVYD_WORKER_NAME", ""), "Worker name (defaults to hostname)")
		maxConcurrent = flag.Int("max-concurrent", getEnvInt("SOLVYD_MAX_CONCURRENT_BUILDS", 2), "Maximum concurrent builds")
		labels        = flag.StringSlice("label", []string{}, "Worker labels (key=value)")
		logLevel      = flag.String("log-level", getEnv("SOLVYD_LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
		isolationType = flag.String("isolation", getEnv("SOLVYD_ISOLATION", "docker"), "Build isolation type (docker, process, vm)")
	)

	flag.Parse()

	// Set log level
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Msg("Starting Solvyd Worker Agent")

	// Generate or use worker name
	if *workerName == "" {
		hostname, _ := os.Hostname()
		*workerName = fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
	}

	// Parse labels
	labelMap := make(map[string]string)
	for _, label := range *labels {
		// Parse label in format key=value
		// Simplified for now
		labelMap["custom"] = label
	}

	// Create config
	cfg := &config.Config{
		APIServer:     *apiServer,
		WorkerName:    *workerName,
		MaxConcurrent: *maxConcurrent,
		Labels:        labelMap,
		IsolationType: *isolationType,
	}

	// Create executor
	exec, err := executor.NewExecutor(cfg.IsolationType)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create executor")
	}

	// Create agent
	agent, err := agent.NewAgent(cfg, exec)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create agent")
	}

	// Start agent
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go agent.Start(ctx)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down worker agent...")
	cancel()
	// Give it time to finish current builds
	time.Sleep(5 * time.Second)

	log.Info().Msg("Worker agent exited")
}

// getEnv gets environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intVal == 1 {
			return defaultValue
		}
	}
	return defaultValue
}
