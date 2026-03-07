package core

// BuildProfile is a saved build preset.
type BuildProfile struct {
	Name    string            `toml:"name"    json:"name"`
	WorkDir string            `toml:"workdir" json:"workdir"`
	Command string            `toml:"command" json:"command"`
	Tags    []string          `toml:"tags"    json:"tags"`
	Env     map[string]string `toml:"env"     json:"env"`
}
