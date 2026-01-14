package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// DockerExecutor executes builds in Docker containers
type DockerExecutor struct {
	workDir string
}

// NewDockerExecutor creates a new Docker executor
func NewDockerExecutor() *DockerExecutor {
	workDir := os.Getenv("SOLVYD_WORK_DIR")
	if workDir == "" {
		workDir = "/tmp/solvyd-builds"
	}
	os.MkdirAll(workDir, 0755)

	return &DockerExecutor{
		workDir: workDir,
	}
}

// Execute runs a build in a Docker container
func (e *DockerExecutor) Execute(ctx context.Context, build *BuildRequest) (*BuildResult, error) {
	startTime := time.Now()

	log.Info().
		Str("build_id", build.BuildID).
		Str("scm_url", build.SCMURL).
		Msg("Starting Docker build execution")

	result := &BuildResult{
		LogLines:  []string{},
		Artifacts: []Artifact{},
	}

	// Create build directory
	buildDir := filepath.Join(e.workDir, build.BuildID)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to create build directory: %v", err)
		return result, err
	}

	// Step 1: Clone repository
	result.LogLines = append(result.LogLines, fmt.Sprintf("[INFO] Cloning repository: %s", build.SCMURL))
	if err := e.cloneRepository(ctx, build, buildDir, result); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to clone repository: %v", err)
		result.ExitCode = 1
		return result, err
	}

	// Step 2: Get build image from config or use default
	buildImage := "ubuntu:22.04"
	if img, ok := build.BuildConfig["image"].(string); ok && img != "" {
		buildImage = img
	}

	// Step 3: Execute build in Docker container
	result.LogLines = append(result.LogLines, fmt.Sprintf("[INFO] Using Docker image: %s", buildImage))

	containerName := fmt.Sprintf("solvyd-build-%s", build.BuildID)

	// Build commands from config
	commands := []string{}
	if cmds, ok := build.BuildConfig["commands"].([]interface{}); ok {
		for _, cmd := range cmds {
			if cmdStr, ok := cmd.(string); ok {
				commands = append(commands, cmdStr)
			}
		}
	}

	// Default commands if none specified
	if len(commands) == 0 {
		commands = []string{
			"echo 'No build commands specified'",
			"ls -la",
		}
	}

	// Combine commands
	combinedCmd := strings.Join(commands, " && ")

	// Run Docker container
	dockerArgs := []string{
		"run",
		"--rm",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", buildDir),
		"-w", "/workspace",
	}

	// Add environment variables
	for key, value := range build.EnvVars {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	dockerArgs = append(dockerArgs, buildImage, "sh", "-c", combinedCmd)

	result.LogLines = append(result.LogLines, fmt.Sprintf("[INFO] Running: docker %s", strings.Join(dockerArgs, " ")))

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	cmd.Dir = buildDir

	// Capture output
	output, err := cmd.CombinedOutput()
	outputLines := strings.Split(string(output), "\n")
	for _, line := range outputLines {
		if line != "" {
			result.LogLines = append(result.LogLines, line)
		}
	}

	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("Build failed with exit code %d", result.ExitCode)
		} else {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("Failed to execute build: %v", err)
			result.ExitCode = 1
		}
	} else {
		result.Success = true
		result.ExitCode = 0
		result.LogLines = append(result.LogLines, "[INFO] Build completed successfully")
	}

	// Step 4: Collect artifacts (if any)
	if artifactsPath, ok := build.BuildConfig["artifacts"].(string); ok {
		e.collectArtifacts(buildDir, artifactsPath, result)
	}

	result.Duration = int(time.Since(startTime).Seconds())

	return result, nil
}

// cloneRepository clones the Git repository
func (e *DockerExecutor) cloneRepository(ctx context.Context, build *BuildRequest, buildDir string, result *BuildResult) error {
	var cmd *exec.Cmd

	if build.CommitSHA != "" {
		// Clone specific commit
		cmd = exec.CommandContext(ctx, "git", "clone", build.SCMURL, ".")
		cmd.Dir = buildDir
		if output, err := cmd.CombinedOutput(); err != nil {
			result.LogLines = append(result.LogLines, string(output))
			return err
		}

		// Checkout specific commit
		cmd = exec.CommandContext(ctx, "git", "checkout", build.CommitSHA)
		cmd.Dir = buildDir
		if output, err := cmd.CombinedOutput(); err != nil {
			result.LogLines = append(result.LogLines, string(output))
			return err
		}
	} else if build.SCMBranch != "" {
		// Clone specific branch
		cmd = exec.CommandContext(ctx, "git", "clone", "-b", build.SCMBranch, "--depth", "1", build.SCMURL, ".")
		cmd.Dir = buildDir
		if output, err := cmd.CombinedOutput(); err != nil {
			result.LogLines = append(result.LogLines, string(output))
			return err
		}
	} else {
		// Clone default branch
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", "1", build.SCMURL, ".")
		cmd.Dir = buildDir
		if output, err := cmd.CombinedOutput(); err != nil {
			result.LogLines = append(result.LogLines, string(output))
			return err
		}
	}

	result.LogLines = append(result.LogLines, "[INFO] Repository cloned successfully")
	return nil
}

// collectArtifacts collects build artifacts
func (e *DockerExecutor) collectArtifacts(buildDir, artifactsPath string, result *BuildResult) {
	fullPath := filepath.Join(buildDir, artifactsPath)

	// Check if it's a glob pattern
	matches, err := filepath.Glob(fullPath)
	if err != nil || len(matches) == 0 {
		// Try as single file
		if info, err := os.Stat(fullPath); err == nil {
			artifact := Artifact{
				Name:      filepath.Base(fullPath),
				Path:      fullPath,
				SizeBytes: info.Size(),
			}
			result.Artifacts = append(result.Artifacts, artifact)
			result.LogLines = append(result.LogLines, fmt.Sprintf("[INFO] Collected artifact: %s (%d bytes)", artifact.Name, artifact.SizeBytes))
		}
		return
	}

	// Collect all matching files
	for _, match := range matches {
		if info, err := os.Stat(match); err == nil && !info.IsDir() {
			artifact := Artifact{
				Name:      filepath.Base(match),
				Path:      match,
				SizeBytes: info.Size(),
			}
			result.Artifacts = append(result.Artifacts, artifact)
			result.LogLines = append(result.LogLines, fmt.Sprintf("[INFO] Collected artifact: %s (%d bytes)", artifact.Name, artifact.SizeBytes))
		}
	}
}

// Cleanup removes Docker container and build directory
func (e *DockerExecutor) Cleanup(ctx context.Context, buildID string) error {
	log.Debug().Str("build_id", buildID).Msg("Cleaning up Docker resources")

	// Remove build directory
	buildDir := filepath.Join(e.workDir, buildID)
	if err := os.RemoveAll(buildDir); err != nil {
		log.Warn().Err(err).Str("build_id", buildID).Msg("Failed to clean up build directory")
	}

	// Stop container if still running
	containerName := fmt.Sprintf("solvyd-build-%s", buildID)
	cmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	cmd.Run() // Ignore errors, container might already be stopped

	return nil
}
