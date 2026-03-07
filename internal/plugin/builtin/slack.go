package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// SlackPlugin implements a Slack webhook notify plugin.
type SlackPlugin struct {
	webhookURL string
	notifyMap  map[string]bool
	client     *http.Client
}

// NewSlackPlugin creates a Slack notify plugin.
func NewSlackPlugin(url string, notifyEvents []string) *SlackPlugin {
	m := make(map[string]bool)
	for _, e := range notifyEvents {
		m[e] = true
	}
	return &SlackPlugin{
		webhookURL: url,
		notifyMap:  m,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *SlackPlugin) Name() string    { return "slack-notify" }
func (s *SlackPlugin) Version() string { return "1.0.0" }
func (s *SlackPlugin) OnLoad() error   { return nil }
func (s *SlackPlugin) OnUnload() error { return nil }

func (s *SlackPlugin) OnBuildEnd(b *core.Build) {
	if !s.notifyMap["build_end"] && !s.notifyMap["build_error"] {
		return
	}
	if b.State == core.StateSuccess && !s.notifyMap["build_end"] {
		return
	}
	if b.State == core.StateFailed && !s.notifyMap["build_error"] {
		return
	}

	color := "#36a64f" // green
	title := "Build Succeeded"
	text := fmt.Sprintf("*%s* (%s) finished in %.1fs", b.Name, b.Tool, b.Duration.Seconds())

	if b.State == core.StateFailed {
		color = "#ff0000" // red
		title = "Build Failed"
		text = fmt.Sprintf("*%s* (%s) failed in %.1fs with %d errors.", b.Name, b.Tool, b.Duration.Seconds(), len(b.Errors))
	}

	msg := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fallback": text,
				"color":    color,
				"pretext":  title,
				"text":     text,
			},
		},
	}
	s.post(msg)
}

func (s *SlackPlugin) post(payload interface{}) {
	b, _ := json.Marshal(payload)
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Printf("[slack-notify] POST error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("[slack-notify] bad status code: %d", resp.StatusCode)
	}
}
