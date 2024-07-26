package slackapicommandhandler

import (
	slackapischeduler "slack-manager/slack-api-scheduler"
	utils "slack-manager/utils"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func CommandHandler(cmd slack.SlashCommand, client *socketmode.Client) ([]slack.Block, string) {
	blocks := []slack.Block{}
	responseType := slack.ResponseTypeInChannel
	switch cmd.Command {
	case "/meeting-order":
		blocks = MeetingOrder(cmd, client)
	case "/meeting-order-v2":
		blocks = MeetingOrderV2(cmd, client)
	case "/api-status":
		blocks = GetAPIStatus(cmd, client)
	case "/get-rotation":
		blocks = slackapischeduler.GetRotationState(cmd, client)
		responseType = slack.ResponseTypeEphemeral
	case "/set-fe-rotation":
		blocks = slackapischeduler.SetFERotation(cmd, client)
		responseType = slack.ResponseTypeEphemeral
	case "/set-be-rotation":
		blocks = slackapischeduler.SetBERotation(cmd, client)
		responseType = slack.ResponseTypeEphemeral
	default:
		blocks = append(blocks, utils.MakeTextSectionBlock("Unknown command :("))
	}
	return blocks, responseType
}

func getShuffledUsersInChannel(cmd slack.SlashCommand, client *socketmode.Client) ([]string, error) {
	users, _, err := client.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: cmd.ChannelID})
	if err == nil {
		utils.Shuffle(users)
	}

	return users, err
}
