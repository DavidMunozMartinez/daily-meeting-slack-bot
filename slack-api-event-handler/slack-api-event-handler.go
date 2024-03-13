package slackapieventshandler

import (
	utils "slack-manager/utils"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const (
	// Code review
	CodeReview = "CR"
)

func EventHandler(client *socketmode.Client, event slackevents.EventsAPIEvent) {
	switch event.InnerEvent.Type {
	case "app_mention":
		data := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
		AppMentionHandler(client, data)
	}
}

func AppMentionHandler(socketClient *socketmode.Client, data *slackevents.AppMentionEvent) {
	var args = strings.Split(data.Text, " ")

	if len(args) > 1 && args[1] != "" {
		switch args[1] {
		case CodeReview:
			TagCodeReviewers(socketClient, data)
		}
	}
}

func TagCodeReviewers(socketClient *socketmode.Client, data *slackevents.AppMentionEvent) {
	var args = strings.Split(data.Text, " ")
	var team = ""
	var count = 0
	if len(args) > 2 {
		team = args[2]
	}

	blocks := []slack.Block{}
	users, _, _ := socketClient.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: data.Channel})

	utils.Shuffle(users)
	var filtered []string
	// We have a parameter
	for i := range users {
		info, _ := socketClient.GetUserInfo(users[i])
		if utils.CanAddToList(info, team, false) && count < 2 && data.User != users[i] {
			filtered = append(filtered, users[i])
			count++
		}
	}
	if len(filtered) == 0 {
		blocks = append(blocks, utils.MakeTextSectionBlock("No available reviewers :c"))
	} else {
		var title = ""
		if team != "" {
			title += "[" + team + "] "
		}
		title += "Code reviewrs: "
		blocks = append(blocks, utils.MakeTextSectionBlock(title))
	}
	for i := range filtered {
		blocks = append(blocks, utils.MakeTextSectionBlock("<@"+filtered[i]+">"))
	}
	socketClient.PostMessage(
		data.Channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionTS(data.ThreadTimeStamp),
		slack.MsgOptionBlocks(blocks...),
	)
}
