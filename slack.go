package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackutilsx"
)

func sendToSlack(format string, args ...interface{}) {
	token := config.Slack.Token
	channelName := config.Slack.Channel
	user := config.Slack.User

	if token == "" {
		println("no slack token")
		return
	}

	if channelName == "" {
		println("no slack channel name")
		return
	}

	if channelName[0] != '#' {
		channelName = "#" + channelName
	}

	api := slack.New(token)
	_, _, err := api.PostMessage(channelName,
		slack.MsgOptionUser(user),
		slack.MsgOptionText(fmt.Sprintf(format, args...), false))
	if err != nil {
		perror(fmt.Errorf("can not post msg to slack with err: %v", err))
	}
}

func formatSectionForSlackOutput(buf *bytes.Buffer, title string, description string) {
	buf.WriteString(fmt.Sprintf("*%s*\n", slackutilsx.EscapeMessage(title)))
	buf.WriteString(fmt.Sprintf("> %s\n", slackutilsx.EscapeMessage(description)))
}

func formatGitHubIssueForSlackOutput(issue github.Issue) string {
	isFromTeam := false
	login := issue.GetUser().GetLogin()
	for _, id := range allMembers {
		if strings.EqualFold(id, login) {
			isFromTeam = true
			break
		}
	}
	var tp string
	if !isFromTeam {
		tp = " _(Community)_"
	}

	s := fmt.Sprintf("[ %s ]%s <%s|%s> by @%s", regexRepo.FindStringSubmatch(issue.GetHTMLURL())[1], tp, issue.GetHTMLURL(), slackutilsx.EscapeMessage(issue.GetTitle()), issue.GetUser().GetLogin())

	if issue.Assignees != nil && len(issue.Assignees) > 0 {
		s += fmt.Sprintf(", assigned to")
		for _, assigne := range issue.Assignees {
			s += fmt.Sprintf(" @%s", assigne.GetLogin())
		}
	}

	return s
}

func formatJiraIssueForSlackOutput(issue jira.Issue) string {
	link := fmt.Sprintf("%sbrowse/%s", config.Jira.Endpoint, issue.Key)
	status := "Unknown"
	if issue.Fields != nil && issue.Fields.Status != nil {
		status = issue.Fields.Status.Name
	}
	priority := "Unknown"
	if issue.Fields != nil && issue.Fields.Priority != nil {
		priority = issue.Fields.Priority.Name
	}
	return fmt.Sprintf("[ %s / %s ] <%s|%s>", status, priority, link, issue.Fields.Summary)
}

func formatGitHubIssuesForSlackOutput(buf *bytes.Buffer, issues []github.Issue) {
	if len(issues) == 0 {
		buf.WriteString("_None_\n")
		return
	}
	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("• %s\n", formatGitHubIssueForSlackOutput(issue)))
	}
}

func formatJiraIssuesForSlackOutput(buf *bytes.Buffer, issues []jira.Issue) {
	if len(issues) == 0 {
		buf.WriteString("_None_\n")
		return
	}
	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("• %s\n", formatJiraIssueForSlackOutput(issue)))
	}
}
