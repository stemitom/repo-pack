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

// Error constants
var (
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrFetchError         = errors.New("could not obtain repository data from the GitHub API")
)

// RepoInfo represents information about a repository
type RepoInfo struct {
	Private bool `json:"private"`
}

// FetchRepoIsPrivate checks if a repository is private or not on GitHub.
func FetchRepoIsPrivate(ctx context.Context, components *model.RepoURLComponents, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", components.Owner, components.Repository)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
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
	case http.StatusOK:
		var repoInfo RepoInfo
		if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
			return false, err
		}
		return repoInfo.Private, nil
	default:
		return false, ErrFetchError
	}

	return false, nil
}

// isLfsResponse checks if the HTTP response potentially contains a Git LFS response.
func isLfsResponse(res *http.Response) bool {
	if contentLength, err := strconv.Atoi(res.Header.Get("Content-Length")); err == nil && 128 < contentLength &&
		contentLength < 140 {
		bufr := make([]byte, 40)
		_, err := io.ReadFull(res.Body, bufr)
		if err != nil {
			return false
		}

		restOfBody, err := io.ReadAll(res.Body)
		if err != nil {
			return false
		}

		res.Body.Close()
		fullBody := append(bufr, restOfBody...)
		res.Body = io.NopCloser(bytes.NewReader(fullBody))

		return strings.HasPrefix(string(bufr), "version https://git-lfs.github.com/spec/v1")
	}
	return false
}

// FetchPublicFile downloads a file from a public GitHub repository, handling Git LFS if necessary and saves it.
func FetchPublicFile(ctx context.Context, path string, components *model.RepoURLComponents) error {
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP error for %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("HTTP %s for %s", resp.Status, path)
	}

	if isLfsResponse(resp) {
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
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("HTTP error for LFS %s: %w", path, err)
		}
	}

	err = helpers.SaveFile(filepath.Base(components.Dir), path, resp.Body)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("error saving file %s %v", path, err)
	}

	return nil
}

