package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

var repoQuery string

func getCreatedIssues(start, end string) []github.Issue {
	opt := github.SearchOptions{
		Sort: "created",
	}

	var allIssues []github.Issue

	query := fmt.Sprintf("%s created:%s..%s is:issue", repoQuery, start, end)

	for {
		issues, resp, err := githubClient.Search.Issues(globalCtx, query, &opt)
		perror(err)

		allIssues = append(allIssues, issues.Issues...)

		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return allIssues
}

func getInvolvesPullRequest(user string, start, end string) []github.Issue {
	opt := github.SearchOptions{
		Sort: "updated",
	}

	var allIssues []github.Issue

	query := fmt.Sprintf("%s updated:%s..%s is:pr involves:%s", repoQuery, start, end, user)

	for {
		issues, resp, err := githubClient.Search.Issues(globalCtx, query, &opt)
		perror(err)

		allIssues = append(allIssues, issues.Issues...)

		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return allIssues
}

func initRepoQuery() {
	s := strings.Join(config.Github.Repos, " repo:")
	repoQuery = "repo:" + s
}
