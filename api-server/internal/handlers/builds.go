package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/models"
)

// BuildHandler handles build-related requests
type BuildHandler struct {
	db *database.Database
}

// NewBuildHandler creates a new build handler
func NewBuildHandler(db *database.Database) *BuildHandler {
	return &BuildHandler{db: db}
}

// ListBuilds returns all builds
func (h *BuildHandler) ListBuilds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Optional filters
	jobID := r.URL.Query().Get("job_id")
	status := r.URL.Query().Get("status")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	query := `
		SELECT b.id, b.job_id, b.build_number, b.status, b.queued_at, 
		       b.started_at, b.completed_at, b.duration_seconds, b.worker_id,
		       b.scm_commit_sha, b.scm_commit_message, b.scm_author, b.branch,
		       b.triggered_by, b.exit_code, b.error_message, b.artifact_count,
		       j.name as job_name
		FROM builds b
		JOIN jobs j ON b.job_id = j.id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if jobID != "" {
		query += ` AND b.job_id = $` + strconv.Itoa(argCount)
		args = append(args, jobID)
		argCount++
	}

	if status != "" {
		query += ` AND b.status = $` + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	query += ` ORDER BY b.queued_at DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.GetConn().QueryContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query builds")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch builds")
		return
	}
	defer rows.Close()

	builds := []map[string]interface{}{}
	for rows.Next() {
		var build models.Build
		var jobName string
		err := rows.Scan(
			&build.ID, &build.JobID, &build.BuildNumber, &build.Status,
			&build.QueuedAt, &build.StartedAt, &build.CompletedAt, &build.Duration,
			&build.WorkerID, &build.CommitSHA, &build.CommitMessage, &build.Author,
			&build.Branch, &build.TriggeredBy, &build.ExitCode, &build.ErrorMessage,
			&build.ArtifactCount, &jobName,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan build row")
			continue
		}

		buildMap := map[string]interface{}{
			"id":            build.ID,
			"job_id":        build.JobID,
			"job_name":      jobName,
			"build_number":  build.BuildNumber,
			"status":        build.Status,
			"queued_at":     build.QueuedAt,
			"started_at":    build.StartedAt,
			"completed_at":  build.CompletedAt,
			"duration":      build.Duration,
			"worker_id":     build.WorkerID,
			"commit_sha":    build.CommitSHA,
			"commit_msg":    build.CommitMessage,
			"author":        build.Author,
			"branch":        build.Branch,
			"triggered_by":  build.TriggeredBy,
			"exit_code":     build.ExitCode,
			"error_message": build.ErrorMessage,
			"artifacts":     build.ArtifactCount,
		}
		builds = append(builds, buildMap)
	}

	SendJSON(w, http.StatusOK, builds)
}

// GetBuild returns a single build
func (h *BuildHandler) GetBuild(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	buildID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid build ID")
		return
	}

	query := `
		SELECT id, job_id, build_number, status, queued_at, started_at, 
		       completed_at, duration_seconds, worker_id, scm_commit_sha,
		       scm_commit_message, scm_author, branch, parameters,
		       environment_vars, triggered_by, trigger_metadata, exit_code,
		       error_message, log_url, artifact_count
		FROM builds
		WHERE id = $1
	`

	var build models.Build
	err = h.db.GetConn().QueryRowContext(ctx, query, buildID).Scan(
		&build.ID, &build.JobID, &build.BuildNumber, &build.Status,
		&build.QueuedAt, &build.StartedAt, &build.CompletedAt, &build.Duration,
		&build.WorkerID, &build.CommitSHA, &build.CommitMessage, &build.Author,
		&build.Branch, &build.Parameters, &build.EnvVars, &build.TriggeredBy,
		&build.TriggerMetadata, &build.ExitCode, &build.ErrorMessage,
		&build.LogURL, &build.ArtifactCount,
	)

	if err == sql.ErrNoRows {
		SendError(w, http.StatusNotFound, nil, "Build not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query build")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch build")
		return
	}

	SendJSON(w, http.StatusOK, build)
}

// CancelBuild cancels a running build
func (h *BuildHandler) CancelBuild(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	buildID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid build ID")
		return
	}

	query := `
		UPDATE builds
		SET status = 'cancelled', completed_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('queued', 'running')
	`

	result, err := h.db.GetConn().ExecContext(ctx, query, buildID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel build")
		SendError(w, http.StatusInternalServerError, err, "Failed to cancel build")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		SendError(w, http.StatusNotFound, nil, "Build not found or already completed")
		return
	}

	log.Info().Str("build_id", buildID.String()).Msg("Build cancelled")
	SendJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetBuildLogs returns build logs
func (h *BuildHandler) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	buildID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid build ID")
		return
	}

	query := `
		SELECT sequence_number, timestamp, log_line, stream
		FROM build_logs
		WHERE build_id = $1
		ORDER BY sequence_number ASC
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query, buildID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query build logs")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch logs")
		return
	}
	defer rows.Close()

	logs := []models.BuildLog{}
	for rows.Next() {
		var log models.BuildLog
		err := rows.Scan(&log.SequenceNumber, &log.Timestamp, &log.LogLine, &log.Stream)
		if err != nil {
			continue
		}
		log.BuildID = buildID
		logs = append(logs, log)
	}

	SendJSON(w, http.StatusOK, logs)
}

// ListArtifacts returns artifacts for a build
func (h *BuildHandler) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	buildID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid build ID")
		return
	}

	query := `
		SELECT id, build_id, name, path, size_bytes, checksum_sha256,
		       content_type, storage_plugin, storage_url, promotion_status,
		       promoted_at, promoted_by, created_at
		FROM artifacts
		WHERE build_id = $1
		ORDER BY created_at ASC
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query, buildID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query artifacts")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch artifacts")
		return
	}
	defer rows.Close()

	artifacts := []models.Artifact{}
	for rows.Next() {
		var artifact models.Artifact
		err := rows.Scan(
			&artifact.ID, &artifact.BuildID, &artifact.Name, &artifact.Path,
			&artifact.SizeBytes, &artifact.ChecksumSHA256, &artifact.ContentType,
			&artifact.StoragePlugin, &artifact.StorageURL, &artifact.PromotionStatus,
			&artifact.PromotedAt, &artifact.PromotedBy, &artifact.CreatedAt,
		)
		if err != nil {
			continue
		}
		artifacts = append(artifacts, artifact)
	}

	SendJSON(w, http.StatusOK, artifacts)
}

// GetWorkerBuilds returns pending builds assigned to a specific worker
func (h *BuildHandler) GetWorkerBuilds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	workerID := vars["worker_id"]

	// Validate worker ID
	if _, err := uuid.Parse(workerID); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid worker ID")
		return
	}

	query := `
		SELECT b.id, b.job_id, b.build_number, b.status, b.queued_at, 
		       b.scm_commit_sha, b.branch, b.triggered_by, b.build_config,
		       j.name as job_name, j.scm_url, j.scm_type
		FROM builds b
		JOIN jobs j ON b.job_id = j.id
		WHERE b.worker_id = $1 AND b.status = 'queued'
		ORDER BY b.queued_at ASC
		LIMIT 10
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query, workerID)
	if err != nil {
		log.Error().Err(err).Str("worker_id", workerID).Msg("Failed to query worker builds")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch builds")
		return
	}
	defer rows.Close()

	builds := []map[string]interface{}{}
	for rows.Next() {
		var build models.Build
		var jobName, scmURL, scmType string
		var buildConfig models.JSONB

		err := rows.Scan(
			&build.ID, &build.JobID, &build.BuildNumber, &build.Status,
			&build.QueuedAt, &build.CommitSHA, &build.Branch,
			&build.TriggeredBy, &buildConfig, &jobName, &scmURL, &scmType,
		)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to scan build row")
			continue
		}

		buildMap := map[string]interface{}{
			"id":           build.ID,
			"job_id":       build.JobID,
			"job_name":     jobName,
			"build_number": build.BuildNumber,
			"status":       build.Status,
			"queued_at":    build.QueuedAt,
			"commit_sha":   build.CommitSHA,
			"branch":       build.Branch,
			"triggered_by": build.TriggeredBy,
			"build_config": buildConfig,
			"scm_url":      scmURL,
			"scm_type":     scmType,
		}

		builds = append(builds, buildMap)
	}

	SendJSON(w, http.StatusOK, builds)
}

// UpdateBuildStatus updates the status and details of a build
func (h *BuildHandler) UpdateBuildStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	buildID := vars["id"]

	// Validate build ID
	if _, err := uuid.Parse(buildID); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid build ID")
		return
	}

	var req struct {
		Status       string  `json:"status"`
		StartedAt    *string `json:"started_at,omitempty"`
		CompletedAt  *string `json:"completed_at,omitempty"`
		ExitCode     *int    `json:"exit_code,omitempty"`
		ErrorMessage *string `json:"error_message,omitempty"`
		Duration     *int    `json:"duration_seconds,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"queued": true, "running": true, "success": true,
		"failure": true, "cancelled": true,
	}
	if !validStatuses[req.Status] {
		SendError(w, http.StatusBadRequest, nil, "Invalid status value")
		return
	}

	// Build dynamic update query
	query := `UPDATE builds SET status = $1, updated_at = NOW()`
	args := []interface{}{req.Status}
	argCount := 2

	if req.StartedAt != nil {
		query += `, started_at = $` + strconv.Itoa(argCount)
		args = append(args, req.StartedAt)
		argCount++
	}

	if req.CompletedAt != nil {
		query += `, completed_at = $` + strconv.Itoa(argCount)
		args = append(args, req.CompletedAt)
		argCount++
	}

	if req.ExitCode != nil {
		query += `, exit_code = $` + strconv.Itoa(argCount)
		args = append(args, req.ExitCode)
		argCount++
	}

	if req.ErrorMessage != nil {
		query += `, error_message = $` + strconv.Itoa(argCount)
		args = append(args, req.ErrorMessage)
		argCount++
	}

	if req.Duration != nil {
		query += `, duration_seconds = $` + strconv.Itoa(argCount)
		args = append(args, req.Duration)
		argCount++
	}

	query += ` WHERE id = $` + strconv.Itoa(argCount)
	args = append(args, buildID)

	result, err := h.db.GetConn().ExecContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Str("build_id", buildID).Msg("Failed to update build status")
		SendError(w, http.StatusInternalServerError, err, "Failed to update build")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		SendError(w, http.StatusNotFound, nil, "Build not found")
		return
	}

	log.Info().
		Str("build_id", buildID).
		Str("status", req.Status).
		Msg("Build status updated")

	SendJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Build status updated successfully",
		"build_id": buildID,
		"status":   req.Status,
	})
}
