package main

import (
	"fmt"
	"strconv"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

const (
	dayFormat  = "2006-01-02"
	dateFormat = "2006-01-02T15:04:05Z07:00"
	// We use one week for a sprint
	sprintDuration = 7 * 24 * time.Hour
)

// Get the board ID by project and boardType.
// Here we assume that you must create a board in the project and
// the function will return the first board ID.
func getBoardID(project string, boardType string) int {
	opts := jira.BoardListOptions{
		BoardType:      boardType,
		ProjectKeyOrID: project,
	}

	boards, _, err := jiraClient.Board.GetAllBoards(&opts)
	perror(err)

	return boards.Values[0].ID
}

func getSprints(boardID int, state string) []jira.Sprint {
	opts := jira.GetAllSprintsOptions{
		State: state,
	}

	// TODO: support pagination
	sprints, _, err := jiraClient.Board.GetAllSprintsWithOptions(boardID, &opts)
	perror(err)

	return sprints.Values
}

// Returns the only active sprint
func getActiveSprint(boardID int) jira.Sprint {
	sprints := getSprints(boardID, "active")
	return sprints[0]
}

func createSprint(boardID int, name string, startDate, endDate string) jira.Sprint {
	apiEndpoint := "rest/agile/1.0/sprint"
	sprint := map[string]string{
		"name":          name,
		"startDate":     startDate,
		"endDate":       endDate,
		"originBoardId": strconv.Itoa(boardID),
	}
	req, err := jiraClient.NewRequest("POST", apiEndpoint, sprint)
	perror(err)

	responseSprint := new(jira.Sprint)
	_, err = jiraClient.Do(req, responseSprint)
	perror(err)

	return *responseSprint
}

func createNextSprint(boardID int) jira.Sprint {
	now := time.Now()
	// Assume now is 2018-09-28 12:00:00
	// The next sprint's start time is 2018-09-28 00:00:00
	// and the end time is 2018-10-05 00:00:00
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.Add(sprintDuration)

	// The sprint name is 2018-09-28 - 2018-10-04
	name := fmt.Sprintf("%s - %s", startDate.Format(dayFormat), endDate.Add(-time.Second).Format(dayFormat))

	sprints := getSprints(boardID, "future")
	for _, sprint := range sprints {
		if sprint.Name == name {
			return sprint
		}
	}

	return createSprint(boardID, name, startDate.Format(dateFormat), endDate.Format(dateFormat))
}

func deleteSprint(sprintID int) {
	apiEndpoint := "rest/agile/1.0/sprint/" + strconv.Itoa(sprintID)
	req, err := jiraClient.NewRequest("DELETE", apiEndpoint, nil)
	perror(err)

	_, err = jiraClient.Do(req, nil)
	perror(err)
}
