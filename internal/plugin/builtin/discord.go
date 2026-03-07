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

// DiscordPlugin implements a Discord webhook notify plugin.
type DiscordPlugin struct {
	webhookURL string
	notifyMap  map[string]bool
	client     *http.Client
}

// NewDiscordPlugin creates a Discord notify plugin.
func NewDiscordPlugin(url string, notifyEvents []string) *DiscordPlugin {
	m := make(map[string]bool)
	for _, e := range notifyEvents {
		m[e] = true
	}
	return &DiscordPlugin{
		webhookURL: url,
		notifyMap:  m,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (d *DiscordPlugin) Name() string    { return "discord-notify" }
func (d *DiscordPlugin) Version() string { return "1.0.0" }
func (d *DiscordPlugin) OnLoad() error   { return nil }
func (d *DiscordPlugin) OnUnload() error { return nil }

func (d *DiscordPlugin) OnBuildEnd(b *core.Build) {
	if !d.notifyMap["build_end"] && !d.notifyMap["build_error"] {
		return
	}
	if b.State == core.StateSuccess && !d.notifyMap["build_end"] {
		return
	}
	if b.State == core.StateFailed && !d.notifyMap["build_error"] {
		return
	}

	color := 3066993 // green
	title := "Build Succeeded"
	desc := fmt.Sprintf("*%s* (%s) finished in %.1fs", b.Name, b.Tool, b.Duration.Seconds())

	if b.State == core.StateFailed {
		color = 15158332 // red
		title = "Build Failed"
		desc = fmt.Sprintf("*%s* (%s) failed in %.1fs with %d errors.", b.Name, b.Tool, b.Duration.Seconds(), len(b.Errors))
	}

	msg := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": desc,
				"color":       color,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}
	d.post(msg)
}

func (d *DiscordPlugin) post(payload interface{}) {
	b, _ := json.Marshal(payload)
	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Printf("[discord-notify] POST error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("[discord-notify] bad status code: %d", resp.StatusCode)
	}
}
