package plugin

import (
	"log"

	"github.com/MemestaVedas/gobuild/internal/config"
	"github.com/MemestaVedas/gobuild/internal/plugin/builtin"
)

// LoadBuiltins instantiates and registers enabled built-in plugins.
func LoadBuiltins(eb *EventBus, cfg map[string]config.PluginConfig) {
	if c, ok := cfg["slack-notify"]; ok && c.Enabled {
		if c.WebhookURL != "" {
			p := builtin.NewSlackPlugin(c.WebhookURL, c.NotifyOn)
			if err := p.OnLoad(); err == nil {
				eb.Register(p)
				log.Printf("Loaded builtin plugin: %s", p.Name())
			} else {
				log.Printf("Failed to load slack-notify: %v", err)
			}
		}
	}

	if c, ok := cfg["discord-notify"]; ok && c.Enabled {
		if c.WebhookURL != "" {
			p := builtin.NewDiscordPlugin(c.WebhookURL, c.NotifyOn)
			if err := p.OnLoad(); err == nil {
				eb.Register(p)
				log.Printf("Loaded builtin plugin: %s", p.Name())
			} else {
				log.Printf("Failed to load discord-notify: %v", err)
			}
		}
	}

	// sound-alert and git-detect are handled elsewhere or via special native logic
}
