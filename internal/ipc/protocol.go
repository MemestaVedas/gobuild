package ipc

// MessageType defines the kinds of messages sent over WebSocket.
type MessageType string

const (
	MsgHello       MessageType = "hello"
	MsgBuildUpdate MessageType = "update"
	MsgBuildStart  MessageType = "build_start"
	MsgBuildEnd    MessageType = "finished"
	MsgBuildFailed MessageType = "failed"
	MsgStatsUpdate MessageType = "stats"
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

// HelloMessage is actually a full update in Android's eyes
type HelloMessage struct {
	Type   MessageType   `json:"type"`
	Builds []*HelloBuild `json:"builds"`
	CPU    float64       `json:"cpu"`
}

type HelloBuild struct {
	Project         string  `json:"project"`
	Tool            string  `json:"tool"`
	Status          string  `json:"status"`
	Progress        float64 `json:"progress"`
	PID             int     `json:"pid"`
	DurationSeconds int     `json:"duration_seconds"`
}

// BuildUpdateMessage matches Android's "update" type
type BuildUpdateMessage struct {
	Type        MessageType   `json:"type"`
	Builds      []*HelloBuild `json:"builds"`
	CPU         float64       `json:"cpu"`
	Timestamp   int64         `json:"timestamp"`
	ActiveCount int           `json:"active_count"`
}

// BuildStartMessage notifies clients a build began.
type BuildStartMessage struct {
	Type MessageType `json:"type"`
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Tool string      `json:"tool"`
}

// BuildEndMessage notifies clients a build finished/failed.
type BuildEndMessage struct {
	Type            MessageType `json:"type"`
	Project         string      `json:"project"`
	Tool            string      `json:"tool"`
	DurationSeconds int         `json:"duration_seconds"`
	Success         bool        `json:"success"`
	ErrorLine       string      `json:"error_line"`
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
