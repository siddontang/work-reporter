package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

func newDailyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "daily",
		Short: "Daily Report",
		Args:  cobra.MinimumNArgs(0),
		Run:   runDailyCommandFunc,
	}

	return m
}

func plainFormatIssue(issue github.Issue) string {
	s := fmt.Sprintf("%s %s", issue.GetHTMLURL(), issue.GetTitle())

	if issue.Assignees != nil {
		for _, assigne := range issue.Assignees {
			s += fmt.Sprintf(" @%s", assigne.GetLogin())
		}
	}

	return s
}

func plainFormatIssues(buf *bytes.Buffer, title string, issues []github.Issue) {
	if len(issues) == 0 {
		return
	}

	buf.WriteString(fmt.Sprintf("# %s\n", title))

	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("+ %s\n", plainFormatIssue(issue)))
	}
}

func runDailyCommandFunc(cmd *cobra.Command, args []string) {
	now := time.Now()
	start := now.Add(-24 * time.Hour).Format(dateFormat)
	end := now.Format(dateFormat)

	var buf bytes.Buffer
	issues := getCreatedIssues(start, end)
	plainFormatIssues(&buf, "New Issues", issues)

	issues = getCreatedPullRequests(start, end)
	plainFormatIssues(&buf, "New Pull Requests", issues)

	msg := buf.String()
	if len(msg) == 0 {
		return
	}

	sendToSlack(buf.String())
}
