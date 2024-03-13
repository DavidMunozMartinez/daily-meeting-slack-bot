package slackapicommandhandler

import (
	"fmt"
	slackapiinteractionhandler "slack-manager/slack-api-interaction-handler"
	utils "slack-manager/utils"
	"strconv"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func MeetingOrder(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	blocks := []slack.Block{}
	users, err := getShuffledUsersInChannel(cmd, client)
	if err != nil {
		blocks = append(blocks, utils.MakeTextSectionBlock("Error: "+err.Error()))
	}

	if cmd.Text != "" && len(users) > 0 {
		blocks = append(blocks, utils.MakeTextSectionBlock(cmd.Text+" Team"))
	}

	count := 0
	for _, user := range users {
		info, err := client.GetUserInfo(user)
		if err != nil {
			blocks = append(blocks, utils.MakeTextSectionBlock("Error: "+err.Error()))
		}

		if utils.CanAddToList(info, cmd.Text, true) {
			count++
			order := strconv.FormatInt(int64(count), 10)
			display := "<@" + user + ">"
			blocks = append(
				blocks,
				utils.MakeTextSectionBlock(order+" - "+display),
			)

			fmt.Println(info.RealName)
		}
	}
	return blocks
}

func MeetingOrderV2(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	blocks := []slack.Block{}
	users, err := getShuffledUsersInChannel(cmd, client)
	if err != nil {
		blocks = append(blocks, utils.MakeTextSectionBlock("Error: "+err.Error()))
		return blocks
	}

	if cmd.Text != "" && len(users) > 0 {
		blocks = append(blocks, utils.MakeTextSectionBlock(cmd.Text+" Team"))
	}

	count := 0
	for _, user := range users {
		info, err := client.GetUserInfo(user)
		if err != nil {
			blocks = append(blocks, utils.MakeTextSectionBlock("Error: "+err.Error()))
		}

		if utils.CanAddToList(info, cmd.Text, true) {
			count++
			order := strconv.FormatInt(int64(count), 10)
			display := "<@" + user + ">"
			blocks = append(
				blocks,
				utils.MakeTextSectionBlock(order+" - "+slackapiinteractionhandler.GetDefaultMarker()+display),
			)
		}
	}

	if count > 0 {
		blocks = append(blocks, utils.MakeButtonSectionBlock("I am in the meeting!", ":slack_call:"))
	}

	return blocks
}
