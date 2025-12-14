package executor

import (
	"context"

	"github.com/rs/zerolog/log"
)

// ProcessExecutor executes builds as separate processes
type ProcessExecutor struct{}

// NewProcessExecutor creates a new process executor
func NewProcessExecutor() *ProcessExecutor {
	return &ProcessExecutor{}
}

// Execute runs a build as a separate process
func (e *ProcessExecutor) Execute(ctx context.Context, build *BuildRequest) (*BuildResult, error) {
	log.Info().
		Str("build_id", build.BuildID).
		Str("scm_url", build.SCMURL).
		Msg("Starting process build execution")

	// TODO: Implement actual process execution
	// 1. Clone repository to work directory
	// 2. Execute build commands as subprocess
	// 3. Collect logs and artifacts
	// 4. Push artifacts to storage

	result := &BuildResult{
		Success:  true,
		ExitCode: 0,
		LogLines: []string{
			"[INFO] Process executor stub",
			"[INFO] This is where actual process build would run",
			"[INFO] Build completed successfully",
		},
		Artifacts: []Artifact{},
	}

	return result, nil
}

// Cleanup removes build work directory
func (e *ProcessExecutor) Cleanup(ctx context.Context, buildID string) error {
	log.Debug().Str("build_id", buildID).Msg("Cleaning up process resources")
	// TODO: Remove work directory
	return nil
}
