package main

import (
	"bytes"
	"fmt"
	"html"

	jira "github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

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
	genReviewPullReques(buf, m.Github, curSprint.StartDate.Format(dayFormat), curSprint.EndDate.Format(dayFormat))
	if nextSprintID > 0 {
		buf.WriteString("<h3>Next Week</h3>")
		buf.WriteString(fmt.Sprintf(jql, config.Jira.Server, config.Jira.ServerID, nextSprintID, m.Email))
	}
}

func genReviewPullReques(buf *bytes.Buffer, user, start, end string) {
	buf.WriteString("<h3>Review PR</h3>")

	issues := getReviewPullRequests(user, start, end)
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

func htmlFormatIssues(buf *bytes.Buffer, title string, issues []github.Issue) {
	if len(issues) == 0 {
		return
	}

	buf.WriteString(fmt.Sprintf("<h1>%s</h1>", title))
	buf.WriteString("<ul>")

	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("<li>%s</li>", htmlFormatIssue(issue)))
	}

	buf.WriteString("</ul>")
}

func genCreatedIssues(buf *bytes.Buffer, title string, start, end string) {
	issues := getCreatedIssues(start, end)
	htmlFormatIssues(buf, title, issues)
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
	// Next sprint starts at the end of the current sprint
	nextSprint := createNextSprint(id, *sprint.EndDate)
	sendToSlack("please fill your next sprint works at %s", nextSprint.Name)

	var body bytes.Buffer

	startDate := sprint.StartDate.Format(dayFormat)
	endDate := sprint.EndDate.Format(dayFormat)

	genToc(&body)
	genOnCall(&body, startDate, endDate)
	genCreatedIssues(&body, "Issues", startDate, endDate)

	for _, team := range config.Teams {
		body.WriteString(fmt.Sprintf("<h1>%s Team</h1>", team.Name))
		for _, m := range team.Members {
			genUserPage(&body, m, sprint, nextSprint)
		}
	}

	title := sprint.Name
	createWeeklyReport(title, body.String())

	issues := getIssuesForSprint(sprint.ID)
	pendingIssues := make([]jira.Issue, 0)
	for _, is := range issues {
		if is.Fields.Status.Name == issuesStatusClosed {
			continue
		}
		pendingIssues = append(pendingIssues, is)
	}
	// Active the next sprint before moving issues.
	updateSprintState(nextSprint.ID, "active")
	moveIssuesToSprint(nextSprint.ID, pendingIssues)
	// Then close the old sprint.
	updateSprintState(sprint.ID, "closed")
	sendToSlack("close current active sprint %s", sprint.Name)
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

	sendToSlack("weekly report for sprint %s is generated, please see %s%s", title, config.Confluence.Endpoint, c.Links.WebUI)
}
