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

func createNextSprint(boardID int, startDate time.Time) jira.Sprint {
	// We assuem the sprint starts at 00:00 and ends at 00:00
	// E.g, current sprint time range is 2018-09-28T00:00:00+08:00 2018-10-05T00:00:00+08:00
	// So the next sprint is 2018-10-05T00:00:00+08:00, 2018-10-12T00:00:00+08:00
	// The sprint name is 2018-10-05 - 2018-10-11
	endDate := startDate.Add(sprintDuration)

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

func updateSprintTime(sprintID int, startDate, endDate string) jira.Sprint {
	return updateSprint(sprintID, map[string]string{
		"startDate": startDate,
		"endDate":   endDate,
	})
}

func updateSprintState(sprintID int, state string) jira.Sprint {
	return updateSprint(sprintID, map[string]string{
		"state": state,
	})
}

func updateSprint(sprintID int, args map[string]string) jira.Sprint {
	apiEndpoint := "rest/agile/1.0/sprint/" + strconv.Itoa(sprintID)

	req, err := jiraClient.NewRequest("POST", apiEndpoint, args)
	perror(err)

	responseSprint := new(jira.Sprint)
	_, err = jiraClient.Do(req, responseSprint)
	perror(err)

	return *responseSprint
}
