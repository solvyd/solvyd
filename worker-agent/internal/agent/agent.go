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
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/vrenjith/ritmo/worker-agent/internal/config"
	"github.com/vrenjith/ritmo/worker-agent/internal/executor"
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

	apiURL := fmt.Sprintf("http://%s", cfg.APIServer)

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

	// In a real implementation, this would query the API server
	// for builds assigned to this worker
	// For now, this is a stub
	log.Debug().Msg("Checking for builds...")
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
