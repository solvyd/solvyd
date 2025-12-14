package executor

import (
	"context"
	"fmt"
)

// Executor defines the interface for build execution
type Executor interface {
	// Execute runs a build in an isolated environment
	Execute(ctx context.Context, build *BuildRequest) (*BuildResult, error)

	// Cleanup cleans up resources after build execution
	Cleanup(ctx context.Context, buildID string) error
}

// BuildRequest contains all information needed to execute a build
type BuildRequest struct {
	BuildID     string
	JobID       string
	SCMType     string
	SCMURL      string
	SCMBranch   string
	CommitSHA   string
	BuildConfig map[string]interface{}
	EnvVars     map[string]string
	WorkDir     string
}

// BuildResult contains the result of a build execution
type BuildResult struct {
	Success      bool
	ExitCode     int
	ErrorMessage string
	Duration     int // seconds
	LogLines     []string
	Artifacts    []Artifact
}

// Artifact represents a build artifact
type Artifact struct {
	Name           string
	Path           string
	SizeBytes      int64
	ChecksumSHA256 string
}

// NewExecutor creates a new executor based on isolation type
func NewExecutor(isolationType string) (Executor, error) {
	switch isolationType {
	case "docker":
		return NewDockerExecutor(), nil
	case "process":
		return NewProcessExecutor(), nil
	case "vm":
		return nil, fmt.Errorf("VM executor not yet implemented")
	default:
		return nil, fmt.Errorf("unknown isolation type: %s", isolationType)
	}
}
