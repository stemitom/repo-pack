package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"repo-pack/gh"
	"repo-pack/helpers"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Parse flags
	repoURL := flag.String("url", "", "GitHub repository URL")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	// Display help if required flags are missing
	if *repoURL == "" {
		flag.Usage()
		return flag.ErrHelp
	}

	// Parse repository URL
	components, err := helpers.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	// Check if repository is private
	ctx := context.Background()
	gh.FetchRepoIsPrivate(ctx, &components, *token)

	// Fetch files from repository
	files, _, err := gh.RepoListingSlashBranchSupport(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents API: %v", err)
	}

	// Display information about the repository
	baseDir := filepath.Base(components.Dir)
	fmt.Printf("[-] Repository: %s/%s\n", components.Owner, components.Repository)
	fmt.Printf("[-] GitHub Directory: %s\n", components.Dir)
	fmt.Printf("[-] Fetching %d files\n", len(files))

	// Concurrently download files
	var wg sync.WaitGroup
	errorsCh := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			content, err := gh.FetchPublicFile(ctx, file, &components)
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file, err)
				return
			}

			if err := helpers.SaveFile(baseDir, file, content); err != nil {
				errorsCh <- err
				return
			}
		}(file)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	// Log any errors encountered during downloads
	for err := range errorsCh {
		log.Println(err)
	}

	fmt.Printf("[-] Successfully downloaded %d files\n", len(files))
	return nil
}
