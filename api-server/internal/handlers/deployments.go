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

// DeploymentHandler handles deployment-related requests
type DeploymentHandler struct {
	db *database.Database
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(db *database.Database) *DeploymentHandler {
	return &DeploymentHandler{db: db}
}

// ListDeployments returns all deployments
func (h *DeploymentHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	environment := r.URL.Query().Get("environment")
	buildID := r.URL.Query().Get("build_id")

	query := `
		SELECT id, build_id, artifact_id, environment, status, target_type,
		       target_url, started_at, completed_at, duration_seconds,
		       deployment_plugin, exit_code, error_message, deployment_url,
		       deployed_by, created_at
		FROM deployments
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if environment != "" {
		query += ` AND environment = $` + string(rune('0'+argCount))
		args = append(args, environment)
		argCount++
	}
	if buildID != "" {
		query += ` AND build_id = $` + string(rune('0'+argCount))
		args = append(args, buildID)
		argCount++
	}

	query += ` ORDER BY started_at DESC LIMIT 100`

	rows, err := h.db.GetConn().QueryContext(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query deployments")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch deployments")
		return
	}
	defer rows.Close()

	deployments := []models.Deployment{}
	for rows.Next() {
		var d models.Deployment
		err := rows.Scan(
			&d.ID, &d.BuildID, &d.ArtifactID, &d.Environment, &d.Status,
			&d.TargetType, &d.TargetURL, &d.StartedAt, &d.CompletedAt,
			&d.Duration, &d.DeploymentPlugin, &d.ExitCode, &d.ErrorMessage,
			&d.DeploymentURL, &d.DeployedBy, &d.CreatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan deployment row")
			continue
		}
		deployments = append(deployments, d)
	}

	SendJSON(w, http.StatusOK, deployments)
}

// GetDeployment returns a single deployment
func (h *DeploymentHandler) GetDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	deploymentID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid deployment ID")
		return
	}

	query := `
		SELECT id, build_id, artifact_id, environment, status, target_type,
		       target_url, target_metadata, started_at, completed_at,
		       duration_seconds, deployment_plugin, exit_code, error_message,
		       deployment_url, rollback_from_deployment_id, deployed_by,
		       deployment_notes, created_at
		FROM deployments
		WHERE id = $1
	`

	var d models.Deployment
	err = h.db.GetConn().QueryRowContext(ctx, query, deploymentID).Scan(
		&d.ID, &d.BuildID, &d.ArtifactID, &d.Environment, &d.Status,
		&d.TargetType, &d.TargetURL, &d.TargetMetadata, &d.StartedAt,
		&d.CompletedAt, &d.Duration, &d.DeploymentPlugin, &d.ExitCode,
		&d.ErrorMessage, &d.DeploymentURL, &d.RollbackFromID, &d.DeployedBy,
		&d.DeploymentNotes, &d.CreatedAt,
	)
	if err == sql.ErrNoRows {
		SendError(w, http.StatusNotFound, nil, "Deployment not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query deployment")
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch deployment")
		return
	}

	SendJSON(w, http.StatusOK, d)
}

// CreateDeployment creates a new deployment
func (h *DeploymentHandler) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		BuildID     uuid.UUID `json:"build_id"`
		ArtifactID  uuid.UUID `json:"artifact_id"`
		Environment string    `json:"environment"`
		TargetType  string    `json:"target_type"`
		TargetURL   string    `json:"target_url"`
		DeployedBy  string    `json:"deployed_by"`
		Notes       string    `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	deploymentID := uuid.New()

	query := `
		INSERT INTO deployments (id, build_id, artifact_id, environment, status,
		                        target_type, target_url, deployed_by, deployment_notes)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8)
		RETURNING id, started_at
	`

	var d struct {
		ID        uuid.UUID `json:"id"`
		StartedAt string    `json:"started_at"`
	}

	err := h.db.GetConn().QueryRowContext(ctx, query,
		deploymentID, req.BuildID, req.ArtifactID, req.Environment,
		req.TargetType, req.TargetURL, req.DeployedBy, req.Notes,
	).Scan(&d.ID, &d.StartedAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create deployment")
		SendError(w, http.StatusInternalServerError, err, "Failed to create deployment")
		return
	}

	log.Info().Str("deployment_id", d.ID.String()).Str("environment", req.Environment).Msg("Deployment created")
	SendJSON(w, http.StatusCreated, d)
}

// RollbackDeployment creates a rollback deployment
func (h *DeploymentHandler) RollbackDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	deploymentID, err := uuid.Parse(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, err, "Invalid deployment ID")
		return
	}

	// Get the deployment to rollback from
	var deployment models.Deployment
	query := `SELECT build_id, environment, target_type, target_url FROM deployments WHERE id = $1`
	err = h.db.GetConn().QueryRowContext(ctx, query, deploymentID).Scan(
		&deployment.BuildID, &deployment.Environment, &deployment.TargetType, &deployment.TargetURL,
	)
	if err != nil {
		SendError(w, http.StatusNotFound, nil, "Deployment not found")
		return
	}

	// Create rollback deployment (logic would find previous successful deployment)
	// For now, this is a stub
	SendJSON(w, http.StatusOK, map[string]string{
		"status":  "rollback_initiated",
		"message": "Rollback functionality to be implemented",
	})
}
