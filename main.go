package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

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
	dryRun := flag.Bool("dry-run", false, "Preview files without downloading")
	outputDir := flag.String("output", ".", "Output directory for downloaded files")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	resume := flag.Bool("resume", false, "Skip files that already exist locally")

	flag.Parse()

	if *repoURL == "" {
		return fmt.Errorf("missing required argument --url\nUsage: repo-pack --url <github_url>\nExample: repo-pack --url https://github.com/owner/repo/tree/main/path/to/directory")
	}

	if *verbose && *quiet {
		return fmt.Errorf("cannot use both --verbose and --quiet flags")
	}

	if *limit <= 0 {
		return fmt.Errorf("concurrent download limit must be greater than 0, got: %d", *limit)
	}
	if *limit > 100 {
		log.Printf("Warning: high concurrent download limit (%d) may cause rate limiting or system issues\n", *limit)
	}
	if *style == "" {
		return fmt.Errorf("progress bar style cannot be empty")
	}
	if len([]rune(*style)) > 1 {
		return fmt.Errorf("progress bar style must be a single character, got: %s", *style)
	}

	cfg.ConcurrentDownloadLimit = *limit
	cfg.ProgressBarStyle = *style

	if *token == "" {
		tokenBytes, err := os.ReadFile(cfg.GithubTokenPath)
		if err == nil {
			*token = string(tokenBytes)
		} else if !os.IsNotExist(err) {
			log.Printf("Warning: token file exists at %s but could not be read: %v\n", cfg.GithubTokenPath, err)
		}
	}

	components, err := helpers.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	// Validate and create output directory if needed
	if *outputDir != "." {
		if err := os.MkdirAll(*outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %v", *outputDir, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nCancelling...")
		cancel()
	}()

	if _, privErr := gh.FetchRepoIsPrivate(ctx, &components, *token); privErr != nil && !*quiet {
		log.Printf("Warning: could not verify repository privacy: %v\n", privErr)
	}

	files, err := gh.RepoListingSlashBranchSupport(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents API: %v", err)
	}

	if !*quiet {
		fmt.Printf("Repository: %s/%s\n", components.Owner, components.Repository)
		if components.Dir != "" {
			fmt.Printf("Directory: %s\n", components.Dir)
		}
		fmt.Printf("Output: %s\n", *outputDir)
		fmt.Printf("Files found: %d\n", len(files))
		if *verbose {
			fmt.Printf("Branch/Ref: %s\n", components.Ref)
		}
	}

	if *dryRun {
		if !*quiet {
			fmt.Printf("\nDry run - %d file(s) ready\n", len(files))
			if *verbose {
				fmt.Println("Files:")
				for i, file := range files {
					fmt.Printf("  %d. %s\n", i+1, file)
				}
			}
		}
		return nil
	}

	var bar *helpers.Bar
	if !*quiet {
		bar = &helpers.Bar{}
		bar.Config(0, int64(len(files)), "Downloading ")
		bar.SetStyle(cfg.ProgressBarStyle)
	}

	downloadStartTime := time.Now()

	var wg sync.WaitGroup
	errorsCh := make(chan error, len(files))
	skippedCh := make(chan string, len(files))
	sem := make(chan struct{}, cfg.ConcurrentDownloadLimit)

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			if *resume {
				exists, err := helpers.FileExists(filepath.Base(components.Dir), file, *outputDir)
				if err != nil {
					if *verbose {
						log.Printf("Warning: could not check if file exists %s: %v\n", file, err)
					}
				} else if exists {
					if *verbose {
						fmt.Printf("[-] Skipping (already exists): %s\n", file)
					}
					skippedCh <- file
					if !*quiet && bar != nil {
						bar.Increment()
					}
					return
				}
			}

			if *verbose {
				fmt.Printf("[-] Downloading: %s\n", file)
			}

			err := gh.FetchPublicFile(ctx, file, &components, *outputDir)
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file, err)
				return
			}
			if !*quiet && bar != nil {
				bar.Increment()
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(errorsCh)
		close(skippedCh)
		if !*quiet && bar != nil {
			bar.Finish()
		}
	}()

	var downloadErrors []error
	var skippedFiles []string

	for err := range errorsCh {
		log.Println(err)
		downloadErrors = append(downloadErrors, err)
	}

	for file := range skippedCh {
		skippedFiles = append(skippedFiles, file)
	}

	if !*quiet {
		downloadDuration := time.Since(downloadStartTime)
		downloadedCount := len(files) - len(downloadErrors) - len(skippedFiles)
		fmt.Printf("\n%d/%d downloaded", downloadedCount, len(files))
		if len(skippedFiles) > 0 {
			fmt.Printf(", %d skipped", len(skippedFiles))
		}
		if len(downloadErrors) > 0 {
			fmt.Printf(", %d failed", len(downloadErrors))
		}
		fmt.Printf(" [%s]\n", downloadDuration.Round(time.Millisecond))
	}

	// Check if there were errors or cancellation
	if len(downloadErrors) > 0 {
		if ctx.Err() != nil {
			return fmt.Errorf("download interrupted by user with %d error(s)", len(downloadErrors))
		}
		return fmt.Errorf("failed to download %d file(s)", len(downloadErrors))
	}

	if ctx.Err() != nil {
		return fmt.Errorf("download cancelled by user")
	}

	return nil
}
