package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

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
