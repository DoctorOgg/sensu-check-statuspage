package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

type StatusPage struct {
	Page struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		URL       string    `json:"url"`
		TimeZone  string    `json:"time_zone"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"page"`
	Incidents []struct {
		CreatedAt       string `json:"created_at"`
		ID              string `json:"id"`
		Impact          string `json:"impact"`
		IncidentUpdates []struct {
			Body       string `json:"body"`
			CreatedAt  string `json:"created_at"`
			DisplayAt  string `json:"display_at"`
			ID         string `json:"id"`
			IncidentID string `json:"incident_id"`
			Status     string `json:"status"`
			UpdatedAt  string `json:"updated_at"`
		} `json:"incident_updates"`
		MonitoringAt interface{} `json:"monitoring_at"`
		Name         string      `json:"name"`
		PageID       string      `json:"page_id"`
		ResolvedAt   interface{} `json:"resolved_at"`
		Shortlink    string      `json:"shortlink"`
		Status       string      `json:"status"`
		UpdatedAt    string      `json:"updated_at"`
	} `json:"incidents"`
}

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Url string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-statuspage",
			Short:    "Check for statuspage.com",
			Keyspace: "sensu.io/plugins/sensu-check-statuspage/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "url",
			Env:       "STATUS_PAGE_URL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "",
			Usage:     "URL of Statuspage to Monitor",
			Value:     &plugin.Url,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(plugin.Url) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--url or STATUS_PAGE_URL environment variable is required")
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	var incidentSummary string
	var currentStatus StatusPage

	res, err := http.Get(plugin.Url + "/api/v2/incidents/unresolved.json")
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&currentStatus)
	if err != nil {
		return sensu.CheckStateCritical, err
	}

	if len(currentStatus.Incidents) > 0 {
		for _, s := range currentStatus.Incidents {
			incidentSummary += strings.ToUpper(s.Impact) + ": " + s.Name + " (" + s.Shortlink + ") " + strings.ToUpper(s.Status) + "\n"
		}
		fmt.Printf("%v: Incidents: %v, Updated at: %v \n%v", currentStatus.Page.Name, len(currentStatus.Incidents), currentStatus.Page.UpdatedAt, incidentSummary)
		return sensu.CheckStateCritical, nil
	} else {
		fmt.Printf("%v: Incidents: %v, Updated at: %v\n", currentStatus.Page.Name, len(currentStatus.Incidents), currentStatus.Page.UpdatedAt)
		return sensu.CheckStateOK, nil
	}
}
