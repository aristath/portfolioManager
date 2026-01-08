// Package embedded provides embedded static assets for the application.
package embedded

import (
	"embed"
)

// Files contains all files embedded in the Go binary:
// - Frontend files (frontend/dist) - served directly via HTTP
// - Display app files (display/app) - extracted to disk at deployment time
// - Display app.yaml (display/app.yaml) - extracted to root of app directory
// - Sketch files (display/sketch) - extracted to disk, compiled and uploaded
//
// Note: Files are copied into pkg/embedded/ during GitHub Actions build
//
//go:embed frontend/dist display/app display/app.yaml display/sketch
var Files embed.FS
