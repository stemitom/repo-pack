package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	fetcher "repo-pack/fetcher"
	"repo-pack/gh"
	parse "repo-pack/parse"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Define flags for repository URL and token
	repoURL := flag.String("url", "", "GitHub repository URL")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	// Check if repository URL flag is provided
	if *repoURL == "" {
		return fmt.Errorf("usage: %s --url <repository_url> --token <personal_access_token>", os.Args[0])
	}

	// Parse the repository URL
	components, err := parse.ParseRepoURL(*repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %v", err)
	}

	// Check if repository is private
	isPrivate, err := fetcher.FetchRepoIsPrivate(ctx, components.Owner, components.Repository, *token)
	if err != nil {
		return fmt.Errorf("error checking repository privacy: %v", err)
	}

	if isPrivate {
		log.Println("The repository is private.")
	} else {
		log.Println("The repository is public.")
	}

	files, truncated, err := gh.ViaTreesAPI(ctx, components, *token)
	if err != nil {
		return fmt.Errorf("failed to get files via contents api: %v", err)
	}

	log.Println(files)
	log.Println(truncated)

	return nil
}
