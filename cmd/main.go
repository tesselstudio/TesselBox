package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/platform"
	"github.com/tesselstudio/TesselBox/pkg/server"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
)

// GameWrapper wraps the GameManager for Ebiten compatibility
type GameWrapper struct {
	manager      *game.GameManager
	screenWidth  int
	screenHeight int
}

func (gw *GameWrapper) Update() error {
	return gw.manager.Update(1.0 / 60.0)
}

func (gw *GameWrapper) Draw(screen *ebiten.Image) {
	gw.manager.Draw(screen)
}

func (gw *GameWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return gw.screenWidth, gw.screenHeight
}

func runClient(worldName string, worldSeed int64, creativeMode bool) error {
	log.Printf("Starting TesselBox client...")
	log.Printf("World: %s, Seed: %d, Creative: %v", worldName, worldSeed, creativeMode)

	// Initialize platform
	isMobile := runtime.GOOS == "android" || runtime.GOOS == "ios"
	if isMobile {
		platform.InitMobile()
	} else {
		platform.InitDesktop()
	}

	// Create game manager
	gm := game.NewGameManager(worldName, worldSeed, creativeMode, ScreenWidth, ScreenHeight)

	// Create wrapper and run
	wrapper := &GameWrapper{
		manager:      gm,
		screenWidth:  ScreenWidth,
		screenHeight: ScreenHeight,
	}

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("TesselBox")

	return ebiten.RunGame(wrapper)
}

func runServer(useTUI bool) error {
	log.Printf("Starting TesselBox server...")

	srv := server.NewServer(server.DefaultConfig())

	if useTUI {
		model := server.TUIModel{
			Server:  srv,
			Choices: []string{"Start Server", "Exit"},
			Cursor:  0,
		}
		p := tea.NewProgram(model)
		_, err := p.Run()
		return err
	}

	return srv.Run()
}

func main() {
	mode := flag.String("mode", "client", "Run mode: client, server")
	world := flag.String("world", "default", "World name")
	seed := flag.Int64("seed", 0, "World seed (0 for random)")
	creative := flag.Bool("creative", false, "Enable creative mode")
	useTUI := flag.Bool("tui", false, "Use TUI for server")
	flag.Parse()

	var err error
	switch *mode {
	case "client":
		err = runClient(*world, *seed, *creative)
	case "server":
		err = runServer(*useTUI)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		fmt.Fprintf(os.Stderr, "Usage: %s [-mode=client|server] [options]\n", os.Args[0])
		os.Exit(1)
	}

	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

