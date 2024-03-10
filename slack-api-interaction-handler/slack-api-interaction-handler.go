package slackapiinteractionhandler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type responseUrlStruct struct {
	Blocks          slack.Blocks `json:"blocks"`
	ReplaceOriginal bool         `json:"replace_original"`
}

func Handler(event slack.InteractionCallback, client *socketmode.Client) {
	action := event.ActionCallback.BlockActions[0]

	switch action.ActionID {
	case "mark-assistance":
		MarkAssistance(event, client)
	}
}

func MarkAssistance(event slack.InteractionCallback, client *socketmode.Client) {
	var userId = event.User.ID
	var newBlocks = event.Message.Blocks
	for _, block := range newBlocks.BlockSet {
		var data = block.(*slack.SectionBlock)
		if strings.Contains(data.Text.Text, userId) {
			var text = data.Text.Text
			var deconstructedText = strings.Split(text, " - ")
			var order = deconstructedText[0]
			var name = deconstructedText[1]
			data.Text.Text = order + " - " + ":white_check_mark:" + name
		}
	}

	json, err := json.Marshal(responseUrlStruct{
		Blocks:          newBlocks,
		ReplaceOriginal: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(event.ResponseURL, "application/json", bytes.NewReader(json))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
