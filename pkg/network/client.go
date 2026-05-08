package network

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// Client represents the game client
type Client struct {
	conn     net.Conn
	protocol *Protocol
	reader   *MessageReader
	writer   *MessageWriter

	// Connection state
	connected  bool
	playerID   uint32
	serverAddr string

	// Client state
	localPlayer  *LocalPlayer
	otherPlayers map[uint32]*RemotePlayer

	// Callbacks
	onConnect     func(uint32)
	onDisconnect  func()
	onPlayerJoin  func(*RemotePlayer)
	onPlayerLeave func(uint32)
	onBlockPlace  func(*BlockPlaceMessage)
	onBlockBreak  func(*BlockBreakMessage)
	onChat        func(*ChatMessage)

	// Synchronization
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Prediction
	predictor *ClientPredictor
}

// LocalPlayer represents the local player
type LocalPlayer struct {
	ID       uint32
	Name     string
	Position types.Vec3
	Rotation types.Vec3
	Velocity types.Vec3
	Health   uint16
}

// RemotePlayer represents a remote player
type RemotePlayer struct {
	ID         uint32
	Name       string
	Position   types.Vec3
	Rotation   types.Vec3
	Velocity   types.Vec3
	Health     uint16
	LastUpdate time.Time
}

// ClientPredictor handles client-side prediction
type ClientPredictor struct {
	inputHistory []InputState
	sequence     uint32
}

// InputState represents a player's input state
type InputState struct {
	Sequence  uint32
	Timestamp uint64
	Position  types.Vec3
	Rotation  types.Vec3
	Velocity  types.Vec3
	InputMask uint32 // Bit mask of input states
}

// ClientConfig holds client configuration
type ClientConfig struct {
	ServerAddr   string
	PlayerName   string
	TickRate     int
	PingInterval time.Duration
	Timeout      time.Duration
}

// NewClient creates a new game client
func NewClient(config ClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		protocol:     NewProtocol(),
		otherPlayers: make(map[uint32]*RemotePlayer),
		serverAddr:   config.ServerAddr,
		ctx:          ctx,
		cancel:       cancel,
		predictor:    &ClientPredictor{},
	}
}

// Connect connects to the game server
func (c *Client) Connect(playerName string) error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	c.reader = NewMessageReader(conn)
	c.writer = NewMessageWriter(conn)
	c.connected = true

	// Send handshake
	handshake := &HandshakeMessage{
		Version:    "1.0.0",
		PlayerName: playerName,
	}

	msg := &Message{
		Type:      MessageTypeHandshake,
		Data:      c.protocol.EncodeHandshake(handshake),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  0,
	}

	err = c.writer.WriteMessage(msg)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Start message reader
	go c.readMessages()

	// Start ping loop
	go c.pingLoop()

	log.Printf("Connected to server at %s", c.serverAddr)
	return nil
}

// Disconnect disconnects from the server
func (c *Client) Disconnect() {
	c.cancel()
	c.connected = false

	if c.conn != nil {
		c.conn.Close()
	}

	if c.onDisconnect != nil {
		c.onDisconnect()
	}

	log.Println("Disconnected from server")
}

// readMessages reads messages from the server
func (c *Client) readMessages() {
	for c.connected {
		msg, err := c.reader.ReadMessage()
		if err != nil {
			log.Printf("Error reading from server: %v", err)
			break
		}

		c.handleMessage(msg)
	}

	c.Disconnect()
}

// handleMessage handles a message from the server
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeHandshakeResponse:
		c.handleHandshakeResponse(msg)

	case MessageTypePlayerJoin:
		c.handlePlayerJoin(msg)

	case MessageTypePlayerLeave:
		c.handlePlayerLeave(msg)

	case MessageTypePlayerMove:
		c.handlePlayerMove(msg)

	case MessageTypeBlockPlace:
		c.handleBlockPlace(msg)

	case MessageTypeBlockBreak:
		c.handleBlockBreak(msg)

	case MessageTypeChat:
		c.handleChat(msg)

	case MessageTypeWorldUpdate:
		c.handleWorldUpdate(msg)

	case MessageTypePong:
		// Handle pong (ping measurement)

	case MessageTypeError:
		c.handleError(msg)

	default:
		log.Printf("Unknown message type: %v", msg.Type)
	}
}

// handleHandshakeResponse handles handshake response from server
func (c *Client) handleHandshakeResponse(msg *Message) {
	response, err := c.protocol.DecodeHandshakeResponse(msg.Data)
	if err != nil {
		log.Printf("Error decoding handshake response: %v", err)
		return
	}

	if !response.Success {
		log.Printf("Server rejected connection: %s", response.Message)
		c.Disconnect()
		return
	}

	c.playerID = response.PlayerID

	// Create local player
	c.mu.Lock()
	c.localPlayer = &LocalPlayer{
		ID:   c.playerID,
		Name: "", // Will be set from handshake
	}
	c.mu.Unlock()

	log.Printf("Connected as player %d", c.playerID)

	if c.onConnect != nil {
		c.onConnect(c.playerID)
	}
}

// handlePlayerJoin handles a player joining the game
func (c *Client) handlePlayerJoin(msg *Message) {
	// Decode player join message
	player := &RemotePlayer{
		ID:         msg.PlayerID,
		Name:       "Unknown", // Will be decoded from message
		Position:   types.NewVec3(0, 0, 0),
		Rotation:   types.NewVec3(0, 0, 0),
		Velocity:   types.NewVec3(0, 0, 0),
		Health:     100,
		LastUpdate: time.Now(),
	}

	c.mu.Lock()
	c.otherPlayers[msg.PlayerID] = player
	c.mu.Unlock()

	log.Printf("Player %d joined", msg.PlayerID)

	if c.onPlayerJoin != nil {
		c.onPlayerJoin(player)
	}
}

// handlePlayerLeave handles a player leaving the game
func (c *Client) handlePlayerLeave(msg *Message) {
	c.mu.Lock()
	delete(c.otherPlayers, msg.PlayerID)
	c.mu.Unlock()

	log.Printf("Player %d left", msg.PlayerID)

	if c.onPlayerLeave != nil {
		c.onPlayerLeave(msg.PlayerID)
	}
}

// handlePlayerMove handles player movement
func (c *Client) handlePlayerMove(msg *Message) {
	move, err := c.protocol.DecodePlayerMove(msg.Data)
	if err != nil {
		log.Printf("Error decoding player move: %v", err)
		return
	}

	c.mu.Lock()
	if move.PlayerID == c.playerID {
		// Update local player (server correction)
		if c.localPlayer != nil {
			c.localPlayer.Position = move.Position
			c.localPlayer.Rotation = move.Rotation
			c.localPlayer.Velocity = move.Velocity
		}
	} else {
		// Update remote player
		if player, exists := c.otherPlayers[move.PlayerID]; exists {
			player.Position = move.Position
			player.Rotation = move.Rotation
			player.Velocity = move.Velocity
			player.LastUpdate = time.Now()
		}
	}
	c.mu.Unlock()
}

// handleBlockPlace handles block placement
func (c *Client) handleBlockPlace(msg *Message) {
	place, err := c.protocol.DecodeBlockPlace(msg.Data)
	if err != nil {
		log.Printf("Error decoding block place: %v", err)
		return
	}

	log.Printf("Block placed at %v by player %d", place.Position, place.PlayerID)

	if c.onBlockPlace != nil {
		c.onBlockPlace(place)
	}
}

// handleBlockBreak handles block breaking
func (c *Client) handleBlockBreak(msg *Message) {
	// Decode block break message
	log.Printf("Block broken by player %d", msg.PlayerID)

	if c.onBlockBreak != nil {
		// c.onBlockBreak(breakMsg)
	}
}

// handleChat handles chat messages
func (c *Client) handleChat(msg *Message) {
	// Decode chat message
	log.Printf("Chat message from player %d", msg.PlayerID)

	if c.onChat != nil {
		// c.onChat(chatMsg)
	}
}

// handleWorldUpdate handles world state updates
func (c *Client) handleWorldUpdate(msg *Message) {
	// Decode world update
	// Apply world changes
}

// handleError handles error messages
func (c *Client) handleError(msg *Message) {
	// Decode error message
	log.Printf("Error from server: %v", msg.Data)
}

// pingLoop sends periodic ping messages
func (c *Client) pingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for c.connected {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			ping := &PingMessage{
				Timestamp: uint64(time.Now().UnixNano()),
			}

			msg := &Message{
				Type:      MessageTypePing,
				Data:      []byte{}, // Simplified
				Timestamp: ping.Timestamp,
				PlayerID:  c.playerID,
			}

			c.writer.WriteMessage(msg)
		}
	}
}

// SendPlayerMove sends player movement to server
func (c *Client) SendPlayerMove(position, rotation, velocity types.Vec3) {
	if !c.connected {
		return
	}

	c.mu.Lock()
	if c.localPlayer != nil {
		c.localPlayer.Position = position
		c.localPlayer.Rotation = rotation
		c.localPlayer.Velocity = velocity
	}
	c.mu.Unlock()

	move := &PlayerMoveMessage{
		PlayerID: c.playerID,
		Position: position,
		Rotation: rotation,
		Velocity: velocity,
	}

	msg := &Message{
		Type:      MessageTypePlayerMove,
		Data:      c.protocol.EncodePlayerMove(move),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  c.playerID,
	}

	c.writer.WriteMessage(msg)
}

// SendBlockPlace sends block placement to server
func (c *Client) SendBlockPlace(blockType uint8, position types.Vec3, rotation int) {
	if !c.connected {
		return
	}

	place := &BlockPlaceMessage{
		PlayerID:  c.playerID,
		BlockType: blockType,
		Position:  position,
		Rotation:  rotation,
	}

	msg := &Message{
		Type:      MessageTypeBlockPlace,
		Data:      c.protocol.EncodeBlockPlace(place),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  c.playerID,
	}

	c.writer.WriteMessage(msg)
}

// SendBlockBreak sends block breaking to server
func (c *Client) SendBlockBreak(position types.Vec3) {
	if !c.connected {
		return
	}

	// Create block break message
	msg := &Message{
		Type:      MessageTypeBlockBreak,
		Data:      []byte{}, // Simplified
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  c.playerID,
	}

	c.writer.WriteMessage(msg)
}

// SendChat sends a chat message
func (c *Client) SendChat(message string) {
	if !c.connected {
		return
	}

	// Create chat message
	msg := &Message{
		Type:      MessageTypeChat,
		Data:      []byte(message), // Simplified
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  c.playerID,
	}

	c.writer.WriteMessage(msg)
}

// GetLocalPlayer returns the local player
func (c *Client) GetLocalPlayer() *LocalPlayer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.localPlayer
}

// GetOtherPlayers returns all other players
func (c *Client) GetOtherPlayers() map[uint32]*RemotePlayer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	players := make(map[uint32]*RemotePlayer)
	for id, player := range c.otherPlayers {
		players[id] = player
	}

	return players
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// GetPlayerID returns the local player ID
func (c *Client) GetPlayerID() uint32 {
	return c.playerID
}

// SetCallbacks sets callback functions
func (c *Client) SetCallbacks(onConnect func(uint32), onDisconnect func(),
	onPlayerJoin func(*RemotePlayer), onPlayerLeave func(uint32),
	onBlockPlace func(*BlockPlaceMessage), onBlockBreak func(*BlockBreakMessage),
	onChat func(*ChatMessage)) {
	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
	c.onPlayerJoin = onPlayerJoin
	c.onPlayerLeave = onPlayerLeave
	c.onBlockPlace = onBlockPlace
	c.onBlockBreak = onBlockBreak
	c.onChat = onChat
}

// PredictMovement predicts client movement for smooth gameplay
func (c *Client) PredictMovement(position, rotation, velocity types.Vec3, inputMask uint32) {
	// Add to prediction history
	input := InputState{
		Sequence:  c.predictor.sequence,
		Timestamp: uint64(time.Now().UnixNano()),
		Position:  position,
		Rotation:  rotation,
		Velocity:  velocity,
		InputMask: inputMask,
	}

	c.predictor.inputHistory = append(c.predictor.inputHistory, input)
	c.predictor.sequence++

	// Keep only recent history
	if len(c.predictor.inputHistory) > 60 { // 1 second at 60 FPS
		c.predictor.inputHistory = c.predictor.inputHistory[1:]
	}
}

// ReconcileWithServer reconciles client prediction with server state
func (c *Client) ReconcileWithServer(serverPosition types.Vec3) {
	// Find the last acknowledged input
	c.mu.RLock()
	localPlayer := c.localPlayer
	c.mu.RUnlock()

	if localPlayer == nil {
		return
	}

	// Check if we need to correct position
	distance := localPlayer.Position.Distance(serverPosition)
	if distance > 0.1 { // Threshold for correction
		// Correct position
		c.mu.Lock()
		localPlayer.Position = serverPosition
		c.mu.Unlock()

		// Replay inputs from correction point
		c.replayInputs()
	}
}

// replayInputs replays inputs from the correction point
func (c *Client) replayInputs() {
	// Find the correction point and replay subsequent inputs
	// This is a simplified implementation
}
