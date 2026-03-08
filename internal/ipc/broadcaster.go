package ipc

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

		ip := getLocalIP()
		log.Printf("Broadcasting discovery beacon on %s:%d (deviceName: %s)", ip, b.port, b.deviceName)

		for {
			select {
			case <-ticker.C:
				// Re-verify IP in case network changed (e.g. laptop move)
				currentIP := getLocalIP()
				// Android expects "GOBUILD_DISCOVERY:$IP"
				msg := fmt.Sprintf("GOBUILD_DISCOVERY:%s", currentIP)
				_, _ = conn.Write([]byte(msg))
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

// getLocalIP attempts to find the primary non-loopback IP address,
// prioritizing common local network subnets to avoid virtual bridges (WSL/Docker).
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	var candidates []string

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}

			ipStr := ip.String()
			// Prioritize common LAN subnets
			if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "172.16.") {
				return ipStr
			}
			candidates = append(candidates, ipStr)
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}
	return "127.0.0.1"
}
