package handlers

import (
	"net/http"

	"github.com/solvyd/solvyd/api-server/internal/gitops"
)

// GitOpsHandler handles GitOps-related requests
type GitOpsHandler struct {
	sync *gitops.SyncService
}

// NewGitOpsHandler creates a new GitOps handler
func NewGitOpsHandler(sync *gitops.SyncService) *GitOpsHandler {
	return &GitOpsHandler{sync: sync}
}

// GetStatus returns the current GitOps sync status
func (h *GitOpsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.sync.GetStatus()
	SendJSON(w, http.StatusOK, status)
}

// TriggerSync triggers a manual synchronization
func (h *GitOpsHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	if err := h.sync.Sync(); err != nil {
		SendError(w, http.StatusInternalServerError, err, "Sync failed")
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "GitOps sync completed",
	})
}
