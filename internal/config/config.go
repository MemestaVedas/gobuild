package config

import (
	"fmt"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config holds the complete application configuration.
type Config struct {
	Server        ServerConfig            `mapstructure:"server"`
	UI            UIConfig                `mapstructure:"ui"`
	Notifications NotificationConfig      `mapstructure:"notifications"`
	Editor        EditorConfig            `mapstructure:"editor"`
	Watch         WatchConfig             `mapstructure:"watch"`
	Plugins       map[string]PluginConfig `mapstructure:"plugins"`
}

type ServerConfig struct {
	WSPort   int    `mapstructure:"ws_port"`
	UDPPort  int    `mapstructure:"udp_port"`
	Hostname string `mapstructure:"hostname"`
}

type UIConfig struct {
	Theme       string `mapstructure:"theme"`
	LogMaxLines int    `mapstructure:"log_max_lines"`
	HistoryMax  int    `mapstructure:"history_max"`
}

type NotificationConfig struct {
	Desktop bool `mapstructure:"desktop"`
	Sound   bool `mapstructure:"sound"`
}

type EditorConfig struct {
	Command string `mapstructure:"command"`
}

type WatchConfig struct {
	Directories []WatchedDir `mapstructure:"directories"`
}

type WatchedDir struct {
	Path     string   `mapstructure:"path"`
	Commands []string `mapstructure:"commands"`
}

type PluginConfig struct {
	Enabled    bool              `mapstructure:"enabled"`
	WebhookURL string            `mapstructure:"webhook_url"`
	NotifyOn   []string          `mapstructure:"notify_on"`
	Extra      map[string]string `mapstructure:",remain"`
}

// DefaultConfig returns the compiled-in defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			WSPort:  7712,
			UDPPort: 7713,
		},
		UI: UIConfig{
			Theme:       "aurora-pastel",
			LogMaxLines: 500,
			HistoryMax:  50,
		},
		Notifications: NotificationConfig{
			Desktop: true,
			Sound:   true,
		},
		Watch: WatchConfig{
			Directories: []WatchedDir{},
		},
		Plugins: map[string]PluginConfig{
			"git-detect":     {Enabled: true},
			"sound-alert":    {Enabled: true},
			"slack-notify":   {Enabled: false},
			"discord-notify": {Enabled: false},
		},
	}
}

// ConfigDir returns the platform-appropriate config directory.
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "gobuild")
	}
	// Linux / macOS: XDG or fallback
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configHome, "gobuild")
}

// Load reads configuration from the config directory, applying defaults.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.ws_port", 7712)
	v.SetDefault("server.udp_port", 7713)
	v.SetDefault("server.hostname", "")
	v.SetDefault("ui.theme", "aurora-pastel")
	v.SetDefault("ui.log_max_lines", 500)
	v.SetDefault("ui.history_max", 50)
	v.SetDefault("notifications.desktop", true)
	v.SetDefault("notifications.sound", true)
	v.SetDefault("editor.command", "")
	v.SetDefault("plugins.git-detect.enabled", true)
	v.SetDefault("plugins.sound-alert.enabled", true)
	v.SetDefault("plugins.slack-notify.enabled", false)
	v.SetDefault("plugins.discord-notify.enabled", false)

	// Config file
	dir := ConfigDir()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(dir)
	v.AddConfigPath(".")

	// Env var overrides: GOBUILD_SERVER_WS_PORT etc.
	v.SetEnvPrefix("GOBUILD")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// Config file not found — use defaults only
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	loadWatchConfig(cfg)
	return cfg, nil
}

func loadWatchConfig(cfg *Config) {
	// We load the watch configuration from a separate watch.json file
	// to make it easier to manage purely for watched directories and commands
	// independent of the main UI/server toml config.
	watchFile := filepath.Join(ConfigDir(), "watch.json")
	
	data, err := os.ReadFile(watchFile)
	if err == nil {
		var watchCfg WatchConfig
		if err := json.Unmarshal(data, &watchCfg); err == nil {
			cfg.Watch = watchCfg
		}
	}
}

// SaveWatchConfig saves the watch configuration to watch.json
func SaveWatchConfig(cfg *Config) error {
	watchFile := filepath.Join(ConfigDir(), "watch.json")
	data, err := json.MarshalIndent(cfg.Watch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal watch config: %w", err)
	}
	
	if err := os.WriteFile(watchFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write watch config: %w", err)
	}
	return nil
}

// EnsureConfigDir creates the config directory if it does not exist.
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}
