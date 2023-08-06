package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	godotenv.Load(".env")

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	api := slack.New(
		token,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(appToken),
	)
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)
					continue
				}

				blocks := []slack.Block{}
				switch cmd.Command {
				case "/meeting-order":
					users, _, err := socketClient.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: cmd.ChannelID})
					if err != nil {
						blocks = append(blocks, MakeSimpleTextSectionBlock("Error: "+err.Error()))
					}
					Shuffle(users)
					if cmd.Text != "" && len(users) > 0 {
						blocks = append(blocks, MakeSimpleTextSectionBlock(cmd.Text+" Team"))
					}
					count := 0
					for _, user := range users {
						info, err := socketClient.GetUserInfo(user)
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
				default:
					blocks = append(blocks, MakeSimpleTextSectionBlock("Unknown command :("))
				}

				payload := map[string]interface{}{
					"blocks":        blocks,
					"response_type": "in_channel",
				}

				socketClient.Ack(*evt.Request, payload)
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	socketClient.Run()
}

func Shuffle(a []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
}

func MakeSimpleTextSectionBlock(text string) slack.Block {
	block := slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: text,
		},
		nil,
		nil,
	)
	return block
}

func CanAddToList(info *slack.User, query string) bool {
	passesQuery := query == "" || query != "" && strings.Contains(info.Profile.Title, query)
	return !info.IsBot && passesQuery
}
