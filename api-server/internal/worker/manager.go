package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/vrenjith/ritmo/api-server/internal/database"
	"github.com/vrenjith/ritmo/api-server/internal/metrics"
)

// Manager handles worker registration and health monitoring
type Manager struct {
	db      *database.Database
	metrics *metrics.Collector
}

// NewManager creates a new worker manager
func NewManager(db *database.Database, m *metrics.Collector) *Manager {
	return &Manager{
		db:      db,
		metrics: m,
	}
}

// Start begins the worker management loop
func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("Worker manager started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Worker manager stopped")
			return
		case <-ticker.C:
			m.checkWorkerHealth(ctx)
		}
	}
}

// checkWorkerHealth monitors worker heartbeats and marks stale workers as offline
func (m *Manager) checkWorkerHealth(ctx context.Context) {
	query := `
		UPDATE workers
		SET status = 'offline', health_status = 'unhealthy'
		WHERE status = 'online'
		  AND last_heartbeat < CURRENT_TIMESTAMP - INTERVAL '2 minutes'
		RETURNING id, name
	`

	rows, err := m.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check worker health")
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		log.Warn().Str("worker_id", id).Str("worker_name", name).Msg("Worker marked as offline due to missed heartbeat")
		count++
	}

	if count > 0 {
		log.Info().Int("count", count).Msg("Marked workers as offline")
	}

	// Update metrics
	m.updateWorkerMetrics(ctx)
}

// updateWorkerMetrics collects and records worker metrics
func (m *Manager) updateWorkerMetrics(ctx context.Context) {
	query := `
		SELECT status, COUNT(*)
		FROM workers
		GROUP BY status
	`

	rows, err := m.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query worker metrics")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		m.metrics.RecordWorkerCount(status, count)
	}
}
