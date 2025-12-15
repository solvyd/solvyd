package handlers

import (
	"net/http"

	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/scheduler"
)

// WebhookHandler handles webhook requests from SCM providers
type WebhookHandler struct {
	db    *database.Database
	sched *scheduler.Scheduler
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(db *database.Database, sched *scheduler.Scheduler) *WebhookHandler {
	return &WebhookHandler{db: db, sched: sched}
}

// HandleWebhook processes incoming webhooks
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement webhook processing for GitHub, GitLab, etc.
	// This would parse webhook payload, verify signatures, and trigger builds
	SendJSON(w, http.StatusOK, map[string]string{
		"status": "webhook received",
		"note":   "Implementation pending",
	})
}
