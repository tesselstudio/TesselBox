package network

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"github.com/tesselstudio/TesselBox/pkg/world"
	"kaijuengine.com/matrix"
)

// Server represents the game server
type Server struct {
	listener     net.Listener
	clients      map[uint32]*ServerClient
	clientsMu    sync.RWMutex
	world        *WorldManager
	nextPlayerID uint32
	protocol     *Protocol
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc

	// Configuration
	config ServerConfig

	// Channels
	connectChan    chan *ServerClient
	disconnectChan chan uint32
	messageChan    chan *ClientMessage
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string
	Port         int
	MaxPlayers   int
	TickRate     int // Updates per second
	WorldSize    int // Size of world in chunks
	ChunkSize    int // Size of each chunk
	ViewDistance int // How many chunks to send to clients
}

// ServerClient represents a connected client
type ServerClient struct {
	ID        uint32
	Name      string
	Conn      net.Conn
	Player    *Player
	Writer    *MessageWriter
	Reader    *MessageReader
	LastPing  time.Time
	Connected bool
}

// Player represents a player in the game
type Player struct {
	ID       uint32
	Name     string
	Position matrix.Vec3
	Rotation matrix.Vec3
	Velocity matrix.Vec3
	Health   uint16
	WorldPos world.HexCoord
}

// ClientMessage represents a message from a client
type ClientMessage struct {
	Client  *ServerClient
	Message *Message
}

// WorldManager manages the game world state
type WorldManager struct {
	chunks   map[world.HexCoord]*Chunk
	chunksMu sync.RWMutex
	blocks   map[matrix.Vec3]*blocks.Block
	blocksMu sync.RWMutex
	config   ServerConfig
}

// Chunk represents a chunk of the world
type Chunk struct {
	Position world.HexCoord
	Blocks   map[matrix.Vec3]*blocks.Block
	Dirty    bool
}

// NewServer creates a new game server
func NewServer(config ServerConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		clients:        make(map[uint32]*ServerClient),
		nextPlayerID:   1,
		protocol:       NewProtocol(),
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
		world:          NewWorldManager(config),
		connectChan:    make(chan *ServerClient, 100),
		disconnectChan: make(chan uint32, 100),
		messageChan:    make(chan *ClientMessage, 1000),
	}
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.running = true

	log.Printf("Server started on %s", addr)

	// Start accepting connections
	go s.acceptConnections()

	// Start game loop
	go s.gameLoop()

	// Start message handler
	go s.handleMessages()

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.cancel()
	s.running = false

	if s.listener != nil {
		s.listener.Close()
	}

	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clientsMu.Unlock()

	log.Println("Server stopped")
}

// acceptConnections accepts new client connections
func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Error accepting connection: %v", err)
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a new client connection
func (s *Server) handleConnection(conn net.Conn) {
	client := &ServerClient{
		Conn:      conn,
		Writer:    NewMessageWriter(conn),
		Reader:    NewMessageReader(conn),
		LastPing:  time.Now(),
		Connected: true,
	}

	// Read handshake message
	msg, err := client.Reader.ReadMessage()
	if err != nil {
		log.Printf("Error reading handshake: %v", err)
		conn.Close()
		return
	}

	if msg.Type != MessageTypeHandshake {
		log.Printf("Expected handshake, got %v", msg.Type)
		conn.Close()
		return
	}

	handshake, err := s.protocol.DecodeHandshake(msg.Data)
	if err != nil {
		log.Printf("Error decoding handshake: %v", err)
		conn.Close()
		return
	}

	// Check if server is full
	s.clientsMu.RLock()
	if len(s.clients) >= s.config.MaxPlayers {
		s.clientsMu.RUnlock()

		// Send server full response
		response := &HandshakeResponseMessage{
			Success:  false,
			PlayerID: 0,
			Message:  "Server is full",
		}

		respMsg := &Message{
			Type:      MessageTypeHandshakeResponse,
			Data:      s.protocol.EncodeHandshakeResponse(response),
			Timestamp: uint64(time.Now().UnixNano()),
			PlayerID:  0,
		}

		client.Writer.WriteMessage(respMsg)
		conn.Close()
		return
	}
	s.clientsMu.RUnlock()

	// Assign player ID
	client.ID = s.nextPlayerID
	s.nextPlayerID++
	// Validate player name
	name := strings.TrimSpace(handshake.PlayerName)
	if len(name) == 0 || len(name) > 32 {
		// Send invalid name response
		response := &HandshakeResponseMessage{
			Success:  false,
			PlayerID: 0,
			Message:  "Player name must be 1-32 characters",
		}

		respMsg := &Message{
			Type:      MessageTypeHandshakeResponse,
			Data:      s.protocol.EncodeHandshakeResponse(response),
			Timestamp: uint64(time.Now().UnixNano()),
			PlayerID:  0,
		}

		client.Writer.WriteMessage(respMsg)
		conn.Close()
		return
	}

	// Sanitize name (allow only alphanumeric, spaces, underscores)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_ ]+$`).MatchString(name)
	if !validName {
		// Send invalid characters response
		response := &HandshakeResponseMessage{
			Success:  false,
			PlayerID: 0,
			Message:  "Player name contains invalid characters",
		}

		respMsg := &Message{
			Type:      MessageTypeHandshakeResponse,
			Data:      s.protocol.EncodeHandshakeResponse(response),
			Timestamp: uint64(time.Now().UnixNano()),
			PlayerID:  0,
		}

		client.Writer.WriteMessage(respMsg)
		conn.Close()
		return
	}

	client.Name = name

	// Create player object
	client.Player = &Player{
		ID:       client.ID,
		Name:     client.Name,
		Position: matrix.NewVec3(0, 10, 0), // Start above ground
		Rotation: matrix.NewVec3(0, 0, 0),
		Velocity: matrix.NewVec3(0, 0, 0),
		Health:   100,
	}

	// Send successful handshake response
	response := &HandshakeResponseMessage{
		Success:  true,
		PlayerID: client.ID,
		Message:  "Connected successfully",
	}

	respMsg := &Message{
		Type:      MessageTypeHandshakeResponse,
		Data:      s.protocol.EncodeHandshakeResponse(response),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  client.ID,
	}

	err = client.Writer.WriteMessage(respMsg)
	if err != nil {
		log.Printf("Error sending handshake response: %v", err)
		conn.Close()
		return
	}

	// Add client to server
	s.clientsMu.Lock()
	s.clients[client.ID] = client
	s.clientsMu.Unlock()

	// Notify other players
	s.broadcastPlayerJoin(client)

	// Send initial world state
	s.sendInitialWorld(client)

	log.Printf("Player %s (%d) connected", client.Name, client.ID)

	// Start reading messages from this client
	go s.readClientMessages(client)

	// Add to connect channel
	s.connectChan <- client
}

// readClientMessages reads messages from a client
func (s *Server) readClientMessages(client *ServerClient) {
	for client.Connected && s.running {
		msg, err := client.Reader.ReadMessage()
		if err != nil {
			log.Printf("Error reading from client %d: %v", client.ID, err)
			break
		}

		// Update last ping
		client.LastPing = time.Now()

		// Queue message for processing
		s.messageChan <- &ClientMessage{
			Client:  client,
			Message: msg,
		}
	}

	// Client disconnected
	s.disconnectChan <- client.ID
}

// handleMessages processes messages from clients
func (s *Server) handleMessages() {
	for {
		select {
		case <-s.ctx.Done():
			return

		case clientMsg := <-s.messageChan:
			s.processClientMessage(clientMsg)

		case playerID := <-s.disconnectChan:
			s.handlePlayerDisconnect(playerID)

		case client := <-s.connectChan:
			s.handlePlayerConnect(client)
		}
	}
}

// processClientMessage processes a message from a client
func (s *Server) processClientMessage(clientMsg *ClientMessage) {
	client := clientMsg.Client
	msg := clientMsg.Message

	switch msg.Type {
	case MessageTypePlayerMove:
		s.handlePlayerMove(client, msg)

	case MessageTypeBlockPlace:
		s.handleBlockPlace(client, msg)

	case MessageTypeBlockBreak:
		s.handleBlockBreak(client, msg)

	case MessageTypeChat:
		s.handleChat(client, msg)

	case MessageTypePing:
		s.handlePing(client, msg)

	default:
		log.Printf("Unknown message type: %v from client %d", msg.Type, client.ID)
	}
}

// handlePlayerMove handles player movement
func (s *Server) handlePlayerMove(client *ServerClient, msg *Message) {
	move, err := s.protocol.DecodePlayerMove(msg.Data)
	if err != nil {
		log.Printf("Error decoding player move: %v", err)
		return
	}

	// Update player position
	client.Player.Position = move.Position
	client.Player.Rotation = move.Rotation
	client.Player.Velocity = move.Velocity

	// Broadcast to other players
	s.broadcastPlayerMove(client)
}

// handleBlockPlace handles block placement
func (s *Server) handleBlockPlace(client *ServerClient, msg *Message) {
	place, err := s.protocol.DecodeBlockPlace(msg.Data)
	if err != nil {
		log.Printf("Error decoding block place: %v", err)
		return
	}

	// Create block
	block, err := blocks.NewBlock(blocks.BlockType(place.BlockType), place.Position, place.Rotation)
	if err != nil {
		log.Printf("Error creating block: %v", err)
		return
	}

	// Add to world
	s.world.AddBlock(block)

	// Broadcast block change
	s.broadcastBlockPlace(place)
}

// handleBlockBreak handles block breaking
func (s *Server) handleBlockBreak(client *ServerClient, msg *Message) {
	// Decode block break message
	// Remove block from world
	// Broadcast block change
}

// handleChat handles chat messages
func (s *Server) handleChat(client *ServerClient, msg *Message) {
	// Decode chat message
	// Broadcast to all players
}

// handlePing handles ping messages
func (s *Server) handlePing(client *ServerClient, msg *Message) {
	// Send pong response
	pong := &PongMessage{
		Timestamp: uint64(time.Now().UnixNano()),
	}

	respMsg := &Message{
		Type:      MessageTypePong,
		Data:      []byte{}, // Simplified
		Timestamp: pong.Timestamp,
		PlayerID:  client.ID,
	}

	client.Writer.WriteMessage(respMsg)
}

// handlePlayerConnect handles a new player connection
func (s *Server) handlePlayerConnect(client *ServerClient) {
	// Player is already added in handleConnection
}

// handlePlayerDisconnect handles a player disconnection
func (s *Server) handlePlayerDisconnect(playerID uint32) {
	s.clientsMu.Lock()
	client, exists := s.clients[playerID]
	if exists {
		delete(s.clients, playerID)
		client.Connected = false
		client.Conn.Close()
	}
	s.clientsMu.Unlock()

	if exists {
		log.Printf("Player %s (%d) disconnected", client.Name, playerID)

		// Notify other players
		s.broadcastPlayerLeave(playerID)
	}
}

// gameLoop runs the main game loop
func (s *Server) gameLoop() {
	ticker := time.NewTicker(time.Second / time.Duration(s.config.TickRate))
	defer ticker.Stop()

	for s.running {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateGame()
		}
	}
}

// updateGame updates the game state
func (s *Server) updateGame() {
	// Update world
	s.world.Update()

	// Send world updates to clients
	s.broadcastWorldUpdate()

	// Check for timeouts
	s.checkTimeouts()
}

// checkTimeouts checks for client timeouts
func (s *Server) checkTimeouts() {
	now := time.Now()
	timeout := 30 * time.Second

	s.clientsMu.RLock()
	for _, client := range s.clients {
		if now.Sub(client.LastPing) > timeout {
			log.Printf("Client %d timed out", client.ID)
			go func(c *ServerClient) {
				c.Conn.Close()
			}(client)
		}
	}
	s.clientsMu.RUnlock()
}

// broadcastPlayerJoin broadcasts a player join message
func (s *Server) broadcastPlayerJoin(client *ServerClient) {
	msg := &Message{
		Type:      MessageTypePlayerJoin,
		Data:      []byte{}, // Simplified encoding
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  client.ID,
	}

	s.broadcastToOthers(client.ID, msg)
}

// broadcastPlayerLeave broadcasts a player leave message
func (s *Server) broadcastPlayerLeave(playerID uint32) {
	msg := &Message{
		Type:      MessageTypePlayerLeave,
		Data:      []byte{}, // Simplified encoding
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  playerID,
	}

	s.broadcastToAll(msg)
}

// broadcastPlayerMove broadcasts player movement
func (s *Server) broadcastPlayerMove(client *ServerClient) {
	move := &PlayerMoveMessage{
		PlayerID: client.ID,
		Position: client.Player.Position,
		Rotation: client.Player.Rotation,
		Velocity: client.Player.Velocity,
	}

	msg := &Message{
		Type:      MessageTypePlayerMove,
		Data:      s.protocol.EncodePlayerMove(move),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  client.ID,
	}

	s.broadcastToOthers(client.ID, msg)
}

// broadcastBlockPlace broadcasts block placement
func (s *Server) broadcastBlockPlace(place *BlockPlaceMessage) {
	msg := &Message{
		Type:      MessageTypeBlockPlace,
		Data:      s.protocol.EncodeBlockPlace(place),
		Timestamp: uint64(time.Now().UnixNano()),
		PlayerID:  place.PlayerID,
	}

	s.broadcastToAll(msg)
}

// broadcastWorldUpdate broadcasts world state updates
func (s *Server) broadcastWorldUpdate() {
	// Collect world changes
	// Send to all clients
}

// sendInitialWorld sends initial world state to a client
func (s *Server) sendInitialWorld(client *ServerClient) {
	// Send nearby chunks
	// Send existing players
	// Send initial player state
}

// broadcastToAll broadcasts a message to all clients
func (s *Server) broadcastToAll(msg *Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.Connected {
			client.Writer.WriteMessage(msg)
		}
	}
}

// broadcastToOthers broadcasts a message to all clients except one
func (s *Server) broadcastToOthers(excludePlayerID uint32, msg *Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.Connected && client.ID != excludePlayerID {
			client.Writer.WriteMessage(msg)
		}
	}
}

// NewWorldManager creates a new world manager
func NewWorldManager(config ServerConfig) *WorldManager {
	return &WorldManager{
		chunks: make(map[world.HexCoord]*Chunk),
		blocks: make(map[matrix.Vec3]*blocks.Block),
		config: config,
	}
}

// AddBlock adds a block to the world
func (wm *WorldManager) AddBlock(block *blocks.Block) {
	wm.blocksMu.Lock()
	defer wm.blocksMu.Unlock()

	wm.blocks[block.Position] = block

	// Mark chunk as dirty
	chunkPos := wm.getChunkPosition(block.Position)
	wm.markChunkDirty(chunkPos)
}

// RemoveBlock removes a block from the world
func (wm *WorldManager) RemoveBlock(position matrix.Vec3) {
	wm.blocksMu.Lock()
	defer wm.blocksMu.Unlock()

	delete(wm.blocks, position)

	// Mark chunk as dirty
	chunkPos := wm.getChunkPosition(position)
	wm.markChunkDirty(chunkPos)
}

// GetBlock gets a block from the world
func (wm *WorldManager) GetBlock(position matrix.Vec3) (*blocks.Block, bool) {
	wm.blocksMu.RLock()
	defer wm.blocksMu.RUnlock()

	block, exists := wm.blocks[position]
	return block, exists
}

// Update updates the world manager
func (wm *WorldManager) Update() {
	// Process dirty chunks
	// Update physics
	// Update entities
}

// getChunkPosition gets the chunk position for a world position
func (wm *WorldManager) getChunkPosition(worldPos matrix.Vec3) world.HexCoord {
	grid := world.NewHexGrid(matrix.Float(wm.config.ChunkSize), world.NewHexCoord(0, 0))
	return grid.FromWorld(worldPos)
}

// markChunkDirty marks a chunk as dirty
func (wm *WorldManager) markChunkDirty(chunkPos world.HexCoord) {
	wm.chunksMu.Lock()
	defer wm.chunksMu.Unlock()

	chunk, exists := wm.chunks[chunkPos]
	if !exists {
		chunk = &Chunk{
			Position: chunkPos,
			Blocks:   make(map[matrix.Vec3]*blocks.Block),
			Dirty:    true,
		}
		wm.chunks[chunkPos] = chunk
	}

	chunk.Dirty = true
}
