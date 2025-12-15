package gitops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/solvyd/solvyd/api-server/internal/config"
	"github.com/solvyd/solvyd/api-server/internal/database"
)

// SyncService handles GitOps synchronization
type SyncService struct {
	cfg      *config.GitOpsConfig
	db       *database.Database
	repoPath string
	lastSync time.Time
	lastHash string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSyncService creates a new GitOps sync service
func NewSyncService(cfg *config.GitOpsConfig, db *database.Database) *SyncService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncService{
		cfg:      cfg,
		db:       db,
		repoPath: filepath.Join(os.TempDir(), "solvyd-gitops-repo"),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the GitOps synchronization loop
func (s *SyncService) Start() error {
	if !s.cfg.Enabled {
		log.Info().Msg("GitOps sync disabled")
		return nil
	}

	log.Info().
		Str("repo", s.cfg.Repository.URL).
		Str("branch", s.cfg.Repository.Branch).
		Int("interval", s.cfg.Sync.Interval).
		Msg("Starting GitOps sync service")

	// Initial sync
	if err := s.Sync(); err != nil {
		log.Error().Err(err).Msg("Initial GitOps sync failed")
		return err
	}

	// Start periodic sync
	go s.syncLoop()

	return nil
}

// Stop stops the sync service
func (s *SyncService) Stop() {
	log.Info().Msg("Stopping GitOps sync service")
	s.cancel()
}

// syncLoop runs periodic synchronization
func (s *SyncService) syncLoop() {
	ticker := time.NewTicker(time.Duration(s.cfg.Sync.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.Sync(); err != nil {
				log.Error().Err(err).Msg("GitOps sync failed")
			}
		}
	}
}

// Sync performs a synchronization from Git to database
func (s *SyncService) Sync() error {
	log.Info().Msg("Starting GitOps sync")

	// Clone or pull repository
	if err := s.cloneOrPull(); err != nil {
		return fmt.Errorf("failed to clone/pull repository: %w", err)
	}

	// Get current commit hash
	hash, err := s.getCurrentCommit()
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}

	// Check if anything changed
	if hash == s.lastHash {
		log.Debug().Msg("No changes detected, skipping sync")
		return nil
	}

	log.Info().
		Str("old_commit", s.lastHash).
		Str("new_commit", hash).
		Msg("Changes detected, applying configuration")

	// Apply configuration
	if err := s.applyConfiguration(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	s.lastHash = hash
	s.lastSync = time.Now()

	log.Info().
		Str("commit", hash).
		Time("sync_time", s.lastSync).
		Msg("GitOps sync completed successfully")

	return nil
}

// cloneOrPull clones the repository or pulls latest changes
func (s *SyncService) cloneOrPull() error {
	if _, err := os.Stat(filepath.Join(s.repoPath, ".git")); os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		log.Info().Str("path", s.repoPath).Msg("Cloning GitOps repository")
		return s.cloneRepository()
	}

	// Repository exists, pull latest changes
	log.Debug().Msg("Pulling latest changes")
	return s.pullRepository()
}

// cloneRepository clones the Git repository
func (s *SyncService) cloneRepository() error {
	// Remove existing directory if it exists
	os.RemoveAll(s.repoPath)

	cmd := exec.Command("git", "clone",
		"--branch", s.cfg.Repository.Branch,
		"--depth", "1",
		s.getAuthenticatedURL(),
		s.repoPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, output)
	}

	return nil
}

// pullRepository pulls latest changes
func (s *SyncService) pullRepository() error {
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = s.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w, output: %s", err, output)
	}

	return nil
}

// getCurrentCommit gets the current commit hash
func (s *SyncService) getCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = s.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return string(output[:40]), nil // First 40 chars = full SHA
}

// getAuthenticatedURL returns the Git URL with authentication
func (s *SyncService) getAuthenticatedURL() string {
	// Insert token into HTTPS URL
	// https://github.com/org/repo -> https://TOKEN@github.com/org/repo
	if s.cfg.Authentication.Type == "token" && s.cfg.Authentication.Token != "" {
		url := s.cfg.Repository.URL
		if len(url) > 8 && url[:8] == "https://" {
			return "https://" + s.cfg.Authentication.Token + "@" + url[8:]
		}
	}

	return s.cfg.Repository.URL
}

// applyConfiguration applies the configuration from Git to database
func (s *SyncService) applyConfiguration() error {
	configPath := filepath.Join(s.repoPath, s.cfg.Repository.Path)

	// Apply jobs
	if err := s.applyJobs(filepath.Join(configPath, "jobs")); err != nil {
		log.Error().Err(err).Msg("Failed to apply jobs")
		if !s.cfg.Sync.DryRun {
			return err
		}
	}

	// Apply credentials
	if err := s.applyCredentials(filepath.Join(configPath, "credentials")); err != nil {
		log.Error().Err(err).Msg("Failed to apply credentials")
		if !s.cfg.Sync.DryRun {
			return err
		}
	}

	// Apply webhooks
	if err := s.applyWebhooks(filepath.Join(configPath, "webhooks")); err != nil {
		log.Error().Err(err).Msg("Failed to apply webhooks")
		if !s.cfg.Sync.DryRun {
			return err
		}
	}

	return nil
}

// applyJobs applies job configurations
func (s *SyncService) applyJobs(jobsPath string) error {
	if _, err := os.Stat(jobsPath); os.IsNotExist(err) {
		log.Debug().Str("path", jobsPath).Msg("Jobs directory not found, skipping")
		return nil
	}

	files, err := filepath.Glob(filepath.Join(jobsPath, "*.yaml"))
	if err != nil {
		return err
	}

	log.Info().Int("count", len(files)).Msg("Applying jobs from GitOps")

	for _, file := range files {
		log.Debug().Str("file", file).Msg("Processing job configuration")
		// TODO: Parse YAML and apply to database
		// This will be implemented in the next iteration
	}

	return nil
}

// applyCredentials applies credential configurations
func (s *SyncService) applyCredentials(credsPath string) error {
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		log.Debug().Str("path", credsPath).Msg("Credentials directory not found, skipping")
		return nil
	}

	files, err := filepath.Glob(filepath.Join(credsPath, "*.yaml"))
	if err != nil {
		return err
	}

	log.Info().Int("count", len(files)).Msg("Applying credentials from GitOps")

	for _, file := range files {
		log.Debug().Str("file", file).Msg("Processing credential configuration")
		// TODO: Parse YAML and apply to database
	}

	return nil
}

// applyWebhooks applies webhook configurations
func (s *SyncService) applyWebhooks(webhooksPath string) error {
	if _, err := os.Stat(webhooksPath); os.IsNotExist(err) {
		log.Debug().Str("path", webhooksPath).Msg("Webhooks directory not found, skipping")
		return nil
	}

	files, err := filepath.Glob(filepath.Join(webhooksPath, "*.yaml"))
	if err != nil {
		return err
	}

	log.Info().Int("count", len(files)).Msg("Applying webhooks from GitOps")

	for _, file := range files {
		log.Debug().Str("file", file).Msg("Processing webhook configuration")
		// TODO: Parse YAML and apply to database
	}

	return nil
}

// GetStatus returns the current sync status
func (s *SyncService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":     s.cfg.Enabled,
		"repository":  s.cfg.Repository.URL,
		"branch":      s.cfg.Repository.Branch,
		"last_sync":   s.lastSync,
		"last_commit": s.lastHash,
		"auto_apply":  s.cfg.Sync.AutoApply,
		"dry_run":     s.cfg.Sync.DryRun,
	}
}
