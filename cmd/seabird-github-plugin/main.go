package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	github "github.com/seabird-chat/seabird-github-plugin"
)

func main() {
	_ = godotenv.Load()

	coreURL := os.Getenv("SEABIRD_HOST")
	coreToken := os.Getenv("SEABIRD_TOKEN")

	if coreURL == "" || coreToken == "" {
		log.Fatal("Missing SEABIRD_HOST or SEABIRD_TOKEN")
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("Missing GITHUB_TOKEN")
	}

	mapping := os.Getenv("GITHUB_REPOS")
	if mapping == "" {
		log.Fatal("Missing GITHUB_REPOS")
	}

	repos := map[string]github.Repo{}
	for _, repoEntry := range strings.Split(mapping, ",") {
		split := strings.SplitN(repoEntry, "=", 2)
		if len(split) != 2 {
			log.Fatal("Malformed repo entry")
		}

		tag := split[0]
		split = strings.Split(split[1], "/")
		if len(split) != 2 {
			log.Fatal("Malformed repo")
		}

		repos[tag] = github.Repo{
			Owner: split[0],
			Name:  split[1],
		}
	}

	c, err := github.NewClient(
		coreURL,
		coreToken,
		githubToken,
		repos,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
