package slackapicommandhandler

import (
	utils "slack-manager/utils"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func CommandHandler(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	blocks := []slack.Block{}
	switch cmd.Command {
	case "/meeting-order":
		blocks = MeetingOrder(cmd, client)
	case "/meeting-order-v2":
		blocks = MeetingOrderV2(cmd, client)
	default:
		blocks = append(blocks, utils.MakeTextSectionBlock("Unknown command :("))
	}
	return blocks
}

func getShuffledUsersInChannel(cmd slack.SlashCommand, client *socketmode.Client) ([]string, error) {
	users, _, err := client.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: cmd.ChannelID})
	if err == nil {
		utils.Shuffle(users)
	}

	return users, err
}
