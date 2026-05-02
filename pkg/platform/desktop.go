//go:build !mobile
// +build !mobile

package platform

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/tesselstudio/TesselBox/internal/game"
	"github.com/tesselstudio/TesselBox/pkg/network"
)

type DesktopGame struct {
	*game.Game
}

func (g *DesktopGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func RunDesktop() {
	var (
		serverMode  = flag.Bool("server", false, "Run as dedicated server only (no game)")
		port        = flag.Uint("port", uint(network.DefaultPort), "Server port")
		serverName  = flag.String("name", "TesselBox Server", "Server name")
		playerName  = flag.String("player", "Player", "Player name")
		connectAddr = flag.String("connect", "", "Connect to server at address (e.g., localhost:25565)")
		discover    = flag.Bool("discover", false, "Enable LAN server discovery")
	)
	flag.Parse()

	if *serverMode {
		log.Printf("Starting TesselBox dedicated server on port %d...", *port)
		server := network.NewServer(*serverName, uint16(*port), network.MaxPlayers)
		server.OnPlayerJoin = func(session *network.ClientSession) {
			log.Printf("[JOIN] %s joined the game", session.Name)
		}
		server.OnPlayerLeave = func(session *network.ClientSession) {
			log.Printf("[LEAVE] %s left the game", session.Name)
		}
		server.OnChatMessage = func(data *network.ChatMessageData) {
			log.Printf("[CHAT] %s: %s", data.PlayerName, data.Message)
		}
		server.OnBlockChange = func(data *network.BlockChangeData) {
			log.Printf("[BLOCK] Player %s changed block at (%d, %d) to type %d",
				data.PlayerID, data.X, data.Y, data.BlockType)
		}
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		log.Printf("Server '%s' is running on port %d", *serverName, *port)
		log.Printf("Press Ctrl+C to stop the server")
		select {}
	}

	g := &game.Game{
		Platform:   "PC",
		PlayerName: *playerName,
	}
	g.InitMultiplayer(*playerName)

	if *discover {
		if err := g.StartDiscovery(); err != nil {
			log.Printf("Failed to start discovery: %v", err)
		} else {
			log.Println("LAN server discovery enabled")
		}
	}

	tt, err := opentype.Parse(goregular.TTF)
	if err == nil {
		g.Font, err = opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    24,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			log.Printf("Failed to create font face: %v", err)
		}
	} else {
		log.Printf("Failed to parse font: %v", err)
	}

	desktopGame := &DesktopGame{Game: g}

	if *connectAddr != "" {
		addr := *connectAddr
		portNum := uint16(network.DefaultPort)
		for i := len(addr) - 1; i >= 0; i-- {
			if addr[i] == ':' {
				portStr := addr[i+1:]
				addr = addr[:i]
				var p int
				if _, err := fmt.Sscanf(portStr, "%d", &p); err == nil && p > 0 && p < 65536 {
					portNum = uint16(p)
				}
				break
			}
		}
		log.Printf("Connecting to server at %s:%d...", addr, portNum)
		if err := g.ConnectToServer(addr, portNum); err != nil {
			log.Printf("Failed to connect: %v", err)
		} else {
			log.Printf("Connected to server!")
		}
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("TesselBox")

	if err := ebiten.RunGame(desktopGame); err != nil {
		log.Fatal(err)
	}

	if g.IsServer {
		g.StopServer()
	}
	if g.IsMultiplayer {
		g.Disconnect()
	}
	if g.DiscoveryBrowser != nil {
		g.StopDiscovery()
	}

	os.Exit(0)
}

func InitMobile() {
	log.Fatal("InitMobile called on non-mobile platform")
}

func InitDesktop() {
	// Desktop initialization is handled in main.go for unified entry point
	log.Println("Desktop platform initialized")
}
