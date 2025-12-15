package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/solvyd/solvyd/plugin-sdk/pkg/sdk"
)

// SlackNotifyPlugin implements notification plugin for Slack
type SlackNotifyPlugin struct {
	webhookURL string
	channel    string
	username   string
}

func (p *SlackNotifyPlugin) Name() string {
	return "slack-notification"
}

func (p *SlackNotifyPlugin) Version() string {
	return "1.0.0"
}

func (p *SlackNotifyPlugin) Type() string {
	return "notification"
}

func (p *SlackNotifyPlugin) Initialize(config map[string]interface{}) error {
	if url, ok := config["webhook_url"].(string); ok {
		p.webhookURL = url
	} else {
		return fmt.Errorf("webhook_url is required")
	}

	if channel, ok := config["channel"].(string); ok {
		p.channel = channel
	}

	if username, ok := config["username"].(string); ok {
		p.username = username
	} else {
		p.username = "Ritmo CI"
	}

	return nil
}

func (p *SlackNotifyPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	// Build notification message from context
	message := &sdk.NotificationMessage{
		Title:   fmt.Sprintf("Build %s", ctx.Parameters["status"]),
		Body:    fmt.Sprintf("Job: %s", ctx.Parameters["job_name"]),
		Level:   ctx.Parameters["level"].(string),
		BuildID: ctx.BuildID,
	}

	if err := p.Notify(message); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: err.Error(),
		}, err
	}

	return &sdk.Result{
		Success:  true,
		ExitCode: 0,
		Output:   "Slack notification sent successfully",
	}, nil
}

func (p *SlackNotifyPlugin) Notify(msg *sdk.NotificationMessage) error {
	color := p.getColor(msg.Level)

	payload := map[string]interface{}{
		"username": p.username,
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  msg.Title,
				"text":   msg.Body,
				"footer": "Ritmo CI",
				"ts":     time.Now().Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Build ID",
						"value": msg.BuildID,
						"short": true,
					},
					{
						"title": "Status",
						"value": msg.Status,
						"short": true,
					},
				},
			},
		},
	}

	if p.channel != "" {
		payload["channel"] = p.channel
	}

	if msg.URL != "" {
		payload["attachments"].([]map[string]interface{})[0]["title_link"] = msg.URL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(p.webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack notification failed with status %d", resp.StatusCode)
	}

	return nil
}

func (p *SlackNotifyPlugin) getColor(level string) string {
	switch level {
	case "success":
		return "good"
	case "warning":
		return "warning"
	case "error":
		return "danger"
	default:
		return "#36a64f"
	}
}

func (p *SlackNotifyPlugin) Cleanup() error {
	return nil
}

// Export the plugin
var Plugin SlackNotifyPlugin

// Main function for standalone testing
func main() {
	fmt.Println("Slack Notification Plugin v1.0.0")
	fmt.Println("This is a plugin and should be loaded by the Ritmo system")
	fmt.Println("To build as a plugin: go build -buildmode=plugin -o slack-notify.so")
}
