//go:build mobile
// +build mobile

package platform

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/mobile"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/tesselstudio/TesselBox-unified/internal/game"
)

type MobileGame struct {
	*game.Game
}

func (g *MobileGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func InitMobile() {
	g := &game.Game{
		Platform:   "Mobile",
		PlayerName: "MobilePlayer",
	}
	g.InitMultiplayer("MobilePlayer")

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

	mobileGame := &MobileGame{Game: g}
	mobile.SetGame(mobileGame)
}

func (g *MobileGame) ConnectToServer(serverAddr string, port uint16) error {
	return g.Game.ConnectToServer(serverAddr, port)
}

func (g *MobileGame) Disconnect() error {
	return g.Game.Disconnect()
}

func (g *MobileGame) StartDiscovery() error {
	return g.Game.StartDiscovery()
}

func (g *MobileGame) StopDiscovery() {
	g.Game.StopDiscovery()
}

func (g *MobileGame) IsMultiplayer() bool {
	return g.Game.IsMultiplayer
}

func (g *MobileGame) IsConnected() bool {
	if g.Game.NetworkClient == nil {
		return false
	}
	return g.Game.NetworkClient.IsConnected()
}
