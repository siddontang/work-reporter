package main

import (
	"bytes"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var regexRepo = regexp.MustCompile("github\\.com\\/([^\\/]+\\/[^\\/]+)\\/")

func newDailyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "daily",
		Short: "Daily Report",
		Args:  cobra.MinimumNArgs(0),
		Run:   runDailyCommandFunc,
	}

	return m
}

func formatDailyInactiveOnCalls(buf *bytes.Buffer) {

}

func runDailyCommandFunc(cmd *cobra.Command, args []string) {
	now := time.Now().UTC()
	start := now.Add(-24 * time.Hour).Format(githubUTCDateFormat)

	var buf bytes.Buffer
	buf.WriteString("\n")

	issues := getCreatedIssues(start, nil)
	formatSectionForSlackOutput(&buf, "New Issues", "New issues in last 24 hours")
	formatGitHubIssuesForSlackOutput(&buf, issues)
	buf.WriteString("\n")

	issues = getCreatedPullRequests(start, nil)
	formatSectionForSlackOutput(&buf, "New Pull Requests", "New PRs in last 24 hours")
	formatGitHubIssuesForSlackOutput(&buf, issues)
	buf.WriteString("\n")

	oncallIssues := queryJiraIssues("project = ONCALL AND created >= \"-1d\"")
	formatSectionForSlackOutput(&buf, "New OnCalls", "New on calls in last 24 hours")
	formatJiraIssuesForSlackOutput(&buf, oncallIssues)
	buf.WriteString("\n")

	oncallIssues = queryJiraIssues("project = ONCALL AND priority = Highest AND resolution = Unresolved AND updated <= \"-3d\"")
	formatSectionForSlackOutput(&buf, "Inactive OnCalls", "Highest priority on calls inactive >= 3 days")
	formatJiraIssuesForSlackOutput(&buf, oncallIssues)
	buf.WriteString("\n")

	sendToSlack(buf.String())
}
