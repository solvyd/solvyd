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

	// GitOps
	GitOps GitOpsConfig
}

// GitOpsConfig holds GitOps configuration
type GitOpsConfig struct {
	Enabled        bool
	Repository     GitOpsRepository
	Authentication GitOpsAuth
	Sync           GitOpsSyncConfig
}

// GitOpsRepository defines the Git repository configuration
type GitOpsRepository struct {
	URL    string
	Branch string
	Path   string
}

// GitOpsAuth defines authentication for Git repository
type GitOpsAuth struct {
	Type  string // token, ssh, github_app
	Token string
}

// GitOpsSyncConfig defines sync behavior
type GitOpsSyncConfig struct {
	Interval  int // seconds
	AutoApply bool
	DryRun    bool
	Prune     bool // Delete resources not in Git
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

	// GitOps defaults
	viper.SetDefault("gitops.enabled", false)
	viper.SetDefault("gitops.repository.branch", "main")
	viper.SetDefault("gitops.repository.path", "/")
	viper.SetDefault("gitops.authentication.type", "token")
	viper.SetDefault("gitops.sync.interval", 60)
	viper.SetDefault("gitops.sync.auto_apply", true)
	viper.SetDefault("gitops.sync.dry_run", false)
	viper.SetDefault("gitops.sync.prune", true)

	// Read from environment
	viper.AutomaticEnv()
	viper.SetEnvPrefix("RITMO")
	viper.BindEnv("gitops.enabled", "RITMO_GITOPS_ENABLED")
	viper.BindEnv("gitops.repository.url", "RITMO_GITOPS_REPO_URL")
	viper.BindEnv("gitops.repository.branch", "RITMO_GITOPS_REPO_BRANCH")
	viper.BindEnv("gitops.repository.path", "RITMO_GITOPS_REPO_PATH")
	viper.BindEnv("gitops.authentication.type", "RITMO_GITOPS_AUTH_TYPE")
	viper.BindEnv("gitops.authentication.token", "RITMO_GITOPS_TOKEN")

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
		GitOps: GitOpsConfig{
			Enabled: viper.GetBool("gitops.enabled"),
			Repository: GitOpsRepository{
				URL:    viper.GetString("gitops.repository.url"),
				Branch: viper.GetString("gitops.repository.branch"),
				Path:   viper.GetString("gitops.repository.path"),
			},
			Authentication: GitOpsAuth{
				Type:  viper.GetString("gitops.authentication.type"),
				Token: viper.GetString("gitops.authentication.token"),
			},
			Sync: GitOpsSyncConfig{
				Interval:  viper.GetInt("gitops.sync.interval"),
				AutoApply: viper.GetBool("gitops.sync.auto_apply"),
				DryRun:    viper.GetBool("gitops.sync.dry_run"),
				Prune:     viper.GetBool("gitops.sync.prune"),
			},
		},
	}

	return cfg, nil
}
