package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ProgressFunc func(downloaded, total int64)

func Download(ctx context.Context, assetURL string, destPath string) error {
	return DownloadWithProgress(ctx, assetURL, destPath, nil)
}

func DownloadWithProgress(ctx context.Context, assetURL string, destPath string, progress ProgressFunc) error {
	u, err := url.Parse(assetURL)
	if err != nil {
		return fmt.Errorf("failed to parse asset URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		return downloadHTTP(ctx, assetURL, destPath, progress)
	case "s3":
		return downloadS3(ctx, u, destPath, progress)
	default:
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}

func downloadHTTP(ctx context.Context, assetURL string, destPath string, progress ProgressFunc) (err error) {
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
	defer func() {
		out.Close()
		if err != nil {
			_ = os.Remove(destPath)
		}
	}()

	total := resp.ContentLength
	if progress != nil {
		progress(0, total)
	}

	_, err = copyWithProgress(ctx, out, resp.Body, total, progress)
	if err != nil {
		return fmt.Errorf("failed to write asset to file: %w", err)
	}

	return nil
}

func downloadS3(ctx context.Context, u *url.URL, destPath string, progress ProgressFunc) (err error) {
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
	defer func() {
		out.Close()
		if err != nil {
			_ = os.Remove(destPath)
		}
	}()

	if progress != nil {
		progress(0, *result.ContentLength)
	}

	_, err = copyWithProgress(ctx, out, result.Body, *result.ContentLength, progress)
	if err != nil {
		return fmt.Errorf("failed to write asset to file: %w", err)
	}

	return nil
}

func copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader, total int64, progress ProgressFunc) (int64, error) {
	const defaultBufferSize = 32 * 1024
	buf := make([]byte, defaultBufferSize)
	var written int64

	for {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return written, err
			}
		}

		n, readErr := src.Read(buf)
		if n > 0 {
			w, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return written, writeErr
			}
			if w != n {
				return written, io.ErrShortWrite
			}
			written += int64(w)
			if progress != nil {
				progress(written, total)
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return written, nil
			}
			return written, readErr
		}
	}
}
