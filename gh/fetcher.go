package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"repo-pack/helpers"
	"repo-pack/model"
)

type ProgressCallback func(bytesDownloaded int64)

var (
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrFetchError         = errors.New("could not obtain repository data from the GitHub API")
)

type RepoInfo struct {
	Private bool `json:"private"`
}

func FetchRepoIsPrivate(ctx context.Context, components *model.RepoURLComponents, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", components.Owner, components.Repository)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := doRequestWithRetry(ctx, req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return false, fmt.Errorf("repo not found: %s/%s", components.Owner, components.Repository)
	case http.StatusUnauthorized:
		return false, ErrInvalidToken
	case http.StatusForbidden:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return false, ErrRateLimitExceeded
		}
		return false, fmt.Errorf("HTTP 403 Forbidden - check repository access and rate limits")
	case http.StatusOK:
		var repoInfo RepoInfo
		if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
			return false, err
		}
		return repoInfo.Private, nil
	default:
		return false, fmt.Errorf("%w: HTTP %d", ErrFetchError, resp.StatusCode)
	}
}

// isLfsResponse checks if the HTTP response potentially contains a Git LFS response.
// It peeks at the response body without consuming it, resetting it for subsequent reads.
func isLfsResponse(res *http.Response) bool {
	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil || contentLength < 128 || contentLength > 140 {
		return false
	}

	// Peek at the beginning of the response
	bufr := make([]byte, 40)
	n, err := io.ReadFull(res.Body, bufr)
	if err != nil && err != io.ErrUnexpectedEOF {
		return false
	}

	// Read the rest of the body
	restOfBody, err := io.ReadAll(res.Body)
	if err != nil {
		return false
	}

	// Check if this is an LFS response
	isLfs := strings.HasPrefix(string(bufr[:n]), "version https://git-lfs.github.com/spec/v1")

	// Reset the body for the caller to read
	res.Body.Close()
	fullBody := append(bufr, restOfBody...)
	res.Body = io.NopCloser(bytes.NewReader(fullBody))

	return isLfs
}

func FetchPublicFile(ctx context.Context, path string, components *model.RepoURLComponents, outputDir string) error {
	return FetchPublicFileWithProgress(ctx, path, "", components, outputDir, nil)
}

func FetchPublicFileWithProgress(ctx context.Context, path, sha string, components *model.RepoURLComponents, outputDir string, onProgress ProgressCallback) error {
	destPath := filepath.Join(outputDir, filepath.Base(components.Dir), path)

	cache := GetCache()
	if sha != "" && cache != nil {
		if found, err := cache.Get(sha, destPath); found && err == nil {
			return nil
		}
	}

	user := components.Owner
	repository := components.Repository
	ref := components.Ref

	rawURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		user,
		repository,
		ref,
		url.PathEscape(path),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("creating request for %s: %w", path, err)
	}

	resp, err := doRequestWithRetry(ctx, req)
	if err != nil {
		return fmt.Errorf("HTTP error for %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("HTTP %s for %s", resp.Status, path)
	}

	if isLfsResponse(resp) {
		resp.Body.Close()

		lfsURL := fmt.Sprintf(
			"https://media.githubusercontent.com/media/%s/%s/%s/%s",
			user,
			repository,
			ref,
			url.PathEscape(path),
		)
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, lfsURL, nil)
		if err != nil {
			return fmt.Errorf("error creating LFS request for %s: %w", path, err)
		}
		resp, err = doRequestWithRetry(ctx, req)
		if err != nil {
			return fmt.Errorf("HTTP error for LFS %s: %w", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("HTTP %s for LFS %s", resp.Status, path)
		}
	}

	err = helpers.SaveFileWithProgress(filepath.Base(components.Dir), path, resp.Body, outputDir, helpers.ProgressCallback(onProgress))
	if err != nil {
		return fmt.Errorf("error saving file %s: %w", path, err)
	}

	if sha != "" && cache != nil {
		_ = cache.Put(sha, destPath)
	}

	return nil
}

func FetchSingleFile(ctx context.Context, components *model.RepoURLComponents, outputPath string, onProgress ProgressCallback) (int64, error) {
	rawURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		components.Owner,
		components.Repository,
		components.Ref,
		url.PathEscape(components.FilePath),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := doRequestWithRetry(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("HTTP error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return 0, fmt.Errorf("HTTP %s", resp.Status)
	}

	contentLength := resp.ContentLength

	if isLfsResponse(resp) {
		resp.Body.Close()

		lfsURL := fmt.Sprintf(
			"https://media.githubusercontent.com/media/%s/%s/%s/%s",
			components.Owner,
			components.Repository,
			components.Ref,
			url.PathEscape(components.FilePath),
		)
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, lfsURL, nil)
		if err != nil {
			return 0, fmt.Errorf("error creating LFS request: %w", err)
		}
		resp, err = doRequestWithRetry(ctx, req)
		if err != nil {
			return 0, fmt.Errorf("HTTP error for LFS: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return 0, fmt.Errorf("HTTP %s for LFS", resp.Status)
		}
		contentLength = resp.ContentLength
	}

	err = helpers.SaveFileDirect(outputPath, resp.Body, ".", helpers.ProgressCallback(onProgress))
	if err != nil {
		return 0, fmt.Errorf("error saving file: %w", err)
	}

	return contentLength, nil
}

func GetFileSize(ctx context.Context, components *model.RepoURLComponents) (int64, error) {
	rawURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		components.Owner,
		components.Repository,
		components.Ref,
		url.PathEscape(components.FilePath),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := doRequestWithRetry(ctx, req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %s", resp.Status)
	}

	return resp.ContentLength, nil
}
