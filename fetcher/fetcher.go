package fetcher

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"repo-pack/model"
)

var (
	// ErrInvalidToken       = errors.New("invalid token")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrFetchError         = errors.New("could not obtain repository data from the github api")
)

type FileInfo struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

type RepoInfo struct {
	Private bool `json:"private"`
}

// FetchRepoIsPrivate checks if a repository is private or not on GitHub.
func FetchRepoIsPrivate(ctx context.Context, components *model.RepoURLComponents, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", components.Owner, components.Repository)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return false, fmt.Errorf("repo not found: %s/%s", components.Owner, components.Repository)
	case http.StatusUnauthorized:
		return false, fmt.Errorf("invalid token: %s", token)
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

// maybeLfsResponse checks if the HTTP response potentially contains a Git LFS response.
// It inspects the Content-Length header and the beginning of the response body to make a determination.
// If the Content-Length is between 128 and 140 bytes (inclusive) and the response body starts with
// "version https://git-lfs.github.com/spec/v1", it returns true.
func maybeLfsResponse(res *http.Response) bool {
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

// fetchPrivateFile downloads a file from a public GitHub repository
func fetchPrivateFile(ctx context.Context, file FileInfo, token string) (io.Reader, error) {
	url := file.URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download error: %s", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("unexpected content type")
	}

	var data struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	contentReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data.Content))

	return contentReader, nil
}

// fetchPublicFile downloads a file from a public GitHub repository, handling Git LFS if necessary.
func FetchPublicFile(ctx context.Context, path string, components *model.RepoURLComponents) ([]byte, error) {
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

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", path, err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP error for %s: %w", path, err)
	}
	defer resp.Body.Close()

	if maybeLfsResponse(resp) {
		lfsURL := fmt.Sprintf(
			"https://media.githubusercontent.com/media/%s/%s/%s/%s",
			user,
			repository,
			ref,
			url.PathEscape(path),
		)
		req, err = http.NewRequestWithContext(ctx, "GET", lfsURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating LFS request for %s: %w", path, err)
		}
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP error for LFS %s: %w", path, err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s for %s", resp.Status, path)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body for %s: %w", path, err)
	}

	return content, nil
}
