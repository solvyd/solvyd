package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/models"
	"github.com/solvyd/solvyd/api-server/internal/worker"
)

// WorkerHandler handles worker-related requests
type WorkerHandler struct {
	db  *database.Database
	mgr *worker.Manager
}

// NewWorkerHandler creates a new worker handler
func NewWorkerHandler(db *database.Database, mgr *worker.Manager) *WorkerHandler {
	return &WorkerHandler{db: db, mgr: mgr}
}

// ListWorkers returns all workers
func (h *WorkerHandler) ListWorkers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := `
		SELECT id, name, hostname, ip_address, max_concurrent_builds,
		       current_builds, cpu_cores, memory_mb, labels, capabilities,
		       status, last_heartbeat, health_status, agent_version,
		       registered_at, updated_at
		FROM workers
		ORDER BY name ASC
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query workers")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch workers")
		return
	}
	defer rows.Close()

	workers := []models.Worker{}
	for rows.Next() {
		var worker models.Worker
		err := rows.Scan(
			&worker.ID, &worker.Name, &worker.Hostname, &worker.IP,
			&worker.MaxConcurrentBuilds, &worker.CurrentBuilds,
			&worker.CPUCores, &worker.MemoryMB, &worker.Labels, &worker.Capabilities,
			&worker.Status, &worker.LastHeartbeat,
			&worker.HealthStatus, &worker.AgentVersion, &worker.RegisteredAt,
			&worker.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan worker row")
			continue
		}
		workers = append(workers, worker)
	}

	SendJSON(w, http.StatusOK, workers)
}

// GetWorker returns a single worker
func (h *WorkerHandler) GetWorker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	workerID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid worker ID")
		return
	}

	query := `
		SELECT id, name, hostname, ip_address, max_concurrent_builds,
		       current_builds, cpu_cores, memory_mb, labels, capabilities,
		       status, last_heartbeat, health_status, agent_version,
		       registered_at, updated_at
		FROM workers
		WHERE id = $1
	`

	var worker models.Worker
	err = h.db.GetConn().QueryRowContext(ctx, query, workerID).Scan(
		&worker.ID, &worker.Name, &worker.Hostname, &worker.IP,
		&worker.MaxConcurrentBuilds, &worker.CurrentBuilds,
		&worker.CPUCores, &worker.MemoryMB, &worker.Labels,
		&worker.Capabilities, &worker.Status, &worker.LastHeartbeat,
		&worker.HealthStatus, &worker.AgentVersion, &worker.RegisteredAt,
		&worker.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		SendError(w, http.StatusNotFound, nil, "Worker not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query worker")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch worker")
		return
	}

	SendJSON(w, http.StatusOK, worker)
}

// UpdateWorker updates worker configuration
func (h *WorkerHandler) UpdateWorker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	workerID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid worker ID")
		return
	}

	var updates struct {
		MaxConcurrentBuilds *int                   `json:"max_concurrent_builds"`
		Labels              map[string]interface{} `json:"labels"`
		Status              *string                `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Build dynamic UPDATE query
	query := `UPDATE workers SET `
	args := []interface{}{workerID}
	argCount := 2
	updateParts := []string{}

	if updates.MaxConcurrentBuilds != nil {
		updateParts = append(updateParts, `max_concurrent_builds = $`+string(rune('0'+argCount)))
		args = append(args, *updates.MaxConcurrentBuilds)
		argCount++
	}
	if updates.Labels != nil {
		labelsJSON, _ := json.Marshal(updates.Labels)
		updateParts = append(updateParts, `labels = $`+string(rune('0'+argCount)))
		args = append(args, labelsJSON)
		argCount++
	}
	if updates.Status != nil {
		updateParts = append(updateParts, `status = $`+string(rune('0'+argCount)))
		args = append(args, *updates.Status)
		argCount++
	}

	if len(updateParts) == 0 {
		SendError(w, http.StatusBadRequest, nil, "No updates provided")
		return
	}

	query += updateParts[0]
	for i := 1; i < len(updateParts); i++ {
		query += `, ` + updateParts[i]
	}
	query += ` WHERE id = $1`

	result, err := h.db.GetConn().ExecContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update worker")
		SendError(w, http.StatusInternalServerError, err, "Failed to update worker")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		SendError(w, http.StatusNotFound, nil, "Worker not found")
		return
	}

	log.Info().Str("worker_id", workerID.String()).Msg("Worker updated")
	SendJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DrainWorker puts a worker into draining mode
func (h *WorkerHandler) DrainWorker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	workerID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid worker ID")
		return
	}

	query := `UPDATE workers SET status = 'draining' WHERE id = $1`
	result, err := h.db.GetConn().ExecContext(ctx, query, workerID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to drain worker")
		SendError(w, http.StatusInternalServerError, err, "Failed to drain worker")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		SendError(w, http.StatusNotFound, nil, "Worker not found")
		return
	}

	log.Info().Str("worker_id", workerID.String()).Msg("Worker set to draining")
	SendJSON(w, http.StatusOK, map[string]string{"status": "draining"})
}

// RegisterWorker registers a new worker
func (h *WorkerHandler) RegisterWorker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name                string                 `json:"name"`
		Hostname            string                 `json:"hostname"`
		IPAddress           string                 `json:"ip_address"`
		MaxConcurrentBuilds int                    `json:"max_concurrent_builds"`
		CPUCores            int                    `json:"cpu_cores"`
		MemoryMB            int                    `json:"memory_mb"`
		Labels              map[string]interface{} `json:"labels"`
		Capabilities        []string               `json:"capabilities"`
		AgentVersion        string                 `json:"agent_version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		SendError(w, http.StatusBadRequest, nil, "Worker name is required")
		return
	}
	if req.MaxConcurrentBuilds <= 0 {
		req.MaxConcurrentBuilds = 2 // Default
	}

	// Convert labels and capabilities to JSON
	labelsJSON, _ := json.Marshal(req.Labels)
	capsJSON, _ := json.Marshal(req.Capabilities)

	// Insert worker into database
	query := `
		INSERT INTO workers (
			name, hostname, ip_address, max_concurrent_builds,
			cpu_cores, memory_mb, labels, capabilities,
			status, health_status, agent_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'online', 'healthy', $9)
		ON CONFLICT (name) 
		DO UPDATE SET 
			hostname = EXCLUDED.hostname,
			ip_address = EXCLUDED.ip_address,
			max_concurrent_builds = EXCLUDED.max_concurrent_builds,
			cpu_cores = EXCLUDED.cpu_cores,
			memory_mb = EXCLUDED.memory_mb,
			labels = EXCLUDED.labels,
			capabilities = EXCLUDED.capabilities,
			agent_version = EXCLUDED.agent_version,
			status = 'online',
			last_heartbeat = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, name, registered_at
	`

	var workerID uuid.UUID
	var workerName string
	var registeredAt interface{}

	err := h.db.GetConn().QueryRowContext(ctx, query,
		req.Name, req.Hostname, req.IPAddress, req.MaxConcurrentBuilds,
		req.CPUCores, req.MemoryMB, labelsJSON, capsJSON, req.AgentVersion,
	).Scan(&workerID, &workerName, &registeredAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to register worker")
		SendError(w, http.StatusInternalServerError, err, "Failed to register worker")
		return
	}

	log.Info().
		Str("worker_id", workerID.String()).
		Str("worker_name", workerName).
		Msg("Worker registered")

	response := map[string]interface{}{
		"id":            workerID,
		"name":          workerName,
		"registered_at": registeredAt,
		"status":        "online",
	}

	SendJSON(w, http.StatusCreated, response)
}

// Heartbeat handles worker heartbeat updates
func (h *WorkerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	workerID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid worker ID")
		return
	}

	var req struct {
		CurrentBuilds int    `json:"current_builds"`
		HealthStatus  string `json:"health_status"`
		CPUUsage      *int   `json:"cpu_usage,omitempty"`
		MemoryUsage   *int   `json:"memory_usage,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Default health status if not provided
	if req.HealthStatus == "" {
		req.HealthStatus = "healthy"
	}

	// Update worker heartbeat and status
	query := `
		UPDATE workers 
		SET last_heartbeat = CURRENT_TIMESTAMP,
		    current_builds = $1,
		    health_status = $2,
		    status = CASE 
		        WHEN status = 'draining' THEN 'draining'
		        ELSE 'online'
		    END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING status, current_builds, max_concurrent_builds
	`

	var status string
	var currentBuilds, maxBuilds int

	err = h.db.GetConn().QueryRowContext(ctx, query, req.CurrentBuilds, req.HealthStatus, workerID).
		Scan(&status, &currentBuilds, &maxBuilds)

	if err == sql.ErrNoRows {
		SendError(w, http.StatusNotFound, nil, "Worker not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update heartbeat")
		SendError(w, http.StatusInternalServerError, err, "Failed to update heartbeat")
		return
	}

	log.Debug().
		Str("worker_id", workerID.String()).
		Int("current_builds", currentBuilds).
		Str("health", req.HealthStatus).
		Msg("Heartbeat received")

	// Check if there are pending builds for this worker
	hasWork := false
	if currentBuilds < maxBuilds && status == "online" {
		// Check for assigned builds
		checkQuery := `
			SELECT EXISTS(
				SELECT 1 FROM builds 
				WHERE worker_id = $1 AND status = 'queued'
				LIMIT 1
			)
		`
		h.db.GetConn().QueryRowContext(ctx, checkQuery, workerID).Scan(&hasWork)
	}

	response := map[string]interface{}{
		"status":         status,
		"current_builds": currentBuilds,
		"max_builds":     maxBuilds,
		"has_work":       hasWork,
	}

	SendJSON(w, http.StatusOK, response)
}
