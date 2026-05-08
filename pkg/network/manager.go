package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// NetworkCallback defines functions that the game implements to handle network events
type NetworkCallback interface {
	OnPlayerJoin(playerID uint32, name string, position types.Vec3)
	OnPlayerLeave(playerID uint32)
	OnPlayerMove(playerID uint32, position, rotation types.Vec3)
	OnBlockPlace(playerID uint32, blockType uint8, position types.Vec3)
	OnBlockBreak(playerID uint32, position types.Vec3)
	OnChatMessage(playerID uint32, message string)
}

// Manager bridges the network layer with the game
type Manager struct {
	mu sync.RWMutex

	// Network client
	client *Client

	// Network server (if hosting)
	server *Server

	// Callback to game
	callback NetworkCallback

	// Host state
	isHost      bool
	isConnected bool

	// Remote players tracked by this manager
	remotePlayers map[uint32]*RemotePlayerState

	// Pending updates to process
	pendingUpdates []NetworkUpdate
}

// RemotePlayerState tracks a remote player's state
type RemotePlayerState struct {
	ID         uint32
	Name       string
	Position   types.Vec3
	Rotation   types.Vec3
	Velocity   types.Vec3
	LastUpdate time.Time
}

// NetworkUpdate represents a pending network update
type NetworkUpdate struct {
	Type     UpdateType
	PlayerID uint32
	Data     interface{}
}

// UpdateType defines the type of network update
type UpdateType int

const (
	UpdatePlayerJoin UpdateType = iota
	UpdatePlayerLeave
	UpdatePlayerMove
	UpdateBlockPlace
	UpdateBlockBreak
	UpdateChat
)

// NewManager creates a new network manager
func NewManager() *Manager {
	return &Manager{
		remotePlayers:  make(map[uint32]*RemotePlayerState),
		pendingUpdates: make([]NetworkUpdate, 0),
	}
}

// SetCallback sets the game callback for network events
func (m *Manager) SetCallback(callback NetworkCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = callback
}

// Connect connects to a multiplayer server
func (m *Manager) Connect(address, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isConnected {
		return fmt.Errorf("already connected")
	}

	config := ClientConfig{
		ServerAddr: address,
		PlayerName: playerName,
		TickRate:   20,
	}

	m.client = NewClient(config)

	// Set up network callbacks
	m.setupClientCallbacks()

	// Connect in background
	go func() {
		if err := m.client.Connect(playerName); err != nil {
			fmt.Printf("Failed to connect: %v\n", err)
			return
		}

		m.mu.Lock()
		m.isConnected = true
		m.mu.Unlock()
	}()

	return nil
}

// HostServer starts a server and connects as host
func (m *Manager) HostServer(port int, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		return fmt.Errorf("server already running")
	}

	config := ServerConfig{
		Host:         "",
		Port:         port,
		MaxPlayers:   8,
		TickRate:     20,
		WorldSize:    16,
		ChunkSize:    16,
		ViewDistance: 8,
	}

	m.server = NewServer(config)

	// Start server in background
	go func() {
		if err := m.server.Start(); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			return
		}
	}()

	// Wait a moment for server to start, then connect
	time.Sleep(100 * time.Millisecond)

	// Connect as client to our own server
	m.isHost = true
	return m.Connect(fmt.Sprintf("localhost:%d", port), playerName)
}

// setupClientCallbacks configures callbacks for network events
func (m *Manager) setupClientCallbacks() {
	m.client.SetCallbacks(
		// OnConnect
		func(playerID uint32) {
			fmt.Printf("Connected as player %d\n", playerID)
		},
		// OnDisconnect
		func() {
			fmt.Println("Disconnected from server")
			m.mu.Lock()
			m.isConnected = false
			m.mu.Unlock()
		},
		// OnPlayerJoin
		func(player *RemotePlayer) {
			m.mu.Lock()
			m.pendingUpdates = append(m.pendingUpdates, NetworkUpdate{
				Type:     UpdatePlayerJoin,
				PlayerID: player.ID,
				Data:     player,
			})
			m.mu.Unlock()
		},
		// OnPlayerLeave
		func(playerID uint32) {
			m.mu.Lock()
			m.pendingUpdates = append(m.pendingUpdates, NetworkUpdate{
				Type:     UpdatePlayerLeave,
				PlayerID: playerID,
			})
			m.mu.Unlock()
		},
		// OnBlockPlace
		func(place *BlockPlaceMessage) {
			m.mu.Lock()
			m.pendingUpdates = append(m.pendingUpdates, NetworkUpdate{
				Type:     UpdateBlockPlace,
				PlayerID: place.PlayerID,
				Data:     place,
			})
			m.mu.Unlock()
		},
		// OnBlockBreak
		func(breakMsg *BlockBreakMessage) {
			m.mu.Lock()
			m.pendingUpdates = append(m.pendingUpdates, NetworkUpdate{
				Type:     UpdateBlockBreak,
				PlayerID: breakMsg.PlayerID,
				Data:     breakMsg,
			})
			m.mu.Unlock()
		},
		// OnChat
		func(chat *ChatMessage) {
			m.mu.Lock()
			m.pendingUpdates = append(m.pendingUpdates, NetworkUpdate{
				Type:     UpdateChat,
				PlayerID: chat.PlayerID,
				Data:     chat,
			})
			m.mu.Unlock()
		},
	)
}

// Update processes pending network updates (call from main thread)
func (m *Manager) Update() {
	m.mu.Lock()
	updates := m.pendingUpdates
	m.pendingUpdates = make([]NetworkUpdate, 0)
	m.mu.Unlock()

	if m.callback == nil {
		return
	}

	// Process updates
	for _, update := range updates {
		switch update.Type {
		case UpdatePlayerJoin:
			if player, ok := update.Data.(*RemotePlayer); ok {
				m.callback.OnPlayerJoin(update.PlayerID, player.Name, player.Position)
			}

		case UpdatePlayerLeave:
			m.callback.OnPlayerLeave(update.PlayerID)

		case UpdateBlockPlace:
			if place, ok := update.Data.(*BlockPlaceMessage); ok {
				m.callback.OnBlockPlace(update.PlayerID, place.BlockType, place.Position)
			}

		case UpdateBlockBreak:
			if breakMsg, ok := update.Data.(*BlockBreakMessage); ok {
				m.callback.OnBlockBreak(update.PlayerID, breakMsg.Position)
			}

		case UpdateChat:
			if chat, ok := update.Data.(*ChatMessage); ok {
				m.callback.OnChatMessage(update.PlayerID, chat.Message)
			}
		}
	}

	// Send player position updates
	if m.client != nil && m.client.IsConnected() {
		if localPlayer := m.client.GetLocalPlayer(); localPlayer != nil {
			m.client.SendPlayerMove(localPlayer.Position, localPlayer.Rotation, localPlayer.Velocity)
		}
	}
}

// SendBlockPlace notifies the server of a block placement
func (m *Manager) SendBlockPlace(blockType uint8, position types.Vec3, rotation int) {
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client != nil && client.IsConnected() {
		client.SendBlockPlace(blockType, position, rotation)
	}
}

// SendBlockBreak notifies the server of a block break
func (m *Manager) SendBlockBreak(position types.Vec3) {
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client != nil && client.IsConnected() {
		client.SendBlockBreak(position)
	}
}

// SendChat sends a chat message
func (m *Manager) SendChat(message string) {
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client != nil && client.IsConnected() {
		client.SendChat(message)
	}
}

// IsConnected returns true if connected to a server
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isConnected
}

// IsHost returns true if this client is hosting the server
func (m *Manager) IsHost() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isHost
}

// GetPlayerID returns the local player ID
func (m *Manager) GetPlayerID() uint32 {
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client != nil {
		return client.GetPlayerID()
	}
	return 0
}

// Disconnect disconnects from the server
func (m *Manager) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		m.client.Disconnect()
		m.client = nil
	}

	if m.server != nil {
		m.server.Stop()
		m.server = nil
	}

	m.isConnected = false
	m.isHost = false
}

// GetRemotePlayers returns a copy of the remote players map
func (m *Manager) GetRemotePlayers() map[uint32]*RemotePlayerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	players := make(map[uint32]*RemotePlayerState)
	for id, player := range m.remotePlayers {
		players[id] = player
	}
	return players
}
