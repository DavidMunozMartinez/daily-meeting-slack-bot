package slackapiinteractionhandler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
)

type ResponseURLBody struct {
	Blocks          slack.Blocks `json:"blocks"`
	ReplaceOriginal bool         `json:"replace_original"`
}

func InteractionHandler(event slack.InteractionCallback) {
	action := event.ActionCallback.BlockActions[0]

	switch action.ActionID {
	case "mark-assistance":
		markAssistance(event)
	}
}

func GetAttendanceMarker() string {
	return ":large_green_square:"
}

func GetDefaultMarker() string {
	return ":white_square:"
}

func markAssistance(event slack.InteractionCallback) {
	var userId = event.User.ID
	var newBlocks = event.Message.Blocks
	var changes = 0
	for _, elem := range newBlocks.BlockSet {
		var block = elem.(*slack.SectionBlock)
		if blockContainsUserId(block, userId) && !hasAttendanceMarker(block) {
			removeBlankMarker(block)
			addAttendanceMarker(block)
			changes++
		}
	}

	// Prevent the api call if no changes where made
	if changes > 0 {
		json, err := json.Marshal(ResponseURLBody{
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
}

func blockContainsUserId(block *slack.SectionBlock, userId string) bool {
	return strings.Contains(block.Text.Text, userId)
}

func hasAttendanceMarker(block *slack.SectionBlock) bool {
	var text = block.Text.Text
	return strings.Contains(text, GetAttendanceMarker())
}

func removeBlankMarker(block *slack.SectionBlock) {
	block.Text.Text = strings.Replace(block.Text.Text, GetDefaultMarker(), "", -1)
}

func addAttendanceMarker(block *slack.SectionBlock) {
	var text = block.Text.Text
	var deconstructedText = strings.Split(text, " - ")
	var order = deconstructedText[0]
	var name = deconstructedText[1]
	block.Text.Text = order + " - " + GetAttendanceMarker() + name
}
