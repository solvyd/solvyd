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
)

// JobHandler handles job-related requests
type JobHandler struct {
	db *database.Database
}

// NewJobHandler creates a new job handler
func NewJobHandler(db *database.Database) *JobHandler {
	return &JobHandler{db: db}
}

// ListJobs returns all jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := `
		SELECT id, name, description, scm_type, scm_url, scm_branch, 
		       build_config, environment_vars, triggers, enabled, 
		       worker_labels, plugins, pipeline_stages, timeout_minutes, 
		       max_retries, created_at, updated_at, created_by
		FROM jobs
		ORDER BY created_at DESC
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query jobs")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch jobs")
		return
	}
	defer rows.Close()

	jobs := []models.Job{}
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.SCMType, &job.SCMURL,
			&job.SCMBranch, &job.BuildConfig, &job.EnvVars, &job.Triggers,
			&job.Enabled, &job.WorkerLabels, &job.Plugins, &job.PipelineStages,
			&job.TimeoutMinutes, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
			&job.CreatedBy,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan job row")
			continue
		}
		jobs = append(jobs, job)
	}

	SendJSON(w, http.StatusOK, jobs)
}

// GetJob returns a single job
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid job ID")
		return
	}

	query := `
		SELECT id, name, description, scm_type, scm_url, scm_branch, 
		       build_config, environment_vars, triggers, enabled, 
		       worker_labels, plugins, pipeline_stages, timeout_minutes, 
		       max_retries, created_at, updated_at, created_by
		FROM jobs
		WHERE id = $1
	`

	var job models.Job
	err = h.db.GetConn().QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.Name, &job.Description, &job.SCMType, &job.SCMURL,
		&job.SCMBranch, &job.BuildConfig, &job.EnvVars, &job.Triggers,
		&job.Enabled, &job.WorkerLabels, &job.Plugins, &job.PipelineStages,
		&job.TimeoutMinutes, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
		&job.CreatedBy,
	)
	if err == sql.ErrNoRows {
		SendError(w, http.StatusNotFound, nil, "Job not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query job")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch job")
		return
	}

	SendJSON(w, http.StatusOK, job)
}

// CreateJob creates a new job
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var job models.Job

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate required fields
	if job.Name == "" {
		SendError(w, http.StatusBadRequest, nil, "Job name is required")
		return
	}

	job.ID = uuid.New()

	query := `
		INSERT INTO jobs (id, name, description, scm_type, scm_url, scm_branch,
		                  build_config, environment_vars, triggers, enabled,
		                  worker_labels, plugins, pipeline_stages, timeout_minutes,
		                  max_retries, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`

	err := h.db.GetConn().QueryRowContext(ctx, query,
		job.ID, job.Name, job.Description, job.SCMType, job.SCMURL, job.SCMBranch,
		job.BuildConfig, job.EnvVars, job.Triggers, job.Enabled,
		job.WorkerLabels, job.Plugins, job.PipelineStages, job.TimeoutMinutes,
		job.MaxRetries, job.CreatedBy,
	).Scan(&job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create job")
		SendError(w, http.StatusInternalServerError, err, "Failed to create job")
		return
	}

	log.Info().Str("job_id", job.ID.String()).Str("job_name", job.Name).Msg("Job created")
	SendJSON(w, http.StatusCreated, job)
}

// UpdateJob updates an existing job
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid job ID")
		return
	}

	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	query := `
		UPDATE jobs
		SET name = $2, description = $3, scm_type = $4, scm_url = $5, scm_branch = $6,
		    build_config = $7, environment_vars = $8, triggers = $9, enabled = $10,
		    worker_labels = $11, plugins = $12, pipeline_stages = $13,
		    timeout_minutes = $14, max_retries = $15
		WHERE id = $1
	`

	result, err := h.db.GetConn().ExecContext(ctx, query,
		jobID, job.Name, job.Description, job.SCMType, job.SCMURL, job.SCMBranch,
		job.BuildConfig, job.EnvVars, job.Triggers, job.Enabled,
		job.WorkerLabels, job.Plugins, job.PipelineStages, job.TimeoutMinutes,
		job.MaxRetries,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to update job")
		SendError(w, http.StatusInternalServerError, err, "Failed to update job")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		SendError(w, http.StatusNotFound, nil, "Job not found")
		return
	}

	log.Info().Str("job_id", jobID.String()).Msg("Job updated")
	SendJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteJob deletes a job
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid job ID")
		return
	}

	query := `DELETE FROM jobs WHERE id = $1`

	result, err := h.db.GetConn().ExecContext(ctx, query, jobID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete job")
		SendError(w, http.StatusInternalServerError, err, "Failed to delete job")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		SendError(w, http.StatusNotFound, nil, "Job not found")
		return
	}

	log.Info().Str("job_id", jobID.String()).Msg("Job deleted")
	SendJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// TriggerJob triggers a manual build for a job
func (h *JobHandler) TriggerJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid job ID")
		return
	}

	// Parse optional parameters
	var params struct {
		Parameters map[string]interface{} `json:"parameters"`
		Branch     string                 `json:"branch"`
	}
	json.NewDecoder(r.Body).Decode(&params)

	// Create a new build
	buildID := uuid.New()

	query := `
		INSERT INTO builds (id, job_id, status, triggered_by, parameters, branch)
		VALUES ($1, $2, 'queued', 'manual', $3, $4)
		RETURNING id, build_number, queued_at
	`

	paramsJSON, _ := json.Marshal(params.Parameters)
	var build struct {
		ID          uuid.UUID `json:"id"`
		BuildNumber int       `json:"build_number"`
		QueuedAt    string    `json:"queued_at"`
	}

	err = h.db.GetConn().QueryRowContext(ctx, query, buildID, jobID, paramsJSON, params.Branch).
		Scan(&build.ID, &build.BuildNumber, &build.QueuedAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to trigger build")
		SendError(w, http.StatusInternalServerError, err, "Failed to trigger build")
		return
	}

	log.Info().
		Str("job_id", jobID.String()).
		Str("build_id", build.ID.String()).
		Int("build_number", build.BuildNumber).
		Msg("Build triggered")

	SendJSON(w, http.StatusCreated, build)
}
