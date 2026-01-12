// Package embedded provides embedded static assets for the application.
package embedded

import (
	"embed"
)

// Files contains all files embedded in the Go binary:
// - Frontend files (frontend/dist) - served directly via HTTP
// - Sketch files (display/sketch) - extracted to disk, compiled and uploaded
//
// Note: Files are copied into pkg/embedded/ during GitHub Actions build
// Note: Python display app has been removed - MCU communication is now handled
//
//	directly by the Go service via Unix socket to arduino-router
//
//go:embed frontend/dist display/sketch
var Files embed.FS
