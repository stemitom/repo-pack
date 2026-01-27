package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"repo-pack/model"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          200,
		MaxConnsPerHost:       50,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false,
		ForceAttemptHTTP2:     true,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

type Item struct {
	Type string `json:"type"`
	Path string `json:"path"`
	URL  string `json:"url,omitempty"`
	SHA  string `json:"sha,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type TreeResponse struct {
	SHA       *string `json:"sha,omitempty"`
	Tree      []Item  `json:"tree"`
	Truncated bool    `json:"truncated"`
}

var ErrNotFound = errors.New("not found")

func apiRequest(ctx context.Context, endpoint, token string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		if resp.StatusCode == 403 {
			rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
			if rateLimitRemaining == "0" {
				resetTime := resp.Header.Get("X-RateLimit-Reset")
				return nil, fmt.Errorf("GitHub API rate limit exceeded (resets at %s). Try using a personal access token with --token", resetTime)
			}
			return nil, fmt.Errorf("HTTP 403 Forbidden - you may need a personal access token for private repositories (use --token)")
		}
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Retry after %s seconds", retryAfter)
		}
		if resp.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// API makes a GET request to the GitHub API with the given endpoint and optional authentication token.
// It returns the response body as a byte slice or an error if the request fails.
func API(ctx context.Context, endpoint, token string) ([]byte, error) {
	resp, err := apiRequest(ctx, endpoint, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ViaContentsAPI(ctx context.Context, urlComponents model.RepoURLComponents, token string) ([]model.FileInfo, error) {
	var files []model.FileInfo
	resp, err := apiRequest(
		ctx,
		fmt.Sprintf(
			"%s/%s/contents/%s?ref=%s",
			urlComponents.Owner,
			urlComponents.Repository,
			urlComponents.Dir,
			urlComponents.Ref,
		),
		token,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var items []Item
	if err = json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	for _, item := range items {
		switch item.Type {
		case "file":
			files = append(files, model.FileInfo{
				Path: item.Path,
				Size: item.Size,
				SHA:  item.SHA,
			})
		case "dir":
			newComponents := urlComponents
			newComponents.Dir = item.Path
			subFiles, err := ViaContentsAPI(ctx, newComponents, token)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		default:
			return nil, fmt.Errorf("ignoring item with unknown type: %s", item.Type)
		}
	}

	return files, nil
}

func ViaTreesAPI(
	ctx context.Context,
	urlComponents model.RepoURLComponents,
	token string,
) (files []model.FileInfo, truncated bool, err error) {
	if !strings.HasSuffix(urlComponents.Dir, "/") {
		urlComponents.Dir += "/"
	}

	resp, err := apiRequest(
		ctx,
		fmt.Sprintf(
			"%s/%s/git/trees/%s?recursive=1",
			urlComponents.Owner,
			urlComponents.Repository,
			urlComponents.Ref,
		),
		token,
	)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	var treeResponse TreeResponse
	if err = json.NewDecoder(resp.Body).Decode(&treeResponse); err != nil {
		return nil, false, fmt.Errorf("failed to parse tree response: %w", err)
	}

	for _, item := range treeResponse.Tree {
		if item.Type == "blob" && strings.HasPrefix(item.Path, urlComponents.Dir) {
			files = append(files, model.FileInfo{
				Path: item.Path,
				Size: item.Size,
				SHA:  item.SHA,
			})
		}
	}

	truncated = treeResponse.Truncated

	return files, truncated, nil
}

func RepoListingSlashBranchSupport(ctx context.Context, components *model.RepoURLComponents, token string) ([]model.FileInfo, error) {
	var files []model.FileInfo
	var isTruncated bool

	dir := components.Dir

	decodedDir, err := url.QueryUnescape(dir)
	if err != nil {
		return nil, fmt.Errorf("error decoding: %s", dir)
	}

	dirParts := strings.Split(decodedDir, "/")

	for len(dirParts) > 0 {
		content, truncated, err := ViaTreesAPI(ctx, *components, token)
		if err == nil {
			files = content
			isTruncated = truncated
			break
		} else if errors.Is(err, ErrNotFound) {
			components.Ref = path.Join(components.Ref, dirParts[0])
			dirParts = dirParts[1:]
			components.Dir = strings.Join(dirParts, "/")
		} else {
			return nil, err
		}
	}

	if len(files) == 0 && isTruncated {
		files, err := ViaContentsAPI(ctx, *components, token)
		if err != nil {
			return nil, err
		}
		return files, nil
	}

	return files, nil
}
