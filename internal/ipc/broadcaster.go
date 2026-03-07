package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// Broadcaster sends UDP beacon packets for Android app local network discovery.
type Broadcaster struct {
	port       int
	wsPort     int
	deviceName string
	stop       chan struct{}
}

// NewBroadcaster creates a UDP discovery beacon service.
func NewBroadcaster(port, wsPort int) *Broadcaster {
	host, _ := os.Hostname()
	return &Broadcaster{
		port:       port,
		wsPort:     wsPort,
		deviceName: host,
		stop:       make(chan struct{}),
	}
}

// Start spawns a goroutine that sends a beacon every 2 seconds.
func (b *Broadcaster) Start() {
	go func() {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", b.port))
		if err != nil {
			log.Printf("UDP invalid address: %v", err)
			return
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Printf("UDP dial failed: %v", err)
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ip := getLocalIP()
				msg := BeaconMessage{
					Type:    "beacon",
					Host:    ip,
					Port:    b.wsPort,
					Version: "1.0",
					Name:    b.deviceName,
				}
				data, err := json.Marshal(msg)
				if err == nil {
					_, _ = conn.Write(data)
				}
			case <-b.stop:
				return
			}
		}
	}()
}

// Stop halts the UDP broadcasting goroutine.
func (b *Broadcaster) Stop() {
	close(b.stop)
}

// getLocalIP attempts to find the primary non-loopback IP address.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
