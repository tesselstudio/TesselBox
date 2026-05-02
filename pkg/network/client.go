package network

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type ConnectionState int32

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

type Client struct {
	ID              string
	PlayerID        string
	Name            string
	ServerAddress   string
	ServerPort      uint16
	udpConn         *UDPConn
	tcpConn         *TCPConn
	state           atomic.Int32
	connected       atomic.Bool
	positionUpdates chan *PlayerPositionData
	blockChanges    chan *BlockChangeData
	chatMessages    chan *ChatMessageData
	playerJoins     chan *PlayerInfo
	playerLeaves    chan *PlayerInfo
	stopChan        chan struct{}
	mu              sync.RWMutex
	lastPosX        float64
	lastPosY        float64
	lastPosTime     time.Time
	sequence        atomic.Uint32
	OnConnected     func()
	OnDisconnected  func(reason string)
	OnPlayerPosition func(data *PlayerPositionData)
	OnBlockChange   func(data *BlockChangeData)
	OnChatMessage   func(data *ChatMessageData)
	OnPlayerJoin    func(info *PlayerInfo)
	OnPlayerLeave   func(info *PlayerInfo)
	OnServerInfo    func(info *ServerInfoData)
	OnPlayerList    func(players []*PlayerInfo)
}

func NewClient(playerName string) *Client {
	return &Client{
		ID:              GenerateID(),
		Name:            playerName,
		ServerPort:      DefaultPort,
		positionUpdates: make(chan *PlayerPositionData, 100),
		blockChanges:    make(chan *BlockChangeData, 100),
		chatMessages:    make(chan *ChatMessageData, 100),
		playerJoins:     make(chan *PlayerInfo, 10),
		playerLeaves:    make(chan *PlayerInfo, 10),
		stopChan:        make(chan struct{}),
	}
}

func (c *Client) Connect(serverAddr string, port uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.GetState() != StateDisconnected {
		return fmt.Errorf("client already connected or connecting")
	}

	c.state.Store(int32(StateConnecting))
	c.ServerAddress = serverAddr
	if port > 0 {
		c.ServerPort = port
	}

	tcpConn, err := NewTCPClient(serverAddr, c.ServerPort)
	if err != nil {
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("failed to connect TCP: %w", err)
	}
	c.tcpConn = tcpConn

	handshake := NewPacket(PacketTypeHandshake, c.ID)
	handshake.Payload = []byte(c.Name)
	if err := c.tcpConn.SendPacket(handshake); err != nil {
		c.tcpConn.Close()
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	var ackPacket *Packet
	select {
	case data := <-c.tcpConn.Receive():
		if data == nil {
			c.tcpConn.Close()
			c.state.Store(int32(StateDisconnected))
			return fmt.Errorf("connection closed during handshake")
		}
		ackPacket, err = DeserializePacket(data)
		if err != nil {
			c.tcpConn.Close()
			c.state.Store(int32(StateDisconnected))
			return fmt.Errorf("failed to deserialize handshake ack: %w", err)
		}
	case <-time.After(10 * time.Second):
		c.tcpConn.Close()
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("handshake timeout")
	}

	if ackPacket.Header.Type != PacketTypeHandshakeAck {
		c.tcpConn.Close()
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("expected handshake ack, got %d", ackPacket.Header.Type)
	}

	c.PlayerID = string(ackPacket.Payload)

	udpConn, err := NewUDPClient(serverAddr, c.ServerPort)
	if err != nil {
		c.tcpConn.Close()
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("failed to connect UDP: %w", err)
	}
	c.udpConn = udpConn

	c.connected.Store(true)
	c.state.Store(int32(StateConnected))

	log.Printf("Connected to server %s:%d as '%s' (ID: %s)", serverAddr, c.ServerPort, c.Name, c.PlayerID)

	go c.tcpReceiveLoop()
	go c.udpReceiveLoop()
	go c.positionBroadcastLoop()

	if c.OnConnected != nil {
		c.OnConnected()
	}

	return nil
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.GetState() == StateDisconnected {
		return nil
	}

	c.state.Store(int32(StateDisconnected))
	c.connected.Store(false)
	close(c.stopChan)

	if c.tcpConn != nil && !c.tcpConn.IsClosed() {
		disconnect := NewPacket(PacketTypeDisconnect, c.ID)
		c.tcpConn.SendPacket(disconnect)
	}

	if c.udpConn != nil {
		c.udpConn.Close()
	}
	if c.tcpConn != nil {
		c.tcpConn.Close()
	}

	log.Printf("Disconnected from server")

	if c.OnDisconnected != nil {
		c.OnDisconnected("user initiated")
	}

	return nil
}

func (c *Client) GetState() ConnectionState {
	return ConnectionState(c.state.Load())
}

func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

func (c *Client) SendPosition(x, y, vx, vy float64) {
	if !c.IsConnected() || c.udpConn == nil {
		return
	}
	if time.Since(c.lastPosTime) < 50*time.Millisecond {
		return
	}
	dx := x - c.lastPosX
	dy := y - c.lastPosY
	if dx*dx+dy*dy < 1.0 {
		return
	}

	data := &PlayerPositionData{
		PlayerID:  c.PlayerID,
		X:         x,
		Y:         y,
		VelocityX: vx,
		VelocityY: vy,
	}

	payload, err := SerializePlayerPosition(data)
	if err != nil {
		return
	}

	packet := NewPacket(PacketTypePlayerPosition, c.ID)
	packet.Payload = payload
	packet.Header.Sequence = c.sequence.Add(1)

	packetData, _ := packet.Serialize()
	c.udpConn.SendToServer(packetData)

	c.lastPosX = x
	c.lastPosY = y
	c.lastPosTime = time.Now()
}

func (c *Client) SendBlockChange(x, y int32, blockType uint8) error {
	if !c.IsConnected() || c.tcpConn == nil {
		return fmt.Errorf("not connected")
	}
	data := &BlockChangeData{
		X:         x,
		Y:         y,
		BlockType: blockType,
		PlayerID:  c.PlayerID,
	}
	payload, err := SerializeBlockChange(data)
	if err != nil {
		return err
	}
	packet := NewPacket(PacketTypeBlockChange, c.ID)
	packet.Payload = payload
	return c.tcpConn.SendPacket(packet)
}

func (c *Client) SendChatMessage(message string) error {
	if !c.IsConnected() || c.tcpConn == nil {
		return fmt.Errorf("not connected")
	}
	data := &ChatMessageData{
		PlayerID:   c.PlayerID,
		PlayerName: c.Name,
		Message:    message,
		Timestamp:  uint64(time.Now().UnixNano()),
	}
	payload, err := SerializeChatMessage(data)
	if err != nil {
		return err
	}
	packet := NewPacket(PacketTypeChatMessage, c.ID)
	packet.Payload = payload
	return c.tcpConn.SendPacket(packet)
}

func (c *Client) tcpReceiveLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		case data := <-c.tcpConn.Receive():
			if data == nil {
				c.handleDisconnect("server closed connection")
				return
			}
			c.handleTCPPacket(data)
		}
	}
}

func (c *Client) udpReceiveLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		case msg := <-c.udpConn.Receive():
			if msg == nil {
				return
			}
			c.handleUDPPacket(msg.Data)
		}
	}
}

func (c *Client) handleTCPPacket(data []byte) {
	packet, err := DeserializePacket(data)
	if err != nil {
		log.Printf("Failed to deserialize TCP packet: %v", err)
		return
	}

	switch packet.Header.Type {
	case PacketTypeServerInfo:
		c.handleServerInfo(packet.Payload)
	case PacketTypePlayerList:
		c.handlePlayerList(packet.Payload)
	case PacketTypePlayerJoin:
		c.handlePlayerJoin(packet.Payload)
	case PacketTypePlayerLeave:
		c.handlePlayerLeave(packet.Payload)
	case PacketTypeBlockChange:
		c.handleBlockChange(packet.Payload)
	case PacketTypeChatMessage:
		c.handleChatMessage(packet.Payload)
	case PacketTypeDisconnect:
		reason := string(packet.Payload)
		if reason == "" {
			reason = "kicked by server"
		}
		c.handleDisconnect(reason)
	}
}

func (c *Client) handleUDPPacket(data []byte) {
	packet, err := DeserializePacket(data)
	if err != nil {
		return
	}
	switch packet.Header.Type {
	case PacketTypePlayerPosition:
		c.handlePlayerPosition(packet.Payload)
	}
}

func (c *Client) handleServerInfo(payload []byte) {
	info, err := DeserializeServerInfo(payload)
	if err != nil {
		return
	}
	if c.OnServerInfo != nil {
		c.OnServerInfo(info)
	}
}

func (c *Client) handlePlayerList(payload []byte) {
	if len(payload) < 1 {
		return
	}
	playerCount := int(payload[0])
	offset := 1
	players := make([]*PlayerInfo, 0, playerCount)

	for i := 0; i < playerCount && offset < len(payload); i++ {
		info, err := DeserializePlayerInfo(payload[offset:])
		if err != nil {
			break
		}
		players = append(players, info)
		offset += len(info.ID) + len(info.Name) + 18
	}
	if c.OnPlayerList != nil {
		c.OnPlayerList(players)
	}
}

func (c *Client) handlePlayerJoin(payload []byte) {
	info, err := DeserializePlayerInfo(payload)
	if err != nil {
		return
	}
	log.Printf("Player joined: %s", info.Name)
	if c.OnPlayerJoin != nil {
		c.OnPlayerJoin(info)
	}
}

func (c *Client) handlePlayerLeave(payload []byte) {
	info, err := DeserializePlayerInfo(payload)
	if err != nil {
		return
	}
	log.Printf("Player left: %s", info.Name)
	if c.OnPlayerLeave != nil {
		c.OnPlayerLeave(info)
	}
}

func (c *Client) handlePlayerPosition(payload []byte) {
	posData, err := DeserializePlayerPosition(payload)
	if err != nil {
		return
	}
	if posData.PlayerID == c.PlayerID {
		return
	}
	if c.OnPlayerPosition != nil {
		c.OnPlayerPosition(posData)
	}
}

func (c *Client) handleBlockChange(payload []byte) {
	blockData, err := DeserializeBlockChange(payload)
	if err != nil {
		return
	}
	if c.OnBlockChange != nil {
		c.OnBlockChange(blockData)
	}
}

func (c *Client) handleChatMessage(payload []byte) {
	chatData, err := DeserializeChatMessage(payload)
	if err != nil {
		return
	}
	if c.OnChatMessage != nil {
		c.OnChatMessage(chatData)
	}
}

func (c *Client) handleDisconnect(reason string) {
	c.state.Store(int32(StateDisconnected))
	c.connected.Store(false)
	if c.udpConn != nil {
		c.udpConn.Close()
	}
	if c.tcpConn != nil {
		c.tcpConn.Close()
	}
	log.Printf("Disconnected: %s", reason)
	if c.OnDisconnected != nil {
		c.OnDisconnected(reason)
	}
}

func (c *Client) positionBroadcastLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
		}
	}
}

func (c *Client) GetPlayerID() string {
	return c.PlayerID
}

func (c *Client) GetServerAddress() string {
	return c.ServerAddress
}

func (c *Client) GetServerPort() uint16 {
	return c.ServerPort
}
