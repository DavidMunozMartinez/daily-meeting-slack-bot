package slackapicommandhandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	gitlabAPIURL = "https://gitlab.com/api/v4"
	projectURL   = "https://gitlab.com/ggallagher/api"
)

func sendGetRequest(url string, privateToken string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("PRIVATE-TOKEN", privateToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	return body, nil
}

func getEnvironments() ([]map[string]interface{}, error) {
	var privateToken = os.Getenv("GITLAB_PRIVATE_TOKEN")
	var projectID = os.Getenv("GITLAB_PROJECT_ID")
	url := fmt.Sprintf("%s/projects/%s/environments", gitlabAPIURL, projectID)

	body, err := sendGetRequest(url, privateToken)
	catchErr(err)

	var environments []map[string]interface{}
	err = json.Unmarshal(body, &environments)
	return environments, err
}

func getDeployments(perPage int, maxPages int) ([]map[string]interface{}, error) {
	var privateToken = os.Getenv("GITLAB_PRIVATE_TOKEN")
	var projectID = os.Getenv("GITLAB_PROJECT_ID")

	var deployments []map[string]interface{}

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s/projects/%s/deployments?per_page=%d&page=%d&order_by=created_at&sort=desc", gitlabAPIURL, projectID, perPage, page)
		body, err := sendGetRequest(url, privateToken)
		catchErr(err)

		var pageDeployments []map[string]interface{}
		err = json.Unmarshal(body, &pageDeployments)
		catchErr(err)

		if len(pageDeployments) == 0 {
			break
		}

		deployments = append(deployments, pageDeployments...)
	}
	return deployments, nil
}

func findFirstMatchingDeployment(deployments []map[string]interface{}, name string) map[string]interface{} {
	for _, deployment := range deployments {
		if deployment["deployable"].(map[string]interface{})["name"] == name {
			return deployment
		}
	}
	return nil
}

func catchErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

// filter environmentDeployments based on environmentID
func filterEnvironmentDeployments(environmentDeployments []map[string]interface{}, environmentID string) []map[string]interface{} {
	var filteredDeployments []map[string]interface{}
	for _, deployment := range environmentDeployments {
		deploymentEnvID := fmt.Sprintf("%v", deployment["environment"].(map[string]interface{})["id"])
		deploymentStatus := fmt.Sprintf("%v", deployment["deployable"].(map[string]interface{})["status"])
		if deploymentEnvID == environmentID && deploymentStatus == "success" {
			filteredDeployments = append(filteredDeployments, deployment)
		}
	}
	return filteredDeployments
}

// getUniqueDeployableNames returns a map of unique deployable names
func getUniqueDeployableNames(environmentDeployments []map[string]interface{}) map[string]bool {
	var uniqueDeployableNames = make(map[string]bool)
	for _, deployment := range environmentDeployments {
		name := fmt.Sprintf("%v", deployment["deployable"].(map[string]interface{})["name"])
		uniqueDeployableNames[name] = true
	}
	return uniqueDeployableNames
}

// identify if it is a merge request or a branch and return the source URL
func getSourceUrlBasedOnRef(ref string) string {
	if strings.HasPrefix(ref, "refs/merge-requests/") {
		return fmt.Sprintf("%v/merge_requests/%v", projectURL, strings.Split(ref, "/")[2])
	}
	return fmt.Sprintf("%v/-/tree/%v", projectURL, strings.Replace(ref, "refs/heads/", "", 1))
}

// format the time to human readable format
func getHumanReadableTime(timeString string) string {
	location, _ := time.LoadLocation("America/Los_Angeles")
	t, err := time.Parse(time.RFC3339, timeString)
	catchErr(err)

	t = t.In(location)
	return t.Format("January 02, 2006, 15:04:05")
}

func filterEnvironments(environments []map[string]interface{}, text string) []map[string]interface{} {
	var filteredEnvironments []map[string]interface{}
	for _, environment := range environments {
		if strings.Contains(fmt.Sprintf("%v", environment["name"]), text) {
			filteredEnvironments = append(filteredEnvironments, environment)
		}
	}
	return filteredEnvironments
}

func GetAPIStatus(cmd slack.SlashCommand, client *socketmode.Client) []slack.Block {
	environments, err := getEnvironments()
	catchErr(err)

	deployments, err := getDeployments(100, 1)
	catchErr(err)

	var blocks []slack.Block

	if cmd.Text != "" {
		environments = filterEnvironments(environments, cmd.Text)
		fmt.Println("environments: ", environments)
	}

	for _, environment := range environments {

		environmentID := fmt.Sprintf("%v", environment["id"])
		var environmentDeployments = filterEnvironmentDeployments(deployments, environmentID)
		var uniqueDeployableNames = getUniqueDeployableNames(environmentDeployments)

		if len(environmentDeployments) > 0 {
			for deployName := range uniqueDeployableNames {

				matchingDeployment := findFirstMatchingDeployment(environmentDeployments, deployName)
				if matchingDeployment != nil {

					var sourceUrl = getSourceUrlBasedOnRef(fmt.Sprintf("%v", matchingDeployment["deployable"].(map[string]interface{})["ref"]))
					var createdAtRedable = getHumanReadableTime(fmt.Sprintf("%v", matchingDeployment["created_at"]))

					var formattedDeployment = map[string]interface{}{
						"env":          strings.Replace(fmt.Sprintf("%v", matchingDeployment["deployable"].(map[string]interface{})["name"]), "deploy-job-", "", 1),
						"user":         fmt.Sprintf("%v", matchingDeployment["user"].(map[string]interface{})["name"]),
						"created_at":   fmt.Sprintf("%v (America/Los_Angeles)", createdAtRedable),
						"pipeline_url": fmt.Sprintf("%v", matchingDeployment["deployable"].(map[string]interface{})["pipeline"].(map[string]interface{})["web_url"]),
						"source_url":   sourceUrl,
					}

					//jFormattedDeployment, _ := json.MarshalIndent(formattedDeployment, "", "\t")
					//fmt.Println(string(jFormattedDeployment))

					deploymentText := fmt.Sprintf("*%s*\n> User: %s\n> Created At: %s\n> [Pipeline URL](%s)\n> [Source URL](%s)",
						formattedDeployment["env"], formattedDeployment["user"], formattedDeployment["created_at"], formattedDeployment["pipeline_url"], formattedDeployment["source_url"])

					blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", deploymentText, false, false), nil, nil))
				}
			}
		}
	}

	return blocks
}
