package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	fetcher "repo-pack/fetcher"
	gh "repo-pack/gh"
	parse "repo-pack/parse"
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
		return fmt.Errorf("usage: %s --url <repository_url> --token <personal_access_token>", os.Args[0])
	}

	components, err := parse.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	isPrivate, err := fetcher.FetchRepoIsPrivate(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("error checking repository privacy: %v", err)
	}

	files, _, err := gh.RepoListingSlashBranchSupport(ctx, &components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents api: %v", err)
	}

	fmt.Printf("[-] Downloading %d files\n", len(files))
	fmt.Printf("[-] Github Directory: %s\n", components.Dir)

	var wg sync.WaitGroup
	wg.Add(len(files))

	errorsCh := make(chan error, len(files))

	for _, file := range files {
		go func(file string, isPrivate bool) {
			defer wg.Done()

			content, err := fetcher.FetchPublicFile(ctx, file, &components)
			fmt.Println(string(content))
			if err != nil {
				errorsCh <- fmt.Errorf("error fetching %s: %v", file, err)
				return
			}
		}(file, isPrivate)
	}

	wg.Wait()
	close(errorsCh)

	select {
	case err := <-errorsCh:
		log.Println(err)
		log.Println("there is an error i guess")
	default:
		log.Println("All downloads completed successfully!")
	}

	return nil
}
