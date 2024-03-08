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

	"github.com/cheggaaa/pb/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	repoURL := flag.String("url", "", "GitHub repository URL")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	if *repoURL == "" {
		flag.Usage()
		return flag.ErrHelp
	}

	components, err := helpers.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	ctx := context.Background()
	gh.FetchRepoIsPrivate(ctx, &components, *token)

	files, _, err := gh.RepoListingSlashBranchSupport(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents API: %v", err)
	}

	baseDir := filepath.Base(components.Dir)
	fmt.Printf("[-] Repository: %s/%s\n", components.Owner, components.Repository)
	fmt.Printf("[-] GitHub Directory: %s\n", components.Dir)
	fmt.Printf("[-] Fetching %d files\n", len(files))

	bar := pb.Full.Start64(0)
	defer bar.Finish()

	var totalDownloaded int64
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

			totalDownloaded += int64(len(content))
			bar.SetTotal(totalDownloaded)
			if err := helpers.SaveFile(baseDir, file, content); err != nil {
				errorsCh <- err
				return
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	for err := range errorsCh {
		log.Println(err)
	}

	fmt.Printf("[-] Successfully downloaded %d files\n", len(files))
	return nil
}
