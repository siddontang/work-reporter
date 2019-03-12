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

var allMembers []string

const (
	// go-github/github incorrectly handles URL escape with "+", so we avoid "+" by using a UTC time
	githubUTCDateFormat = "2006-01-02T15:04:05Z"
)

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

func generateDateRangeQuery(start *string, end *string) string {
	if start != nil && end != nil {
		return fmt.Sprintf("%s..%s", *start, *end)
	} else if start != nil && end == nil {
		return fmt.Sprintf(">=%s", *start)
	} else if start == nil && end != nil {
		return fmt.Sprintf("<%s", *end)
	} else {
		panic("start and end can not be nil at the same time")
	}
}

func getCreatedIssues(start *string, end *string) []github.Issue {
	return getIssues("created", map[string]string{
		"is":      "issue",
		"created": generateDateRangeQuery(start, end),
	})
}

func getCreatedPullRequests(start *string, end *string) []github.Issue {
	return getIssues("created", map[string]string{
		"is":      "pr",
		"created": generateDateRangeQuery(start, end),
	})
}

func getMergedPullRequests(start *string, end *string) []github.Issue {
	return getIssues("created", map[string]string{
		"is":     "merged",
		"merged": generateDateRangeQuery(start, end),
	})
}

func getReviewPullRequests(user string, start *string, end *string) []github.Issue {
	return getIssues("updated", map[string]string{
		"is":        "pr",
		"commenter": user,
		"-author":   user,
		"updated":   generateDateRangeQuery(start, end),
	})
}

func getInactiveCommunityPullRequests(start *string, end *string) []github.Issue {
	openPullRequests := getIssues("updated", map[string]string{
		"is":      "pr",
		"state":   "open",
		"updated": generateDateRangeQuery(start, end),
	})

	communityPullRequests := make([]github.Issue, 0, len(openPullRequests))
nextOpenIssue:
	for _, issue := range openPullRequests {
		login := issue.GetUser().GetLogin()
		for _, id := range allMembers {
			if strings.EqualFold(id, login) {
				continue nextOpenIssue
			}
		}
		communityPullRequests = append(communityPullRequests, issue)
	}
	return communityPullRequests
}

func initRepoQuery() {
	s := strings.Join(config.Github.Repos, " repo:")
	repoQuery = "repo:" + s
}

func initTeamMembers() {
	for _, team := range config.Teams {
		for _, member := range team.Members {
			allMembers = append(allMembers, member.Github)
		}
	}
}
