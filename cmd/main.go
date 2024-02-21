package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"repo-pack/cmd/repopack"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Set up HTTP client
	client := &http.Client{}

	// Set repository URL (you need to pass this as an argument or from some configuration)
	repoURL := "https://github.com/user/repo/tree/ref/dir"

	// Extract user, repository, ref, and dir from the URL
	user, repository, ref, dir, err := parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	// Fetch repository information
	repoInfo, err := repopack.FetchRepoInfo(ctx, user, repository)
	if err != nil {
		return err
	}

	// Check if repository is private
	repoIsPrivate := repoInfo.Private

	// Set repository listing configuration
	repoListingConfig := repopack.APIArgs{
		User:       user,
		Repository: repository,
		Ref:        ref,
		Directory:  dir,
		Token:      os.Getenv("GITHUB_TOKEN"), // Set your GitHub token as an environment variable
	}

	// Fetch files
	files, ref, err := repopack.FetchFiles(ctx, repoListingConfig, client)
	if err != nil {
		return err
	}

	// Handle files based on repo privacy
	if repoIsPrivate {
		// Handle private repository
	} else {
		// Handle public repository
	}

	// Download files
	if err := downloadFiles(files, ref, client); err != nil {
		return err
	}

	fmt.Println("Download successful")
	return nil
}

func parseRepoURL(repoURL string) (user, repository, ref, dir string, err error) {
	// Extract user, repository, ref, and dir from the URL
	parts := strings.Split(repoURL, "/")
	if len(parts) < 6 {
		err = fmt.Errorf("invalid repository URL: %s", repoURL)
		return
	}

	user = parts[3]
	repository = parts[4]
	ref = parts[6]
	dir = strings.Join(parts[7:], "/")

	return
}

func downloadFiles(files []string, ref string, client *http.Client) error {
	// Implement file download logic here
	return nil
}
