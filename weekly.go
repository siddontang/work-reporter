package main

import (
	"bytes"
	"fmt"
	"html"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

const jiraLabelColorGrey = "Grey"
const jiraLabelColorRed = "Red"
const jiraLabelColorYellow = "Yellow"
const jiraLabelColorGreen = "Green"
const jiraLabelColorBlue = "Blue"

func newWeeklyReportCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "report",
		Short: "Create Weekly Report",
		Run:   runWeelyReportCommandFunc,
	}
	return m
}

func newRotateSprintCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "rotate-sprint",
		Short: "Rotate Current Week Sprint",
		Run:   runRotateSprintCommandFunc,
	}
	return m
}

func newWeeklyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "weekly",
		Short: "Weely Tasks",
	}
	m.AddCommand(newWeeklyReportCommand())
	m.AddCommand(newRotateSprintCommand())
	return m
}

func runWeelyReportCommandFunc(cmd *cobra.Command, args []string) {
	boardID := getBoardID(config.Jira.Project, "scrum")
	sprints := getSprints(boardID, jira.GetAllSprintsOptions{})
	lastSprint := getLatestPassedSprint(sprints)

	var body bytes.Buffer

	startDate := lastSprint.StartDate.Format(dayFormat)
	endDate := lastSprint.EndDate.Format(dayFormat)

	githubStartDate := lastSprint.StartDate.UTC().Format(githubUTCDateFormat)
	githubEndDate := lastSprint.EndDate.UTC().Format(githubUTCDateFormat)

	formatPageBeginForHtmlOutput(&body)

	genWeeklyReportToc(&body)
	genWeeklyReportOnCall(&body, startDate, endDate)
	genWeeklyReportIssuesPRs(&body, githubStartDate, githubEndDate)
	genWeeklyReportProjects(&body, lastSprint)

	formatPageEndForHtmlOutput(&body)

	createWeeklyReport(lastSprint, body.String())
}

func runRotateSprintCommandFunc(cmd *cobra.Command, args []string) {
	boardID := getBoardID(config.Jira.Project, "scrum")
	activeSprint := getActiveSprint(boardID)
	nextSprint := createNextSprint(boardID, *activeSprint.EndDate)

	// Close the old sprint.
	updateSprintState(activeSprint.ID, "closed")
	// Active the next sprint.
	updateSprintState(nextSprint.ID, "active")
	sendToSlack("Current active Sprint %s is closed", activeSprint.Name)
}

func formatPageBeginForHtmlOutput(buf *bytes.Buffer) {
	buf.WriteString(`<ac:layout>`)
}

func formatPageEndForHtmlOutput(buf *bytes.Buffer) {
	buf.WriteString(`</ac:layout>`)
}

func formatSectionBeginForHtmlOutput(buf *bytes.Buffer) {
	buf.WriteString(`<ac:layout-section ac:type="single"><ac:layout-cell><hr/>`)
	buf.WriteString("\n")
}

func formatSectionEndForHtmlOutput(buf *bytes.Buffer) {
	buf.WriteString(`</ac:layout-cell></ac:layout-section>`)
	buf.WriteString("\n")
}

func formatLabelForHtmlOutput(name string, color string) string {
	s := fmt.Sprintf(`
	<ac:structured-macro ac:macro-id="9f29312a-2730-48f0-ab6d-91d6bef3f016" ac:name="status" ac:schema-version="1">
		<ac:parameter ac:name="colour">%s</ac:parameter>
		<ac:parameter ac:name="title">%s</ac:parameter>
	</ac:structured-macro>`, color, html.EscapeString(name))
	return s
}

func formatGitHubIssueForHtmlOutput(issue github.Issue) string {
	isFromTeam := false
	login := issue.GetUser().GetLogin()

	for _, id := range allMembers {
		if strings.EqualFold(id, login) {
			isFromTeam = true
			break
		}
	}

	var labelColor = jiraLabelColorGrey
	if issue.GetState() == "closed" {
		labelColor = jiraLabelColorGreen
	}

	s := fmt.Sprintf(
		`%s <a href="%s">%s</a> by @%s`,
		formatLabelForHtmlOutput(regexRepo.FindStringSubmatch(issue.GetHTMLURL())[1], labelColor),
		issue.GetHTMLURL(),
		html.EscapeString(issue.GetTitle()),
		html.EscapeString(issue.GetUser().GetLogin()),
	)

	if issue.Assignees != nil && len(issue.Assignees) > 0 {
		s += fmt.Sprintf(", assigned to")
		for _, assigne := range issue.Assignees {
			s += fmt.Sprintf(" @%s", assigne.GetLogin())
		}
	}

	if !isFromTeam {
		s += " " + formatLabelForHtmlOutput("Community", jiraLabelColorBlue)
	}

	return s
}

func formatGitHubIssuesForHtmlOutput(buf *bytes.Buffer, issues []github.Issue) {
	if len(issues) == 0 {
		buf.WriteString("<p><i>None</i></p>\n")
		return
	}
	buf.WriteString("<ul>")
	for _, issue := range issues {
		buf.WriteString(fmt.Sprintf("<li>%s</li>\n", formatGitHubIssueForHtmlOutput(issue)))
	}
	buf.WriteString("</ul>")
}

func genWeeklyUserPage(buf *bytes.Buffer, m Member, sprint *jira.Sprint) {
	formatPageBeginForHtmlOutput(buf)

	formatSectionBeginForHtmlOutput(buf)
	buf.WriteString("\n<h3>Work</h3>\n")
	buf.WriteString("\n<blockquote>A summary of my work in this week</blockquote>\n")
	buf.WriteString("\n<p>Please fill this section</p>\n")
	buf.WriteString("\n<h3>Next Week</h3>\n")
	buf.WriteString("\n<blockquote>A plan of the next week</blockquote>\n")
	buf.WriteString("\n<p>Please fill this section</p>\n")
	formatSectionEndForHtmlOutput(buf)

	formatSectionBeginForHtmlOutput(buf)
	buf.WriteString("\n<h3>Issues in this week</h3>\n")
	template := `
<ac:structured-macro ac:name="jira">
  <ac:parameter ac:name="columns">key,summary,created,updated,status</ac:parameter>
  <ac:parameter ac:name="server">%s</ac:parameter>
  <ac:parameter ac:name="serverId">%s</ac:parameter>
  <ac:parameter ac:name="jqlQuery">project = %s AND Sprint = %d AND assignee = "%s"</ac:parameter>
</ac:structured-macro>`
	buf.WriteString(fmt.Sprintf(template, config.Jira.Server, config.Jira.ServerID, config.Jira.Project, sprint.ID, m.Email))
	formatSectionEndForHtmlOutput(buf)

	formatPageEndForHtmlOutput(buf)
}

func genReviewPullRequests(buf *bytes.Buffer, user, start, end string) {
	buf.WriteString("<h3>Review PR</h3>")
	issues := getReviewPullRequests(user, &start, &end)
	formatGitHubIssuesForHtmlOutput(buf, issues)
}

func genWeeklyReportOnCall(buf *bytes.Buffer, start, end string) {
	formatSectionBeginForHtmlOutput(buf)

	buf.WriteString("\n<h1>New OnCall</h1>\n")
	buf.WriteString(fmt.Sprintf("\n<blockquote>Newly created OnCalls (created &gt;= %s AND created &lt; %s)</blockquote>\n", start, end))
	html := `
<ac:structured-macro ac:name="jira">
  <ac:parameter ac:name="columns">key,summary,created,updated,assignee,status</ac:parameter>
  <ac:parameter ac:name="server">%s</ac:parameter>
  <ac:parameter ac:name="serverId">%s</ac:parameter>
  <ac:parameter ac:name="jqlQuery">project = %s AND created &gt;= %s AND created &lt; %s</ac:parameter>
</ac:structured-macro>
`
	buf.WriteString(fmt.Sprintf(html, config.Jira.Server, config.Jira.ServerID, config.Jira.OnCall, start, end))

	buf.WriteString("\n<h1>Highest Priority</h1>\n")
	buf.WriteString("\n<blockquote>Unresolved highest priority OnCalls (priority = Highest AND resolution = Unresolved)</blockquote>\n")
	html = `
<ac:structured-macro ac:name="jira">
  <ac:parameter ac:name="columns">key,summary,created,updated,assignee,status</ac:parameter>
  <ac:parameter ac:name="server">%s</ac:parameter>
  <ac:parameter ac:name="serverId">%s</ac:parameter>
  <ac:parameter ac:name="jqlQuery">project = %s AND priority = Highest AND resolution = Unresolved</ac:parameter>
</ac:structured-macro>
`
	buf.WriteString(fmt.Sprintf(html, config.Jira.Server, config.Jira.ServerID, config.Jira.OnCall))

	formatSectionEndForHtmlOutput(buf)
}

func genWeeklyReportIssuesPRs(buf *bytes.Buffer, start, end string) {
	formatSectionBeginForHtmlOutput(buf)
	issues := getCreatedIssues(&start, &end)
	buf.WriteString("\n<h1>New Issues</h1>\n")
	buf.WriteString(fmt.Sprintf("\n<blockquote>New GitHub issues (created: %s..%s)</blockquote>\n", start, end))
	formatGitHubIssuesForHtmlOutput(buf, issues)
	prs := getMergedPullRequests(&start, &end)
	buf.WriteString("\n<h1>Merged PRs</h1>\n")
	buf.WriteString(fmt.Sprintf("\n<blockquote>Merged GitHub PRs (merged: %s..%s)</blockquote>\n", start, end))
	formatGitHubIssuesForHtmlOutput(buf, prs)
	formatSectionEndForHtmlOutput(buf)
}

func genWeeklyReportProjects(buf *bytes.Buffer, sprint *jira.Sprint) {
	formatSectionBeginForHtmlOutput(buf)
	buf.WriteString("\n<h1>Projects</h1>\n")
	epicQuery := `project = %s and "Epic Link" is not EMPTY and Sprint = %d`
	epicIssues := queryJiraIssues(fmt.Sprintf(epicQuery, config.Jira.Project, sprint.ID))
	// An epic link set.
	epics := make(map[string]struct{})
	for _, is := range epicIssues {
		// The magic name of epic link field.
		const epicLinkField = "customfield_10100"
		epicLink := is.Fields.Unknowns[epicLinkField].(string)
		epics[epicLink] = struct{}{}
	}

	projects := `
<table class="relative-table wrapped">
  <tbody>
  <tr>
    <th>Name</th>
    <th>Purpose</th>
    <th><p>Links</p></th>
    <th>Manager</th>
    <th>Collaborators</th>
    <th>Start Date</th>
    <th><p>Description</p></th>
  </tr>
  %s
  </tbody>
</table>`

	projectsBuf := bytes.Buffer{}
	for ep := range epics {
		epIssuesTemplate := `
    <ac:structured-macro ac:name="jira">
      <ac:parameter ac:name="columns">key,summary,assignee,status</ac:parameter>
      <ac:parameter ac:name="server">%s</ac:parameter>
      <ac:parameter ac:name="serverId">%s</ac:parameter>
      <ac:parameter ac:name="jqlQuery">project = %s and "Epic Link" = %s and Sprint = %d</ac:parameter>
    </ac:structured-macro>`
		epIssues := fmt.Sprintf(epIssuesTemplate,
			config.Jira.Server, config.Jira.ServerID, config.Jira.Project, ep, sprint.ID)

		projectTemplate := `
    <tr>
      <td>%s</td>
      <td>%s</td>
      <td><div class="content-wrapper">%s</div></td>
      <td><ac:link><ri:user ri:username="%s" /></ac:link></td>
      <td><br /></td>
      <td><time datetime="%s"/></td>
      <td><br /></td>
    </tr>`
		epic, _, err := jiraClient.Issue.Get(ep, nil)
		perror(err)
		// The magic name of epic name field.
		const epicNameField = "customfield_10102"
		epicName := html.EscapeString(epic.Fields.Unknowns[epicNameField].(string))
		projectsBuf.WriteString(fmt.Sprintf(projectTemplate,
			epicName, epic.Fields.Summary, epIssues, epic.Fields.Assignee.Name, time.Time(epic.Fields.Created).Format(dayFormat)))
	}

	buf.WriteString(fmt.Sprintf(projects, projectsBuf.String()))
	formatSectionEndForHtmlOutput(buf)
}

func genWeeklyReportToc(buf *bytes.Buffer) {
	formatSectionBeginForHtmlOutput(buf)

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

	formatSectionEndForHtmlOutput(buf)
}

func createWeeklyReport(sprint *jira.Sprint, value string) {
	title := sprint.Name
	space := config.Confluence.Space
	c := getContentByTitle(space, title)

	if c.Id != "" {
		c = updateContent(c, value)
	} else {
		parent := getContentByTitle(space, config.Confluence.WeeklyPath)
		c = createContent(space, parent.Id, title, value)
		for _, team := range config.Teams {
			for _, m := range team.Members {
				body := bytes.Buffer{}
				genWeeklyUserPage(&body, m, sprint)
				userTitle := fmt.Sprintf("%s - %s", m.Name, title)
				createContent(space, c.Id, userTitle, body.String())
			}
		}
	}

	sendToSlack("Weekly report for sprint %s is generated: %s%s", title, config.Confluence.Endpoint, c.Links.WebUI)
}
