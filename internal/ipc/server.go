package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/gorilla/websocket"
)

// Server implements the WebSocket server pushing data to the Android companion.
type Server struct {
	port       int
	clients    map[*websocket.Conn]bool
	clientsMu  sync.Mutex
	bm         *core.BuildManager
	upgrader   websocket.Upgrader
	onRunBuild func(profileName string)
	onKill     func(id string)
}

// NewServer creates a new IPC Server.
func NewServer(port int, bm *core.BuildManager, onRun func(string), onKill func(string)) *Server {
	return &Server{
		port:       port,
		clients:    make(map[*websocket.Conn]bool),
		bm:         bm,
		onRunBuild: onRun,
		onKill:     onKill,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // allow connection from local mobile devices
			},
		},
	}
}

// Start opens the HTTP port and begins accepting connections on /ws.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", s.port)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("IPC Server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	s.sendHello(conn)

	// Keep-alive loop & read commands
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var base BaseMessage
		if err := json.Unmarshal(msg, &base); err != nil {
			continue
		}

		switch base.Type {
		case MsgPing:
			_ = conn.WriteJSON(BaseMessage{Type: MsgPong})
		case MsgRunBuild:
			var req ClientRunBuildMessage
			if err := json.Unmarshal(msg, &req); err == nil && s.onRunBuild != nil {
				s.onRunBuild(req.Profile)
			}
		case MsgKillBuild:
			var req ClientKillBuildMessage
			if err := json.Unmarshal(msg, &req); err == nil && s.onKill != nil {
				s.onKill(req.ID)
			}
		}
	}
}

func (s *Server) sendHello(conn *websocket.Conn) {
	active := s.bm.Active()
	hello := HelloMessage{
		Type:   MsgHello,
		Builds: make([]*HelloBuild, len(active)),
	}

	for i, b := range active {
		hello.Builds[i] = &HelloBuild{
			ID:       b.ID,
			Name:     b.Name,
			Tool:     b.Tool.String(),
			State:    b.State.String(),
			Progress: b.Progress,
			ElapsedS: b.Elapsed().Seconds(),
		}
	}

	_ = conn.WriteJSON(hello)
}

// Broadcast sends a JSON payload to all connected Android clients.
func (s *Server) Broadcast(msg any) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for conn := range s.clients {
		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		err := conn.WriteJSON(msg)
		if err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}
