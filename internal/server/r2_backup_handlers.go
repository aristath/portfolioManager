package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aristath/sentinel/internal/reliability"
	"github.com/aristath/sentinel/internal/work"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// R2BackupHandlers handles R2 cloud backup operations
type R2BackupHandlers struct {
	r2BackupService *reliability.R2BackupService
	restoreService  *reliability.RestoreService
	workProcessor   *work.Processor
	log             zerolog.Logger
}

// validateBackupFilename validates that a filename is safe and follows expected format
// Returns sanitized filename or error
func validateBackupFilename(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename is empty")
	}

	// Clean the filename to prevent path traversal
	filename = filepath.Base(filename)

	// Check for any remaining path separators (defense in depth)
	if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
		return "", fmt.Errorf("invalid filename: contains path separators")
	}

	// Validate filename format: sentinel-backup-YYYY-MM-DD-HHMMSS.tar.gz
	if !strings.HasPrefix(filename, "sentinel-backup-") || !strings.HasSuffix(filename, ".tar.gz") {
		return "", fmt.Errorf("invalid filename format")
	}

	// Additional length check to prevent extremely long filenames
	if len(filename) > 255 {
		return "", fmt.Errorf("filename too long")
	}

	return filename, nil
}

// NewR2BackupHandlers creates new R2 backup handlers
func NewR2BackupHandlers(
	r2BackupService *reliability.R2BackupService,
	restoreService *reliability.RestoreService,
	workProcessor *work.Processor,
	log zerolog.Logger,
) *R2BackupHandlers {
	return &R2BackupHandlers{
		r2BackupService: r2BackupService,
		restoreService:  restoreService,
		workProcessor:   workProcessor,
		log:             log.With().Str("handler", "r2_backup").Logger(),
	}
}

// HandleListBackups lists all backups from R2
// GET /api/backups/r2
func (h *R2BackupHandlers) HandleListBackups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.r2BackupService == nil {
		h.writeJSON(w, map[string]string{
			"error": "R2 backup service not configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	backups, err := h.r2BackupService.ListBackups(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list R2 backups")
		http.Error(w, fmt.Sprintf("Failed to list backups: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"backups": backups,
		"count":   len(backups),
	})
}

// HandleCreateBackup triggers an immediate backup to R2
// POST /api/backups/r2
func (h *R2BackupHandlers) HandleCreateBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.r2BackupService == nil {
		h.writeJSON(w, map[string]string{
			"error": "R2 backup service not configured",
		})
		return
	}

	// Execute R2 backup via Work Processor
	if h.workProcessor == nil {
		h.log.Error().Msg("Work processor not available")
		http.Error(w, "Work processor not available", http.StatusInternalServerError)
		return
	}

	if err := h.workProcessor.ExecuteNow("maintenance:r2-backup", ""); err != nil {
		h.log.Error().Err(err).Msg("Failed to execute R2 backup")
		http.Error(w, fmt.Sprintf("Failed to execute backup: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Info().Msg("Manual R2 backup triggered")

	h.writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Backup executed",
	})
}

// HandleTestConnection tests R2 connection and credentials
// POST /api/backups/r2/test
func (h *R2BackupHandlers) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.r2BackupService == nil {
		h.writeJSON(w, map[string]string{
			"error": "R2 backup service not configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Test connection via R2 client
	if err := h.r2BackupService.GetR2Client().TestConnection(ctx); err != nil {
		h.log.Error().Err(err).Msg("R2 connection test failed")
		h.writeJSON(w, map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Connection failed: %v", err),
		})
		return
	}

	h.writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Connection successful",
	})
}

// HandleDeleteBackup deletes a specific backup from R2
// DELETE /api/backups/r2/{filename}
func (h *R2BackupHandlers) HandleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.r2BackupService == nil {
		h.writeJSON(w, map[string]string{
			"error": "R2 backup service not configured",
		})
		return
	}

	filename := chi.URLParam(r, "filename")

	// Validate and sanitize filename
	validatedFilename, err := validateBackupFilename(filename)
	if err != nil {
		h.log.Warn().Err(err).Str("filename", filename).Msg("Invalid filename in delete request")
		http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
		return
	}
	filename = validatedFilename

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
	defer cancel()

	if err := h.r2BackupService.GetR2Client().Delete(ctx, filename); err != nil {
		h.log.Error().Err(err).Str("filename", filename).Msg("Failed to delete R2 backup")
		http.Error(w, fmt.Sprintf("Failed to delete backup: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Info().Str("filename", filename).Msg("R2 backup deleted")

	h.writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Backup deleted",
	})
}

// HandleStageRestore stages a restore and restarts the service
// POST /api/backups/r2/restore
func (h *R2BackupHandlers) HandleStageRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.restoreService == nil {
		h.writeJSON(w, map[string]string{
			"error": "Restore service not configured",
		})
		return
	}

	// Parse request body
	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize filename
	validatedFilename, err := validateBackupFilename(req.Filename)
	if err != nil {
		h.log.Warn().Err(err).Str("filename", req.Filename).Msg("Invalid filename in restore request")
		http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
		return
	}
	req.Filename = validatedFilename

	h.log.Info().Str("filename", req.Filename).Msg("Staging restore from R2")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	// Stage restore (downloads, validates, creates flag file)
	if err := h.restoreService.StageRestoreFromR2(ctx, req.Filename); err != nil {
		h.log.Error().Err(err).Str("filename", req.Filename).Msg("Failed to stage restore")
		http.Error(w, fmt.Sprintf("Failed to stage restore: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Info().Msg("Restore staged successfully, restarting service")

	// Restart the service
	// Use systemctl or supervisorctl depending on setup
	restartCmd := "systemctl restart sentinel"

	// Respond before restarting
	h.writeJSON(w, map[string]string{
		"status":          "success",
		"message":         "Restore staged, restarting service...",
		"restart_command": restartCmd,
	})

	// Give the response time to be sent
	go func() {
		time.Sleep(500 * time.Millisecond)
		cmd := exec.Command("sh", "-c", restartCmd)
		if err := cmd.Run(); err != nil {
			h.log.Error().Err(err).Msg("Failed to restart service")
		}
	}()
}

// HandleCancelRestore cancels a staged restore
// DELETE /api/backups/r2/restore/staged
func (h *R2BackupHandlers) HandleCancelRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.restoreService == nil {
		h.writeJSON(w, map[string]string{
			"error": "Restore service not configured",
		})
		return
	}

	if err := h.restoreService.CancelStagedRestore(); err != nil {
		h.log.Error().Err(err).Msg("Failed to cancel staged restore")
		http.Error(w, fmt.Sprintf("Failed to cancel restore: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Info().Msg("Staged restore canceled")

	h.writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Staged restore canceled",
	})
}

// HandleDownloadBackup downloads a backup archive from R2
// GET /api/backups/r2/{filename}/download
func (h *R2BackupHandlers) HandleDownloadBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.r2BackupService == nil {
		http.Error(w, "R2 backup service not configured", http.StatusServiceUnavailable)
		return
	}

	filename := chi.URLParam(r, "filename")

	// Validate and sanitize filename
	validatedFilename, err := validateBackupFilename(filename)
	if err != nil {
		h.log.Warn().Err(err).Str("filename", filename).Msg("Invalid filename in download request")
		http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
		return
	}
	filename = validatedFilename

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "r2-download-*.tar.gz")
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create temp file")
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download from R2
	writerAt := &reliability.FileWriterAt{File: tmpFile, Offset: 0}
	_, err = h.r2BackupService.GetR2Client().Download(ctx, filename, writerAt)
	if err != nil {
		h.log.Error().Err(err).Str("filename", filename).Msg("Failed to download from R2")
		http.Error(w, fmt.Sprintf("Failed to download: %v", err), http.StatusInternalServerError)
		return
	}

	// Seek back to start
	if _, err := tmpFile.Seek(0, 0); err != nil {
		http.Error(w, fmt.Sprintf("Failed to seek file: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Stream file to response
	http.ServeContent(w, r, filename, time.Now(), tmpFile)
}

// writeJSON writes a JSON response
func (h *R2BackupHandlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
