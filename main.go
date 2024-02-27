package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"repo-pack/fetcher"
	"repo-pack/gh"
	"repo-pack/parse"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	repoURL := flag.String("url", "", "GitHub repository URL")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	if *repoURL == "" {
		return fmt.Errorf("usage: %s --url <repository_url> [--token <personal_access_token>]", os.Args[0])
	}

	components, err := parse.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	files, _, err := gh.RepoListingSlashBranchSupport(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents api: %v", err)
	}

	filesLength := len(files)
	baseDir := filepath.Base(components.Dir)

	fmt.Printf("[-] Downloading %d files\n", filesLength)
	fmt.Printf("[-] Github Directory: %s\n", components.Dir)

	var wg sync.WaitGroup
	wg.Add(filesLength)
	errorsCh := make(chan error, len(files))

	for _, file := range files {
		// TODO: isPrivate should be used once privatefile function can be properly tested
		go func(file string) {
			defer wg.Done()

			content, err := fetcher.FetchPublicFile(ctx, file, &components)
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file, err)
				return
			}

			if err := saveFile(baseDir, file, content); err != nil {
				errorsCh <- err
				return
			}
		}(file)
	}

	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		// TODO: Let error be easily identifiable
		log.Println(err)
	}

	fmt.Printf("Downloaded :%d files\n", filesLength)
	return nil
}

func saveFile(baseDir string, filePath string, content []byte) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	baseDirIndex := strings.Index(filePath, baseDir+"/")
	if baseDirIndex == -1 {
		return fmt.Errorf("base directory %s not found in file path %s", baseDir, filePath)
	}

	adjustedFilePath := filePath[baseDirIndex:]
	fullPath := filepath.Join(currentDir, adjustedFilePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating output folder for %s: %w", fullPath, err)
	}

	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return fmt.Errorf("error saving file %s: %w", fullPath, err)
	}

	fmt.Printf("[-] Downloaded: %s\n", adjustedFilePath)
	return nil
}
