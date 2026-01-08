package reliability

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/rs/zerolog"
)

// TestNewR2Client tests R2 client creation
func TestNewR2Client(t *testing.T) {
	log := zerolog.New(io.Discard)

	tests := []struct {
		name            string
		accountID       string
		accessKeyID     string
		secretAccessKey string
		bucketName      string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "valid credentials",
			accountID:       "test-account-id",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			expectError:     false,
		},
		{
			name:            "missing account ID",
			accountID:       "",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			expectError:     true,
			errorContains:   "r2 credentials incomplete",
		},
		{
			name:            "missing access key",
			accountID:       "test-account-id",
			accessKeyID:     "",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			expectError:     true,
			errorContains:   "r2 credentials incomplete",
		},
		{
			name:            "missing secret key",
			accountID:       "test-account-id",
			accessKeyID:     "test-access-key",
			secretAccessKey: "",
			bucketName:      "test-bucket",
			expectError:     true,
			errorContains:   "r2 credentials incomplete",
		},
		{
			name:            "missing bucket name",
			accountID:       "test-account-id",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "",
			expectError:     true,
			errorContains:   "r2 credentials incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewR2Client(tt.accountID, tt.accessKeyID, tt.secretAccessKey, tt.bucketName, log)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorContains)
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
				if client != nil {
					if client.bucket != tt.bucketName {
						t.Errorf("expected bucket %q, got %q", tt.bucketName, client.bucket)
					}
					if client.client == nil {
						t.Error("expected S3 client to be initialized")
					}
					if client.uploader == nil {
						t.Error("expected uploader to be initialized")
					}
					if client.downloader == nil {
						t.Error("expected downloader to be initialized")
					}
				}
			}
		})
	}
}

// TestR2ClientMethods tests basic method signatures and structure
// Note: These are structure tests only. Integration tests with real R2 would be separate.
func TestR2ClientMethods(t *testing.T) {
	log := zerolog.New(io.Discard)

	client, err := NewR2Client("test-account", "test-key", "test-secret", "test-bucket", log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test that methods exist and have correct signatures
	ctx := context.Background()

	t.Run("Upload method exists", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		// We expect this to fail since we're not connected to real R2,
		// but we're testing that the method signature is correct
		_ = client.Upload(ctx, "test-key", reader, 9)
	})

	t.Run("Download method exists", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		writerAt := &WriterAtWrapper{w: buffer}
		// We expect this to fail since we're not connected to real R2
		_, _ = client.Download(ctx, "test-key", writerAt)
	})

	t.Run("List method exists", func(t *testing.T) {
		// We expect this to fail since we're not connected to real R2
		_, _ = client.List(ctx, "")
	})

	t.Run("Delete method exists", func(t *testing.T) {
		// We expect this to fail since we're not connected to real R2
		_ = client.Delete(ctx, "test-key")
	})

	t.Run("TestConnection method exists", func(t *testing.T) {
		// We expect this to fail since we're not connected to real R2
		_ = client.TestConnection(ctx)
	})

	t.Run("GetObjectMetadata method exists", func(t *testing.T) {
		// We expect this to fail since we're not connected to real R2
		_, _ = client.GetObjectMetadata(ctx, "test-key")
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WriterAtWrapper wraps an io.Writer to implement io.WriterAt
type WriterAtWrapper struct {
	w      io.Writer
	offset int64
}

func (w *WriterAtWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	if off != w.offset {
		return 0, errors.New("WriterAtWrapper only supports sequential writes")
	}
	n, err = w.w.Write(p)
	w.offset += int64(n)
	return n, err
}
