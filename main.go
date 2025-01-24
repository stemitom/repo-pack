package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"repo-pack/config"
	"repo-pack/gh"
	"repo-pack/helpers"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	repoURL := flag.String("url", "", "GitHub repository URL")
	token := flag.String("token", "", "GitHub personal access token")
	limit := flag.Int("limit", cfg.ConcurrentDownloadLimit, "Concurrent download limit")
	style := flag.String("style", cfg.ProgressBarStyle, "Progress bar style")
	flag.Parse()

	if *repoURL == "" {
		return fmt.Errorf("missing argument for repoURL")
	}

	cfg.ConcurrentDownloadLimit = *limit
	cfg.ProgressBarStyle = *style

	if *token == "" {
		tokenBytes, err := os.ReadFile(cfg.GithubTokenPath)
		if err == nil {
			*token = string(tokenBytes)
		}
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

	fmt.Printf("[-] Repository: %s/%s\n", components.Owner, components.Repository)
	fmt.Printf("[-] GitHub Directory: %s\n", components.Dir)
	fmt.Printf("[-] Fetching %d files\n", len(files))

	bar := &helpers.Bar{}
	bar.Config(0, int64(len(files)), "[-] Progress: ")
	bar.SetStyle(cfg.ProgressBarStyle)

	var wg sync.WaitGroup
	errorsCh := make(chan error, len(files))
	sem := make(chan struct{}, cfg.ConcurrentDownloadLimit)

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			err := gh.FetchPublicFile(ctx, file, &components)
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file, err)
				return
			}
			bar.Update(bar.Cur + 1)
		}(file)
	}

	go func() {
		wg.Wait()
		close(errorsCh)
		bar.Finish()
	}()

	for err := range errorsCh {
		log.Println(err)
	}

	return nil
}
