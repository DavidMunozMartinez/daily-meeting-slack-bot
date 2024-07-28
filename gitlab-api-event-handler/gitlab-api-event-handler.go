package gitlabapieventhandler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/xanzy/go-gitlab"
)

func Init() {
	var provateToken = os.Getenv("GITLAB_PRIVATE_TOKEN")
	var projectID = os.Getenv("GITLAB_PROJECT_ID")

	client, err := gitlab.NewClient(provateToken)
	if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Get project events
    events, _, err := client.Events.ListProjectVisibleEvents(projectID, &gitlab.ListProjectVisibleEventsOptions{})
    if err != nil {
        log.Fatalf("Failed to list project events: %v", err)
    }

	// Information on which gitlab events are visible to our project
    for _, event := range events {
        fmt.Printf("Event: %s\n", event.ActionName)
    }

	options := &gitlab.AddProjectHookOptions{
        URL:                   gitlab.Ptr("https://your-webhook-url/gitlab-api"),
        PushEvents:            gitlab.Ptr(true),
        IssuesEvents:          gitlab.Ptr(true),
        MergeRequestsEvents:   gitlab.Ptr(true),
        TagPushEvents:         gitlab.Ptr(true),
        NoteEvents:            gitlab.Ptr(true),
        PipelineEvents:        gitlab.Ptr(true),
        WikiPageEvents:        gitlab.Ptr(true),
        EnableSSLVerification: gitlab.Ptr(false),
    }

	hook, _, err := client.Projects.AddProjectHook(projectID, options)
    if err != nil {
        log.Fatalf("Failed to add project hook: %v", err)
    }

    fmt.Printf("Webhook created: %v\n", hook)

    // Start an HTTP server to handle webhook events
    http.HandleFunc("/gitlab-api", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Webhook received")
        // Process the event here
    })

    log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}