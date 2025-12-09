package client

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

// ProgressFunc reports cumulative bytes downloaded and the expected total.
type ProgressFunc func(downloaded, total int64)

// DownloadAsset retrieves the asset at assetURL and writes it to destPath.
func (c *Client) DownloadAsset(ctx context.Context, assetURL, destPath string) error {
	return c.DownloadAssetWithProgress(ctx, assetURL, destPath, nil)
}

// DownloadAssetWithProgress downloads an asset while reporting progress.
func (c *Client) DownloadAssetWithProgress(
	ctx context.Context,
	assetURL string,
	destPath string,
	progress ProgressFunc,
) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}

	u, err := url.Parse(assetURL)
	if err != nil {
		return fmt.Errorf("failed to parse asset URL: %w", err)
	}

	if u.Scheme == "" {
		u = c.baseURL.ResolveReference(u)
	}

	switch u.Scheme {
	case "http", "https":
		return c.downloadHTTP(ctx, u.String(), destPath, progress)
	case "s3":
		return downloadS3(ctx, u, destPath, progress)
	default:
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}

func (c *Client) downloadHTTP(ctx context.Context, assetURL string, destPath string, progress ProgressFunc) (err error) {
	resp, err := c.doRequest(ctx, http.MethodGet, assetURL, nil)
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

	var total int64
	if result.ContentLength != nil {
		total = *result.ContentLength
	}

	if progress != nil {
		progress(0, total)
	}

	_, err = copyWithProgress(ctx, out, result.Body, total, progress)
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
