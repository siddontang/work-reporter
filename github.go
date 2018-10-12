package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

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

	retryCount := 0
	for {
		issues, resp, err := githubClient.Search.Issues(globalCtx, query.String(), &opt)
		if err1, ok := err.(*github.RateLimitError); ok {
			dur := err1.Rate.Reset.Time.Sub(time.Now())
			if dur < 0 {
				dur = time.Minute
			}
			retryCount++
			if retryCount <= 10 {
				fmt.Printf("meet RateLimitError, wait %s and retry %d\n", dur, retryCount)
				time.Sleep(dur)
				continue
			}
		}

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
