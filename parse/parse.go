package parse

import (
	"fmt"
	"net/url"
	"regexp"

	"repo-pack/model"
)

// ParseRepoURL validates that URL is valid and then extracts user, repository, ref, and directory
func ParseRepoURL(urlStr string) (urlComponents model.RepoURLComponents, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		err = fmt.Errorf("invalid URL: %s", urlStr)
		return
	}

	urlPath := parsedURL.Path
	urlParserRegex := regexp.MustCompile(`^/([^/]+)/([^/]+)/tree/([^/]+)/(.*)`)
	match := urlParserRegex.FindStringSubmatch(urlPath)

	if len(match) != 5 {
		err = fmt.Errorf("invalid URL format: %s", urlStr)
		return
	}

	owner := match[1]
	repository := match[2]
	ref := match[3]
	dir := match[4]

	urlComponents = model.RepoURLComponents{
		Owner:      owner,
		Repository: repository,
		Ref:        ref,
		Dir:        dir,
	}
	return urlComponents, nil
}
