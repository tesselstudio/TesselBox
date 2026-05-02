package network

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultPort               uint16 = 25565
	MaxPlayers                uint8  = 16
	ClientTimeout                    = 30 * time.Second
	PositionBroadcastInterval        = 50 * time.Millisecond
)

type Server struct {
	ID              string
	Name            string
	Port            uint16
	MaxPlayersValue uint8
	udpConn         *UDPConn
	tcpListener     *TCPListener
	clients         map[string]*ClientSession
	clientsMu       sync.RWMutex
	playerCount     atomic.Uint32
	discoveryServer *DiscoveryServer
	blockChanges    chan *BlockChangeData
	chatMessages    chan *ChatMessageData
	stopChan        chan struct{}
	running         bool
	runningMu       sync.RWMutex
	worldSeed       int64
	startTime       time.Time
	OnPlayerJoin    func(session *ClientSession)
	OnPlayerLeave   func(session *ClientSession)
	OnBlockChange   func(data *BlockChangeData)
	OnChatMessage   func(data *ChatMessageData)
}

func NewServer(name string, port uint16, maxPlayers uint8) *Server {
	if maxPlayers == 0 {
		maxPlayers = MaxPlayers
	}
	if port == 0 {
		port = DefaultPort
	}
	return &Server{
		ID:              GenerateID(),
		Name:            name,
		Port:            port,
		MaxPlayersValue: maxPlayers,
		clients:         make(map[string]*ClientSession),
		blockChanges:    make(chan *BlockChangeData, 100),
		chatMessages:    make(chan *ChatMessageData, 100),
		stopChan:        make(chan struct{}),
		startTime:       time.Now(),
	}
}

func (s *Server) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		return fmt.Errorf("server already running")
	}

	udpConn, err := NewUDPListener(s.Port)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	s.udpConn = udpConn

	tcpListener, err := NewTCPListener(s.Port)
	if err != nil {
		s.udpConn.Close()
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	s.tcpListener = tcpListener
	s.tcpListener.OnConnect(s.handleTCPConnection)

	discoveryServer, err := NewDiscoveryServer(s.Port, s.Name, s.MaxPlayersValue)
	if err != nil {
		s.udpConn.Close()
		s.tcpListener.Close()
		return fmt.Errorf("failed to start discovery server: %w", err)
	}
	s.discoveryServer = discoveryServer

	s.running = true
	go s.udpReceiveLoop()
	go s.broadcastLoop()
	go s.cleanupLoop()

	log.Printf("Server '%s' started on port %d (UDP/TCP)", s.Name, s.Port)
	return nil
}

func (s *Server) Stop() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if !s.running {
		return nil
	}

	s.running = false
	close(s.stopChan)

	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Disconnect()
	}
	s.clients = make(map[string]*ClientSession)
	s.clientsMu.Unlock()

	if s.discoveryServer != nil {
		s.discoveryServer.Close()
	}
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}
	if s.udpConn != nil {
		s.udpConn.Close()
	}

	log.Printf("Server '%s' stopped", s.Name)
	return nil
}

func (s *Server) IsRunning() bool {
	s.runningMu.RLock()
	defer s.runningMu.RUnlock()
	return s.running
}

func (s *Server) GetPlayerCount() uint8 {
	return uint8(s.playerCount.Load())
}

func (s *Server) GetServerInfo() *ServerInfoData {
	return &ServerInfoData{
		Name:        s.Name,
		PlayerCount: s.GetPlayerCount(),
		MaxPlayers:  s.MaxPlayersValue,
		Address:     "",
		Port:        s.Port,
	}
}

func (s *Server) GetClients() []*ClientSession {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	clients := make([]*ClientSession, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *Server) udpReceiveLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		case msg := <-s.udpConn.Receive():
			if msg == nil {
				return
			}
			s.handleUDPPacket(msg)
		}
	}
}

func (s *Server) handleUDPPacket(msg *UDPMessage) {
	packet, err := DeserializePacket(msg.Data)
	if err != nil {
		log.Printf("Failed to deserialize UDP packet: %v", err)
		return
	}

	switch packet.Header.Type {
	case PacketTypePlayerPosition:
		s.handlePlayerPosition(packet, msg.Addr)
	case PacketTypePing:
		s.handlePing(packet, msg.Addr)
	}
}

func (s *Server) handleTCPConnection(conn *TCPConn) {
	var handshakePacket *Packet
	select {
	case data := <-conn.Receive():
		if data == nil {
			conn.Close()
			return
		}
		var err error
		handshakePacket, err = DeserializePacket(data)
		if err != nil {
			log.Printf("Failed to deserialize handshake: %v", err)
			conn.Close()
			return
		}
	case <-time.After(5 * time.Second):
		log.Printf("Handshake timeout")
		conn.Close()
		return
	}

	if handshakePacket.Header.Type != PacketTypeHandshake {
		log.Printf("Expected handshake packet, got %d", handshakePacket.Header.Type)
		conn.Close()
		return
	}

	if s.GetPlayerCount() >= s.MaxPlayersValue {
		resp := NewPacket(PacketTypeDisconnect, s.ID)
		resp.Payload = []byte("Server full")
		conn.SendPacket(resp)
		conn.Close()
		return
	}

	clientID := handshakePacket.Header.SenderID
	if clientID == "" {
		clientID = GenerateID()
	}

	playerID := GenerateID()
	playerName := string(handshakePacket.Payload)
	if playerName == "" {
		playerName = "Player"
	}

	client := NewClientSession(clientID, playerName, playerID)
	client.TCPConn = conn

	s.clientsMu.Lock()
	s.clients[clientID] = client
	s.clientsMu.Unlock()
	s.playerCount.Add(1)

	ack := NewPacket(PacketTypeHandshakeAck, s.ID)
	ack.Payload = []byte(playerID)
	conn.SendPacket(ack)

	serverInfo := NewPacket(PacketTypeServerInfo, s.ID)
	infoData, _ := SerializeServerInfo(s.GetServerInfo())
	serverInfo.Payload = infoData
	conn.SendPacket(serverInfo)

	s.sendPlayerList(client)
	s.BroadcastPlayerJoin(client)

	log.Printf("Player '%s' joined (ID: %s)", playerName, clientID)

	conn.OnDisconnect(func() {
		s.handleClientDisconnect(clientID)
	})

	go s.handleClientTCP(client, conn)

	if s.OnPlayerJoin != nil {
		s.OnPlayerJoin(client)
	}
}

func (s *Server) handleClientTCP(client *ClientSession, conn *TCPConn) {
	for data := range conn.Receive() {
		if data == nil {
			return
		}
		packet, err := DeserializePacket(data)
		if err != nil {
			log.Printf("Failed to deserialize TCP packet: %v", err)
			continue
		}
		client.LastSeen = time.Now()

		switch packet.Header.Type {
		case PacketTypeBlockChange:
			s.handleBlockChange(packet, client)
		case PacketTypeChatMessage:
			s.handleChatMessage(packet, client)
		case PacketTypePing:
			s.sendPong(client)
		}
	}
}

func (s *Server) handleClientDisconnect(clientID string) {
	s.clientsMu.Lock()
	client, exists := s.clients[clientID]
	if exists {
		delete(s.clients, clientID)
	}
	s.clientsMu.Unlock()

	if !exists {
		return
	}

	client.IsConnected = false
	s.playerCount.Add(^uint32(0))
	s.BroadcastPlayerLeave(client)

	log.Printf("Player '%s' left", client.Name)

	if s.OnPlayerLeave != nil {
		s.OnPlayerLeave(client)
	}
}

func (s *Server) handlePlayerPosition(packet *Packet, addr *net.UDPAddr) {
	s.clientsMu.RLock()
	client, exists := s.clients[packet.Header.SenderID]
	s.clientsMu.RUnlock()

	if !exists {
		return
	}

	if client.UDPAddr == nil || client.UDPAddr.String() != addr.String() {
		client.UDPAddr = addr
	}

	posData, err := DeserializePlayerPosition(packet.Payload)
	if err != nil {
		log.Printf("Failed to deserialize position: %v", err)
		return
	}

	client.UpdatePosition(posData.X, posData.Y, posData.VelocityX, posData.VelocityY)
}

func (s *Server) handlePing(packet *Packet, addr *net.UDPAddr) {
	pong := NewPacket(PacketTypePong, s.ID)
	s.udpConn.Send(pong.SerializeOrNil(), addr)
}

func (s *Server) sendPong(client *ClientSession) {
	pong := NewPacket(PacketTypePong, s.ID)
	client.SendReliable(pong)
}

func (s *Server) handleBlockChange(packet *Packet, client *ClientSession) {
	blockData, err := DeserializeBlockChange(packet.Payload)
	if err != nil {
		log.Printf("Failed to deserialize block change: %v", err)
		return
	}
	blockData.PlayerID = client.PlayerID
	s.BroadcastBlockChange(blockData, client.ID)
	if s.OnBlockChange != nil {
		s.OnBlockChange(blockData)
	}
}

func (s *Server) handleChatMessage(packet *Packet, client *ClientSession) {
	chatData, err := DeserializeChatMessage(packet.Payload)
	if err != nil {
		log.Printf("Failed to deserialize chat message: %v", err)
		return
	}
	chatData.PlayerID = client.PlayerID
	chatData.PlayerName = client.Name
	chatData.Timestamp = uint64(time.Now().UnixNano())
	s.BroadcastChatMessage(chatData)
	log.Printf("[Chat] %s: %s", client.Name, chatData.Message)
	if s.OnChatMessage != nil {
		s.OnChatMessage(chatData)
	}
}

func (s *Server) BroadcastPosition(sender *ClientSession, x, y, vx, vy float64) {
	data := &PlayerPositionData{
		PlayerID:  sender.PlayerID,
		X:         x,
		Y:         y,
		VelocityX: vx,
		VelocityY: vy,
	}
	payload, err := SerializePlayerPosition(data)
	if err != nil {
		return
	}
	packet := NewPacket(PacketTypePlayerPosition, s.ID)
	packet.Payload = payload
	packetData, _ := packet.Serialize()

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for id, client := range s.clients {
		if id == sender.ID || !client.IsConnected {
			continue
		}
		if client.UDPAddr != nil {
			s.udpConn.Send(packetData, client.UDPAddr)
		}
	}
}

func (s *Server) BroadcastBlockChange(data *BlockChangeData, excludeClientID string) {
	payload, err := SerializeBlockChange(data)
	if err != nil {
		return
	}
	packet := NewPacket(PacketTypeBlockChange, s.ID)
	packet.Payload = payload

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for id, client := range s.clients {
		if id == excludeClientID || !client.IsConnected {
			continue
		}
		client.SendReliable(packet)
	}
}

func (s *Server) BroadcastChatMessage(data *ChatMessageData) {
	payload, err := SerializeChatMessage(data)
	if err != nil {
		return
	}
	packet := NewPacket(PacketTypeChatMessage, s.ID)
	packet.Payload = payload

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if !client.IsConnected {
			continue
		}
		client.SendReliable(packet)
	}
}

func (s *Server) BroadcastPlayerJoin(newClient *ClientSession) {
	data := newClient.ToPlayerInfo()
	payload, _ := SerializePlayerInfo(data)
	packet := NewPacket(PacketTypePlayerJoin, s.ID)
	packet.Payload = payload

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for id, client := range s.clients {
		if id == newClient.ID || !client.IsConnected {
			continue
		}
		client.SendReliable(packet)
	}
}

func (s *Server) BroadcastPlayerLeave(leavingClient *ClientSession) {
	data := leavingClient.ToPlayerInfo()
	payload, _ := SerializePlayerInfo(data)
	packet := NewPacket(PacketTypePlayerLeave, s.ID)
	packet.Payload = payload

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if !client.IsConnected {
			continue
		}
		client.SendReliable(packet)
	}
}

func (s *Server) sendPlayerList(client *ClientSession) {
	s.clientsMu.RLock()
	clients := make([]*ClientSession, 0, len(s.clients))
	for _, c := range s.clients {
		if c.IsConnected {
			clients = append(clients, c)
		}
	}
	s.clientsMu.RUnlock()

	buf := make([]byte, 0)
	buf = append(buf, byte(len(clients)))

	for _, c := range clients {
		info := c.ToPlayerInfo()
		infoData, _ := SerializePlayerInfo(info)
		buf = append(buf, infoData...)
	}

	packet := NewPacket(PacketTypePlayerList, s.ID)
	packet.Payload = buf
	client.SendReliable(packet)
}

func (s *Server) broadcastLoop() {
	ticker := time.NewTicker(PositionBroadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.broadcastPositions()
		}
	}
}

func (s *Server) broadcastPositions() {
	s.clientsMu.RLock()
	clients := make([]*ClientSession, 0, len(s.clients))
	for _, c := range s.clients {
		if c.IsConnected {
			clients = append(clients, c)
		}
	}
	s.clientsMu.RUnlock()

	for _, client := range clients {
		s.BroadcastPosition(client, client.PositionX, client.PositionY,
			client.VelocityX, client.VelocityY)
	}
}

func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.removeTimedOutClients()
		}
	}
}

func (s *Server) removeTimedOutClients() {
	s.clientsMu.Lock()
	for id, client := range s.clients {
		if client.IsTimedOut(ClientTimeout) {
			log.Printf("Client '%s' timed out", client.Name)
			client.Disconnect()
			delete(s.clients, id)
			s.playerCount.Add(^uint32(0))
			go s.BroadcastPlayerLeave(client)
			if s.OnPlayerLeave != nil {
				go s.OnPlayerLeave(client)
			}
		}
	}
	s.clientsMu.Unlock()
}

func (p *Packet) SerializeOrNil() []byte {
	data, err := p.Serialize()
	if err != nil {
		return nil
	}
	return data
}

func SerializePlayerInfo(info *PlayerInfo) ([]byte, error) {
	buf := make([]byte, 0)
	idBytes := []byte(info.ID)
	buf = append(buf, byte(len(idBytes)))
	buf = append(buf, idBytes...)
	nameBytes := []byte(info.Name)
	buf = append(buf, byte(len(nameBytes)))
	buf = append(buf, nameBytes...)
	posBuf := make([]byte, 16)
	binary.LittleEndian.PutUint64(posBuf[0:8], uint64(info.X))
	binary.LittleEndian.PutUint64(posBuf[8:16], uint64(info.Y))
	buf = append(buf, posBuf...)
	return buf, nil
}

func DeserializePlayerInfo(data []byte) (*PlayerInfo, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("data too short")
	}
	info := &PlayerInfo{}
	offset := 0
	idLen := int(data[offset])
	offset++
	if len(data) < offset+idLen {
		return nil, fmt.Errorf("invalid ID length")
	}
	info.ID = string(data[offset : offset+idLen])
	offset += idLen
	if len(data) <= offset {
		return nil, fmt.Errorf("missing name length")
	}
	nameLen := int(data[offset])
	offset++
	if len(data) < offset+nameLen {
		return nil, fmt.Errorf("invalid name length")
	}
	info.Name = string(data[offset : offset+nameLen])
	offset += nameLen
	if len(data) < offset+16 {
		return nil, fmt.Errorf("missing position data")
	}
	info.X = float64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	info.Y = float64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	return info, nil
}
