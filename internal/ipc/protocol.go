package ipc

import "github.com/MemestaVedas/gobuild/internal/core"

// MessageType defines the kinds of messages sent over WebSocket.
type MessageType string

const (
	MsgHello       MessageType = "hello"
	MsgBuildUpdate MessageType = "build_update"
	MsgBuildStart  MessageType = "build_start"
	MsgBuildEnd    MessageType = "build_end"
	MsgStatsUpdate MessageType = "stats_update"
	MsgRunBuild    MessageType = "run_build"
	MsgKillBuild   MessageType = "kill_build"
	MsgPing        MessageType = "ping"
	MsgPong        MessageType = "pong"
)

// BaseMessage is the structure all WebSocket messages share.
type BaseMessage struct {
	Type MessageType `json:"type"`
}

// BeaconMessage is sent over UDP for local network discovery.
type BeaconMessage struct {
	Type    string `json:"type"` // always "beacon"
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	Name    string `json:"name"`
}

// HelloMessage provides full state on initial connect.
type HelloMessage struct {
	Type   MessageType   `json:"type"`
	Builds []*HelloBuild `json:"builds"`
}

type HelloBuild struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Tool     string  `json:"tool"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"`
	ElapsedS float64 `json:"elapsed_s"`
}

// BuildUpdateMessage is sent periodically for active builds.
type BuildUpdateMessage struct {
	Type     MessageType       `json:"type"`
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Tool     string            `json:"tool"`
	State    string            `json:"state"`
	Progress float64           `json:"progress"`
	ElapsedS float64           `json:"elapsed_s"`
	LogTail  []string          `json:"log_tail"`
	Errors   []core.BuildError `json:"errors"`
}

// BuildStartMessage notifies clients a build began.
type BuildStartMessage struct {
	Type MessageType `json:"type"`
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Tool string      `json:"tool"`
}

// BuildEndMessage notifies clients a build finished.
type BuildEndMessage struct {
	Type         MessageType `json:"type"`
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	State        string      `json:"state"` // success or failed
	DurationS    float64     `json:"duration_s"`
	ErrorCount   int         `json:"error_count"`
	WarningCount int         `json:"warning_count"`
}

// StatsUpdateMessage provides current PC resource utilisation.
type StatsUpdateMessage struct {
	Type  MessageType `json:"type"`
	CPU   float64     `json:"cpu"`
	RAM   uint64      `json:"ram"` // used bytes
	NetUp uint64      `json:"net_up"`
	NetDn uint64      `json:"net_dn"`
}

// ClientRunBuildMessage is received from Android to start a build.
type ClientRunBuildMessage struct {
	Type    MessageType `json:"type"`
	Profile string      `json:"profile"`
}

// ClientKillBuildMessage is received from Android to terminate a build.
type ClientKillBuildMessage struct {
	Type MessageType `json:"type"`
	ID   string      `json:"id"`
}
