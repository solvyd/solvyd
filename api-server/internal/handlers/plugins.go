package handlers

import (
	"net/http"

	"github.com/solvyd/solvyd/api-server/internal/database"
	"github.com/solvyd/solvyd/api-server/internal/models"
)

// PluginHandler handles plugin-related requests
type PluginHandler struct {
	db *database.Database
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(db *database.Database) *PluginHandler {
	return &PluginHandler{db: db}
}

// ListPlugins returns all plugins
func (h *PluginHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := `
		SELECT id, name, type, version, description, author, homepage_url,
		       enabled, installed_at, updated_at
		FROM plugins
		ORDER BY type, name
	`

	rows, err := h.db.GetConn().QueryContext(ctx, query)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err, "Failed to fetch plugins")
		return
	}
	defer rows.Close()

	plugins := []models.Plugin{}
	for rows.Next() {
		var p models.Plugin
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.Version, &p.Description,
			&p.Author, &p.HomepageURL, &p.Enabled, &p.InstalledAt, &p.UpdatedAt,
		)
		if err != nil {
			continue
		}
		plugins = append(plugins, p)
	}

	SendJSON(w, http.StatusOK, plugins)
}

// GetPlugin returns a single plugin
func (h *PluginHandler) GetPlugin(w http.ResponseWriter, r *http.Request) {
	SendJSON(w, http.StatusOK, map[string]string{"status": "stub"})
}

// InstallPlugin installs a new plugin
func (h *PluginHandler) InstallPlugin(w http.ResponseWriter, r *http.Request) {
	SendJSON(w, http.StatusOK, map[string]string{"status": "stub - plugin installation"})
}
