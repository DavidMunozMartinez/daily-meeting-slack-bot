package main

import (
	"strconv"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func MeetingOrder(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	blocks := []slack.Block{}
	users, _, err := client.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: cmd.ChannelID})
	if err != nil {
		blocks = append(blocks, MakeSimpleTextSectionBlock("Error: "+err.Error()))
	}
	Shuffle(users)
	if cmd.Text != "" && len(users) > 0 {
		blocks = append(blocks, MakeSimpleTextSectionBlock(cmd.Text+" Team"))
	}
	count := 0
	for _, user := range users {
		info, err := client.GetUserInfo(user)
		if err != nil {
			blocks = append(blocks, MakeSimpleTextSectionBlock("Error: "+err.Error()))
		}

		if CanAddToList(info, cmd.Text) {
			count++
			order := strconv.FormatInt(int64(count), 10)
			display := "<@" + user + ">"
			blocks = append(
				blocks,
				MakeSimpleTextSectionBlock(order+" - "+display),
			)
		}
	}
	return blocks
}
