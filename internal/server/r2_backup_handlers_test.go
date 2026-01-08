package server

import (
	"testing"
)

func TestValidateBackupFilename(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantErr     bool
		expectedOut string
	}{
		{
			name:        "valid filename",
			filename:    "sentinel-backup-2026-01-08-143022.tar.gz",
			wantErr:     false,
			expectedOut: "sentinel-backup-2026-01-08-143022.tar.gz",
		},
		{
			name:     "empty filename",
			filename: "",
			wantErr:  true,
		},
		{
			name:     "path traversal attempt",
			filename: "../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:        "path traversal with valid prefix",
			filename:    "../sentinel-backup-2026-01-08-143022.tar.gz",
			wantErr:     false, // filepath.Base will extract the filename
			expectedOut: "sentinel-backup-2026-01-08-143022.tar.gz",
		},
		{
			name:     "absolute path",
			filename: "/tmp/sentinel-backup-2026-01-08-143022.tar.gz",
			wantErr:  false, // filepath.Base will extract the filename
			expectedOut: "sentinel-backup-2026-01-08-143022.tar.gz",
		},
		{
			name:     "invalid prefix",
			filename: "malicious-backup-2026-01-08-143022.tar.gz",
			wantErr:  true,
		},
		{
			name:     "invalid suffix",
			filename: "sentinel-backup-2026-01-08-143022.zip",
			wantErr:  true,
		},
		{
			name:     "missing tar.gz extension",
			filename: "sentinel-backup-2026-01-08-143022",
			wantErr:  true,
		},
		{
			name:     "windows path separator",
			filename: "sentinel-backup-2026-01-08-143022.tar.gz\\file",
			wantErr:  true,
		},
		{
			name:     "unix path separator in name",
			filename: "sentinel-backup/2026-01-08-143022.tar.gz",
			wantErr:  true,
		},
		{
			name:     "extremely long filename",
			filename: "sentinel-backup-" + string(make([]byte, 300)) + ".tar.gz",
			wantErr:  true,
		},
		{
			name:        "valid filename with path (should extract basename)",
			filename:    "some/path/sentinel-backup-2026-01-08-143022.tar.gz",
			wantErr:     false,
			expectedOut: "sentinel-backup-2026-01-08-143022.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateBackupFilename(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateBackupFilename() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("validateBackupFilename() unexpected error: %v", err)
				return
			}

			if tt.expectedOut != "" && got != tt.expectedOut {
				t.Errorf("validateBackupFilename() = %v, want %v", got, tt.expectedOut)
			}
		})
	}
}
