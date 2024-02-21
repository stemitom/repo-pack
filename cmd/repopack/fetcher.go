package repopack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrInvalidToken       = errors.New("invalid token")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrFetchError         = errors.New("could not obtain repository data from the github api")
)

type RepoInfo struct {
	Private bool `json:"private"`
}

// FetchRepoInfo fetches repository information from GitHub API
func FetchRepoInfo(ctx context.Context, owner string, repository string, token string) (*RepoInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repository)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401:
		return nil, fmt.Errorf("invalid token: %s", token)
	case 403:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return nil, ErrRateLimitExceeded
		}
	case 404:
		return nil, ErrRepositoryNotFound
	default:
	}

	var repoInfo RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}

	return &repoInfo, nil
}
