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

// Use getSlackClient() to access it with lazy initialize feature.
var slackClient *slack.Client = nil

var slackMemberInit = false
var slackMembers = map[string]string{}

func getSlackClient() *slack.Client {
	if slackClient == nil {
		slackClient = slack.New(config.Slack.Token)
	}
	return slackClient
}

func initSlackMemberCache() {
	if slackMemberInit {
		return
	}
	users, err := getSlackClient().GetUsers()
	perror(err)
	if len(users) == 0 {
		perror(fmt.Errorf("cannot retrieve slack user list. slack app must be granted `users:read` and `users:read.email` permission"))
	}

	for _, user := range users {
		slackMembers[strings.ToLower(user.Profile.Email)] = user.ID
	}
	slackMemberInit = true
}

func buildSlackMention(email string) string {
	initSlackMemberCache()
	id, ok := slackMembers[strings.ToLower(email)]
	if !ok {
		return slackutilsx.EscapeMessage(email)
	}
	return fmt.Sprintf("<@%s>", id)
}

func sendToSlack(format string, args ...interface{}) {
	channelName := config.Slack.Channel
	user := config.Slack.User

	if channelName == "" {
		println("no slack channel name")
		return
	}

	if channelName[0] != '#' {
		channelName = "#" + channelName
	}

	_, _, err := getSlackClient().PostMessage(channelName,
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
	var closed string
	if issue.GetState() == "closed" {
		closed = " _(Closed)_"
	}

	s := fmt.Sprintf(
		"[ %s ]%s%s <%s|%s> by @%s",
		slackutilsx.EscapeMessage(regexRepo.FindStringSubmatch(issue.GetHTMLURL())[1]),
		slackutilsx.EscapeMessage(closed),
		slackutilsx.EscapeMessage(tp),
		issue.GetHTMLURL(),
		slackutilsx.EscapeMessage(issue.GetTitle()),
		slackutilsx.EscapeMessage(issue.GetUser().GetLogin()),
	)

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
	assignment := ""
	if issue.Fields != nil && issue.Fields.Assignee != nil {
		assignment = fmt.Sprintf("assigned to %s", buildSlackMention(issue.Fields.Assignee.EmailAddress))
	}
	return fmt.Sprintf(
		"[ %s / %s ] <%s|%s> %s",
		slackutilsx.EscapeMessage(status),
		slackutilsx.EscapeMessage(priority),
		link,
		slackutilsx.EscapeMessage(issue.Fields.Summary),
		assignment,
	)
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
