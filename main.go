package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"repo-pack/config"
	"repo-pack/gh"
	"repo-pack/helpers"
	"repo-pack/model"
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

	repoURL := flag.String("url", "", "GitHub repository URL (file or directory)")
	token := flag.String("token", "", "GitHub personal access token")
	limit := flag.Int("limit", cfg.ConcurrentDownloadLimit, "Concurrent download limit")
	style := flag.String("style", cfg.ProgressBarStyle, "Progress bar style")
	dryRun := flag.Bool("dry-run", false, "Preview files without downloading")
	outputDir := flag.String("output", ".", "Output directory for downloaded files")
	outputFile := flag.String("output-file", "", "Custom filename for single file downloads")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	resume := flag.Bool("resume", false, "Skip files that already exist locally")
	noColor := flag.Bool("no-color", false, "Disable colored output")

	flag.Parse()

	if *repoURL == "" {
		return fmt.Errorf("missing required argument --url\nUsage: repo-pack --url <github_url>\nExamples:\n  Directory: repo-pack --url https://github.com/owner/repo/tree/main/path/to/dir\n  File:      repo-pack --url https://github.com/owner/repo/blob/main/path/to/file.txt")
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

	if *noColor {
		helpers.SetColorEnabled(false)
	}

	cfg.ConcurrentDownloadLimit = *limit
	cfg.ProgressBarStyle = *style

	if *token == "" {
		tokenBytes, err := os.ReadFile(cfg.GithubTokenPath)
		if err == nil {
			*token = strings.TrimSpace(string(tokenBytes))
		} else if !os.IsNotExist(err) {
			log.Printf("Warning: token file exists at %s but could not be read: %v\n", cfg.GithubTokenPath, err)
		}
	}

	components, err := helpers.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	if *outputFile != "" && !components.IsFile {
		return fmt.Errorf("--output-file can only be used with single file URLs (/blob/ or raw URLs)")
	}

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
		cancel()
	}()

	if _, privErr := gh.FetchRepoIsPrivate(ctx, &components, *token); privErr != nil && !*quiet {
		log.Printf("Warning: could not verify repository privacy: %v\n", privErr)
	}

	if components.IsFile {
		return downloadSingleFile(ctx, &components, *outputDir, *outputFile, *dryRun, *resume, *quiet, *verbose)
	}

	return downloadDirectory(ctx, &components, *token, *outputDir, cfg, *dryRun, *resume, *quiet, *verbose)
}

func downloadSingleFile(ctx context.Context, components *model.RepoURLComponents, outputDir, outputFile string, dryRun, resume, quiet, verbose bool) error {
	filename := filepath.Base(components.FilePath)
	if outputFile != "" {
		filename = outputFile
	}

	outputPath := filepath.Join(outputDir, filename)

	if !quiet {
		fmt.Printf("Repository: %s/%s\n", components.Owner, components.Repository)
		fmt.Printf("File:       %s", components.FilePath)
		if outputFile != "" {
			fmt.Printf(" → %s", filename)
		}
		fmt.Println()
		fmt.Printf("Output:     %s\n", outputPath)
		if verbose {
			fmt.Printf("Branch/Ref: %s\n", components.Ref)
		}
	}

	if dryRun {
		if !quiet {
			fmt.Printf("\nDry run - 1 file ready: %s\n", components.FilePath)
		}
		return nil
	}

	if resume {
		exists, err := helpers.FileExistsDirect(filename, outputDir)
		if err != nil && verbose {
			log.Printf("Warning: could not check if file exists: %v\n", err)
		} else if exists {
			if !quiet {
				fmt.Printf("\n%s Skipped (already exists): %s\n", helpers.Colorize("⏭", helpers.Yellow), filename)
			}
			return nil
		}
	}

	fileSize, err := gh.GetFileSize(ctx, components)
	if err != nil && verbose {
		log.Printf("Warning: could not get file size: %v\n", err)
	}

	var progress *helpers.SingleFileProgress
	if !quiet {
		fmt.Println()
		progress = helpers.NewSingleFileProgress(filename, fileSize)
	}

	downloadStart := time.Now()

	finalSize, err := gh.FetchSingleFile(ctx, components, outputPath, func(downloaded int64) {
		if progress != nil {
			progress.Update(downloaded)
		}
	})
	if err != nil {
		if ctx.Err() != nil {
			fmt.Printf("\n%s Download cancelled\n", helpers.Colorize("✗", helpers.Red))
			os.Exit(1)
		}
		return fmt.Errorf("failed to download file: %v", err)
	}

	if progress != nil {
		progress.Update(finalSize)
		progress.Finish()
	} else if !quiet {
		fmt.Printf("Downloaded %s (%s) in %s\n", filename, helpers.FormatBytes(finalSize), time.Since(downloadStart).Round(time.Millisecond))
	}

	return nil
}

func downloadDirectory(ctx context.Context, components *model.RepoURLComponents, token, outputDir string, cfg config.Config, dryRun, resume, quiet, verbose bool) error {
	files, err := gh.RepoListingSlashBranchSupport(ctx, components, token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents API: %v", err)
	}

	var totalBytes int64
	for _, f := range files {
		totalBytes += f.Size
	}

	if !quiet {
		fmt.Printf("Repository: %s/%s\n", components.Owner, components.Repository)
		if components.Dir != "" {
			fmt.Printf("Directory:  %s\n", components.Dir)
		}
		fmt.Printf("Output:     %s\n", outputDir)
		fmt.Printf("Files:      %d", len(files))
		if totalBytes > 0 {
			fmt.Printf(" (%s)", helpers.FormatBytes(totalBytes))
		}
		fmt.Println()
		if verbose {
			fmt.Printf("Branch/Ref: %s\n", components.Ref)
		}
	}

	if dryRun {
		if !quiet {
			fmt.Printf("\nDry run - %d file(s) ready\n", len(files))
			if verbose {
				fmt.Println("Files:")
				for i, file := range files {
					sizeStr := ""
					if file.Size > 0 {
						sizeStr = fmt.Sprintf(" (%s)", helpers.FormatBytes(file.Size))
					}
					fmt.Printf("  %d. %s%s\n", i+1, file.Path, sizeStr)
				}
			}
		}
		return nil
	}

	var tracker *helpers.ProgressTracker
	if !quiet {
		fmt.Println()
		tracker = helpers.NewProgressTracker(int64(len(files)), totalBytes)
		tracker.SetStyle(cfg.ProgressBarStyle)
	}

	downloadStartTime := time.Now()

	var wg sync.WaitGroup
	errorsCh := make(chan error, len(files))
	skippedCh := make(chan string, len(files))
	sem := make(chan struct{}, cfg.ConcurrentDownloadLimit)

	for _, file := range files {
		wg.Add(1)
		go func(file model.FileInfo) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			if resume {
				exists, err := helpers.FileExists(filepath.Base(components.Dir), file.Path, outputDir)
				if err != nil {
					if verbose {
						log.Printf("Warning: could not check if file exists %s: %v\n", file.Path, err)
					}
				} else if exists {
					if verbose {
						fmt.Printf("[-] Skipping (already exists): %s\n", file.Path)
					}
					skippedCh <- file.Path
					if tracker != nil {
						tracker.SkipFile(file.Path)
					}
					return
				}
			}

			if tracker != nil {
				tracker.StartFile(file.Path, file.Size)
			}

			err := gh.FetchPublicFileWithProgress(ctx, file.Path, file.SHA, components, outputDir, func(downloaded int64) {
				if tracker != nil {
					tracker.UpdateFileProgress(file.Path, downloaded)
				}
			})
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file.Path, err)
				if tracker != nil {
					tracker.FailFile(file.Path, err)
				}
				return
			}

			if tracker != nil {
				tracker.CompleteFile(file.Path, file.Size)
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(errorsCh)
		close(skippedCh)
	}()

	var downloadErrors []error
	var skippedFiles []string
	wasCancelled := false

	for err := range errorsCh {
		if ctx.Err() != nil {
			wasCancelled = true
		} else if verbose {
			log.Println(err)
		}
		downloadErrors = append(downloadErrors, err)
	}

	for file := range skippedCh {
		skippedFiles = append(skippedFiles, file)
	}

	if tracker != nil {
		tracker.Finish()
	} else if !quiet {
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

	if len(downloadErrors) > 0 {
		if wasCancelled {
			fmt.Printf("Download cancelled by user with %d incomplete download(s)\n", len(downloadErrors))
			os.Exit(1)
		}
		return fmt.Errorf("failed to download %d file(s)", len(downloadErrors))
	}

	if ctx.Err() != nil {
		fmt.Println("Download cancelled by user")
		os.Exit(1)
	}

	return nil
}
