package repopack

import (
	"fmt"
	"net/url"
	"regexp"
)

type RepoURLComponents struct {
	Owner      string
	Repository string
	Ref        string
	Dir        string
}

// ParseRepoURL extracts user, repository, ref, and dir from the URL
func ParseRepoURL(urlStr string) (urlComponents RepoURLComponents, err error) {
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

	urlComponents = RepoURLComponents{
		Owner:      owner,
		Repository: repository,
		Ref:        ref,
		Dir:        dir,
	}
	return urlComponents, nil
}
