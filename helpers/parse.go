package helpers

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"repo-pack/model"
)

var (
	// /owner/repo/tree/ref/path - directory URL
	treeRegex = regexp.MustCompile(`^/([^/]+)/([^/]+)/tree/([^/]+)/(.*)`)
	// /owner/repo/blob/ref/path - single file URL
	blobRegex = regexp.MustCompile(`^/([^/]+)/([^/]+)/blob/([^/]+)/(.+)`)
	// /owner/repo/ref/path - raw.githubusercontent.com URL format
	rawRegex = regexp.MustCompile(`^/([^/]+)/([^/]+)/([^/]+)/(.+)`)
)

func ParseRepoURL(urlStr string) (urlComponents model.RepoURLComponents, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		err = fmt.Errorf("invalid URL: %s", urlStr)
		return
	}

	host := strings.ToLower(parsedURL.Host)
	urlPath := parsedURL.Path

	switch host {
	case "raw.githubusercontent.com":
		return parseRawURL(urlPath, urlStr)
	case "github.com", "www.github.com":
		return parseGitHubURL(urlPath, urlStr)
	default:
		err = fmt.Errorf("unsupported host: %s\nSupported: github.com, raw.githubusercontent.com", host)
		return
	}
}

func parseGitHubURL(urlPath, originalURL string) (model.RepoURLComponents, error) {
	if match := blobRegex.FindStringSubmatch(urlPath); len(match) == 5 {
		decodedPath, err := url.QueryUnescape(match[4])
		if err != nil {
			decodedPath = match[4]
		}
		return model.RepoURLComponents{
			Owner:      match[1],
			Repository: match[2],
			Ref:        match[3],
			Dir:        path.Dir(decodedPath),
			FilePath:   decodedPath,
			IsFile:     true,
		}, nil
	}

	if match := treeRegex.FindStringSubmatch(urlPath); len(match) == 5 {
		decodedDir, err := url.QueryUnescape(match[4])
		if err != nil {
			decodedDir = match[4]
		}
		return model.RepoURLComponents{
			Owner:      match[1],
			Repository: match[2],
			Ref:        match[3],
			Dir:        decodedDir,
			IsFile:     false,
		}, nil
	}

	return model.RepoURLComponents{}, fmt.Errorf(
		"invalid GitHub URL format: %s\nExpected formats:\n"+
			"  Directory: https://github.com/owner/repo/tree/branch/path/to/dir\n"+
			"  File:      https://github.com/owner/repo/blob/branch/path/to/file.ext",
		originalURL,
	)
}

func parseRawURL(urlPath, originalURL string) (model.RepoURLComponents, error) {
	match := rawRegex.FindStringSubmatch(urlPath)
	if len(match) != 5 {
		return model.RepoURLComponents{}, fmt.Errorf(
			"invalid raw URL format: %s\nExpected: https://raw.githubusercontent.com/owner/repo/ref/path/to/file",
			originalURL,
		)
	}

	decodedPath, err := url.QueryUnescape(match[4])
	if err != nil {
		decodedPath = match[4]
	}

	return model.RepoURLComponents{
		Owner:      match[1],
		Repository: match[2],
		Ref:        match[3],
		Dir:        path.Dir(decodedPath),
		FilePath:   decodedPath,
		IsFile:     true,
	}, nil
}
