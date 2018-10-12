package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

var repoQuery string

// IssueSlice is the slice of issues
type IssueSlice []github.Issue

func (s IssueSlice) Len() int      { return len(s) }
func (s IssueSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s IssueSlice) Less(i, j int) bool {
	return s[i].GetHTMLURL() < s[j].GetHTMLURL()
}

func getIssues(bySort string, queryArgs map[string]string) IssueSlice {
	opt := github.SearchOptions{
		Sort: bySort,
	}

	var allIssues IssueSlice

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

	sort.Sort(allIssues)
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

func getReviewPullRequests(user string, start, end string) []github.Issue {
	return getIssues("upated", map[string]string{
		"is":        "pr",
		"commenter": user,
		"-author":   user,
		"updated":   fmt.Sprintf("%s..%s", start, end),
	})
}

func initRepoQuery() {
	s := strings.Join(config.Github.Repos, " repo:")
	repoQuery = "repo:" + s
}
