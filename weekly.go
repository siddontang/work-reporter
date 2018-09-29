package main

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/andygrunwald/go-jira"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

var (
	emailToGithub = map[string]string{}
	githubToEmail = map[string]string{}
)

func initAccountMapping() {
	update := func(m map[string]string, key string, value string) map[string]string {
		key = strings.ToLower(key)
		if _, ok := m[key]; ok {
			perrmsg(fmt.Sprintf("duplicated %s %s", key, value))
		}
		m[key] = value
		return m
	}
	for _, team := range config.Teams {
		for _, member := range team.Members {
			emailToGithub = update(emailToGithub, member.Email, member.Github)
			githubToEmail = update(githubToEmail, member.Github, member.Email)
		}
	}
}

func newWeeklyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "weekly",
		Short: "Weely Report",
		Args:  cobra.MinimumNArgs(0),
		Run:   runWeelyCommandFunc,
	}

	return m
}

func genUserPage(buf *bytes.Buffer, m Member, curSprint jira.Sprint, nextSprint jira.Sprint) {
	sprintID := curSprint.ID
	nextSprintID := nextSprint.ID

	jql := `
<ac:structured-macro ac:name="jira">
  <ac:parameter ac:name="columns">key,summary,created,updated,priority,status</ac:parameter>
  <ac:parameter ac:name="server">%s</ac:parameter>
  <ac:parameter ac:name="serverId">%s</ac:parameter>
  <ac:parameter ac:name="jqlQuery">project = TIKV AND sprint = %d AND assignee = "%s"</ac:parameter>
</ac:structured-macro>
`

	buf.WriteString(fmt.Sprintf("<h2>%s</h2>", m.Name))
	buf.WriteString("<h3>Work</h3>")
	buf.WriteString(fmt.Sprintf(jql, config.Jira.Server, config.Jira.ServerID, sprintID, m.Email))
	genInvolvesPullReques(buf, m.Github, curSprint.StartDate.Format(dateFormat), curSprint.EndDate.Format(dateFormat))
	if nextSprintID > 0 {
		buf.WriteString("<h3>Next Week</h3>")
		buf.WriteString(fmt.Sprintf(jql, config.Jira.Server, config.Jira.ServerID, nextSprintID, m.Email))
	}
}

func genInvolvesPullReques(buf *bytes.Buffer, user, start, end string) {
	buf.WriteString("<h3>Review PR</h3>")

	issues := getInvolvesPullRequest(user, start, end)
	if len(issues) == 0 {
		return
	}

	buf.WriteString("<ul>")
	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("<li>%s</li>", htmlFormatIssueTitle(&issue)))
	}
	buf.WriteString("</ul>")
}

func genOnCall(buf *bytes.Buffer, start, end string) {
	buf.WriteString("<h1>OnCall</h1>")
	jql := `
<ac:structured-macro ac:name="jira">
  <ac:parameter ac:name="columns">key,summary,created,updated,assignee</ac:parameter>
  <ac:parameter ac:name="server">%s</ac:parameter>
  <ac:parameter ac:name="serverId">%s</ac:parameter>
  <ac:parameter ac:name="jqlQuery">project = %s AND created >= %s AND created &lt; %s</ac:parameter>
</ac:structured-macro>
`

	buf.WriteString(fmt.Sprintf(jql, config.Jira.Server, config.Jira.ServerID, config.Jira.OnCall, start, end))
}

func htmlFormatIssueTitle(issue *github.Issue) string {
	s := fmt.Sprintf("<a href=\"%s\">%s</a> %s", issue.GetHTMLURL(), issue.GetHTMLURL(), html.EscapeString(issue.GetTitle()))
	return s
}

func htmlFormatIssue(issue github.Issue) string {
	s := htmlFormatIssueTitle(&issue)

	if issue.Assignees != nil {
		for _, assigne := range issue.Assignees {
			s += fmt.Sprintf(" <a href=\"https://github.com/%s\">@%s</a>", assigne.GetLogin(), assigne.GetLogin())
		}
	}

	return s
}

func genCreatedIssues(buf *bytes.Buffer, start, end string) {
	buf.WriteString("<h1>Issues</h1>")

	issues := getCreatedIssues(start, end)
	if len(issues) == 0 {
		return
	}

	buf.WriteString("<ul>")

	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("<li>%s</li>", htmlFormatIssue(issue)))
	}

	buf.WriteString("</ul>")
}

func genToc(buf *bytes.Buffer) {
	toc := `
<ac:structured-macro ac:name="toc">
  <ac:parameter ac:name="printable">true</ac:parameter>
  <ac:parameter ac:name="style">square</ac:parameter>
  <ac:parameter ac:name="maxLevel">2</ac:parameter>
  <ac:parameter ac:name="class">bigpink</ac:parameter>
  <ac:parameter ac:name="type">list</ac:parameter>
</ac:structured-macro>
	`
	buf.WriteString(toc)
}

func runWeelyCommandFunc(cmd *cobra.Command, args []string) {
	id := getBoardID(config.Jira.Project, "scrum")
	sprint := getActiveSprint(id)
	nextSprint := createNextSprint(id)

	var body bytes.Buffer

	startDate := sprint.StartDate.Format(dayFormat)
	endDate := sprint.EndDate.Format(dayFormat)

	genToc(&body)
	genOnCall(&body, startDate, endDate)
	genCreatedIssues(&body, startDate, endDate)

	for _, team := range config.Teams {
		body.WriteString(fmt.Sprintf("<h1>%s Team</h1>", team.Name))
		for _, m := range team.Members {
			genUserPage(&body, m, sprint, nextSprint)
		}
	}

	sendToSlack(fmt.Sprintf("please fill your next sprint works at %s", nextSprint.Name))
	title := sprint.Name
	createWeeklyReport(title, body.String())
}

func createWeeklyReport(title string, value string) {
	space := config.Confluence.Space
	c := getContentByTitle(space, title)

	if c.Id != "" {
		c = updateContent(c, value)
	} else {
		parent := getContentByTitle(space, config.Confluence.WeeklyPath)
		c = createContent(space, parent.Id, title, value)
	}

	sendToSlack(fmt.Sprintf("weekly report for sprint %s is generated, please see %s%s", title, config.Confluence.Endpoint, c.Links.WebUI))
}
