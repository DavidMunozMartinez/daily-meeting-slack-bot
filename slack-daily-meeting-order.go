package main

import (
	"fmt"
	"log"
	"os"

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
					blocks = MeetingOrder(cmd, socketClient)
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
