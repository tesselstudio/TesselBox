package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	DiscoveryPort         uint16 = 25566
	DiscoveryInterval            = 2 * time.Second
	DiscoveryTimeout             = 5 * time.Second
	DiscoveryQueryTimeout        = 10 * time.Second
)

type DiscoveryQueryData struct {
	QueryID string
}

type DiscoveredServer struct {
	Info     *ServerInfoData
	LastSeen time.Time
	Address  string
}

type DiscoveryServer struct {
	conn         *net.UDPConn
	closed       bool
	closeMu      sync.RWMutex
	stopChan     chan struct{}
	serverPort   uint16
	serverName   string
	maxPlayers   uint8
	playerCount  func() uint8
}

func NewDiscoveryServer(serverPort uint16, serverName string, maxPlayers uint8) (*DiscoveryServer, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve discovery address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on discovery port: %w", err)
	}

	ds := &DiscoveryServer{
		conn:       conn,
		stopChan:   make(chan struct{}),
		serverPort: serverPort,
		serverName: serverName,
		maxPlayers: maxPlayers,
	}

	go ds.listenLoop()
	log.Printf("Discovery server started on port %d", DiscoveryPort)
	return ds, nil
}

func (ds *DiscoveryServer) listenLoop() {
	buffer := make([]byte, 1024)
	for {
		ds.closeMu.RLock()
		if ds.closed {
			ds.closeMu.RUnlock()
			return
		}
		ds.closeMu.RUnlock()

		ds.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, addr, err := ds.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			ds.closeMu.RLock()
			closed := ds.closed
			ds.closeMu.RUnlock()
			if !closed {
				log.Printf("Discovery listen error: %v", err)
			}
			return
		}

		if n > 0 {
			packet, err := DeserializePacket(buffer[:n])
			if err != nil {
				continue
			}
			if packet.Header.Type == PacketTypeDiscoveryQuery {
				ds.handleQuery(addr)
			}
		}
	}
}

func (ds *DiscoveryServer) handleQuery(addr *net.UDPAddr) {
	playerCount := uint8(0)
	if ds.playerCount != nil {
		playerCount = ds.playerCount()
	}

	info := &ServerInfoData{
		Name:        ds.serverName,
		PlayerCount: playerCount,
		MaxPlayers:  ds.maxPlayers,
		Address:     addr.IP.String(),
		Port:        ds.serverPort,
	}

	response := NewPacket(PacketTypeDiscoveryResponse, "discovery")
	payload, _ := SerializeServerInfo(info)
	response.Payload = payload

	data, _ := response.Serialize()
	ds.conn.WriteToUDP(data, addr)
}

func (ds *DiscoveryServer) SetPlayerCountFunc(fn func() uint8) {
	ds.playerCount = fn
}

func (ds *DiscoveryServer) Close() error {
	ds.closeMu.Lock()
	defer ds.closeMu.Unlock()
	if ds.closed {
		return nil
	}
	ds.closed = true
	close(ds.stopChan)
	return ds.conn.Close()
}

type DiscoveryClient struct {
	conn          *net.UDPConn
	closed        bool
	closeMu       sync.RWMutex
	stopChan      chan struct{}
	servers       map[string]*DiscoveredServer
	serversMu     sync.RWMutex
	OnServerFound func(server *DiscoveredServer)
	OnServerLost  func(address string)
}

func NewDiscoveryClient() (*DiscoveryClient, error) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve discovery address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery socket: %w", err)
	}

	dc := &DiscoveryClient{
		conn:     conn,
		stopChan: make(chan struct{}),
		servers:  make(map[string]*DiscoveredServer),
	}

	go dc.listenLoop()
	return dc, nil
}

func (dc *DiscoveryClient) listenLoop() {
	buffer := make([]byte, 1024)
	for {
		dc.closeMu.RLock()
		if dc.closed {
			dc.closeMu.RUnlock()
			return
		}
		dc.closeMu.RUnlock()

		dc.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, addr, err := dc.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			dc.closeMu.RLock()
			closed := dc.closed
			dc.closeMu.RUnlock()
			if !closed {
				log.Printf("Discovery client listen error: %v", err)
			}
			return
		}

		if n > 0 {
			packet, err := DeserializePacket(buffer[:n])
			if err != nil {
				continue
			}
			if packet.Header.Type == PacketTypeDiscoveryResponse {
				dc.handleResponse(packet.Payload, addr)
			}
		}
	}
}

func (dc *DiscoveryClient) handleResponse(payload []byte, addr *net.UDPAddr) {
	info, err := DeserializeServerInfo(payload)
	if err != nil {
		return
	}

	serverAddr := fmt.Sprintf("%s:%d", addr.IP.String(), info.Port)

	dc.serversMu.Lock()
	defer dc.serversMu.Unlock()

	existing, found := dc.servers[serverAddr]
	if found {
		existing.LastSeen = time.Now()
		existing.Info = info
	} else {
		server := &DiscoveredServer{
			Info:     info,
			LastSeen: time.Now(),
			Address:  serverAddr,
		}
		dc.servers[serverAddr] = server
		log.Printf("Discovered server: %s at %s (%d/%d players)",
			info.Name, serverAddr, info.PlayerCount, info.MaxPlayers)
		if dc.OnServerFound != nil {
			go dc.OnServerFound(server)
		}
	}
}

func (dc *DiscoveryClient) Query() error {
	dc.closeMu.RLock()
	if dc.closed {
		dc.closeMu.RUnlock()
		return fmt.Errorf("discovery client closed")
	}
	dc.closeMu.RUnlock()

	broadcastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	if err != nil {
		return fmt.Errorf("failed to resolve broadcast address: %w", err)
	}

	query := NewPacket(PacketTypeDiscoveryQuery, GenerateID())
	data, _ := query.Serialize()

	_, err = dc.conn.WriteToUDP(data, broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to send discovery query: %w", err)
	}
	return nil
}

func (dc *DiscoveryClient) StartDiscovery() {
	go dc.discoveryLoop()
}

func (dc *DiscoveryClient) discoveryLoop() {
	ticker := time.NewTicker(DiscoveryInterval)
	defer ticker.Stop()
	dc.Query()
	for {
		select {
		case <-dc.stopChan:
			return
		case <-ticker.C:
			dc.Query()
			dc.cleanupStaleServers()
		}
	}
}

func (dc *DiscoveryClient) cleanupStaleServers() {
	dc.serversMu.Lock()
	defer dc.serversMu.Unlock()

	now := time.Now()
	for addr, server := range dc.servers {
		if now.Sub(server.LastSeen) > DiscoveryQueryTimeout {
			delete(dc.servers, addr)
			log.Printf("Lost server: %s", addr)
			if dc.OnServerLost != nil {
				go dc.OnServerLost(addr)
			}
		}
	}
}

func (dc *DiscoveryClient) GetServers() []*DiscoveredServer {
	dc.serversMu.RLock()
	defer dc.serversMu.RUnlock()

	servers := make([]*DiscoveredServer, 0, len(dc.servers))
	for _, server := range dc.servers {
		servers = append(servers, server)
	}
	return servers
}

func (dc *DiscoveryClient) Close() error {
	dc.closeMu.Lock()
	defer dc.closeMu.Unlock()
	if dc.closed {
		return nil
	}
	dc.closed = true
	close(dc.stopChan)
	return dc.conn.Close()
}

func (dc *DiscoveryClient) IsClosed() bool {
	dc.closeMu.RLock()
	defer dc.closeMu.RUnlock()
	return dc.closed
}

type DiscoveryBrowser struct {
	client       *DiscoveryClient
	discoveredCB func([]*DiscoveredServer)
	servers      []*DiscoveredServer
	serversMu    sync.RWMutex
}

func NewDiscoveryBrowser() (*DiscoveryBrowser, error) {
	client, err := NewDiscoveryClient()
	if err != nil {
		return nil, err
	}

	db := &DiscoveryBrowser{client: client}
	client.OnServerFound = func(server *DiscoveredServer) {
		db.addServer(server)
	}
	client.OnServerLost = func(address string) {
		db.removeServer(address)
	}
	return db, nil
}

func (db *DiscoveryBrowser) Start() {
	db.client.StartDiscovery()
}

func (db *DiscoveryBrowser) Stop() {
	db.client.Close()
}

func (db *DiscoveryBrowser) Refresh() error {
	return db.client.Query()
}

func (db *DiscoveryBrowser) GetServers() []*DiscoveredServer {
	db.serversMu.RLock()
	defer db.serversMu.RUnlock()
	result := make([]*DiscoveredServer, len(db.servers))
	copy(result, db.servers)
	return result
}

func (db *DiscoveryBrowser) OnDiscovered(callback func([]*DiscoveredServer)) {
	db.discoveredCB = callback
}

func (db *DiscoveryBrowser) addServer(server *DiscoveredServer) {
	db.serversMu.Lock()
	defer db.serversMu.Unlock()
	for _, s := range db.servers {
		if s.Address == server.Address {
			s.Info = server.Info
			s.LastSeen = server.LastSeen
			return
		}
	}
	db.servers = append(db.servers, server)
	if db.discoveredCB != nil {
		db.discoveredCB(db.getServersCopy())
	}
}

func (db *DiscoveryBrowser) removeServer(address string) {
	db.serversMu.Lock()
	defer db.serversMu.Unlock()
	for i, s := range db.servers {
		if s.Address == address {
			db.servers = append(db.servers[:i], db.servers[i+1:]...)
			break
		}
	}
	if db.discoveredCB != nil {
		db.discoveredCB(db.getServersCopy())
	}
}

func (db *DiscoveryBrowser) getServersCopy() []*DiscoveredServer {
	result := make([]*DiscoveredServer, len(db.servers))
	copy(result, db.servers)
	return result
}
