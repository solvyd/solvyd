package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSuccess   JobStatus = "success"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusTimeout   JobStatus = "timeout"
)

// WorkerStatus represents the status of a worker
type WorkerStatus string

const (
	WorkerStatusOnline      WorkerStatus = "online"
	WorkerStatusOffline     WorkerStatus = "offline"
	WorkerStatusDraining    WorkerStatus = "draining"
	WorkerStatusMaintenance WorkerStatus = "maintenance"
)

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusInProgress DeploymentStatus = "in_progress"
	DeploymentStatusSuccess    DeploymentStatus = "success"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
)

// JSONB is a custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Job represents a CI/CD job
type Job struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	// SCM configuration
	SCMType        string     `json:"scm_type"`
	SCMURL         string     `json:"scm_url"`
	SCMBranch      string     `json:"scm_branch"`
	SCMCredentials *uuid.UUID `json:"scm_credentials_id,omitempty"`
	// Build configuration
	BuildConfig JSONB `json:"build_config"`
	EnvVars     JSONB `json:"environment_vars"`
	// Scheduling
	Triggers       JSONB `json:"triggers"`
	Enabled        bool  `json:"enabled"`
	WorkerLabels   JSONB `json:"worker_labels"`
	Plugins        JSONB `json:"plugins"`
	PipelineStages JSONB `json:"pipeline_stages"`
	// Timeout and retry
	TimeoutMinutes int `json:"timeout_minutes"`
	MaxRetries     int `json:"max_retries"`
	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
}

// Build represents a single build execution
type Build struct {
	ID          uuid.UUID `json:"id"`
	JobID       uuid.UUID `json:"job_id"`
	BuildNumber int       `json:"build_number"`
	// Status
	Status JobStatus `json:"status"`
	// Timing
	QueuedAt    time.Time  `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Duration    *int       `json:"duration_seconds,omitempty"`
	// Worker
	WorkerID *uuid.UUID `json:"worker_id,omitempty"`
	// SCM context
	CommitSHA     string `json:"scm_commit_sha"`
	CommitMessage string `json:"scm_commit_message"`
	Author        string `json:"scm_author"`
	Branch        string `json:"branch"`
	// Build context
	Parameters JSONB `json:"parameters"`
	EnvVars    JSONB `json:"environment_vars"`
	// Trigger
	TriggeredBy     string `json:"triggered_by"`
	TriggerMetadata JSONB  `json:"trigger_metadata"`
	// Results
	ExitCode      *int      `json:"exit_code,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	LogURL        string    `json:"log_url,omitempty"`
	ArtifactCount int       `json:"artifact_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// Worker represents a worker node
type Worker struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip_address"`
	// Capacity
	MaxConcurrentBuilds int `json:"max_concurrent_builds"`
	CurrentBuilds       int `json:"current_builds"`
	CPUCores            int `json:"cpu_cores"`
	MemoryMB            int `json:"memory_mb"`
	// Labels and capabilities
	Labels       JSONB `json:"labels"`
	Capabilities JSONB `json:"capabilities"`
	// Status
	Status        WorkerStatus `json:"status"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	HealthStatus  string       `json:"health_status"`
	AgentVersion  string       `json:"agent_version"`
	RegisteredAt  time.Time    `json:"registered_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// Artifact represents a build artifact
type Artifact struct {
	ID             uuid.UUID `json:"id"`
	BuildID        uuid.UUID `json:"build_id"`
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	SizeBytes      int64     `json:"size_bytes"`
	ChecksumSHA256 string    `json:"checksum_sha256"`
	ContentType    string    `json:"content_type"`
	// Storage
	StoragePlugin   string `json:"storage_plugin"`
	StorageURL      string `json:"storage_url"`
	StorageMetadata JSONB  `json:"storage_metadata"`
	// Promotion
	PromotionStatus string     `json:"promotion_status"`
	PromotedAt      *time.Time `json:"promoted_at,omitempty"`
	PromotedBy      string     `json:"promoted_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	Metadata        JSONB      `json:"metadata"`
}

// Deployment represents a deployment record
type Deployment struct {
	ID          uuid.UUID        `json:"id"`
	BuildID     uuid.UUID        `json:"build_id"`
	ArtifactID  *uuid.UUID       `json:"artifact_id,omitempty"`
	Environment string           `json:"environment"`
	Status      DeploymentStatus `json:"status"`
	// Target
	TargetType     string `json:"target_type"`
	TargetURL      string `json:"target_url,omitempty"`
	TargetMetadata JSONB  `json:"target_metadata"`
	// Timing
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	Duration         *int       `json:"duration_seconds,omitempty"`
	DeploymentPlugin string     `json:"deployment_plugin"`
	// Results
	ExitCode      *int   `json:"exit_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	DeploymentURL string `json:"deployment_url,omitempty"`
	// Rollback
	RollbackFromID  *uuid.UUID `json:"rollback_from_deployment_id,omitempty"`
	DeployedBy      string     `json:"deployed_by"`
	DeploymentNotes string     `json:"deployment_notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Plugin represents an installed plugin
type Plugin struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Version        string    `json:"version"`
	BinaryPath     string    `json:"binary_path"`
	BinaryChecksum string    `json:"binary_checksum"`
	Description    string    `json:"description"`
	Author         string    `json:"author"`
	HomepageURL    string    `json:"homepage_url"`
	ConfigSchema   JSONB     `json:"config_schema"`
	Enabled        bool      `json:"enabled"`
	InstalledAt    time.Time `json:"installed_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BuildLog represents a log line from a build
type BuildLog struct {
	ID             uuid.UUID `json:"id"`
	BuildID        uuid.UUID `json:"build_id"`
	SequenceNumber int       `json:"sequence_number"`
	Timestamp      time.Time `json:"timestamp"`
	LogLine        string    `json:"log_line"`
	Stream         string    `json:"stream"` // stdout or stderr
}
