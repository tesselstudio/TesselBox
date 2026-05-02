package game

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"

	"github.com/tesselstudio/TesselBox-unified/pkg/network"
)

type OtherPlayer struct {
	ID                   string
	Name                 string
	X, Y                 float64
	VelocityX, VelocityY float64
	LastSync             time.Time
}

type Game struct {
	Count            int
	Font             font.Face
	Platform         string
	NetworkClient    *network.Client
	ServerInstance   *network.Server
	DiscoveryBrowser *network.DiscoveryBrowser
	IsMultiplayer    bool
	IsServer         bool
	IsHost           bool
	PlayerName       string
	ServerAddress    string
	ServerPort       uint16
	OtherPlayers     map[string]*OtherPlayer
	OtherPlayersMu   sync.RWMutex
	ChatMessages     []ChatMessage
	ChatMu           sync.RWMutex
	PlayerX, PlayerY       float64
	PlayerVelocityX, PlayerVelocityY float64
	lastPositionSend       time.Time
}

type ChatMessage struct {
	PlayerID   string
	PlayerName string
	Message    string
	Timestamp  time.Time
}

func (g *Game) Update() error {
	g.Count++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x10, 0x10, 0x30, 0xff})
	if g.Font != nil {
		msg := fmt.Sprintf("TesselBox %s", g.Platform)
		text.Draw(screen, msg, g.Font, 50, 100, color.White)
		fpsMsg := "Game is running!"
		text.Draw(screen, fpsMsg, g.Font, 50, 150, color.RGBA{0x80, 0xff, 0x80, 0xff})
		countMsg := fmt.Sprintf("Frame: %d", g.Count%1000)
		text.Draw(screen, countMsg, g.Font, 50, 200, color.RGBA{0xff, 0x80, 0x80, 0xff})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func (g *Game) InitMultiplayer(playerName string) {
	g.PlayerName = playerName
	if g.OtherPlayers == nil {
		g.OtherPlayers = make(map[string]*OtherPlayer)
	}
	g.ChatMessages = make([]ChatMessage, 0)
}

func (g *Game) HostServer(serverName string, port uint16) error {
	if g.Platform != "PC" {
		return fmt.Errorf("hosting is only available on PC")
	}
	g.ServerInstance = network.NewServer(serverName, port, 16)
	g.ServerInstance.OnPlayerJoin = func(session *network.ClientSession) {
		g.OtherPlayersMu.Lock()
		g.OtherPlayers[session.PlayerID] = &OtherPlayer{
			ID:       session.PlayerID,
			Name:     session.Name,
			X:        session.PositionX,
			Y:        session.PositionY,
			LastSync: time.Now(),
		}
		g.OtherPlayersMu.Unlock()
	}
	g.ServerInstance.OnPlayerLeave = func(session *network.ClientSession) {
		g.OtherPlayersMu.Lock()
		delete(g.OtherPlayers, session.PlayerID)
		g.OtherPlayersMu.Unlock()
	}
	g.ServerInstance.OnChatMessage = func(data *network.ChatMessageData) {
		g.AddChatMessage(data.PlayerID, data.PlayerName, data.Message)
	}
	if err := g.ServerInstance.Start(); err != nil {
		return err
	}
	g.IsServer = true
	g.IsHost = true
	g.IsMultiplayer = true
	g.ServerPort = port
	return nil
}

func (g *Game) StopServer() error {
	if g.ServerInstance == nil {
		return nil
	}
	if err := g.ServerInstance.Stop(); err != nil {
		return err
	}
	g.ServerInstance = nil
	g.IsServer = false
	g.IsHost = false
	g.IsMultiplayer = false
	return nil
}

func (g *Game) ConnectToServer(serverAddr string, port uint16) error {
	g.NetworkClient = network.NewClient(g.PlayerName)
	g.NetworkClient.OnConnected = func() {
		g.IsMultiplayer = true
		g.IsServer = false
	}
	g.NetworkClient.OnDisconnected = func(reason string) {
		g.IsMultiplayer = false
		g.OtherPlayersMu.Lock()
		g.OtherPlayers = make(map[string]*OtherPlayer)
		g.OtherPlayersMu.Unlock()
	}
	g.NetworkClient.OnPlayerPosition = func(data *network.PlayerPositionData) {
		g.OtherPlayersMu.Lock()
		if player, exists := g.OtherPlayers[data.PlayerID]; exists {
			player.X = data.X
			player.Y = data.Y
			player.VelocityX = data.VelocityX
			player.VelocityY = data.VelocityY
			player.LastSync = time.Now()
		} else {
			g.OtherPlayers[data.PlayerID] = &OtherPlayer{
				ID:        data.PlayerID,
				X:         data.X,
				Y:         data.Y,
				VelocityX: data.VelocityX,
				VelocityY: data.VelocityY,
				LastSync:  time.Now(),
			}
		}
		g.OtherPlayersMu.Unlock()
	}
	g.NetworkClient.OnPlayerJoin = func(info *network.PlayerInfo) {
		g.OtherPlayersMu.Lock()
		g.OtherPlayers[info.ID] = &OtherPlayer{
			ID:       info.ID,
			Name:     info.Name,
			X:        info.X,
			Y:        info.Y,
			LastSync: time.Now(),
		}
		g.OtherPlayersMu.Unlock()
	}
	g.NetworkClient.OnPlayerLeave = func(info *network.PlayerInfo) {
		g.OtherPlayersMu.Lock()
		delete(g.OtherPlayers, info.ID)
		g.OtherPlayersMu.Unlock()
	}
	g.NetworkClient.OnChatMessage = func(data *network.ChatMessageData) {
		g.AddChatMessage(data.PlayerID, data.PlayerName, data.Message)
	}
	return g.NetworkClient.Connect(serverAddr, port)
}

func (g *Game) Disconnect() error {
	if g.NetworkClient == nil {
		return nil
	}
	if err := g.NetworkClient.Disconnect(); err != nil {
		return err
	}
	g.NetworkClient = nil
	g.IsMultiplayer = false
	g.IsServer = false
	g.OtherPlayersMu.Lock()
	g.OtherPlayers = make(map[string]*OtherPlayer)
	g.OtherPlayersMu.Unlock()
	return nil
}

func (g *Game) SendPosition(x, y, vx, vy float64) {
	if !g.IsMultiplayer || g.NetworkClient == nil {
		return
	}
	g.PlayerX = x
	g.PlayerY = y
	g.PlayerVelocityX = vx
	g.PlayerVelocityY = vy
	g.NetworkClient.SendPosition(x, y, vx, vy)
}

func (g *Game) SendBlockChange(x, y int32, blockType uint8) error {
	if !g.IsMultiplayer || g.NetworkClient == nil {
		return fmt.Errorf("not connected")
	}
	return g.NetworkClient.SendBlockChange(x, y, blockType)
}

func (g *Game) SendChatMessage(message string) error {
	if !g.IsMultiplayer || g.NetworkClient == nil {
		return fmt.Errorf("not connected")
	}
	return g.NetworkClient.SendChatMessage(message)
}

func (g *Game) AddChatMessage(playerID, playerName, message string) {
	g.ChatMu.Lock()
	defer g.ChatMu.Unlock()
	g.ChatMessages = append(g.ChatMessages, ChatMessage{
		PlayerID:   playerID,
		PlayerName:   playerName,
		Message:      message,
		Timestamp:    time.Now(),
	})
	if len(g.ChatMessages) > 100 {
		g.ChatMessages = g.ChatMessages[len(g.ChatMessages)-100:]
	}
}

func (g *Game) GetOtherPlayers() map[string]*OtherPlayer {
	g.OtherPlayersMu.RLock()
	defer g.OtherPlayersMu.RUnlock()
	result := make(map[string]*OtherPlayer)
	for k, v := range g.OtherPlayers {
		result[k] = v
	}
	return result
}

func (g *Game) StartDiscovery() error {
	browser, err := network.NewDiscoveryBrowser()
	if err != nil {
		return err
	}
	g.DiscoveryBrowser = browser
	browser.Start()
	return nil
}

func (g *Game) StopDiscovery() {
	if g.DiscoveryBrowser != nil {
		g.DiscoveryBrowser.Stop()
		g.DiscoveryBrowser = nil
	}
}

func (g *Game) GetDiscoveredServers() []*network.DiscoveredServer {
	if g.DiscoveryBrowser == nil {
		return nil
	}
	return g.DiscoveryBrowser.GetServers()
}
