package config

import (
	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Port        int
	LogLevel    string
	DatabaseURL string

	// CORS
	CORSAllowedOrigins []string

	// Worker management
	WorkerHeartbeatTimeout int // seconds
	MaxWorkersPerJob       int

	// Scheduling
	SchedulerTickInterval int // seconds
	MaxConcurrentBuilds   int

	// Plugins
	PluginDirectory string

	// Storage
	ArtifactStorageType   string // s3, gcs, local, minio
	ArtifactStorageConfig map[string]string

	// Security
	JWTSecret string
}

// Load reads configuration from environment and config file
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/ritmo")

	// Set defaults
	viper.SetDefault("port", 8080)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("database_url", "postgres://ritmo:ritmo_dev_password@localhost:5432/ritmo?sslmode=disable")
	viper.SetDefault("cors_allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	viper.SetDefault("worker_heartbeat_timeout", 60)
	viper.SetDefault("max_workers_per_job", 10)
	viper.SetDefault("scheduler_tick_interval", 5)
	viper.SetDefault("max_concurrent_builds", 100)
	viper.SetDefault("plugin_directory", "./plugins")
	viper.SetDefault("artifact_storage_type", "s3")
	viper.SetDefault("jwt_secret", "dev-secret-change-in-production")

	// Read from environment
	viper.AutomaticEnv()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found; using defaults and env vars
	}

	cfg := &Config{
		Port:                   viper.GetInt("port"),
		LogLevel:               viper.GetString("log_level"),
		DatabaseURL:            viper.GetString("database_url"),
		CORSAllowedOrigins:     viper.GetStringSlice("cors_allowed_origins"),
		WorkerHeartbeatTimeout: viper.GetInt("worker_heartbeat_timeout"),
		MaxWorkersPerJob:       viper.GetInt("max_workers_per_job"),
		SchedulerTickInterval:  viper.GetInt("scheduler_tick_interval"),
		MaxConcurrentBuilds:    viper.GetInt("max_concurrent_builds"),
		PluginDirectory:        viper.GetString("plugin_directory"),
		ArtifactStorageType:    viper.GetString("artifact_storage_type"),
		JWTSecret:              viper.GetString("jwt_secret"),
	}

	return cfg, nil
}
