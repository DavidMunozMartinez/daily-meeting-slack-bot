package slackapischeduler

import (
	"fmt"
	"log"
	"os"
	"slack-manager/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// User represents a Slack user with ID and real name
type User struct {
	ID       string
	RealName string
}

var (
	FE_List []User
	BE_List []User
)

// Indices for each list
var (
	FE_Index int
	BE_Index int
	api      *slack.Client
	mu       sync.Mutex // Mutex to synchronize access to indices and lists
)


func Init(a* slack.Client) {
	api = a
	channelID := os.Getenv("SCHEDULER_CHANNEL_ID")
	if channelID == "" {
		log.Fatalf("SCHEDULER_CHANNEL_ID environment variable is not set")
	}

	err := fetchChannelMembers(channelID)
	if err != nil {
		log.Fatalf("Error fetching channel members: %v", err)
	}

	c := cron.New()
	c.AddFunc("0 9 * * 1", scheduledTeamRotation) // Runs every Monday at 9:00 AM
	if err != nil {
		log.Fatalf("Error scheduling job: %v", err)
	}
	c.Start()
}

func fetchChannelMembers(channelID string) error {
	users, _, err := api.GetUsersInConversation(&slack.GetUsersInConversationParameters{
		ChannelID: channelID,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch channel members: %w", err)
	}

	for _, userID := range users {
		user, err := api.GetUserInfo(userID)
		if err != nil {
			return fmt.Errorf("failed to fetch user info: %w", err)
		}
		slackUser := User{ID: userID, RealName: user.RealName}
		if strings.Contains(user.Profile.Title, "FE") {
			FE_List = append(FE_List, slackUser)
		} else if strings.Contains(user.Profile.Title, "BE") {
			BE_List = append(BE_List, slackUser)
		}
	}

	return nil
}
func postMessageToSlack(channelID string, message string) error {
	channelID, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		return err
	}
	log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// This function runs every Monday at 9:00 AM
func scheduledTeamRotation() {
	channelID := os.Getenv("SCHEDULER_CHANNEL_ID")
	if channelID == "" {
		log.Fatalf("SCHEDULER_CHANNEL_ID environment variable is not set")
	}

	mu.Lock()
	defer mu.Unlock()
	
	FE_User := "No users"
	if len(FE_List) > 0 {
		FE_Index = (FE_Index + 1) % len(FE_List)
		FE_User = fmt.Sprintf("<@%s>", FE_List[FE_Index].ID)
	}

	BE_User := "No users"
	if len(BE_List) > 0 {
		BE_Index = (BE_Index + 1) % len(BE_List)
		BE_User = fmt.Sprintf("<@%s>", BE_List[BE_Index].ID)
	}

	message := fmt.Sprintf(
		"This week's sentry maintainers will be :saluting_face:\n%s from the FE team\n%s from the BE team\nThanks!",
		FE_User, BE_User,
	)
	err := postMessageToSlack(channelID, message)
	if err != nil {
		log.Fatalf("Error posting message: %v", err)
	}
}

func GetRotationState(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	mu.Lock()
	defer mu.Unlock()

	var blocks []slack.Block

	headerSection := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "*Current Sentry Maintainers*", false, false),
		nil,
		nil,
	)
	blocks = append(blocks, headerSection)

	FE_Text := ""
	for i, user := range FE_List {
		if i == FE_Index {
			FE_Text += fmt.Sprintf("*%d - <@%s>* :meow_code: \n", i, user.ID)
		} else {
			FE_Text += fmt.Sprintf("%d - <@%s> \n", i, user.ID)
		}
	}

	BE_Text := ""
	for i, user := range BE_List {
		if i == BE_Index {
			BE_Text += fmt.Sprintf("*%d - <@%s>* :meow_code: \n", i, user.ID)
		} else {
			BE_Text += fmt.Sprintf("%d - <@%s>\n", i, user.ID)
		}
	}

	feSection := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Front-End*\n%s", FE_Text), false, false),
		nil,
		nil,
	)
	beSection := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Back-End*\n%s", BE_Text), false, false),
		nil,
		nil,
	)

	blocks = append(blocks, feSection, beSection)

	// Append current user rotation with index
	blocks = append(blocks, utils.MakeTextSectionBlock(
		fmt.Sprintf("Current FE: %d - %s \n Current BE: %d - %s", FE_Index, FE_List[FE_Index].RealName, BE_Index, BE_List[BE_Index].RealName),
	))

	// Add instructions on how to set the current index for each list
	fe_instructions := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "*To override this weeks's FE maintained use:*\n/set-fe-rotation *index*", false, false),
		nil,
		nil,
	)

	be_instructions := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "*To override this weeks's BE maintained use:*\n/set-be-rotation *index*", false, false),
		nil,
		nil,
	)

	blocks = append(blocks, fe_instructions, be_instructions)

	return blocks
}

func SetFERotation(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	mu.Lock()
	defer mu.Unlock()

	blocks := []slack.Block{}
	indexStr := cmd.Text
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || (index >= len(FE_List)) {
		blocks = append(blocks, utils.MakeTextSectionBlock("Invalid index"))
		blocks = append(blocks, utils.MakeTextSectionBlock(fmt.Sprintf("Index must be between 0 and %d, to see lists and indexes use /get-rotation", len(FE_List)-1)))
		return blocks
	}

	FE_Index = index
	
	questionText := slack.NewTextBlockObject("mrkdwn", "Successfully updated!", false, false)
	yesButtonText := slack.NewTextBlockObject("plain_text", "Click here to send update", false, false)
	yesButton := slack.NewButtonBlockElement("send_rotation", "send_rotation", yesButtonText)
	questionSection := slack.NewSectionBlock(questionText, nil, slack.NewAccessory(yesButton))
	blocks = append(blocks, questionSection)

	return blocks
}

func SetBERotation(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	mu.Lock()
	defer mu.Unlock()

	blocks := []slack.Block{}
	indexStr := cmd.Text
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || (index >= len(BE_List)) {
		blocks = append(blocks, utils.MakeTextSectionBlock("Invalid index"))
	}

	BE_Index = index

	questionText := slack.NewTextBlockObject("mrkdwn", "Successfully updated!", false, false)
	yesButtonText := slack.NewTextBlockObject("plain_text", "Click here to send update", false, false)
	yesButton := slack.NewButtonBlockElement("send_rotation", "send_rotation", yesButtonText)
	questionSection := slack.NewSectionBlock(questionText, nil, slack.NewAccessory(yesButton))
	blocks = append(blocks, questionSection)

	return blocks
} 

func PostCurrentRotation() {
	channelID := os.Getenv("SCHEDULER_CHANNEL_ID")
	if channelID == "" {
		log.Fatalf("SCHEDULER_CHANNEL_ID environment variable is not set")
	}

	mu.Lock()
	defer mu.Unlock()
	
	FE_User := "No users"
	if len(FE_List) > 0 {
		FE_User = fmt.Sprintf("<@%s>", FE_List[FE_Index].ID)
	}

	BE_User := "No users"
	if len(BE_List) > 0 {
		BE_User = fmt.Sprintf("<@%s>", BE_List[BE_Index].ID)
	}

	message := fmt.Sprintf(
		"This week's sentry maintainers have been updated :saluting_face:\n%s from the FE team\n%s from the BE team",
		FE_User, BE_User,
	)
	err := postMessageToSlack(channelID, message)
	if err != nil {
		log.Fatalf("Error posting message: %v", err)
	}
}