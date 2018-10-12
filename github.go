package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

var repoQuery string

func getIssues(sort string, queryArgs map[string]string) []github.Issue {
	opt := github.SearchOptions{
		Sort: sort,
	}

	var allIssues []github.Issue

	query := bytes.NewBufferString(repoQuery)

	for key, value := range queryArgs {
		query.WriteString(fmt.Sprintf(" %s:%s", key, value))
	}

	for {
		issues, resp, err := githubClient.Search.Issues(globalCtx, query.String(), &opt)
		perror(err)

		allIssues = append(allIssues, issues.Issues...)

		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return allIssues
}

func getCreatedIssues(start string, end string) []github.Issue {
	return getIssues("created", map[string]string{
		"is":      "issue",
		"created": fmt.Sprintf("%s..%s", start, end),
	})
}

func getCreatedPullRequests(start string, end string) []github.Issue {
	return getIssues("created", map[string]string{
		"is":      "pr",
		"created": fmt.Sprintf("%s..%s", start, end),
	})
}

func getInvolvesPullRequests(user string, start, end string) []github.Issue {
	return getIssues("upated", map[string]string{
		"is":       "pr",
		"involves": user,
		"-author":  user,
		"updated":  fmt.Sprintf("%s..%s", start, end),
	})
}

func initRepoQuery() {
	s := strings.Join(config.Github.Repos, " repo:")
	repoQuery = "repo:" + s
}
