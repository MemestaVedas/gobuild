package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/config"
)

// DaemonServer listens on a UNIX domain socket for proxied commands.
type DaemonServer struct {
	SocketPath string
	listener   net.Listener
	onCommand  func(cmd string, args []string, cwd string) error
	quit       chan struct{}
}

// ProxyRequest represents a payload sent from the shell wrapper to the daemon.
type ProxyRequest struct {
	CWD     string   `json:"cwd"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// NewDaemonServer initializes the local daemon server hook.
func NewDaemonServer(onCommand func(cmd string, args []string, cwd string) error) *DaemonServer {
	return &DaemonServer{
		SocketPath: filepath.Join(config.ConfigDir(), "daemon.sock"),
		onCommand:  onCommand,
		quit:       make(chan struct{}),
	}
}

// Start opens the socket and begins listening for shell wrapper connections.
func (d *DaemonServer) Start() error {
	// Clean up old socket if it exists
	if _, err := os.Stat(d.SocketPath); err == nil {
		os.Remove(d.SocketPath)
	}

	l, err := net.Listen("unix", d.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on daemon socket %s: %w", d.SocketPath, err)
	}
	d.listener = l
	
	// Ensure permissions are open only to the user
	os.Chmod(d.SocketPath, 0600)

	log.Printf("Daemon server listening on %s", d.SocketPath)

	go d.acceptLoop()
	return nil
}

// acceptLoop handles incoming connections.
func (d *DaemonServer) acceptLoop() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.quit:
				return
			default:
				log.Printf("Daemon accept error: %v", err)
				continue
			}
		}

		go d.handleConnection(conn)
	}
}

// handleConnection reads a ProxyRequest and triggers the callback.
func (d *DaemonServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var req ProxyRequest
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode daemon request: %v", err)
		return
	}

	if d.onCommand != nil {
		// Log internal for now, this should be piped to PTY manager
		log.Printf("Daemon received proxied command: %s %v from %s", req.Command, req.Args, req.CWD)
		
		// Note: At the moment this is basic IPC. We need to attach standard I/O to a PTY 
		// if we want to pipe back to the client. For standard "run and monitor in gobuild",
		// this triggers goBuild PTY manager and returns success to client.
		
		if err := d.onCommand(req.Command, req.Args, req.CWD); err != nil {
			fmt.Fprintf(conn, "ERROR: %v\n", err)
		} else {
			fmt.Fprintf(conn, "SUCCESS: Command spawned in goBuild dashboard.\n")
		}
	}
}

// Stop shuts down the daemon server.
func (d *DaemonServer) Stop() {
	close(d.quit)
	if d.listener != nil {
		d.listener.Close()
	}
	os.Remove(d.SocketPath)
}

// SendProxyRequest is a utility for the `gobuild proxy` client command to send to daemon.
func SendProxyRequest(cwd, command string, args []string) error {
	socketPath := filepath.Join(config.ConfigDir(), "daemon.sock")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("could not connect to gobuild daemon (is it running?): %w", err)
	}
	defer conn.Close()

	req := ProxyRequest{
		CWD:     cwd,
		Command: command,
		Args:    args,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&req); err != nil {
		return fmt.Errorf("failed to send proxy request: %w", err)
	}

	// Wait for response
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading response from daemon: %w", err)
	}

	resp = strings.TrimSpace(resp)
	if strings.HasPrefix(resp, "ERROR:") {
		return fmt.Errorf("daemon error: %s", resp)
	}

	// Print daemon success standard message to stdout
	fmt.Println(resp)
	return nil
}
