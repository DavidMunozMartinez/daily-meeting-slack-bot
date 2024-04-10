package utils

import (
	"fmt"
	"math/rand"
	titletags "slack-manager/title-tags"
	"slices"
	"strings"

	"github.com/slack-go/slack"
)

func Shuffle(a []string) {
	l := len(a) - 1
	for i := 0; i <= l; i++ {
		n := rand.Intn(l)
		fmt.Println(n)
		// swap
		x := a[i]
		a[i] = a[n]
		a[n] = x
	}
}

func MakeTextSectionBlock(text string) slack.Block {
	return slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: text,
		},
		nil,
		nil,
	)
}

func MakeCheckboxSectionBlock(display string, userId string) slack.Block {

	option := &slack.OptionBlockObject{
		Text: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Assistance",
		},
		Value: "false",
	}
	checkbox := &slack.CheckboxGroupsBlockElement{
		Type:     "checkboxes",
		ActionID: "mark-assistance-" + userId,
		Options:  []*slack.OptionBlockObject{option},
	}

	return slack.NewSectionBlock(
		&slack.TextBlockObject{Type: slack.MarkdownType, Text: display},
		nil,
		&slack.Accessory{
			CheckboxGroupsBlockElement: checkbox,
		},
	)
}

func MakeButtonSectionBlock(title string, text string) slack.Block {
	button := &slack.ButtonBlockElement{
		Type: "button",
		Text: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: title,
		},
		ActionID: "mark-assistance",
		Style:    "primary",
	}

	return slack.NewSectionBlock(
		&slack.TextBlockObject{
			Text: text,
			Type: slack.MarkdownType,
		},
		nil,
		&slack.Accessory{
			ButtonElement: button,
		},
	)
}

func CanAddToList(info *slack.User, query string, useInclusiveRoles bool) bool {
	var queryInTitle = strings.Contains(info.Profile.Title, query)
	var hasQuery = query != ""
	var IsInclusiveTitle = IsInclusiveTitle(info, query) && useInclusiveRoles

	passesQuery :=
		!hasQuery ||
			hasQuery && queryInTitle ||
			IsInclusiveTitle

	return !info.IsBot && passesQuery
}

// /*
// Check is current title query includes other titles and if the user complies with any
// */
func IsInclusiveTitle(info *slack.User, query string) bool {
	if titletags.TITLE_INCLUSION[query] == nil {
		return false
	} else {
		var incldedTitles = titletags.TITLE_INCLUSION[query]
		var userTitle = GetUserTitle(info)

		return slices.Contains(incldedTitles, userTitle)
	}
}

func GetUserTitle(user *slack.User) string {
	var value = ""
	for _, title := range titletags.TITLES {
		if strings.Contains(user.Profile.Title, title) && value == "" {
			value = title
		}
	}
	return value
}
