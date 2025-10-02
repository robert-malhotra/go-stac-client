package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Download(ctx context.Context, assetURL string, destPath string) error {
	u, err := url.Parse(assetURL)
	if err != nil {
		return fmt.Errorf("failed to parse asset URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		return downloadHTTP(ctx, assetURL, destPath)
	case "s3":
		return downloadS3(ctx, u, destPath)
	default:
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}

func downloadHTTP(ctx context.Context, assetURL string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: unexpected status code %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write asset to file: %w", err)
	}

	return nil
}

func downloadS3(ctx context.Context, u *url.URL, destPath string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	bucket := u.Host
	key := strings.TrimPrefix(u.Path, "/")

	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write asset to file: %w", err)
	}

	return nil
}
