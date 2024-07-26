package main

import (
	"fmt"
	"log"
	"os"
	slackapicommandhandler "slack-manager/slack-api-command-handler"
	slackapieventshandler "slack-manager/slack-api-event-handler"
	slackapiinteractionhandler "slack-manager/slack-api-interaction-handler"
	slackapischeduler "slack-manager/slack-api-scheduler"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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

	slackapischeduler.Init(api)

	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)
					continue
				}

				blocks, responseType := slackapicommandhandler.CommandHandler(cmd, socketClient)
				payload := map[string]interface{}{
					"blocks":        blocks,
					"response_type": responseType,
				}
				fmt.Print(payload)
				socketClient.Ack(*evt.Request, payload)

			case socketmode.EventTypeEventsAPI:
				event, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)
					continue
				}

				slackapieventshandler.EventHandler(socketClient, event)
				socketClient.Ack(*evt.Request)

			case socketmode.EventTypeInteractive:
				event, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)
					continue
				}
				slackapiinteractionhandler.InteractionHandler(event)
				socketClient.Ack(*evt.Request)

			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	socketClient.Run()
}
