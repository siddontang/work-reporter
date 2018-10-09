package main

import (
	"fmt"

	"github.com/nlopes/slack"
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
		slack.MsgOptionText(fmt.Sprintf(format, args...), true))
	if err != nil {
		perror(fmt.Errorf("can not post msg to slack with err: %v", err))
	}
}
