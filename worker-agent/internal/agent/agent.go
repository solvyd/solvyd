package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/worker-agent/internal/config"
	"github.com/solvyd/solvyd/worker-agent/internal/executor"
)

// Agent represents the worker agent
type Agent struct {
	config        *config.Config
	executor      executor.Executor
	workerID      uuid.UUID
	client        *http.Client
	apiURL        string
	currentBuilds int
}

// NewAgent creates a new worker agent
func NewAgent(cfg *config.Config, exec executor.Executor) (*Agent, error) {
	// Auto-detect system info
	cfg.CPUCores = runtime.NumCPU()
	cfg.Hostname, _ = os.Hostname()
	cfg.IPAddress = getOutboundIP()

	// Estimate memory (simplified)
	cfg.MemoryMB = 8192 // TODO: Actually detect memory

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Use API server URL directly if it already has a scheme, otherwise add http://
	apiURL := cfg.APIServer
	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		apiURL = "http://" + apiURL
	}

	return &Agent{
		config:   cfg,
		executor: exec,
		client:   client,
		apiURL:   apiURL,
	}, nil
}

// Start begins the agent main loop
func (a *Agent) Start(ctx context.Context) {
	log.Info().
		Str("worker_name", a.config.WorkerName).
		Str("api_server", a.config.APIServer).
		Msg("Worker agent started")

	// Register with API server
	if err := a.register(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to register with API server")
	}

	// Start heartbeat
	go a.heartbeatLoop(ctx)

	// Start polling for builds
	a.pollLoop(ctx)
}

// register registers the worker with the API server
func (a *Agent) register(ctx context.Context) error {
	payload := map[string]interface{}{
		"name":                  a.config.WorkerName,
		"hostname":              a.config.Hostname,
		"ip_address":            a.config.IPAddress,
		"max_concurrent_builds": a.config.MaxConcurrent,
		"cpu_cores":             a.config.CPUCores,
		"memory_mb":             a.config.MemoryMB,
		"labels":                a.config.Labels,
		"agent_version":         "1.0.0",
		"capabilities": map[string]bool{
			"docker":     a.config.IsolationType == "docker",
			"kubernetes": false,
			"vm":         a.config.IsolationType == "vm",
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL+"/api/v1/workers/register", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if id, ok := result["id"].(string); ok {
		workerID, err := uuid.Parse(id)
		if err == nil {
			a.workerID = workerID
		}
	}

	log.Info().
		Str("worker_id", a.workerID.String()).
		Str("worker_name", a.config.WorkerName).
		Msg("Worker registered successfully")

	return nil
}

// heartbeatLoop sends periodic heartbeats to the API server
func (a *Agent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to send heartbeat")
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the API server
func (a *Agent) sendHeartbeat(ctx context.Context) error {
	if a.workerID == uuid.Nil {
		return nil // Not registered yet
	}

	payload := map[string]interface{}{
		"current_builds": a.currentBuilds,
		"health_status":  "healthy",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/workers/%s/heartbeat", a.apiURL, a.workerID.String())

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status %d", resp.StatusCode)
	}

	// Parse response to check for work
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Check if there's work available
	if hasWork, ok := result["has_work"].(bool); ok && hasWork {
		log.Debug().Msg("Work available for this worker")
		// Trigger immediate poll
		go a.checkForBuilds(ctx)
	}

	log.Debug().Msg("Heartbeat sent")
	return nil
}

// pollLoop polls for new builds to execute
func (a *Agent) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if a.currentBuilds < a.config.MaxConcurrent {
				a.checkForBuilds(ctx)
			}
		}
	}
}

// checkForBuilds checks if there are builds assigned to this worker
func (a *Agent) checkForBuilds(ctx context.Context) {
	if a.workerID == uuid.Nil {
		return
	}

	// Fetch builds assigned to this worker
	url := fmt.Sprintf("%s/api/v1/workers/%s/builds", a.apiURL, a.workerID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create builds request")
		return
	}

	resp, err := a.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch builds")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Msg("Failed to fetch builds")
		return
	}

	var builds []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		log.Error().Err(err).Msg("Failed to decode builds response")
		return
	}

	if len(builds) == 0 {
		log.Debug().Msg("No pending builds")
		return
	}

	log.Info().Int("count", len(builds)).Msg("Found pending builds")

	// Execute builds (up to max concurrent limit)
	for _, buildData := range builds {
		if a.currentBuilds >= a.config.MaxConcurrent {
			log.Warn().Msg("Max concurrent builds reached, stopping")
			break
		}

		a.currentBuilds++
		go a.executeBuild(ctx, buildData)
	}
}

// executeBuild executes a single build
func (a *Agent) executeBuild(ctx context.Context, buildData map[string]interface{}) {
	defer func() {
		a.currentBuilds--
	}()

	buildID := buildData["id"].(string)
	log.Info().Str("build_id", buildID).Msg("Starting build execution")

	// Update build status to running
	if err := a.updateBuildStatus(ctx, buildID, "running", map[string]interface{}{
		"started_at": time.Now().Format(time.RFC3339),
	}); err != nil {
		log.Error().Err(err).Str("build_id", buildID).Msg("Failed to update build status to running")
	}

	// Extract build information
	buildConfig := buildData["build_config"].(map[string]interface{})

	buildRequest := &executor.BuildRequest{
		BuildID:     buildID,
		SCMURL:      buildData["scm_url"].(string),
		SCMBranch:   getStringOrEmpty(buildData, "branch"),
		CommitSHA:   getStringOrEmpty(buildData, "commit_sha"),
		BuildConfig: buildConfig,
		EnvVars:     make(map[string]string),
	}

	// Execute the build
	result, err := a.executor.Execute(ctx, buildRequest)

	// Update build status based on result
	status := "success"
	statusData := map[string]interface{}{
		"completed_at":     time.Now().Format(time.RFC3339),
		"duration_seconds": result.Duration,
	}

	if err != nil || !result.Success {
		status = "failure"
		statusData["exit_code"] = result.ExitCode
		if result.ErrorMessage != "" {
			statusData["error_message"] = result.ErrorMessage
		}
		log.Error().
			Err(err).
			Str("build_id", buildID).
			Int("exit_code", result.ExitCode).
			Msg("Build failed")
	} else {
		log.Info().
			Str("build_id", buildID).
			Int("duration", result.Duration).
			Msg("Build completed successfully")
	}

	// Update final build status
	if err := a.updateBuildStatus(ctx, buildID, status, statusData); err != nil {
		log.Error().Err(err).Str("build_id", buildID).Msg("Failed to update final build status")
	}

	// TODO: Upload logs to API server
	// TODO: Upload artifacts to storage (MinIO/S3)

	// Cleanup
	if err := a.executor.Cleanup(ctx, buildID); err != nil {
		log.Warn().Err(err).Str("build_id", buildID).Msg("Failed to cleanup build resources")
	}
}

// updateBuildStatus updates the status of a build
func (a *Agent) updateBuildStatus(ctx context.Context, buildID string, status string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/v1/builds/%s/status", a.apiURL, buildID)

	payload := map[string]interface{}{
		"status": status,
	}
	for k, v := range data {
		payload[k] = v
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status update failed with code %d", resp.StatusCode)
	}

	return nil
}

// getStringOrEmpty safely extracts string from map
func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getOutboundIP gets the preferred outbound IP of this machine
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
