package scheduler

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/metrics"
	"github.com/solvyd/solvyd/api-server/internal/worker"
)

// Scheduler handles job scheduling and build assignment
type Scheduler struct {
	db        *database.Database
	workerMgr *worker.Manager
	metrics   *metrics.Collector
}

// NewScheduler creates a new scheduler
func NewScheduler(db *database.Database, workerMgr *worker.Manager, m *metrics.Collector) *Scheduler {
	return &Scheduler{
		db:        db,
		workerMgr: workerMgr,
		metrics:   m,
	}
}

// Start begins the scheduler loop
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("Scheduler started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Scheduler stopped")
			return
		case <-ticker.C:
			s.schedulePendingBuilds(ctx)
		}
	}
}

// schedulePendingBuilds assigns queued builds to available workers
func (s *Scheduler) schedulePendingBuilds(ctx context.Context) {
	// Get queued builds
	query := `
		SELECT id, job_id
		FROM builds
		WHERE status = 'queued'
		ORDER BY queued_at ASC
		LIMIT 10
	`

	rows, err := s.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query queued builds")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var buildID, jobID uuid.UUID
		if err := rows.Scan(&buildID, &jobID); err != nil {
			continue
		}

		// Try to assign to a worker
		if err := s.assignBuildToWorker(ctx, buildID, jobID); err != nil {
			log.Debug().Err(err).Str("build_id", buildID.String()).Msg("Could not assign build to worker")
		}
	}
}

// assignBuildToWorker finds an available worker and assigns the build
func (s *Scheduler) assignBuildToWorker(ctx context.Context, buildID, jobID uuid.UUID) error {
	// Find available worker
	query := `
		SELECT id
		FROM workers
		WHERE status = 'online'
		  AND current_builds < max_concurrent_builds
		ORDER BY current_builds ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var workerID uuid.UUID
	err := s.db.GetConn().QueryRowContext(ctx, query).Scan(&workerID)
	if err == sql.ErrNoRows {
		return nil // No workers available, will retry next tick
	}
	if err != nil {
		return err
	}

	// Assign build to worker
	updateBuild := `
		UPDATE builds
		SET status = 'running', worker_id = $1, started_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND status = 'queued'
	`
	if _, err := s.db.GetConn().ExecContext(ctx, updateBuild, workerID, buildID); err != nil {
		return err
	}

	// Increment worker's current_builds count
	updateWorker := `
		UPDATE workers
		SET current_builds = current_builds + 1
		WHERE id = $1
	`
	if _, err := s.db.GetConn().ExecContext(ctx, updateWorker, workerID); err != nil {
		log.Error().Err(err).Msg("Failed to increment worker build count")
	}

	log.Info().
		Str("build_id", buildID.String()).
		Str("worker_id", workerID.String()).
		Msg("Build assigned to worker")

	s.metrics.RecordBuildScheduled()

	return nil
}
