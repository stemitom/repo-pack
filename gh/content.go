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

	"repo-pack/model"
)

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

// API makes a GET request to the GitHub API with the given endpoint and optional authentication token.
// It returns the response body as a byte slice or an error if the request fails.
func API(ctx context.Context, endpoint, token string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// ViaContentsAPI retrieves a list of files in a GitHub repository directory using the Contents API.
// It handles both files and subdirectories recursively.
func ViaContentsAPI(ctx context.Context, urlComponents model.RepoURLComponents, token string) ([]string, error) {
	files := []string{}
	contents, err := API(
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

	var items []Item
	err = json.Unmarshal(contents, &items)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		switch item.Type {
		case "file":
			files = append(files, item.Path)
		case "dir":
			subFiles, err := ViaContentsAPI(ctx, urlComponents, token)
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

// ViaTreesAPI retrieves a list of files in a GitHub repository directory using the Git Trees API.
// It handles both files and subdirectories recursively, and indicates if the response was truncated.
func ViaTreesAPI(
	ctx context.Context,
	urlComponents model.RepoURLComponents,
	token string,
) (files []string, truncated bool, err error) {
	if !strings.HasSuffix(urlComponents.Dir, "/") {
		urlComponents.Dir += "/"
	}

	files = []string{}
	contents, err := API(
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

	var treeResponse TreeResponse
	err = json.Unmarshal(contents, &treeResponse)
	if err != nil {
		return nil, false, err
	}

	for _, item := range treeResponse.Tree {
		if item.Type == "blob" && strings.HasPrefix(item.Path, urlComponents.Dir) {
			files = append(files, item.Path)
		}
	}

	truncated = treeResponse.Truncated

	return files, truncated, nil
}

// RepoListingSlashBranchSupport fetches repository listing recursively.
// It uses the provided context, repository components, and token for authentication.
// It returns the list of files, the final reference, and an error (if any).
func RepoListingSlashBranchSupport(ctx context.Context, components *model.RepoURLComponents, token string) ([]string, string, error) {
	var files []string
	var isTruncated bool

	ref := components.Ref
	dir := components.Dir

	decodedDir, err := url.QueryUnescape(dir)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding: %s", dir)
	}

	dirParts := strings.Split(decodedDir, "/")

	for len(dirParts) > 0 {
		content, truncated, err := ViaTreesAPI(ctx, *components, token)
		if err == nil {
			files = content
			isTruncated = truncated
			break
		} else if errors.Is(err, ErrNotFound) {
			ref = path.Join(ref, dirParts[0])
			dirParts = dirParts[1:]
			components.Dir = strings.Join(dirParts, "/")
		} else {
			return nil, "", err
		}
	}

	if len(files) == 0 && isTruncated {
		files, err := ViaContentsAPI(ctx, *components, token)
		if err != nil {
			return nil, "", err
		}
		return files, ref, nil
	}

	return files, ref, nil
}
