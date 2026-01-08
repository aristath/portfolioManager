// Package reliability provides database backup, restore, and health monitoring services.
//
// The package includes:
// - Local backup service for SQLite databases
// - Cloudflare R2 cloud backup integration
// - Two-phase restore system with safety backups
// - Database health monitoring and integrity checks
package reliability

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog"
)

// R2Client wraps AWS S3 SDK to interact with Cloudflare R2.
// Cloudflare R2 is S3-compatible object storage, so we use the AWS SDK
// with a custom endpoint resolver pointing to R2's API.
type R2Client struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	log        zerolog.Logger
}

// NewR2Client creates a new R2 client configured for Cloudflare's R2 endpoint
func NewR2Client(accountID, accessKeyID, secretAccessKey, bucketName string, log zerolog.Logger) (*R2Client, error) {
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("r2 credentials incomplete")
	}

	// Create R2-specific endpoint resolver
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
			HostnameImmutable: true,
			SigningRegion:     "auto",
		}, nil
	})

	// Load config with R2 credentials and endpoint
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Create uploader and downloader with custom settings
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10 MB parts
		u.Concurrency = 5
	})

	downloader := manager.NewDownloader(client, func(d *manager.Downloader) {
		d.PartSize = 10 * 1024 * 1024 // 10 MB parts
		d.Concurrency = 5
	})

	return &R2Client{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     bucketName,
		log:        log.With().Str("component", "r2_client").Logger(),
	}, nil
}

// Upload uploads a file to R2
func (r *R2Client) Upload(ctx context.Context, key string, reader io.Reader, contentLength int64) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	r.log.Info().
		Str("key", key).
		Int64("size", contentLength).
		Msg("Starting upload to R2")

	_, err := r.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(contentLength),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to r2: %w", err)
	}

	r.log.Info().
		Str("key", key).
		Msg("Successfully uploaded to R2")

	return nil
}

// Download downloads a file from R2
func (r *R2Client) Download(ctx context.Context, key string, writer io.WriterAt) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	r.log.Info().
		Str("key", key).
		Msg("Starting download from R2")

	bytesDownloaded, err := r.downloader.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to download from r2: %w", err)
	}

	r.log.Info().
		Str("key", key).
		Int64("bytes", bytesDownloaded).
		Msg("Successfully downloaded from R2")

	return bytesDownloaded, nil
}

// List lists all objects in the R2 bucket with the given prefix
func (r *R2Client) List(ctx context.Context, prefix string) ([]types.Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	r.log.Debug().
		Str("prefix", prefix).
		Msg("Listing objects from R2")

	var objects []types.Object

	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list r2 objects: %w", err)
		}
		objects = append(objects, page.Contents...)
	}

	r.log.Debug().
		Int("count", len(objects)).
		Msg("Successfully listed objects from R2")

	return objects, nil
}

// Delete deletes an object from R2
func (r *R2Client) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	r.log.Info().
		Str("key", key).
		Msg("Deleting object from R2")

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from r2: %w", err)
	}

	r.log.Info().
		Str("key", key).
		Msg("Successfully deleted from R2")

	return nil
}

// TestConnection tests the connection to R2 by attempting to head the bucket
func (r *R2Client) TestConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	r.log.Debug().Msg("Testing R2 connection")

	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		return fmt.Errorf("r2 connection test failed: %w", err)
	}

	r.log.Info().Msg("R2 connection test successful")
	return nil
}

// GetObjectMetadata retrieves metadata for an object without downloading it
func (r *R2Client) GetObjectMetadata(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	output, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return output, nil
}
